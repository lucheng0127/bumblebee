package host

import (
	"net"
	"os"
)

func GetHostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "UNKNOW"
	}

	return name
}

func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:4789")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	laddr := conn.LocalAddr().(*net.UDPAddr)
	return laddr.IP.To4(), nil
}
