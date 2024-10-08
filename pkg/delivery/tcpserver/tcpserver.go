package tcpserver

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/MinamiKotoriCute/serr"
)

type OnConnctedFuncType func(conn net.Conn) error
type OnDisconnctedFuncType func(conn net.Conn, closeReason string, closeType int32, closeReasonObject interface{})
type OnReceiveFuncType func(conn net.Conn, data []byte) ([]byte, error)

type TcpServer struct {
	listen    net.Listener
	wg        sync.WaitGroup
	serveWg   sync.WaitGroup
	config    *Config
	conns     map[net.Conn]*Connection
	connsLock sync.RWMutex
	log       *slog.Logger
}

func NewTcpServer(config *Config) *TcpServer {
	c := &Config{
		PacketSizeLimit:      config.PacketSizeLimit,
		ReadBufferSize:       config.ReadBufferSize,
		QueuePacketSizeLimit: config.QueuePacketSizeLimit,
		QueuePacketNumLimit:  config.QueuePacketNumLimit,
		X509CertPath:         config.X509CertPath,
		X509KeyPath:          config.X509KeyPath,
		OnConnctedFunc:       config.OnConnctedFunc,
		OnDisconnctedFunc:    config.OnDisconnctedFunc,
		OnReceiveFunc:        config.OnReceiveFunc,
	}

	if c.PacketSizeLimit == 0 {
		c.PacketSizeLimit = 1024 * 1024 // 1MB
	}
	if c.ReadBufferSize == 0 {
		c.ReadBufferSize = 1024 // 1KB
	}
	if c.QueuePacketSizeLimit == 0 {
		c.QueuePacketSizeLimit = 10 * 1024 * 1024 // 10MB
	}
	if c.QueuePacketNumLimit == 0 {
		c.QueuePacketNumLimit = 10
	}
	if c.Log == nil {
		c.Log = slog.Default()
	}

	return &TcpServer{
		config: c,
		conns:  make(map[net.Conn]*Connection),
		log:    c.Log,
	}
}

func (o *TcpServer) Start(address string) error {
	var listen net.Listener

	if o.config.X509CertPath != "" && o.config.X509KeyPath != "" {
		cert, err := tls.LoadX509KeyPair(o.config.X509CertPath, o.config.X509KeyPath)
		if err != nil {
			return serr.Wrap(err)
		}
		config := tls.Config{Certificates: []tls.Certificate{cert}}
		config.Rand = rand.Reader

		tempListen, err := tls.Listen("tcp", address, &config)
		if err != nil {
			return serr.Wrap(err)
		}
		listen = tempListen
	} else {
		tempListen, err := net.Listen("tcp", address)
		if err != nil {
			return serr.Wrap(err)
		}
		listen = tempListen
	}

	o.listen = listen
	o.serveWg.Add(1)
	go o.serve()

	return nil
}

// stop accept new connection, and close all connection read stream
// wait all connection current handle finished
func (o *TcpServer) Stop(ctx context.Context) error {
	if err := o.listen.Close(); err != nil {
		return serr.Wrap(err)
	}
	o.serveWg.Wait()

	o.connsLock.Lock()
	for _, connection := range o.conns {
		connection.DisconnectFromServer("tcp server stop", int32(CloseTypeTcpServerStop), nil)
	}
	o.connsLock.Unlock()
	o.wg.Wait()

	return nil
}

func (o *TcpServer) serve() {
	defer o.serveWg.Done()

	for {
		conn, err := o.listen.Accept()
		connection := &Connection{
			Conn: conn,
		}

		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				o.log.Warn("tcp server accept error", slog.Any("err", serr.ToJSON(err, true)))
			}
			break
		}

		o.connsLock.Lock()
		o.conns[conn] = connection
		o.connsLock.Unlock()

		o.wg.Add(1)
		go func() {
			if err := o.handleConnection(connection); err != nil {
				o.log.Warn("tcp server handle connection error",
					slog.Any("err", serr.ToJSON(err, true)),
					slog.String("remote_address", conn.RemoteAddr().String()))
			}
		}()
	}
}

