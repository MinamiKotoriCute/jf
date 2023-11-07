package tcpserver

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
)

type OnConnctedFuncType func(conn net.Conn) error
type OnDisconnctedFuncType func(conn net.Conn, closeReason string)
type OnReceiveFuncType func(conn net.Conn, data []byte) ([]byte, error)

type TcpServer struct {
	listen            net.Listener
	wg                sync.WaitGroup
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
		return eris.Wrap(err, "")
	}

	o.listen = listen
	o.wg.Add(1)
	go o.serve()

	return nil
}

// stop accept new connection, and close all connection read stream
// wait all connection current handle finished
func (o *TcpServer) Stop(ctx context.Context) error {
	if err := o.listen.Close(); err != nil {
		return eris.Wrap(err, "")
	}

	o.connsLock.Lock()
	for _, connection := range o.conns {
		connection.DisconnectFromServer("tcp server stop")
	}
	o.connsLock.Unlock()
	o.wg.Wait()

	return nil
}

func (o *TcpServer) serve() {
	defer o.wg.Done()

	for {
		conn, err := o.listen.Accept()
		connection := &Connection{
			Conn: conn,
		}

		o.connsLock.Lock()
		o.conns[conn] = connection
		o.connsLock.Unlock()

		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				logrus.WithField("error", eris.ToJSON(err, true)).Warning()
			}
			break
		}

		o.wg.Add(1)
		go func() {
			if err := o.handleConnection(connection); err != nil {
				logrus.WithField("error", eris.ToJSON(err, true)).Warning()
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
			return err
		}
	}

	defer func() {
		if o.onDisconnctedFunc != nil {
			if connection.CloseReason == "" {
				connection.CloseReason = "close by client"
			}
			o.onDisconnctedFunc(conn, connection.CloseReason)
		}
	}()

	for {
		n, err := conn.Read(readBuffer)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			connection.CloseReason = "handle read error"
			return eris.Wrap(err, "")
		}

		if len(tempBuffer)+n > int(o.config.PacketSizeLimit) {
			return eris.Errorf("packet size too large. size=%d", len(tempBuffer)+n)
		}

		tempBuffer = append(tempBuffer, readBuffer[:n]...)
		if len(tempBuffer) < 8 {
			continue
		}

		if packetSize == 0 {
			packetSize = binary.BigEndian.Uint64(tempBuffer[:8])
			if packetSize > uint64(o.config.PacketSizeLimit)-8 {
				connection.CloseReason = "handle read error"
				return eris.Errorf("packet size too large. size=%d", packetSize)
			}
		}

		if len(tempBuffer) < int(packetSize+8) {
			continue
		}

		if rspData, err := o.onReceiveFunc(conn, tempBuffer[8:packetSize+8]); err != nil {
			connection.CloseReason = "handle func error"
			return err
		} else {
			o.SendToUser(conn, rspData)
		}

		tempBuffer = tempBuffer[packetSize+8:]
		packetSize = 0
	}
}

func (o *TcpServer) SendToUser(conn net.Conn, data []byte) error {
	if len(data) == 0 {
		return nil
	}

	writeBuffer := wrapPacket(data)
	if _, err := conn.Write(writeBuffer); err != nil {
		return eris.Wrap(err, "")
	}

	return nil
}

func wrapPacket(data []byte) []byte {
	writeBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(writeBuffer, uint64(len(data)))
	writeBuffer = append(writeBuffer, data...)
	return writeBuffer
}

func (o *TcpServer) DisconnectConnection(conn net.Conn, reason string) {
	o.connsLock.Lock()
	defer o.connsLock.Unlock()
	if connection, ok := o.conns[conn]; ok {
		connection.DisconnectFromServer(reason)
	}
}
