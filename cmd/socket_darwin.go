//go:build darwin

package arp

import (
	"fmt"
	"net"
	"syscall"
	"time"
	"unsafe"
)

// openRawConn opens a BPF device on macOS for raw Ethernet capture.
// Requires root or appropriate entitlements.
func openRawConn(ifc *net.Interface) (net.Conn, error) {
	var fd int
	var err error

	// Open the first available BPF device.
	for i := 0; i < 256; i++ {
		name := fmt.Sprintf("/dev/bpf%d", i)
		fd, err = syscall.Open(name, syscall.O_RDWR, 0)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("no BPF device available: %w", err)
	}

	// Bind to network interface.
	ifReq := struct {
		Name [syscall.IFNAMSIZ]byte
	}{}
	copy(ifReq.Name[:], ifc.Name)
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), 0x8020426c /* BIOCSETIF */, uintptr(unsafe.Pointer(&ifReq))); errno != 0 {
		syscall.Close(fd)
		return nil, fmt.Errorf("BIOCSETIF: %w", errno)
	}

	// Enable immediate mode so reads return as soon as data is available.
	enable := uint32(1)
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), 0x80044266 /* BIOCIMMEDIATE */, uintptr(unsafe.Pointer(&enable))); errno != 0 {
		syscall.Close(fd)
		return nil, fmt.Errorf("BIOCIMMEDIATE: %w", errno)
	}

	// Set BPF filter for ARP frames only.
	// BPF instructions: ldh [12], jeq 0x0806 pass, ret 65535, ret 0
	type bpfInsn struct{ Code, Jt, Jf uint16; K uint32 }
	filter := []bpfInsn{
		{0x28, 0, 0, 12},
		{0x15, 0, 1, 0x0806},
		{0x06, 0, 0, 65535},
		{0x06, 0, 0, 0},
	}
	bpfProg := struct {
		Len   uint32
		Pad   uint32
		Insns uintptr
	}{
		Len:   uint32(len(filter)),
		Insns: uintptr(unsafe.Pointer(&filter[0])),
	}
	syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), 0x80104267 /* BIOCSETF */, uintptr(unsafe.Pointer(&bpfProg))) //nolint:errcheck

	return &bpfConn{fd: fd, ifc: ifc}, nil
}

type bpfConn struct {
	fd       int
	ifc      *net.Interface
	deadline time.Time
	buf      []byte
	bufStart int
	bufEnd   int
}

func (c *bpfConn) Read(b []byte) (int, error) {
	if len(c.buf) == 0 {
		c.buf = make([]byte, 65536)
	}
	if c.bufStart < c.bufEnd {
		n := copy(b, c.buf[c.bufStart:c.bufEnd])
		c.bufStart += n
		return n, nil
	}

	if !c.deadline.IsZero() {
		tv := syscall.NsecToTimeval(time.Until(c.deadline).Nanoseconds())
		syscall.SetsockoptTimeval(c.fd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv) //nolint:errcheck
	}

	n, err := syscall.Read(c.fd, c.buf)
	if err != nil {
		return 0, &net.OpError{Op: "read", Net: "raw", Err: err}
	}
	// BPF prepends a header; skip it (bpf_hdr is typically 18 bytes aligned).
	// For simplicity we skip the standard 18-byte BPF header.
	const bpfHdrLen = 18
	if n < bpfHdrLen {
		return 0, nil
	}
	c.bufStart = bpfHdrLen
	c.bufEnd = n
	copied := copy(b, c.buf[c.bufStart:c.bufEnd])
	c.bufStart += copied
	return copied, nil
}

func (c *bpfConn) Write(b []byte) (int, error) {
	return syscall.Write(c.fd, b)
}

func (c *bpfConn) Close() error               { return syscall.Close(c.fd) }
func (c *bpfConn) LocalAddr() net.Addr        { return &net.IPAddr{} }
func (c *bpfConn) RemoteAddr() net.Addr       { return &net.IPAddr{} }
func (c *bpfConn) SetDeadline(t time.Time) error      { c.deadline = t; return nil }
func (c *bpfConn) SetReadDeadline(t time.Time) error  { c.deadline = t; return nil }
func (c *bpfConn) SetWriteDeadline(_ time.Time) error { return nil }
