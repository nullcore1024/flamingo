package net

import (
	"net"
	"strconv"
)

type InetAddress struct {
	addr *net.TCPAddr
}

func NewInetAddress(ip string, port int) (*InetAddress, error) {
	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(ip, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}
	return &InetAddress{addr: addr}, nil
}

func NewInetAddressFromAddr(addr *net.TCPAddr) *InetAddress {
	return &InetAddress{addr: addr}
}

func (ia *InetAddress) IP() string {
	return ia.addr.IP.String()
}

func (ia *InetAddress) Port() int {
	return ia.addr.Port
}

func (ia *InetAddress) Addr() *net.TCPAddr {
	return ia.addr
}

func (ia *InetAddress) String() string {
	return ia.addr.String()
}
