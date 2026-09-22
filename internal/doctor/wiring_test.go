package doctor

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
)

const placeholderText = "not available yet"

var placeholderTexts = []string{placeholderText, "not implemented yet"}

// TestProductionProbesNotPlaceholders checks that the probes Command uses
// reach the real system: with no restic on PATH and no daemon socket, each
// probe reports prerequisite_missing instead of a placeholder.
func TestProductionProbesNotPlaceholders(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", "")
	t.Setenv("HOME", dir)
	t.Setenv("XDG_RUNTIME_DIR", dir)
	ctx := context.Background()
	p := realProbes()

	t.Run("Repos", func(t *testing.T) {
		repo := config.Repository{ID: "main", Repository: dir, ResticBinary: "restic", PasswordFile: dir + "/pw"}
		validator, lister := p.Repos(repo)
		if validator == nil || lister == nil {
			t.Fatalf("realProbes().Repos(%q) = (%v, %v), want a restic validator and lister", repo.ID, validator, lister)
		}
		_, err := validator.Validate(ctx)
		if err == nil {
			t.Fatalf("Validate() with empty PATH = nil error, want %s", errcode.PrereqMissing)
		}
		if got, want := errcode.Of(err), errcode.PrereqMissing; got != want {
			t.Errorf("errcode.Of(Validate()) = %q, want %q (err: %v)", got, want, err)
		}
		if !strings.Contains(err.Error(), "restic") {
			t.Errorf("Validate() error = %q, want it to name restic", err)
		}
		if strings.Contains(err.Error(), placeholderText) {
			t.Errorf("Validate() error = %q, want no placeholder text", err)
		}
	})

	t.Run("DialStatus", func(t *testing.T) {
		_, err := p.DialStatus(ctx)
		if err == nil {
			t.Fatalf("DialStatus() against a missing socket = nil error, want %s", errcode.PrereqMissing)
		}
		if got, want := errcode.Of(err), errcode.PrereqMissing; got != want {
			t.Errorf("errcode.Of(DialStatus()) = %q, want %q (err: %v)", got, want, err)
		}
		if strings.Contains(err.Error(), placeholderText) {
			t.Errorf("DialStatus() error = %q, want no placeholder text", err)
		}
	})

	t.Run("MountTest", func(t *testing.T) {
		err := p.MountTest(ctx)
		if err == nil {
			t.Fatalf("MountTest() with empty PATH = nil error, want %s", errcode.PrereqMissing)
		}
		if got, want := errcode.Of(err), errcode.PrereqMissing; got != want {
			t.Errorf("errcode.Of(MountTest()) = %q, want %q (err: %v)", got, want, err)
		}
		for _, text := range placeholderTexts {
			if strings.Contains(err.Error(), text) {
				t.Errorf("MountTest() error = %q, want no %q placeholder text", err, text)
			}
		}
	})
}

// TestNoPlaceholderText guards that no production file in the package still
// carries the placeholder wording.
func TestNoPlaceholderText(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("os.ReadDir(.) = %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("os.ReadFile(%q) = %v", name, err)
		}
		for _, text := range placeholderTexts {
			if strings.Contains(string(b), text) {
				t.Errorf("%s contains %q, want the probe wired to the real system", name, text)
			}
		}
	}
}
