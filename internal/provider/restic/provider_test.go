package restic

import (
	"os"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/provider"
)

func TestProviderSatisfiesSnapshotProvider(t *testing.T) {
	p, err := New(validOptions())
	if err != nil {
		t.Fatalf("New(validOptions()) error = %v, want nil", err)
	}
	if p == nil {
		t.Fatal("New(validOptions()) = nil, want non-nil *Provider")
	}
	// A non-nil *Provider held in the interface is a non-nil SnapshotProvider.
	var _ provider.SnapshotProvider = p

	src, err := os.ReadFile("provider.go")
	if err != nil {
		t.Fatalf("read provider.go: %v", err)
	}
	const want = "var _ provider.SnapshotProvider = (*Provider)(nil)"
	if !strings.Contains(string(src), want) {
		t.Errorf("provider.go does not contain %q", want)
	}
}
