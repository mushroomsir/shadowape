package pkg

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"sync"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/mushroomsir/logger/alog"
)

// Client ...
type Client struct {
	session            quic.Session
	socks5Lis, httpLis net.Listener
	config             *ClientConfig
	proxyhttp          *ProxyHTTPServer
	lock               sync.Mutex
	tlsConfig          *tls.Config
}

// NewClient ...
func NewClient(config *ClientConfig) (*Client, error) {
	var err error
	if config.Socks5ServerAddr == "" {
		return nil, errors.New("the socks5_Server_addr is empty")
	}
	c := &Client{config: config}
	if config.CertFile == "" {
		c.tlsConfig = &tls.Config{InsecureSkipVerify: true}
	} else {
		rootPEM, err := ioutil.ReadFile(config.CertFile)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(rootPEM)
		c.tlsConfig = &tls.Config{RootCAs: caCertPool}
	}
	c.session, err = getConn(config.Socks5ServerAddr, c.tlsConfig)
	if err != nil {
		return nil, err
	}
	c.socks5Lis, err = net.Listen("tcp", config.Socks5ListenAddr)
	if err != nil {
		return nil, err
	}
	c.proxyhttp, err = NewProxyHTTPServer(config, &quicForward{session: c.getSession})
	return c, err
}

// Run ...
func (c *Client) Run() {
	go c.proxyhttp.runhttp()
	for {
		conn, err := c.socks5Lis.Accept()
		if alog.Check(err) {
			continue
		}
		go c.handleConn(conn)
	}
}

func (c *Client) handleConn(conn net.Conn) {
	for {
		stream, err := c.getSession().OpenStreamSync()
		if alog.Check(err) {
			time.Sleep(time.Second)
			alog.Info("start reconnection:", c.genNewSession())
			continue
		}
		alog.Infof("%s -> %s, streamID: %v", conn.RemoteAddr().String(), c.session.RemoteAddr().String(), stream.StreamID())
		transfer(conn, stream)
		stream.Close()
		conn.Close()
		return
	}
}

func (c *Client) getSession() quic.Session {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.session
}

func (c *Client) setSession(session quic.Session) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.session = session
}
func (c *Client) genNewSession() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.session.Close(ErrReconnection)
	session, err := getConn(c.config.Socks5ServerAddr, c.tlsConfig)
	if alog.Check(err) {
		return err
	}
	c.session = session
	return nil
}
