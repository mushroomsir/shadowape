package pkg

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/mushroomsir/logger/alog"
)

var (
	errAuthExtraData = errors.New("socks authentication get extra data")
)

const (
	socks5Version = 0x05
	authNone      = 0x00
	socks5Connect = 0x01

	atypeIPV4   = 0x01
	atypeDomain = 0x03
	atypeIPV6   = 0x04
)

func handshake(rw io.ReadWriter) error {
	const (
		idVer     = 0
		idNmethod = 1
	)
	buf := make([]byte, 258)
	n, err := io.ReadAtLeast(rw, buf, idNmethod+1)
	if err != nil {
		return fmt.Errorf("socks5 read version error: %s", err.Error())
	}
	version := buf[idVer]
	if version != socks5Version {
		return fmt.Errorf("socks5 unsupport version: %d", version)
	}
	// NmeThod
	nmethod := int(buf[idNmethod])
	msgLen := nmethod + 2
	if n == msgLen { // handshake done, common case
		// do nothing, jump directly to send confirmation
	} else if n < msgLen { // has more methods to read, rare case
		if _, err = io.ReadFull(rw, buf[n:msgLen]); err != nil {
			return err
		}
	} else { // error, should not get extra data
		return errAuthExtraData
	}
	_, err = rw.Write([]byte{socks5Version, authNone})
	return err
}

func getAddr(rw io.ReadWriter) (string, error) {
	const (
		idVer  = 0
		idCmd  = 1
		idType = 3 // address type index
	)
	// refer to getRequest in server.go for why set buffer size to 263
	buf := make([]byte, 263)
	_, err := io.ReadAtLeast(rw, buf, 5)
	if err != nil {
		return "", err
	}
	// check version and cmd
	if buf[idVer] != socks5Version {
		return "", fmt.Errorf("socks5 read version error: %s", err.Error())
	}
	if buf[idCmd] != socks5Connect {
		return "", fmt.Errorf("socks5 unsupport cmd: %d", buf[idCmd])
	}
	atype := buf[idType]
	var host string
	var hostEnd int
	switch atype {
	case atypeIPV4:
		hostEnd = 4 + net.IPv4len
		host = net.IP(buf[4:hostEnd]).String()
	case atypeIPV6:
		hostEnd = 4 + net.IPv6len
		host = net.IP(buf[4:hostEnd]).String()
	case atypeDomain:
		hostEnd = int(buf[4]) + 5
		host = string(buf[5:hostEnd])
	default:
		return "", fmt.Errorf("socks5 address type: %d", atype)
	}
	port := binary.BigEndian.Uint16(buf[hostEnd : hostEnd+2])
	host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	// Sending connection established message immediately to client.
	// This some round trip time for creating socks connection with the client.
	// But if connection failed, the client will get connection reset error.
	_, err = rw.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x08, 0x43})
	alog.Check(err)
	return host, nil
}
