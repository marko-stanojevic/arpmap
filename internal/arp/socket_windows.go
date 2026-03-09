//go:build windows

package arp

import (
	"fmt"
	"net"
)

func openRawConn(_ *net.Interface) (net.Conn, error) {
	return nil, fmt.Errorf("raw ARP sockets are not supported on windows")
}
