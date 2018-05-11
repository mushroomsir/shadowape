package pkg

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/mushroomsir/logger/alog"

	quic "github.com/lucas-clemente/quic-go"
	xproxy "golang.org/x/net/proxy"
)

// ProxyHTTPServer ...
type ProxyHTTPServer struct {
	config        *ClientConfig
	session       quic.Session
	httpTransport *http.Transport
	dialer        xproxy.Dialer
}

// NewProxyHTTPServer ...
func NewProxyHTTPServer(config *ClientConfig, session quic.Session) (*ProxyHTTPServer, error) {
	dialer, err := xproxy.SOCKS5("tcp", config.Socks5ListenAddr, nil, &quicForward{session: session})
	if err != nil {
		return nil, err
	}
	httpTransport := &http.Transport{
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	httpTransport.Dial = dialer.Dial
	return &ProxyHTTPServer{
		config:        config,
		session:       session,
		httpTransport: httpTransport,
		dialer:        dialer,
	}, nil
}

func (p *ProxyHTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	alog.Infof("HTTP(%s): %s -> [%s->%s] -> %s -> %s", r.Method, r.RemoteAddr, p.config.HTTPListenAddr, p.config.Socks5ListenAddr, p.config.Socks5ServerAddr, r.Host)
	if r.Method == http.MethodConnect {
		p.handleTunneling(w, r)
	} else {
		p.handleHTTP(w, r)
	}
}

func (p *ProxyHTTPServer) runhttp() {
	server := &http.Server{
		Addr:    p.config.HTTPListenAddr,
		Handler: http.HandlerFunc(p.ServeHTTP),
	}
	log.Fatal(server.ListenAndServe())
}

// handleTunneling ...
func (p *ProxyHTTPServer) handleTunneling(w http.ResponseWriter, r *http.Request) {
	destConn, err := p.dialer.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(destConn, clientConn)
}
func (p *ProxyHTTPServer) handleHTTP(w http.ResponseWriter, req *http.Request) {
	resp, err := p.httpTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("shadowape-proxy", "true")
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
