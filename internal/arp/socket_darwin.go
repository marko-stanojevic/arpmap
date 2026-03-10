//go:build darwin

package arp

import (
	"encoding/binary"
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
	var lastErr error

	// Open the first available BPF device.
	for i := 0; i < 256; i++ {
		name := fmt.Sprintf("/dev/bpf%d", i)
		fd, err = syscall.Open(name, syscall.O_RDWR, 0)
		if err == nil {
			break
		}

		switch err {
		case syscall.EBUSY:
			lastErr = fmt.Errorf("%s is busy", name)
			continue
		case syscall.EACCES, syscall.EPERM:
			return nil, fmt.Errorf("opening %s: %w (run as root or grant access to /dev/bpf*)", name, err)
		case syscall.ENOENT:
			lastErr = fmt.Errorf("%s does not exist", name)
			continue
		default:
			lastErr = fmt.Errorf("opening %s: %w", name, err)
		}
	}
	if err != nil {
		if lastErr != nil {
			return nil, fmt.Errorf("no BPF device available: %w", lastErr)
		}
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
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), 0x80044270 /* BIOCIMMEDIATE */, uintptr(unsafe.Pointer(&enable))); errno != 0 {
		syscall.Close(fd)
		return nil, fmt.Errorf("BIOCIMMEDIATE: %w", errno)
	}

	// Enable promiscuous mode to capture all packets.
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), 0x20004269 /* BIOCPROMISC */, 0); errno != 0 {
		syscall.Close(fd)
		return nil, fmt.Errorf("BIOCPROMISC: %w", errno)
	}

	// Set header complete mode (we provide full Ethernet headers).
	hdrCmplt := uint32(1)
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), 0x80044275 /* BIOCSHDRCMPLT */, uintptr(unsafe.Pointer(&hdrCmplt))); errno != 0 {
		syscall.Close(fd)
		return nil, fmt.Errorf("BIOCSHDRCMPLT: %w", errno)
	}

	// Set non-blocking mode for reads.
	if err := syscall.SetNonblock(fd, true); err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("SetNonblock: %w", err)
	}

	// Set BPF filter for ARP frames only.
	// BPF instructions: ldh [12], jeq 0x0806 pass, ret 65535, ret 0
	type bpfInsn struct {
		Code uint16
		Jt   uint8
		Jf   uint8
		K    uint32
	}
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
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), 0x80104267 /* BIOCSETF */, uintptr(unsafe.Pointer(&bpfProg))); errno != 0 {
		syscall.Close(fd)
		return nil, fmt.Errorf("BIOCSETF: %w", errno)
	}

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
		c.buf = make([]byte, 4096)
	}

	deadline := c.deadline
	for {
		if pkt, ok := c.nextPacket(); ok {
			n := copy(b, pkt)
			return n, nil
		}

		n, err := syscall.Read(c.fd, c.buf)
		if err != nil {
			if errno, ok := err.(syscall.Errno); ok && (errno == syscall.EAGAIN || errno == syscall.EWOULDBLOCK) {
				// Non-blocking read with no data - check deadline
				if !deadline.IsZero() && time.Now().After(deadline) {
					return 0, &net.OpError{Op: "read", Net: "raw", Err: syscall.ETIMEDOUT}
				}
				// Sleep briefly and retry
				time.Sleep(10 * time.Millisecond)
				continue
			}
			return 0, &net.OpError{Op: "read", Net: "raw", Err: err}
		}
		if n <= 0 {
			continue
		}

		c.bufStart = 0
		c.bufEnd = n
	}
}

func (c *bpfConn) nextPacket() ([]byte, bool) {
	for c.bufStart < c.bufEnd {
		record := c.buf[c.bufStart:c.bufEnd]
		hdrLen, capLen, alignedLen, ok := parseBPFRecord(record)
		if !ok {
			c.bufStart = c.bufEnd
			return nil, false
		}

		packetStart := c.bufStart + hdrLen
		packetEnd := packetStart + capLen
		c.bufStart += alignedLen

		if packetEnd > c.bufEnd || packetStart >= packetEnd {
			continue
		}

		return c.buf[packetStart:packetEnd], true
	}

	return nil, false
}

func parseBPFRecord(record []byte) (hdrLen int, capLen int, alignedLen int, ok bool) {
	if len(record) < 18 {
		return 0, 0, 0, false
	}

	// Common Darwin/BSD bpf_hdr layout (8-byte timeval):
	// caplen @ +8, datalen @ +12, hdrlen @ +16.
	if len(record) >= 18 {
		cap := int(binary.LittleEndian.Uint32(record[8:12]))
		hdr := int(binary.LittleEndian.Uint16(record[16:18]))
		if validBPFRecord(len(record), hdr, cap) {
			align := bpfWordAlign(hdr + cap)
			if align <= len(record) {
				return hdr, cap, align, true
			}
		}
	}

	// 64-bit timeval variant fallback:
	// caplen @ +16, datalen @ +20, hdrlen @ +24.
	if len(record) >= 26 {
		cap := int(binary.LittleEndian.Uint32(record[16:20]))
		hdr := int(binary.LittleEndian.Uint16(record[24:26]))
		if validBPFRecord(len(record), hdr, cap) {
			align := bpfWordAlign(hdr + cap)
			if align <= len(record) {
				return hdr, cap, align, true
			}
		}
	}

	return 0, 0, 0, false
}

func validBPFRecord(total, hdr, cap int) bool {
	if hdr <= 0 || cap <= 0 {
		return false
	}
	if hdr > total {
		return false
	}
	if cap > total-hdr {
		return false
	}
	return true
}

func bpfWordAlign(n int) int {
	const alignment = 4
	return (n + (alignment - 1)) &^ (alignment - 1)
}

func (c *bpfConn) Write(b []byte) (int, error) {
	return syscall.Write(c.fd, b)
}

func (c *bpfConn) Close() error                       { return syscall.Close(c.fd) }
func (c *bpfConn) LocalAddr() net.Addr                { return &net.IPAddr{} }
func (c *bpfConn) RemoteAddr() net.Addr               { return &net.IPAddr{} }
func (c *bpfConn) SetDeadline(t time.Time) error      { c.deadline = t; return nil }
func (c *bpfConn) SetReadDeadline(t time.Time) error  { c.deadline = t; return nil }
func (c *bpfConn) SetWriteDeadline(_ time.Time) error { return nil }
