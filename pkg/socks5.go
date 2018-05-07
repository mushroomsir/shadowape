package pkg

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
)

const (
	socks5Version = 0x05
	authNone      = 0x00
	cmdTCPConnect = 0x01

	atypeIPV4   = 0x01
	atypeDomain = 0x03
	atypeIPV6   = 0x04
)

func handshake(rw io.ReadWriter) error {
	buf := make([]byte, 1)
	_, err := io.ReadAtLeast(rw, buf, 1)
	if err != nil {
		return fmt.Errorf("socks5 read version error: %s", err.Error())
	}
	version := buf[0]
	if version != socks5Version {
		return fmt.Errorf("socks5 unsupport version: %d", version)
	}
	_, err = rw.Write([]byte{socks5Version, authNone})
	return err
}

func getAddr(rw io.ReadWriter) (string, error) {
	buf := make([]byte, 5)
	_, err := io.ReadAtLeast(rw, buf, 5)
	if err != nil {
		return "", err
	}
	if buf[1] != cmdTCPConnect {
		return "", fmt.Errorf("socks5 unsupport cmd: %d", buf[1])
	}
	atype := buf[3]
	var host string
	var port uint16
	var hostEnd int
	switch atype {
	case atypeIPV4:
		hostEnd = 4 + net.IPv4len
		host = net.IP(buf[4:hostEnd]).String()
	case atypeDomain:
		hostEnd = int(buf[4]) + 5
		host = string(buf[5:hostEnd])
	case atypeIPV6:
		hostEnd = 4 + net.IPv6len
		host = net.IP(buf[4:hostEnd]).String()
	default:
		return "", fmt.Errorf("socks5 address type: %d", atype)
	}
	port = binary.BigEndian.Uint16(buf[hostEnd : hostEnd+2])
	return net.JoinHostPort(host, strconv.Itoa(int(port))), nil
}

func ok(rw io.ReadWriter) error {
	_, err := rw.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x08, 0x43})
	return err
}
