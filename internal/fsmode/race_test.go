package fsmode_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/adeelahmad/snapback/internal/fsmode"
)

const concurrentCreators = 16

// TestMkdirAllToleratesAConcurrentCreatorOfTheSamePath pins the daemon/web-server
// race (FSMODE-RACE): two processes both call MkdirAll for the same missing state
// dir, and the loser's os.Mkdir sees the winner's directory already there.
func TestMkdirAllToleratesAConcurrentCreatorOfTheSamePath(t *testing.T) {
	skipIfRoot(t)

	path := filepath.Join(t.TempDir(), "a", "b", "c")

	var wg sync.WaitGroup
	errs := make([]error, concurrentCreators)
	for i := 0; i < concurrentCreators; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = fsmode.MkdirAll(path, fsmode.Secure())
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("MkdirAll(%q) call %d returned error: %v", path, i, err)
		}
	}

	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("Lstat(%q) returned error: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%q is not a directory", path)
	}
	if got := info.Mode().Perm(); got != fsmode.SecureDir {
		t.Errorf("mode of %q = %#o, want %#o", path, uint32(got), uint32(fsmode.SecureDir))
	}
}
