package tcpserver

import "net"

type Connection struct {
	Conn        net.Conn
	CloseReason string
}

func (o *Connection) DisconnectFromServer(reason string) {
	o.CloseReason = reason
	if tcpConn, ok := o.Conn.(*net.TCPConn); ok {
		tcpConn.CloseRead()
	}
}
