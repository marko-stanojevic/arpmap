//go:build windows

package arp

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"syscall"
	"unsafe"

	"github.com/marko-stanojevic/arpmap/internal/iface"
	"github.com/marko-stanojevic/arpmap/internal/output"
)

const (
	errorGenFailure      syscall.Errno = 31
	errorBadNetName      syscall.Errno = 67
	errorSemTimeout      syscall.Errno = 121
	errorNetworkUnreach  syscall.Errno = 1231
	errorHostUnreach     syscall.Errno = 1232
	errorElementNotFound syscall.Errno = 1168
)

var (
	iphlpapiDLL    = syscall.NewLazyDLL("iphlpapi.dll")
	sendARPProc    = iphlpapiDLL.NewProc("SendARP")
	initSendARPOne sync.Once
	initSendARPErr error
)

func scanWindows(info iface.Info) ([]output.Device, error) {
	if err := initializeSendARP(); err != nil {
		return nil, fmt.Errorf("initializing SendARP: %w", err)
	}

	targets := hostsFromNets(info.Nets)
	if len(targets) == 0 {
		return nil, nil
	}

	sem := make(chan struct{}, workerCount)
	var wg sync.WaitGroup
	var mu sync.Mutex
	devicesByIP := make(map[string]string, len(targets))
	errCh := make(chan error, 1)

	for _, ip := range targets {
		target := cloneIP(ip)
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			mac, found, err := resolveMACWithSendARP(target)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
			if !found {
				return
			}

			mu.Lock()
			devicesByIP[target.String()] = mac.String()
			mu.Unlock()
		}()
	}

	wg.Wait()
	close(errCh)
	if err, ok := <-errCh; ok {
		return nil, fmt.Errorf("scanning subnet hosts: %w", err)
	}

	devices := make([]output.Device, 0, len(devicesByIP))
	for ip, mac := range devicesByIP {
		devices = append(devices, output.Device{IP: ip, MAC: mac})
	}

	return devices, nil
}

func initializeSendARP() error {
	initSendARPOne.Do(func() {
		initSendARPErr = sendARPProc.Find()
	})
	return initSendARPErr
}

func resolveMACWithSendARP(target net.IP) (net.HardwareAddr, bool, error) {
	ip4 := target.To4()
	if ip4 == nil {
		return nil, false, fmt.Errorf("target %q is not IPv4", target.String())
	}

	var mac [8]byte
	macLen := uint32(len(mac))
	r1, _, _ := sendARPProc.Call(
		uintptr(binary.BigEndian.Uint32(ip4)),
		0,
		uintptr(unsafe.Pointer(&mac[0])),
		uintptr(unsafe.Pointer(&macLen)),
	)
	if r1 != 0 {
		errno := syscall.Errno(r1)
		if isNoARPResponseError(errno) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("SendARP %s: %w", target.String(), errno)
	}
	if macLen == 0 || macLen > uint32(len(mac)) {
		return nil, false, nil
	}

	resolved := make(net.HardwareAddr, macLen)
	copy(resolved, mac[:macLen])
	return resolved, true, nil
}

func isNoARPResponseError(errno syscall.Errno) bool {
	switch errno {
	case errorGenFailure, errorBadNetName, errorSemTimeout, errorNetworkUnreach, errorHostUnreach, errorElementNotFound:
		return true
	default:
		return false
	}
}
