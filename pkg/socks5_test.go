package pkg

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	xproxy "golang.org/x/net/proxy"
)

func TestSocks5(t *testing.T) {
	require := require.New(t)
	addr := ":8020"
	reqURL := "127.0.0.1:80"

	go func() {
		dailer, err := xproxy.SOCKS5("tcp", addr, nil, &xproxy.Direct)
		require.Nil(err)
		conn, err := dailer.Dial("tcp", reqURL)
		require.Nil(err)
		defer conn.Close()
	}()

	lis, err := net.Listen("tcp", addr)
	require.Nil(err)
	for {
		conn, err := lis.Accept()
		require.Nil(err)

		err = handshake(conn)
		require.Nil(err)

		host, err := getAddr(conn)
		require.Nil(err)
		require.Equal(reqURL, host)
		return
	}
}
