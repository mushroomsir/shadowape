package pkg

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	config, err := ParseConfig("../config/server.json")
	checkErr(err)
	server, err := NewServer(config.ServerConfig)
	checkErr(err)

	go server.Run()
	ret := m.Run()
	os.Exit(ret)
}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
