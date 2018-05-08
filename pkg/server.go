package pkg

import (
	"net"
	"time"

	quic "github.com/lucas-clemente/quic-go"
	"github.com/mushroomsir/logger/alog"
)

// Server ...
type Server struct {
	lis    quic.Listener
	config *ServerConfig
}

// NewServer ...
func NewServer(config *ServerConfig) (*Server, error) {
	lis, err := quic.ListenAddr(config.ServerAddr, generateTLSConfig(), &quic.Config{
		MaxIncomingStreams:                    65535,
		IdleTimeout:                           time.Hour,
		MaxReceiveStreamFlowControlWindow:     100 * (1 << 20),
		MaxReceiveConnectionFlowControlWindow: 1000 * (1 << 20),
	})
	if alog.Check(err) {
		return nil, err
	}
	return &Server{
		lis:    lis,
		config: config,
	}, nil
}

// Run ...
func (c *Server) Run() {
	for {
		conn, err := c.lis.Accept()
		alog.Check(err)
		alog.Infof("new conntion : %s\n", conn.RemoteAddr())
		go c.handleConn(conn)
	}
}

func (c *Server) handleConn(conn quic.Session) {
	for {
		stream, err := conn.AcceptStream()
		if alog.Check(err) {
			conn.Close(err)
			return
		}
		go c.handleStream(stream)
	}
}

func (c *Server) handleStream(stream quic.Stream) {
	defer stream.Close()
	err := handshake(stream)
	if alog.Check(err) {
		return
	}
	addr, err := getAddr(stream)
	if alog.Check(err) {
		return
	}
	conn, err := net.Dial("tcp", addr)
	if alog.Check(err) {
		return
	}
	alog.Infof("->: %s\n", addr)
	err = transfer(stream, conn)
	alog.Check(err)
	conn.Close()
}
