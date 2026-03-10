// Package arp implements ARP-based host discovery for local subnets.
//
// It crafts raw ARP request frames, broadcasts them over the wire, and
// collects ARP replies using a read loop with a configurable timeout.
// All host addresses within each subnet are probed concurrently with a
// bounded worker pool to avoid exhausting OS file-descriptor limits.
package arp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/marko-stanojevic/arpmap/internal/iface"
	"github.com/marko-stanojevic/arpmap/internal/output"
)

const (
	// workerCount limits concurrent ARP senders to avoid fd exhaustion.
	workerCount = 256
	// debugSampleIPLogCap limits the number of sample IPs printed in debug output.
	debugSampleIPLogCap = 3
	// defaultProbeAttempts is the number of ARP probes sent per target by default.
	defaultProbeAttempts = 1
	// probeTimeout is how long we wait for lingering ARP replies after sending.
	probeTimeout = 2 * time.Second
	// readDeadline is the rolling per-read timeout on the raw socket.
	readDeadline = 200 * time.Millisecond
	// probeRetryDelay spaces repeated probes for the same target.
	probeRetryDelay = 150 * time.Millisecond
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

// ScanConfig holds options for the Scan operation.
type ScanConfig struct {
	Debug           bool
	Workers         int
	Attempts        int
	StopAtFreeCount int // Stop scanning when this many free IPs are found (0 = scan all)
}

// ScanOption is a functional option for configuring Scan.
type ScanOption func(*ScanConfig)

// WithDebug enables or disables debug logging for scanning operations.
func WithDebug(enabled bool) ScanOption {
	return func(cfg *ScanConfig) {
		cfg.Debug = enabled
	}
}

// WithWorkers sets the number of concurrent probe workers.
// Values <= 0 use platform defaults.
func WithWorkers(workers int) ScanOption {
	return func(cfg *ScanConfig) {
		cfg.Workers = workers
	}
}

// WithAttempts sets the number of ARP probe attempts per target.
// Values <= 0 use the default (1).
func WithAttempts(attempts int) ScanOption {
	return func(cfg *ScanConfig) {
		cfg.Attempts = attempts
	}
}

// WithStopAtFreeCount enables early exit when finding N free IPs.
// When N > 0, the scan stops as soon as N free (non-responding) IPs are identified.
func WithStopAtFreeCount(n int) ScanOption {
	return func(cfg *ScanConfig) {
		cfg.StopAtFreeCount = n
	}
}

// Scan sends ARP requests to every host in every subnet of the given interface
// and returns the set of responding devices.
func Scan(info iface.Info, opts ...ScanOption) ([]output.Device, error) {
	cfg := &ScanConfig{Debug: false, Workers: 0, Attempts: defaultProbeAttempts}
	for _, opt := range opts {
		opt(cfg)
	}
	return scan(info, cfg)
}

// scan performs the actual ARP scanning with the given configuration.
func scan(info iface.Info, cfg *ScanConfig) ([]output.Device, error) {
	debug := cfg.Debug
	scanStart := time.Now()
	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Scan started | interface=%s mac=%s subnets=%s\n", info.Name, info.Iface.HardwareAddr.String(), info.CIDRs)
	}

	if runtime.GOOS == "windows" {
		devices, err := scanWindows(info, debug, cfg.Workers, cfg.Attempts, cfg.StopAtFreeCount)
		if err != nil {
			return nil, fmt.Errorf("scanning on windows: %w", err)
		}
		sortDevices(devices)
		return devices, nil
	}

	conn, err := openRawConn(info.Iface)
	if err != nil {
		return nil, fmt.Errorf("opening raw socket on %s: %w", info.Name, err)
	}
	defer conn.Close()

	targets := hostsFromNets(info.Nets)
	attempts := cfg.Attempts
	if attempts <= 0 {
		attempts = defaultProbeAttempts
	}
	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Scan parameters | targets=%d attempts=%d\n", len(targets), attempts)
	}

	var (
		mu         sync.Mutex
		devices    = make(map[string]string) // ip → mac
		probesSent int
	)

	// Background reader goroutine — collects ARP replies.
	stop := make(chan struct{})
	readerDone := make(chan struct{})
	earlyChan := make(chan struct{}) // Signal early exit condition
	readCount := 0
	timeoutCount := 0
	go func() {
		defer close(readerDone)
		defer func() {
			if debug {
				fmt.Fprintf(os.Stderr, "[DEBUG] Reader summary | reads=%d timeouts=%d unique_devices=%d\n", readCount, timeoutCount, len(devices))
			}
		}()
		buf := make([]byte, 4096) // BPF buffer must match system BPF buffer size
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
					timeoutCount++
					continue
				}
				return
			}
			readCount++
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

			if debug {
				fmt.Fprintf(os.Stderr, "[DEBUG] ARP response | sender_ip=%s sender_mac=%s frame_size=%d\n", senderIP, senderMAC, n)
			}

			mu.Lock()
			devices[senderIP] = senderMAC
			// Check for early exit: if we have N confirmed free IPs (probed but no response).
			if cfg.StopAtFreeCount > 0 && probesSent > 0 {
				inUseCount := len(devices)
				confirmedFreeCount := probesSent - inUseCount
				if debug && confirmedFreeCount >= cfg.StopAtFreeCount {
					fmt.Fprintf(os.Stderr, "[DEBUG] Early-exit threshold reached | probed=%d responded=%d confirmed_free=%d required=%d\n", probesSent, inUseCount, confirmedFreeCount, cfg.StopAtFreeCount)
				}
				if confirmedFreeCount >= cfg.StopAtFreeCount {
					select {
					case <-earlyChan:
					default:
						close(earlyChan)
					}
				}
			}
			mu.Unlock()
		}
	}()

	// Send ARP requests with bounded concurrency.
	workers := workerCount
	if cfg.Workers > 0 {
		workers = cfg.Workers
	}
	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Dispatch settings | workers=%d\n", workers)
	}
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	for _, ip := range targets {
		// Check if early exit has been triggered
		if cfg.StopAtFreeCount > 0 {
			select {
			case <-earlyChan:
				mu.Lock()
				sent := probesSent
				mu.Unlock()
				if debug {
					fmt.Fprintf(os.Stderr, "[DEBUG] Probe dispatch stopped early | dispatched=%d total_targets=%d\n", sent, len(targets))
				}
				goto waitForProbes
			default:
			}
		}

		wg.Add(1)
		sem <- struct{}{}
		mu.Lock()
		probesSent++
		mu.Unlock()
		go func(target net.IP) {
			defer wg.Done()
			defer func() { <-sem }()
			sendARPWithRetries(conn, info.Iface, target, attempts, debug)
		}(ip)
	}
