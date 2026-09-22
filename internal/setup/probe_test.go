package setup

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
)

const (
	probeRestic   = "/opt/snapback/bin/restic"
	probeRepo     = "sftp:backup@example.com:/srv/restic"
	probePassword = "/etc/snapback/restic-password"
)

// probeCall records one invocation the probe asked its runner to make.
type probeCall struct {
	name string
	args []string
}

// fakeProbeRunner records every call and replays a canned answer.
type fakeProbeRunner struct {
	stdout []byte
	err    error
	calls  []probeCall
}

func (f *fakeProbeRunner) run(_ context.Context, name string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, probeCall{name: name, args: slices.Clone(args)})
	return f.stdout, f.err
}

const twoSnapshots = `[
  {"id":"1111111111111111111111111111111111111111111111111111111111111111",
   "time":"2026-01-02T03:04:05Z","hostname":"media-01","paths":["/srv/media"]},
  {"id":"2222222222222222222222222222222222222222222222222222222222222222",
   "time":"2026-03-04T05:06:07Z","hostname":"media-02","paths":["/srv/photos","/srv/music"]}
]`

func TestProbeRunsOneReadOnlyCommand(t *testing.T) {
	fr := &fakeProbeRunner{stdout: []byte("[]")}

	if _, err := Probe(context.Background(), fr.run, probeRestic, probeRepo, probePassword); err != nil {
		t.Fatalf("Probe() error = %v, want nil", err)
	}

	if len(fr.calls) != 1 {
		t.Fatalf("Probe() made %d calls, want 1: %+v", len(fr.calls), fr.calls)
	}
	call := fr.calls[0]
	if call.name != probeRestic {
		t.Errorf("Probe() ran %q, want %q", call.name, probeRestic)
	}
	want := []string{"-r", probeRepo, "--password-file", probePassword, "--no-lock", "snapshots", "--json"}
	if !slices.Equal(call.args, want) {
		t.Errorf("Probe() args = %q, want %q", call.args, want)
	}
}

func TestProbeNeverRunsAMutatingSubcommand(t *testing.T) {
	fr := &fakeProbeRunner{stdout: []byte("[]")}

	if _, err := Probe(context.Background(), fr.run, probeRestic, probeRepo, probePassword); err != nil {
		t.Fatalf("Probe() error = %v, want nil", err)
	}

	for _, call := range fr.calls {
		for _, arg := range call.args {
			for _, banned := range []string{"init", "backup", "prune", "unlock", "forget"} {
				if arg == banned {
					t.Errorf("Probe() ran %q with %q, want a read-only command only", call.name, banned)
				}
			}
		}
	}
}

func TestProbeReturnsSnapshotsNewestFirst(t *testing.T) {
	fr := &fakeProbeRunner{stdout: []byte(twoSnapshots)}

	got, err := Probe(context.Background(), fr.run, probeRestic, probeRepo, probePassword)
	if err != nil {
		t.Fatalf("Probe() error = %v, want nil", err)
	}

	want := []Snapshot{
		{
			ID:       "2222222222222222222222222222222222222222222222222222222222222222",
			Hostname: "media-02",
			Paths:    []string{"/srv/photos", "/srv/music"},
			Time:     time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC),
		},
		{
			ID:       "1111111111111111111111111111111111111111111111111111111111111111",
			Hostname: "media-01",
			Paths:    []string{"/srv/media"},
			Time:     time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		},
	}
	if len(got) != len(want) {
		t.Fatalf("Probe() returned %d snapshots, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].ID != want[i].ID || got[i].Hostname != want[i].Hostname ||
			!got[i].Time.Equal(want[i].Time) || !slices.Equal(got[i].Paths, want[i].Paths) {
			t.Errorf("Probe()[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestProbeEmptyRepositoryIsNotAnError(t *testing.T) {
	fr := &fakeProbeRunner{stdout: []byte("[]")}

	got, err := Probe(context.Background(), fr.run, probeRestic, probeRepo, probePassword)
	if err != nil {
		t.Fatalf("Probe(empty) error = %v, want nil", err)
	}
	if got != nil {
		t.Errorf("Probe(empty) = %+v, want nil", got)
	}
}

func TestProbeWrapsRunnerError(t *testing.T) {
	sentinel := errors.New("exit status 1")
	fr := &fakeProbeRunner{err: sentinel}

	_, err := Probe(context.Background(), fr.run, probeRestic, probeRepo, probePassword)
	if err == nil {
		t.Fatalf("Probe(failing runner) error = nil, want an error")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("Probe(failing runner) error = %v, want it to wrap %v", err, sentinel)
	}
	if !strings.HasPrefix(err.Error(), "setup: probe: ") {
		t.Errorf("Probe(failing runner) error = %q, want the prefix %q", err.Error(), "setup: probe: ")
	}
}