func (o *TcpServer) handleConnection(connection *Connection) error {
	defer func() {
		o.connsLock.Lock()
		delete(o.conns, connection.Conn)
		o.connsLock.Unlock()
		connection.Conn.Close()
		o.wg.Done()
	}()

	conn := connection.Conn
	readBuffer := make([]byte, o.config.ReadBufferSize)
	tempBuffer := []byte{}
	packetSize := uint64(0)

	if o.config.OnConnctedFunc != nil {
		if err := o.config.OnConnctedFunc(conn); err != nil {
			connection.AppendCloseReason("onConnctedFunc error", int32(CloseTypeError))
			return err
		}
	}

	queuePacketCh := make(chan []byte, o.config.QueuePacketNumLimit)
	queuePacketSize := atomic.Int64{}

	go func() {
		for packet := range queuePacketCh {
			queuePacketSize.Add(int64(-len(packet)))

			if rspData, err := o.config.OnReceiveFunc(conn, packet); err != nil {
				o.DisconnectConnection(conn, "onReceiveFunc error", int32(CloseTypeError), nil)
				o.log.Warn("OnReceiveFunc error",
					slog.Any("err", serr.ToJSON(err, true)),
					slog.String("remote_address", conn.RemoteAddr().String()))
				return
			} else {
				o.SendToUser(conn, rspData)
			}
		}
	}()

	defer func() {
		if o.config.OnDisconnctedFunc != nil {
			if connection.CloseType == int32(CloseTypeEmpty) {
				connection.CloseType = int32(CloseTypeError)
				connection.CloseReason = "unknown"
			}
			o.config.OnDisconnctedFunc(conn, connection.CloseReason, connection.CloseType, connection.CloseReasonObject)
		}
	}()

	defer close(queuePacketCh)

	for {
		n, err := conn.Read(readBuffer)
		if err != nil {
			if err == io.EOF {
				if connection.CloseType == int32(CloseTypeEmpty) {
					connection.CloseType = int32(CloseTypeDisconnect)
					connection.CloseReason = "close by client at read"
				}
				return nil
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				connection.AppendCloseReason("close by client at read timeout", int32(CloseTypeDisconnectOnRead))
				return nil // read tcp 10.88.1.32:8081->10.140.0.19:17449: read: connection timed out
			}
			if errors.Is(err, syscall.ECONNRESET) {
				connection.AppendCloseReason("close by client at read reset", int32(CloseTypeDisconnectOnRead))
				return nil // example: read tcp 10.88.1.26:8081->10.140.0.19:27151: read: connection reset by peer
			}
			connection.AppendCloseReason("handle read error", int32(CloseTypeError))
			return serr.Wrap(err)
		}

		if len(tempBuffer)+n > int(o.config.PacketSizeLimit) {
			connection.AppendCloseReason("packet size too large", int32(CloseTypeError))
			return serr.Errorf("packet size too large. size=%d", len(tempBuffer)+n)
		}

		tempBuffer = append(tempBuffer, readBuffer[:n]...)

		for {
			if packetSize == 0 {
				if len(tempBuffer) < int(packetSize+8) {
					break
				}

				packetSize = binary.BigEndian.Uint64(tempBuffer[:8])
				if packetSize > uint64(o.config.PacketSizeLimit)-8 {
					connection.AppendCloseReason("packet size too large", int32(CloseTypeError))
					return serr.Errorf("packet size too large. size=%d", packetSize)
				}

				tempBuffer = tempBuffer[8:]
			}

			if len(tempBuffer) < int(packetSize) {
				break
			}

			if uint64(queuePacketSize.Load())+packetSize > o.config.QueuePacketSizeLimit {
				connection.AppendCloseReason("queue packet size too large", int32(CloseTypeError))
				return serr.Errorf("queue packet size too large. current=%d new=%d", queuePacketSize.Load(), packetSize)
			}

			select {
			case queuePacketCh <- tempBuffer[:packetSize]:
				queuePacketSize.Add(int64(packetSize))
			default:
				connection.AppendCloseReason("queue packet number too many", int32(CloseTypeError))
				return serr.Errorf("queue packet number too many")
			}

			tempBuffer = tempBuffer[packetSize:]
			packetSize = 0
		}
	}
}

func (o *TcpServer) SendToUser(conn net.Conn, data []byte) error {
	if len(data) == 0 {
		return nil
	}

	writeBuffer := wrapPacket(data)
	if _, err := conn.Write(writeBuffer); err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// example: write tcp 10.88.1.32:8081->10.88.1.1:2890: write: connection timed out
			o.DisconnectConnection(conn, "close by client at write timeout", int32(CloseTypeDisconnectOnWrite), nil)
		} else if errors.Is(err, syscall.EPIPE) {
			// example: write tcp 10.88.1.32:8081->10.140.0.19:47207: write: broken pipe
			o.DisconnectConnection(conn, "close by client at write broken", int32(CloseTypeDisconnectOnWrite), nil)
		} else {
			o.DisconnectConnection(conn, "handle write error", int32(CloseTypeDisconnectOnWrite), nil)
		}

		return serr.Wrap(err)
	}

	return nil
}

func wrapPacket(data []byte) []byte {
	writeBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(writeBuffer, uint64(len(data)))
	writeBuffer = append(writeBuffer, data...)
	return writeBuffer
}

func (o *TcpServer) DisconnectConnection(conn net.Conn, reason string, closeType int32, closeReasonObject interface{}) {
	o.connsLock.Lock()
	defer o.connsLock.Unlock()
	if connection, ok := o.conns[conn]; ok {
		connection.DisconnectFromServer(reason, closeType, closeReasonObject)
	}
}
