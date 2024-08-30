package tcpserver

import (
	"fmt"
	"net"
)

type CloseType int32

const (
	CloseTypeEmpty CloseType = iota
	CloseTypeDisconnect
	CloseTypeDisconnectOnRead
	CloseTypeDisconnectOnWrite
	CloseTypeTcpServerStop
	CloseTypeError
)

type Connection struct {
	Conn              net.Conn
	CloseType         int32
	CloseReason       string
	CloseReasonObject interface{}
	CloseError        error
	IsClose           bool
}

func (o *Connection) DisconnectFromServer(reason string, closeType int32, closeReasonObject interface{}) {
	if o.IsClose {
		return
	}
	o.IsClose = true
	o.CloseReason = reason
	o.CloseType = closeType
	o.CloseReasonObject = closeReasonObject
	if tcpConn, ok := o.Conn.(*net.TCPConn); ok {
		o.CloseError = tcpConn.CloseRead()
	}
}

func (o *Connection) AppendCloseReason(reason string, closeType int32) {
	if o.CloseType == 0 {
		o.CloseType = closeType
		o.CloseReason = reason
		return
	}

	if o.CloseReason != "" {
		o.CloseReason += " "
	}
	o.CloseReason += fmt.Sprintf("extra_close_type:%d(%s)", closeType, reason)
}
