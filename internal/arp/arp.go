// Package arp implements ARP-based host discovery for local subnets.
//
// It crafts raw ARP request frames, broadcasts them over the wire, and
// collects ARP replies using a read loop with a configurable timeout.
// All host addresses within each subnet are probed concurrently with a
// bounded worker pool to avoid exhausting OS file-descriptor limits.
package arp

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/marko-stanojevic/arpmap/internal/iface"
	"github.com/marko-stanojevic/arpmap/internal/output"
)

const (
	// workerCount limits concurrent ARP senders to avoid fd exhaustion.
	workerCount = 256
	// probeTimeout is how long we wait for lingering ARP replies after sending.
	probeTimeout = 2 * time.Second
	// readDeadline is the rolling per-read timeout on the raw socket.
	readDeadline = 200 * time.Millisecond
)

const etherTypeARP = 0x0806

// arpPacket represents an ARP frame payload for IPv4/Ethernet (28 bytes).
type arpPacket struct {
	HType uint16
	PType uint16
	HLen  uint8
	PLen  uint8
	Op    uint16
	SHA   [6]byte // Sender hardware address
	SPA   [4]byte // Sender protocol address
	THA   [6]byte // Target hardware address
	TPA   [4]byte // Target protocol address
}

func marshalARP(p arpPacket) []byte {
	b := make([]byte, 28)
	binary.BigEndian.PutUint16(b[0:], p.HType)
	binary.BigEndian.PutUint16(b[2:], p.PType)
	b[4] = p.HLen
	b[5] = p.PLen
	binary.BigEndian.PutUint16(b[6:], p.Op)
	copy(b[8:], p.SHA[:])
	copy(b[14:], p.SPA[:])
	copy(b[20:], p.THA[:])
	copy(b[24:], p.TPA[:])
	return b
}

func buildEthernetFrame(src, dst net.HardwareAddr, payload []byte) []byte {
	frame := make([]byte, 14+len(payload))
	copy(frame[0:6], dst)
	copy(frame[6:12], src)
	binary.BigEndian.PutUint16(frame[12:14], etherTypeARP)
	copy(frame[14:], payload)
	return frame
}

// Scan sends ARP requests to every host in every subnet of the given interface
// and returns the set of responding devices.
func Scan(info iface.Info) ([]output.Device, error) {
	conn, err := openRawConn(info.Iface)
	if err != nil {
		return nil, fmt.Errorf("opening raw socket on %s: %w", info.Name, err)
	}
	defer conn.Close()

	targets := hostsFromNets(info.Nets)

	var (
		mu      sync.Mutex
		devices = make(map[string]string) // ip → mac
	)

	// Background reader goroutine — collects ARP replies.
	stop := make(chan struct{})
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		buf := make([]byte, 1500)
		for {
			select {
			case <-stop:
				return
			default:
			}

			if err := conn.SetReadDeadline(time.Now().Add(readDeadline)); err != nil {
				return
			}
			n, err := conn.Read(buf)
			if err != nil {
				if isTimeout(err) {
					continue
				}
				return
			}
			if n < 42 {
				continue
			}
			// Ethernet header is 14 bytes; ARP starts at offset 14.
			if binary.BigEndian.Uint16(buf[12:14]) != etherTypeARP {
				continue
			}
			arpOp := binary.BigEndian.Uint16(buf[20:22])
			if arpOp != 2 { // 2 = ARP reply
				continue
			}
			senderMAC := net.HardwareAddr(append([]byte{}, buf[22:28]...)).String()
			senderIP := net.IP(append([]byte{}, buf[28:32]...)).String()

			mu.Lock()
			devices[senderIP] = senderMAC
			mu.Unlock()
		}
	}()

	// Send ARP requests with bounded concurrency.
	sem := make(chan struct{}, workerCount)
	var wg sync.WaitGroup
	for _, ip := range targets {
		wg.Add(1)
		sem <- struct{}{}
		go func(target net.IP) {
			defer wg.Done()
			defer func() { <-sem }()
			sendARP(conn, info.Iface, target)
		}(ip)
	}
	wg.Wait()

	// Allow extra time for replies to arrive.
	time.Sleep(probeTimeout)
	close(stop)
	<-readerDone

	var result []output.Device
	for ip, mac := range devices {
		result = append(result, output.Device{IP: ip, MAC: mac})
	}
	return result, nil
}

// FindFree scans the subnet and returns IP addresses that did NOT respond.
// If max > 0 the result is capped at max addresses.
func FindFree(info iface.Info, max int) ([]string, error) {
	devices, err := Scan(info)
	if err != nil {
		return nil, err
	}

	used := make(map[string]struct{}, len(devices))
	for _, d := range devices {
		used[d.IP] = struct{}{}
	}
	for _, network := range info.Nets {
		// Reserve network address and broadcast.
		used[network.IP.String()] = struct{}{}
		used[broadcastAddr(network).String()] = struct{}{}
	}

	var free []string
	for _, network := range info.Nets {
		for _, ip := range hostsFromNet(network) {
			if _, inUse := used[ip.String()]; !inUse {
				free = append(free, ip.String())
				if max > 0 && len(free) >= max {
					return free, nil
				}
			}
		}
	}
	return free, nil
}

func broadcastAddr(network *net.IPNet) net.IP {
	ip := network.IP.To4()
	mask := network.Mask
	bcast := make(net.IP, 4)
	for i := range bcast {
		bcast[i] = ip[i] | ^mask[i]
	}
	return bcast
}

func hostsFromNets(nets []*net.IPNet) []net.IP {
	var all []net.IP
	for _, n := range nets {
		all = append(all, hostsFromNet(n)...)
	}
	return all
}

func hostsFromNet(network *net.IPNet) []net.IP {
	ip := cloneIP(network.IP.To4())
	incIP(ip) // skip network address

	bcast := broadcastAddr(network)
	var hosts []net.IP
	for !ip.Equal(bcast) {
		hosts = append(hosts, cloneIP(ip))
		incIP(ip)
	}
	return hosts
}

func cloneIP(ip net.IP) net.IP {
	c := make(net.IP, len(ip))
	copy(c, ip)
	return c
}

func incIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}

func sendARP(conn net.Conn, ifc *net.Interface, target net.IP) {
	srcMAC := ifc.HardwareAddr
	if len(srcMAC) == 0 {
		return
	}

	addrs, err := ifc.Addrs()
	if err != nil {
		return
	}
	var srcIP net.IP
	for _, a := range addrs {
		if ipNet, ok := a.(*net.IPNet); ok {
			if ip4 := ipNet.IP.To4(); ip4 != nil {
				srcIP = ip4
				break
			}
		}
	}
	if srcIP == nil {
		return
	}

	pkt := arpPacket{HType: 1, PType: 0x0800, HLen: 6, PLen: 4, Op: 1}
	copy(pkt.SHA[:], srcMAC)
	copy(pkt.SPA[:], srcIP)
	copy(pkt.TPA[:], target.To4())

	payload := marshalARP(pkt)
	bcast := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	frame := buildEthernetFrame(srcMAC, bcast, payload)
	if _, err := conn.Write(frame); err != nil {
		return
	}
}

func isTimeout(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}
	return false
}
