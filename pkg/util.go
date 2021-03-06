package pkg

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"time"

	quic "github.com/lucas-clemente/quic-go"
)

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}

func getConn(remoteAddr string, tlsConfig *tls.Config) (quic.Session, error) {
	return quic.DialAddr(remoteAddr, tlsConfig, &quic.Config{
		IdleTimeout:                           time.Hour,
		MaxReceiveStreamFlowControlWindow:     100 * (1 << 20),
		MaxReceiveConnectionFlowControlWindow: 1000 * (1 << 20),
	})
}

func transfer(src, dst io.ReadWriter) error {
	go func() {
		io.Copy(dst, src)
	}()
	_, err := io.Copy(src, dst)
	return err
}

type fakeConn struct {
	session getSession
	quic.Stream
}

func (s *fakeConn) LocalAddr() net.Addr {
	return s.session().LocalAddr()
}

func (s *fakeConn) RemoteAddr() net.Addr {
	return s.session().RemoteAddr()
}

type getSession func() quic.Session

type quicForward struct {
	session getSession
}

func (f *quicForward) Dial(network, addr string) (c net.Conn, err error) {
	stream, err := f.session().OpenStreamSync()
	if err != nil {
		return nil, err
	}
	return &fakeConn{
		session: f.session,
		Stream:  stream,
	}, nil
}
