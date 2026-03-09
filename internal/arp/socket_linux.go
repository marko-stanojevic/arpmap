//go:build linux

package arp

import (
	"fmt"
	"net"
	"syscall"
	"time"
	"unsafe"
)

// openRawConn opens an AF_PACKET raw socket bound to the given interface.
// Requires CAP_NET_RAW or root privileges.
func openRawConn(ifc *net.Interface) (net.Conn, error) {
	// htons(ETH_P_ALL) = 0x0300
	const ethPALL = 0x0003

	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(ethPALL)))
	if err != nil {
		return nil, fmt.Errorf("socket(AF_PACKET): %w", err)
	}

	sll := syscall.SockaddrLinklayer{
		Protocol: htons(ethPALL),
		Ifindex:  ifc.Index,
	}
	if err := syscall.Bind(fd, &sll); err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("bind: %w", err)
	}

	// BPF filter: only pass ARP frames (ethertype 0x0806).
	filter := []syscall.SockFilter{
		{Code: 0x28, Jt: 0, Jf: 0, K: 12},     // ldh [12]
		{Code: 0x15, Jt: 0, Jf: 1, K: 0x0806}, // jeq #0x0806
		{Code: 0x06, Jt: 0, Jf: 0, K: 65535},  // ret #65535
		{Code: 0x06, Jt: 0, Jf: 0, K: 0},      // ret #0
	}
	prog := syscall.SockFprog{
		Len:    uint16(len(filter)),
		Filter: &filter[0],
	}
	syscall.Syscall6( //nolint:errcheck
		syscall.SYS_SETSOCKOPT,
		uintptr(fd),
		uintptr(syscall.SOL_SOCKET),
		uintptr(syscall.SO_ATTACH_FILTER),
		uintptr(unsafe.Pointer(&prog)),
		uintptr(unsafe.Sizeof(prog)),
		0,
	)

	return &rawConn{fd: fd}, nil
}

func htons(i uint16) uint16 { return (i<<8)&0xff00 | i>>8 }

// rawConn wraps an AF_PACKET fd as a net.Conn.
type rawConn struct {
	fd       int
	deadline time.Time
}

func (c *rawConn) Read(b []byte) (int, error) {
	if !c.deadline.IsZero() {
		tv := syscall.NsecToTimeval(time.Until(c.deadline).Nanoseconds())
		syscall.SetsockoptTimeval(c.fd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv) //nolint:errcheck
	}
	n, err := syscall.Read(c.fd, b)
	if err != nil {
		return 0, &net.OpError{Op: "read", Net: "raw", Err: err}
	}
	return n, nil
}

func (c *rawConn) Write(b []byte) (int, error) {
	n, err := syscall.Write(c.fd, b)
	return n, err
}

func (c *rawConn) Close() error                       { return syscall.Close(c.fd) }
func (c *rawConn) LocalAddr() net.Addr                { return &net.IPAddr{} }
func (c *rawConn) RemoteAddr() net.Addr               { return &net.IPAddr{} }
func (c *rawConn) SetDeadline(t time.Time) error      { c.deadline = t; return nil }
func (c *rawConn) SetReadDeadline(t time.Time) error  { c.deadline = t; return nil }
func (c *rawConn) SetWriteDeadline(_ time.Time) error { return nil }
