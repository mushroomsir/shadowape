package pkg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	require := require.New(t)
	client, err := NewClient(nil)
	require.Nil(client)
	require.NotNil(err)

	client, err = NewClient(&ClientConfig{Socks5ServerAddr: "x", Socks5ListenAddr: "x"})
	require.Nil(client)
	require.NotNil(err)
}
