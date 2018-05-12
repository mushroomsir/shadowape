package pkg

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProxy(t *testing.T) {
	require := require.New(t)

	config, err := ParseConfig("../config/client.json")
	require.Nil(err)
	client, err := NewClient(config.ClientConfig)
	go client.Run()

	proxyURL, err := url.Parse(fmt.Sprintf("http://%v", config.ClientConfig.HTTPListenAddr))
	require.Nil(err)
	myClient := &http.Client{Transport: &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	resp, err := myClient.Get("http://www.baidu.com")
	require.Nil(err)
	require.Equal(200, resp.StatusCode)
	by, err := ioutil.ReadAll(resp.Body)
	require.Nil(err)
	require.NotEmpty(string(by))
}
