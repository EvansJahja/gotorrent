package domain

import "net"

type Host struct {
	IP   net.IP
	Port uint16
}

func (h Host) Equal(another Host) bool {
	return h.Port == another.Port && h.IP.Equal(another.IP)
}