waitForProbes:
	wg.Wait()

	// Allow extra time for replies to arrive, unless we've found enough free IPs.
	if cfg.StopAtFreeCount > 0 {
		select {
		case <-earlyChan:
			if debug {
				fmt.Fprintf(os.Stderr, "[DEBUG] Early exit activated | sufficient free addresses identified\n")
			}
			// Brief delay for any final replies.
			time.Sleep(100 * time.Millisecond)
		default:
			// Normal timeout if early exit not yet triggered.
			time.Sleep(probeTimeout)
		}
	} else {
		time.Sleep(probeTimeout)
	}
	close(stop)
	<-readerDone

	result := make([]output.Device, 0, len(devices))
	for ip, mac := range devices {
		result = append(result, output.Device{IP: ip, MAC: mac})
	}
	sortDevices(result)

	responseSampleIPs := make([]string, 0, debugSampleIPLogCap)
	noResponseSampleIPs := make([]string, 0, debugSampleIPLogCap)
	for _, target := range targets {
		targetIP := target.String()
		if _, ok := devices[targetIP]; ok {
			if len(responseSampleIPs) < debugSampleIPLogCap {
				responseSampleIPs = append(responseSampleIPs, targetIP)
			}
			continue
		}
		if len(noResponseSampleIPs) < debugSampleIPLogCap {
			noResponseSampleIPs = append(noResponseSampleIPs, targetIP)
		}
		if len(responseSampleIPs) >= debugSampleIPLogCap && len(noResponseSampleIPs) >= debugSampleIPLogCap {
			break
		}
	}

	if debug {
		duration := time.Since(scanStart)
		fmt.Fprintf(os.Stderr, "[DEBUG] Scan completed | duration=%v dispatched=%d total_targets=%d\n", duration, probesSent, len(targets))
		if probesSent > 0 {
			fmt.Fprintf(os.Stderr, "[DEBUG] Response metrics | responded=%d dispatched=%d response_rate=%.1f%%\n", len(devices), probesSent, float64(len(devices))*100/float64(probesSent))
		}
		if len(responseSampleIPs) > 0 {
			fmt.Fprintf(os.Stderr, "[DEBUG] Sample IP addresses responding to ARP requests: %v\n", responseSampleIPs)
		}
		if len(noResponseSampleIPs) > 0 {
			fmt.Fprintf(os.Stderr, "[DEBUG] Sample IP addresses with no ARP response: %v\n", noResponseSampleIPs)
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] Scan finished\n")
	}

	return result, nil
}

// FindFree scans the subnet and returns IP addresses that did NOT respond.
// If max > 0 the result is capped at max addresses.
func FindFree(info iface.Info, max int, opts ...ScanOption) ([]string, error) {
	cfg := &ScanConfig{Debug: false, Workers: 0, Attempts: defaultProbeAttempts}
	for _, opt := range opts {
		opt(cfg)
	}
	return findFree(info, max, cfg)
}

// findFree performs the actual free IP lookup with the given configuration.
func findFree(info iface.Info, max int, cfg *ScanConfig) ([]string, error) {
	// Apply early-exit optimization if a limit is requested.
	scanCfg := &ScanConfig{
		Debug:           cfg.Debug,
		Workers:         cfg.Workers,
		Attempts:        cfg.Attempts,
		StopAtFreeCount: max, // Apply early exit when a limit is set
	}

	devices, err := scan(info, scanCfg)
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

	free := make([]string, 0)
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

func sendARP(conn net.Conn, ifc *net.Interface, target net.IP, debug bool) {
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

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] ARP request: who-has %s tell %s (from %s) (packet size: %d bytes)\n", target, srcIP, srcMAC.String(), len(frame))
	}
}

func sendARPWithRetries(conn net.Conn, ifc *net.Interface, target net.IP, attempts int, debug bool) {
	if attempts <= 0 {
		attempts = defaultProbeAttempts
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		sendARP(conn, ifc, target, debug)
		if attempt < attempts {
			time.Sleep(probeRetryDelay)
		}
	}
}

func isTimeout(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}
	return false
}

func sortDevices(devices []output.Device) {
	sort.Slice(devices, func(i, j int) bool {
		left := net.ParseIP(devices[i].IP).To4()
		right := net.ParseIP(devices[j].IP).To4()
		if left != nil && right != nil {
			return bytes.Compare(left, right) < 0
		}
		return devices[i].IP < devices[j].IP
	})
}
