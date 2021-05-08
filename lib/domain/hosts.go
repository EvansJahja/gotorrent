package domain

import "net"

type Host struct {
	IP   net.IP
	Port uint16
}
