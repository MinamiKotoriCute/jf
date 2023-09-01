package tcpserver

import (
	"encoding/binary"
	"io"
	"net"
	"sync"

	"github.com/golang/glog"
	"github.com/rotisserie/eris"
)

type OnConnctedFuncType func(conn net.Conn) error
type OnDisconnctedFuncType func(conn net.Conn)
type OnReceiveFuncType func(conn net.Conn, data []byte) ([]byte, error)

type TcpServer struct {
	listen            net.Listener
	wg                sync.WaitGroup
	config            *Config
	onConnctedFunc    OnConnctedFuncType
	onDisconnctedFunc OnDisconnctedFuncType
	onReceiveFunc     OnReceiveFuncType
}

func NewTcpServer(config *Config,
	onConnctedFunc OnConnctedFuncType,
	onDisconnctedFunc OnDisconnctedFuncType,
	onReceiveFunc OnReceiveFuncType) *TcpServer {
	return &TcpServer{
		config:            config,
		onConnctedFunc:    onConnctedFunc,
		onDisconnctedFunc: onDisconnctedFunc,
		onReceiveFunc:     onReceiveFunc,
	}
}

func (o *TcpServer) Start(address string) error {
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return eris.Wrap(err, "")
	}

	o.listen = listen
	go o.serve()

	return nil
}

func (o *TcpServer) Stop() error {
	if err := o.listen.Close(); err != nil {
		return eris.Wrap(err, "")
	}
	o.wg.Wait()

	return nil
}

func (o *TcpServer) serve() {
	defer o.wg.Done()

	for {
		conn, err := o.listen.Accept()
		if err != nil {
			glog.Warning(err)
			break
		}

		o.wg.Add(1)
		go func() {
			if err := o.handleConnection(conn); err != nil {
				glog.Warning(eris.ToString(err, false))
			}
		}()
	}
}

func (o *TcpServer) handleConnection(conn net.Conn) error {
	defer func() {
		conn.Close()
		o.wg.Done()
	}()

	readBuffer := make([]byte, 256)
	tempBuffer := []byte{}
	packetSize := uint64(0)

	if o.onConnctedFunc != nil {
		if err := o.onConnctedFunc(conn); err != nil {
			return err
		}
	}

	for {
		n, err := conn.Read(readBuffer)
		if err != nil {
			if err == io.EOF {
				if o.onDisconnctedFunc != nil {
					o.onDisconnctedFunc(conn)
				}
				return nil
			}
			return eris.Wrap(err, "")
		}

		if len(tempBuffer)+n > int(o.config.TcpPacketSizeLimit) {
			return eris.Errorf("packet size too large. size=%d", len(tempBuffer)+n)
		}

		tempBuffer = append(tempBuffer, readBuffer[:n]...)
		if len(tempBuffer) < 8 {
			continue
		}

		if packetSize == 0 {
			packetSize = binary.BigEndian.Uint64(tempBuffer[:8])
			if packetSize > uint64(len(tempBuffer))-8 {
				return eris.Errorf("packet size too large. size=%d", packetSize)
			}
		}

		if len(tempBuffer) < int(packetSize+8) {
			continue
		}

		if rspData, err := o.onReceiveFunc(conn, tempBuffer[8:packetSize+8]); err != nil {
			return err
		} else if len(rspData) > 0 {
			writeBuffer := make([]byte, 8)
			binary.BigEndian.PutUint64(writeBuffer, uint64(len(rspData)))
			writeBuffer = append(writeBuffer, rspData...)
			if _, err := conn.Write(writeBuffer); err != nil {
				return eris.Wrap(err, "")
			}
		}

		tempBuffer = tempBuffer[packetSize+8:]
	}
}
