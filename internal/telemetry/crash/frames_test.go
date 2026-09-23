package crash

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// fixtureStack is a literal runtime/debug.Stack()-style dump mixing a
// snapback frame, a snapback closure frame, a stdlib frame, a third-party
// (pkg/mod) frame and a package-main frame, to prove the filter keeps only
// this module's frames.
const fixtureStack = `goroutine 1 [running]:
runtime/debug.Stack()
	/usr/local/go/src/runtime/debug/stack.go:24 +0x64
github.com/adeelahmad/snapback/internal/mount.(*Catalog).Refresh(0xc0001a2000)
	/Users/x/work/snapback/internal/mount/catalog.go:123 +0x1f
github.com/adeelahmad/snapback/internal/mount.(*Catalog).run.func1()
	/Users/x/work/snapback/internal/mount/catalog.go:88 +0x2a
github.com/adeelahmad/snapback/internal/resolver.Resolve(...)
	/Users/x/work/snapback/internal/resolver/resolve.go:45 +0x100
golang.org/x/sync/errgroup.(*Group).Go.func1()
	/Users/x/go/pkg/mod/golang.org/x/sync@v0.5.0/errgroup/errgroup.go:75 +0x56
main.main()
	/Users/x/work/snapback/cmd/snapback/main.go:12 +0x30
`

func TestFrames_KeepsOnlySnapbackModulesInOrder(t *testing.T) {
	want := []Frame{
		{Module: "github.com/adeelahmad/snapback/internal/mount", Function: "(*Catalog).Refresh"},
		{Module: "github.com/adeelahmad/snapback/internal/mount", Function: "(*Catalog).run.func1"},
		{Module: "github.com/adeelahmad/snapback/internal/resolver", Function: "Resolve"},
	}

	got := Frames([]byte(fixtureStack))

	if len(got) != len(want) {
		t.Fatalf("Frames() = %+v (len %d), want %+v (len %d)", got, len(got), want, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("frame %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestFrames_NoFilePathLineOrArgValues(t *testing.T) {
	got := Frames([]byte(fixtureStack))
	if len(got) == 0 {
		t.Fatal("Frames() returned no frames to check")
	}
	for _, f := range got {
		if strings.Contains(f.Module, "/Users/") || strings.Contains(f.Function, "/Users/") {
			t.Errorf("frame %+v leaks a file path", f)
		}
		if strings.Contains(f.Module, ".go:") || strings.Contains(f.Function, ".go:") {
			t.Errorf("frame %+v leaks a file:line", f)
		}
		if strings.Contains(f.Module, "0x") || strings.Contains(f.Function, "0x") {
			t.Errorf("frame %+v leaks an argument value", f)
		}
	}
}

func TestFrames_EmptyOrGarbageYieldsNil(t *testing.T) {
	if got := Frames(nil); got != nil {
		t.Errorf("Frames(nil) = %+v, want nil", got)
	}
	if got := Frames([]byte("")); got != nil {
		t.Errorf("Frames(\"\") = %+v, want nil", got)
	}
	if got := Frames([]byte("not a stack trace\njust garbage text\n")); got != nil {
		t.Errorf("Frames(garbage) = %+v, want nil", got)
	}
}

func TestFrames_MarshaledResultPassesScanProhibited(t *testing.T) {
	got := Frames([]byte(fixtureStack))

	b, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	findings := telemetry.ScanProhibited(b)
	if len(findings) != 0 {
		t.Errorf("ScanProhibited(%s) = %+v, want zero findings", b, findings)
	}
}
