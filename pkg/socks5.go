package pkg

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
)

var (
	errAuthExtraData = errors.New("socks authentication get extra data")
	emptyString      = ""
)

const (
	socks5Version = 0x05
	authNone      = 0x00
	socks5Connect = 0x01

	atypeIPV4   = 0x01
	atypeDomain = 0x03
	atypeIPV6   = 0x04
)

// https://www.ietf.org/rfc/rfc1928.txt
// +----+----------+----------+
// |VER | NMETHODS | METHODS  |
// +----+----------+----------+
// | 1  |    1     | 1 to 255 |
// +----+----------+----------+
func handshake(rw io.ReadWriter) error {
	buf := make([]byte, 1+1)
	_, err := io.ReadAtLeast(rw, buf, 1+1)
	if err != nil {
		return fmt.Errorf("socks5 read version error: %s", err.Error())
	}
	version := buf[0]
	if version != socks5Version {
		return fmt.Errorf("socks5 unsupport version: %d", version)
	}
	nmethods := int(buf[1])
	bufMethods := make([]byte, nmethods)
	_, err = io.ReadAtLeast(rw, bufMethods, nmethods)
	if err != nil {
		return fmt.Errorf("socks5 read version error: %s", err.Error())
	}
	_, err = rw.Write([]byte{socks5Version, authNone})
	return err
}

// +----+-----+-------+------+----------+----------+
// |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
// +----+-----+-------+------+----------+----------+
// | 1  |  1  | X'00' |  1   | Variable |    2     |
// +----+-----+-------+------+----------+----------+
func getAddr(rw io.ReadWriter) (string, error) {
	buf := make([]byte, 4)
	_, err := io.ReadAtLeast(rw, buf, 4)
	if err != nil {
		return "", err
	}
	if buf[0] != socks5Version {
		return "", fmt.Errorf("socks5 read version error: %s", err.Error())
	}
	if buf[1] != socks5Connect {
		return "", fmt.Errorf("socks5 unsupport cmd: %d", buf[1])
	}
	atype := buf[3]
	addrLen := 0
	switch atype {
	case atypeIPV4:
		addrLen = net.IPv4len
	case atypeIPV6:
		addrLen = net.IPv6len
	case atypeDomain:
		buf := make([]byte, 1)
		_, err := io.ReadAtLeast(rw, buf, 1)
		if err != nil {
			return emptyString, err
		}
		addrLen = int(buf[0])
	default:
		return emptyString, fmt.Errorf("socks5 address type: %d", atype)
	}
	buf = make([]byte, addrLen+2)
	_, err = io.ReadAtLeast(rw, buf, addrLen+2)
	if err != nil {
		return emptyString, err
	}
	host := net.IP(buf[:addrLen]).String()
	port := binary.BigEndian.Uint16(buf[addrLen : addrLen+2])
	host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	_, err = rw.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x08, 0x43})
	return host, err
}
