package tcpserver

import "log/slog"

type Config struct {
	PacketSizeLimit      uint64
	ReadBufferSize       int
	QueuePacketSizeLimit uint64
	QueuePacketNumLimit  int
	X509CertPath         string
	X509KeyPath          string
	OnConnctedFunc       OnConnctedFuncType
	OnDisconnctedFunc    OnDisconnctedFuncType
	OnReceiveFunc        OnReceiveFuncType
	Log                  *slog.Logger
}
