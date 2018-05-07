package pkg

import (
	"encoding/json"
	"io/ioutil"
)

type ServerConfig struct {
	// ServerAddr
	ServerAddr string `json:"server_addr"`
}
type ClientConfig struct {
	// Socks5ListenAddr ...
	Socks5ListenAddr string `json:"socks5_listen_addr"`
	// HTTPListenAddr ...
	HTTPListenAddr string `json:"http_listen_addr"`
	// Socks5ServerAddr ...
	Socks5ServerAddr string `json:"socks5_Server_addr"`
}

// Config ...
type Config struct {
	ClientConfig *ClientConfig `json:"client_config"`
	ServerConfig *ServerConfig `json:"server_config"`
}

// ParseConfig parse config both from file or env variable
func ParseConfig(configPath string) (*Config, error) {
	config := new(Config)
	data, err := ioutil.ReadFile(configPath)
	if err == nil {
		return nil, err
	}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}
	if config.ClientConfig != nil {
		if config.ClientConfig.HTTPListenAddr == "" {
			config.ClientConfig.HTTPListenAddr = "0.0.0.0:8010"
		}
		if config.ClientConfig.Socks5ListenAddr == "" {
			config.ClientConfig.Socks5ListenAddr = "0.0.0.0:8020"
		}
	} else {
		config.ServerConfig.ServerAddr = "0.0.0.0:8060"

	}
	return config, nil
}
