package tcpserver

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"syscall"

	"github.com/MinamiKotoriCute/serr"
	"github.com/sirupsen/logrus"
)

type OnConnctedFuncType func(conn net.Conn) error
type OnDisconnctedFuncType func(conn net.Conn, closeReason string, closeType int32, closeReasonObject interface{})
type OnReceiveFuncType func(conn net.Conn, data []byte) ([]byte, error)

type TcpServer struct {
	listen            net.Listener
	wg                sync.WaitGroup
	serveWg           sync.WaitGroup
	config            *Config
	onConnctedFunc    OnConnctedFuncType
	onDisconnctedFunc OnDisconnctedFuncType
	onReceiveFunc     OnReceiveFuncType
	conns             map[net.Conn]*Connection
	connsLock         sync.RWMutex
}

func NewTcpServer(config *Config,
	onConnctedFunc OnConnctedFuncType,
	onDisconnctedFunc OnDisconnctedFuncType,
	onReceiveFunc OnReceiveFuncType) *TcpServer {

	if config.PacketSizeLimit == 0 {
		config.PacketSizeLimit = 1024 * 1024 // 1MB
	}
	if config.ReadBufferSize == 0 {
		config.ReadBufferSize = 1024 // 1KB
	}

	return &TcpServer{
		config:            config,
		onConnctedFunc:    onConnctedFunc,
		onDisconnctedFunc: onDisconnctedFunc,
		onReceiveFunc:     onReceiveFunc,
		conns:             make(map[net.Conn]*Connection),
	}
}

func (o *TcpServer) Start(address string) error {
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return serr.Wrap(err)
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
				logrus.WithField("error", serr.ToJSON(err, true)).Warning("tcp server accept error")
			}
			break
		}

		o.connsLock.Lock()
		o.conns[conn] = connection
		o.connsLock.Unlock()

		o.wg.Add(1)
		go func() {
			if err := o.handleConnection(connection); err != nil {
				logrus.WithFields(logrus.Fields{
					"error":          serr.ToJSON(err, true),
					"remote_address": conn.RemoteAddr().String(),
				}).Warning("tcp server handle connection error")
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

	if o.onConnctedFunc != nil {
		if err := o.onConnctedFunc(conn); err != nil {
			connection.AppendCloseReason("onConnctedFunc error", int32(CloseTypeError))
			return err
		}
	}

	defer func() {
		if o.onDisconnctedFunc != nil {
			if connection.CloseType == int32(CloseTypeEmpty) {
				connection.CloseType = int32(CloseTypeError)
				connection.CloseReason = "unknown"
			}
			o.onDisconnctedFunc(conn, connection.CloseReason, connection.CloseType, connection.CloseReasonObject)
		}
	}()

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

			if rspData, err := o.onReceiveFunc(conn, tempBuffer[:packetSize]); err != nil {
				connection.AppendCloseReason("onReceiveFunc error", int32(CloseTypeError))
				return err
			} else {
				o.SendToUser(conn, rspData)
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
