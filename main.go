package main

import (
	"flag"

	"github.com/mushroomsir/logger/alog"
	"github.com/mushroomsir/shadowape/pkg"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "config.json", "config file")
	flag.Parse()
	var config *pkg.Config
	var err error
	if configFile == "" {
		config = &pkg.Config{ServerConfig: &pkg.ServerConfig{ServerAddr: "0.0.0.0:8060"}}
	} else {
		config, err = pkg.ParseConfig(configFile)
		if alog.Check(err) {
			return
		}
	}
	alog.Info(*config)
	if config.ClientConfig != nil {
		client, err := pkg.NewClient(config.ClientConfig)
		if err != nil {
			panic(err)
		}
		client.Run()
	} else {
		server, err := pkg.NewServer(config.ServerConfig)
		if err != nil {
			panic(err)
		}
		server.Run()
	}
}
