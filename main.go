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
	config, err := pkg.ParseConfig(configFile)
	if alog.Check(err) {
		return
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
