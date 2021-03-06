package pkg

import (
	"encoding/json"
	"io/ioutil"
)

// ServerConfig ...
type ServerConfig struct {
	// ServerAddr
	ServerAddr string `json:"server_addr"`
	// CertFile
	CertFile string `json:"cert_file"`
	// KeyFile
	KeyFile string `json:"key_file"`
}

// ClientConfig ...
type ClientConfig struct {
	// Socks5ListenAddr ...
	Socks5ListenAddr string `json:"socks5_listen_addr"`
	// HTTPListenAddr ...
	HTTPListenAddr string `json:"http_listen_addr"`
	// Socks5ServerAddr ...
	Socks5ServerAddr string `json:"socks5_Server_addr"`
	// CertFile
	CertFile string `json:"root_cert_file"`
}

// Config ...
type Config struct {
	ClientConfig *ClientConfig `json:"client_config,omitempty"`
	ServerConfig *ServerConfig `json:"server_config,omitempty"`
}

// ParseConfig parse config both from file or env variable
func ParseConfig(configPath string) (*Config, error) {
	config := new(Config)
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
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
