package iface

import (
	"strings"
	"testing"
)

func TestResolveUnknownInterface(t *testing.T) {
	t.Parallel()

	const missing = "definitely-not-a-real-interface-name"
	_, err := Resolve(missing)
	if err == nil {
		t.Fatal("Resolve() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), missing) {
		t.Fatalf("Resolve() error = %q, want interface name %q in error", err.Error(), missing)
	}
}
