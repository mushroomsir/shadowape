package pkg

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/mushroomsir/logger/alog"
	xproxy "golang.org/x/net/proxy"

	"github.com/lucas-clemente/quic-go"
)

// Client ...
type Client struct {
	session            quic.Session
	socks5Lis, httpLis net.Listener
	config             *ClientConfig
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
	httpLis, err := net.Listen("tcp", config.HTTPListenAddr)
	if err != nil {
		return nil, err
	}
	c := &Client{
		session:   session,
		socks5Lis: socks5Lis,
		httpLis:   httpLis,
		config:    config,
	}
	return c, nil
}

// Run ...
func (c *Client) Run() {
	go c.runHTTP()
	for {
		conn, err := c.socks5Lis.Accept()
		if alog.Check(err) {
			continue
		}
		go c.handleConn(conn)
	}
}
func (c *Client) runHTTP() error {
	proxy, err := c.httpBridge()
	if alog.Check(err) {
		return err
	}
	return http.Serve(c.httpLis, proxy)
}

func (c *Client) handleConn(conn net.Conn) {
	alog.Info("LocalAddr", conn.LocalAddr(), "RemoteAddr", conn.RemoteAddr())
	defer conn.Close()
	for {
		stream, err := c.session.OpenStreamSync()
		if alog.Check(err) {
			time.Sleep(time.Second)
			c.session, err = getConn(c.config.Socks5ServerAddr)
			alog.Check(err)
			continue
		}
		err = transfer(conn, stream)
		alog.Check(err)
		stream.Close()
	}
}

func (c *Client) httpBridge() (*goproxy.ProxyHttpServer, error) {
	socks5Dailer, err := xproxy.SOCKS5("tcp", c.config.Socks5ListenAddr, nil, &quicForward{session: c.session})
	if alog.Check(err) {
		return nil, err
	}
	httpProxy := goproxy.NewProxyHttpServer()
	httpProxy.ConnectDial = socks5Dailer.Dial
	return httpProxy, nil
}
