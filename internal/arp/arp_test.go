package arp

import (
	"net"
	"testing"
	"time"

	"github.com/marko-stanojevic/arpmap/internal/output"
)

func TestBroadcastAddr(t *testing.T) {
	t.Parallel()

	network := &net.IPNet{
		IP:   net.IPv4(192, 168, 1, 0),
		Mask: net.CIDRMask(24, 32),
	}

	got := broadcastAddr(network)
	want := net.IPv4(192, 168, 1, 255)
	if !got.Equal(want) {
		t.Fatalf("broadcastAddr() = %v, want %v", got, want)
	}
}

func TestHostsFromNet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		network *net.IPNet
		want    []string
	}{
		{
			name: "slash 30",
			network: &net.IPNet{
				IP:   net.IPv4(192, 168, 1, 0),
				Mask: net.CIDRMask(30, 32),
			},
			want: []string{"192.168.1.1", "192.168.1.2"},
		},
		{
			name: "slash 31",
			network: &net.IPNet{
				IP:   net.IPv4(10, 0, 0, 0),
				Mask: net.CIDRMask(31, 32),
			},
			want: nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			hosts := hostsFromNet(test.network)
			if len(hosts) != len(test.want) {
				t.Fatalf("hostsFromNet() len = %d, want %d", len(hosts), len(test.want))
			}
			for i, host := range hosts {
				if host.String() != test.want[i] {
					t.Fatalf("hostsFromNet()[%d] = %s, want %s", i, host.String(), test.want[i])
				}
			}
		})
	}
}

func TestRetryProbeAttempts(t *testing.T) {
	t.Parallel()

	var attempts int
	done, err := retryProbeAttempts(3, 0, func() (bool, error) {
		attempts++
		return attempts == 2, nil
	})
	if err != nil {
		t.Fatalf("retryProbeAttempts() unexpected error: %v", err)
	}
	if !done {
		t.Fatal("retryProbeAttempts() done = false, want true")
	}
	if attempts != 2 {
		t.Fatalf("retryProbeAttempts() attempts = %d, want 2", attempts)
	}
}

func TestRetryProbeAttemptsReturnsError(t *testing.T) {
	t.Parallel()

	wantErr := net.InvalidAddrError("boom")
	_, err := retryProbeAttempts(2, time.Millisecond, func() (bool, error) {
		return false, wantErr
	})
	if err == nil {
		t.Fatal("retryProbeAttempts() error = nil, want non-nil")
	}
	if err != wantErr {
		t.Fatalf("retryProbeAttempts() error = %v, want %v", err, wantErr)
	}
}

func TestSortDevices(t *testing.T) {
	t.Parallel()

	devices := []output.Device{
		{IP: "192.168.1.20", MAC: "aa:bb:cc:dd:ee:20"},
		{IP: "192.168.1.3", MAC: "aa:bb:cc:dd:ee:03"},
		{IP: "192.168.1.100", MAC: "aa:bb:cc:dd:ee:64"},
	}

	sortDevices(devices)
	want := []string{"192.168.1.3", "192.168.1.20", "192.168.1.100"}
	for i, device := range devices {
		if device.IP != want[i] {
			t.Fatalf("sortDevices()[%d] = %s, want %s", i, device.IP, want[i])
		}
	}
}
