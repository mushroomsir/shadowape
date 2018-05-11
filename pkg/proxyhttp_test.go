package pkg

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/mushroomsir/logger/alog"
	"github.com/stretchr/testify/require"
)

func TestProxy(t *testing.T) {
	require := require.New(t)

	config, err := ParseConfig("../test.json")
	if alog.Check(err) {
		return
	}
	client, err := NewClient(config.ClientConfig)
	go client.Run()

	proxyURL, err := url.Parse(fmt.Sprintf("http://%v", config.ClientConfig.HTTPListenAddr))
	require.Nil(err)
	myClient := &http.Client{Transport: &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	resp, err := myClient.Get("http://google.com")
	require.Nil(err)
	require.Equal(200, resp.StatusCode)
	require.Equal("true", resp.Header.Get("shadowape-proxy"))
}
