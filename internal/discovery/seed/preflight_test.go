package seed

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

type fakeStat struct {
	st    Statfs
	err   error
	calls []string
}

func (f *fakeStat) stat(path string) (Statfs, error) {
	f.calls = append(f.calls, path)
	return f.st, f.err
}

func TestPreflightAllowsWithinBudget(t *testing.T) {
	r := t.TempDir()
	f := &fakeStat{st: Statfs{Files: 1000, FreeFiles: 500}}
	p := Plan{Dirs: []string{r}, Count: 100}

	if err := Preflight(p, f.stat, 0.90, 500000, false); err != nil {
		t.Errorf("Preflight(Count=100) = %v, want nil", err)
	}
	if len(f.calls) != 1 || f.calls[0] != r {
		t.Errorf("fsStat calls = %q, want [%q]", f.calls, r)
	}
}

func TestPreflightRefusesInodeThreshold(t *testing.T) {
	r := t.TempDir()
	f := &fakeStat{st: Statfs{Files: 1000, FreeFiles: 500}}

	err := Preflight(Plan{Dirs: []string{r}, Count: 450}, f.stat, 0.90, 500000, false)
	if got := errcode.Of(err); got != errcode.InodeBudgetExceeded {
		t.Fatalf("errcode.Of(Preflight(Count=450)) = %q, want %q (err %v)", got, errcode.InodeBudgetExceeded, err)
	}
	for _, want := range []string{"95", "90"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Preflight(Count=450) error = %q, want it to contain %q", err.Error(), want)
		}
	}
	if err := Preflight(Plan{Dirs: []string{r}, Count: 350}, f.stat, 0.90, 500000, false); err != nil {
		t.Errorf("Preflight(Count=350) = %v, want nil", err)
	}
}

func TestPreflightRefusesMaxLinks(t *testing.T) {
	r := t.TempDir()
	f := &fakeStat{st: Statfs{Files: 1e9, FreeFiles: 1e9}}

	err := Preflight(Plan{Dirs: []string{r}, Count: 11}, f.stat, 0.90, 10, false)
	if got := errcode.Of(err); got != errcode.InodeBudgetExceeded {
		t.Fatalf("errcode.Of(Preflight(Count=11, maxLinks=10)) = %q, want %q (err %v)", got, errcode.InodeBudgetExceeded, err)
	}
	for _, want := range []string{"11", "10"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Preflight(Count=11) error = %q, want it to contain %q", err.Error(), want)
		}
	}
	if err := Preflight(Plan{Dirs: []string{r}, Count: 10}, f.stat, 0.90, 10, false); err != nil {
		t.Errorf("Preflight(Count=10, maxLinks=10) = %v, want nil", err)
	}
}

func TestPreflightForceOverrides(t *testing.T) {
	r := t.TempDir()
	tests := []struct {
		name     string
		st       Statfs
		count    int
		maxLinks int
	}{
		{name: "inode threshold", st: Statfs{Files: 1000, FreeFiles: 500}, count: 450, maxLinks: 500000},
		{name: "max links", st: Statfs{Files: 1e9, FreeFiles: 1e9}, count: 11, maxLinks: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeStat{st: tt.st}
			p := Plan{Dirs: []string{r}, Count: tt.count}
			if err := Preflight(p, f.stat, 0.90, tt.maxLinks, false); err == nil {
				t.Fatalf("Preflight(force=false) = nil, want a refusal (precondition)")
			}
			if err := Preflight(p, f.stat, 0.90, tt.maxLinks, true); err != nil {
				t.Errorf("Preflight(force=true) = %v, want nil", err)
			}
		})
	}
}

func TestPreflightStatfsError(t *testing.T) {
	r := t.TempDir()
	boomErr := errors.New("boom")
	boom := &fakeStat{err: boomErr}
	tests := []struct {
		name    string
		dir     string
		fsStat  func(string) (Statfs, error)
		wantErr error
	}{
		{name: "fake error", dir: r, fsStat: boom.stat, wantErr: boomErr},
		{name: "real missing", dir: filepath.Join(r, "missing"), fsStat: StatfsOf},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Preflight(Plan{Dirs: []string{tt.dir}, Count: 1}, tt.fsStat, 0.9, 10, false)
			if err == nil {
				t.Fatalf("Preflight(%q) = nil, want error", tt.dir)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("Preflight(%q) = %v, want errors.Is %v", tt.dir, err, tt.wantErr)
			}
		})
	}

	st, err := StatfsOf(r)
	if err != nil {
		t.Fatalf("StatfsOf(%q) error = %v, want nil", r, err)
	}
	if st.Files == 0 {
		t.Errorf("StatfsOf(%q).Files = 0, want > 0", r)
	}
}
