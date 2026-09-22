package gofuse

import (
	"os/exec"
	"slices"
	"strings"
	"testing"
)

const gofusePkg = "github.com/adeelahmad/snapback/internal/mount/gofuse"

func TestGofuseDepsIncludeFsAndFuse(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", gofusePkg).CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps %s: %v\n%s", gofusePkg, err, out)
	}
	deps := strings.Fields(string(out))
	if len(deps) == 0 {
		t.Fatalf("go list -deps %s printed no packages", gofusePkg)
	}
	for _, want := range []string{
		"github.com/hanwen/go-fuse/v2/fs",
		"github.com/hanwen/go-fuse/v2/fuse",
	} {
		if !slices.Contains(deps, want) {
			t.Errorf("go list -deps %s lacks %s", gofusePkg, want)
		}
	}
}
