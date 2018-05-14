package pkg

import (
	"crypto/tls"
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
	var tlsConfig *tls.Config
	if config.CertFile != "" && config.KeyFile != "" {
		cer, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if alog.Check(err) {
			return nil, err
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cer}}
	} else {
		tlsConfig = generateTLSConfig()
	}
	lis, err := quic.ListenAddr(config.ServerAddr, tlsConfig, &quic.Config{
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
		alog.Infof("new conntion: %s", conn.RemoteAddr())
		go c.handleConn(conn)
	}
}

func (c *Server) handleConn(conn quic.Session) {
	for {
		stream, err := conn.AcceptStream()
		if alog.Check(err) {
			alog.Infof("close conntion: %s", conn.RemoteAddr())
			conn.Close(err)
			return
		}
		go c.handleStream(stream, conn.RemoteAddr().String())
	}
}

func (c *Server) handleStream(stream quic.Stream, remoteAddr string) {
	defer stream.Close()
	err := handshake(stream)
	if alog.Check(err) {
		return
	}
	addr, err := getAddr(stream)
	if alog.Check(err) {
		return
	}
	alog.Infof("new streamID: %v, %s -> %s", stream.StreamID(), remoteAddr, addr)
	conn, err := net.Dial("tcp", addr)
	if alog.Check(err) {
		return
	}
	err = transfer(stream, conn)
	alog.Check(err)
	conn.Close()
	alog.Infof("close streamID: %v, %s -> %s", stream.StreamID(), remoteAddr, addr)
}
