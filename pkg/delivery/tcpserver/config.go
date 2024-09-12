package tcpserver

type Config struct {
	PacketSizeLimit uint64
	ReadBufferSize  int
	X509CertPath    string
	X509KeyPath     string
}
