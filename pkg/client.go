package pkg

import (
	"errors"
	"net"
	"sync"

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
}

// NewClient ...
func NewClient(config *ClientConfig) (*Client, error) {
	if config.Socks5ServerAddr == "" {
		return nil, errors.New("the socks5_Server_addr is empty")
	}
	session, err := getConn(config.Socks5ServerAddr)
	if err != nil {
		return nil, err
	}
	socks5Lis, err := net.Listen("tcp", config.Socks5ListenAddr)
	if err != nil {
		return nil, err
	}
	c := &Client{
		session:   session,
		socks5Lis: socks5Lis,
		config:    config,
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
			alog.Info("start reconnection")
			err = c.genNewSession()
			if err != nil {
				alog.Infof("reconnection failed:%v", err)
			}
			continue
		}
		alog.Infof("%s -> %s -> %s, streamID: %v", conn.RemoteAddr().String(), c.session.RemoteAddr().String(), c.config.Socks5ServerAddr, stream.StreamID())
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
func (c *Client) genNewSession() (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.session.Close(ErrReconnection)
	c.session, err = getConn(c.config.Socks5ServerAddr)
	return
}
