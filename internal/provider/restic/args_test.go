package restic

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

func TestGlobalArgsGolden(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(o *Options)
		want   []string
	}{
		{name: "default", mutate: func(*Options) {}, want: []string{"--password-file", pw}},
		{name: "cache dir", mutate: func(o *Options) { o.CacheDir = "/c" }, want: []string{"--password-file", pw, "--cache-dir", "/c"}},
		{name: "no cache", mutate: func(o *Options) { o.NoCache = true }, want: []string{"--password-file", pw, "--no-cache"}},
		{name: "no lock with cache dir", mutate: func(o *Options) { o.NoLock = true; o.CacheDir = "/c" }, want: []string{"--password-file", pw, "--cache-dir", "/c", "--no-lock"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := validOptions()
			tt.mutate(&opts)
			p := mustNew(t, opts)
			if got := p.globalArgs(); !slices.Equal(got, tt.want) {
				t.Errorf("globalArgs() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCommandArgsGolden(t *testing.T) {
	p := mustNew(t, validOptions())
	tests := []struct {
		name string
		got  []string
		want []string
	}{
		{name: "validateArgs()", got: p.validateArgs(), want: []string{"--password-file", pw, "cat", "config", "--json"}},
		{name: "listArgs()", got: p.listArgs(), want: []string{"--password-file", pw, "snapshots", "--json"}},
		{name: `mountArgs("/m")`, got: p.mountArgs("/m"), want: []string{"--password-file", pw, "mount", "--path-template", "ids/%I", "/m"}},
		{name: "lsArgs(id1)", got: p.lsArgs(id1), want: []string{"--password-file", pw, "ls", "--json", id1}},
	}
	for _, tt := range tests {
		if !slices.Equal(tt.got, tt.want) {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestSnapArgsGolden(t *testing.T) {
	opts := validOptions()
	opts.NoLock = true
	opts.CacheDir = "/c"
	p := mustNew(t, opts)
	req := provider.SnapRequest{
		Path:     "-weird dir;$(x)",
		Host:     "h1",
		Tags:     []string{"snapback:adhoc", "pre-migrate"},
		Excludes: []string{".snapshot", "/state"},
	}

	got := p.snapArgs(req)
	want := []string{
		"--password-file", pw, "--cache-dir", "/c",
		"backup", "--json", "--host", "h1",
		"--tag", "snapback:adhoc", "--tag", "pre-migrate",
		"--exclude", ".snapshot", "--exclude", "/state",
		"--", "-weird dir;$(x)",
	}
	if !slices.Equal(got, want) {
		t.Errorf("snapArgs(%+v) = %q, want %q", req, got, want)
	}
	if slices.Contains(got, "--no-lock") {
		t.Errorf("snapArgs(%+v) = %q, want no --no-lock", req, got)
	}
}

func TestNoSecretInAnyArgv(t *testing.T) {
	opts := validOptions()
	opts.Env = map[string]string{"AWS_SECRET_ACCESS_KEY": "topsecret"}
	p := mustNew(t, opts)
	req := provider.SnapRequest{Path: "/home/alex/project", Host: "h1", Tags: []string{"pre"}, Excludes: []string{".snapshot"}}

	argvs := map[string][]string{
		"validateArgs": p.validateArgs(),
		"listArgs":     p.listArgs(),
		"mountArgs":    p.mountArgs("/m"),
		"lsArgs":       p.lsArgs(id1),
		"snapArgs":     p.snapArgs(req),
	}
	for name, argv := range argvs {
		// M-002: every command must carry its password-file flag, so an
		// empty or shim argv cannot pass the negative checks vacuously.
		if !slices.Contains(argv, "--password-file") {
			t.Errorf("%s() = %q, want a non-empty argv with --password-file", name, argv)
		}
		for _, a := range argv {
			if strings.Contains(a, repo) || strings.Contains(a, "secret") || strings.Contains(a, "topsecret") ||
				strings.Contains(a, "--repo") || a == "-r" {
				t.Errorf("%s() element %q, want no repository, secret or --repo/-r flag", name, a)
			}
		}
	}
}

func TestFakeRunnerRecords(t *testing.T) {
	fr := &fakeRunner{results: map[string]fakeResult{"snapshots": {stdout: []byte("[]")}}}
	opts := validOptions()
	opts.Runner = fr
	p := mustNew(t, opts)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	stdout, _, err := fr.Run(ctx, bin, p.listArgs(), p.childEnv())
	if err != nil || string(stdout) != "[]" {
		t.Errorf("fakeRunner.Run(snapshots) = %q, %v, want %q, nil", stdout, err, "[]")
	}

	calls := fr.runCalls()
	if len(calls) != 1 {
		t.Fatalf("fakeRunner recorded %d Run calls, want 1", len(calls))
	}
	c := calls[0]
	if c.name != bin {
		t.Errorf("recorded name = %q, want %q", c.name, bin)
	}
	if want := []string{"--password-file", pw, "snapshots", "--json"}; !slices.Equal(c.args, want) {
		t.Errorf("recorded argv = %q, want %q", c.args, want)
	}
	if !slices.Contains(c.env, "RESTIC_REPOSITORY="+repo) {
		t.Errorf("recorded env = %q, want it to contain RESTIC_REPOSITORY=<repo>", c.env)
	}
	if !c.hasDeadline {
		t.Errorf("recorded hasDeadline = false, want true")
	}
	if got := fr.peakConcurrent(); got != 1 {
		t.Errorf("fakeRunner.peakConcurrent() = %d, want 1", got)
	}

	proc, err := fr.Start(bin, p.mountArgs("/m"), p.childEnv())
	if err != nil {
		t.Fatalf("fakeRunner.Start(mount) error = %v, want nil", err)
	}
	all := fr.recorded()
	if len(all) != 2 || !all[1].start || all[1].hasDeadline {
		t.Fatalf("fakeRunner recorded %+v, want a second call marked start without deadline", all)
	}
	fp := fr.processes()[0]
	fp.exit(nil)
	if err := proc.Wait(); err != nil {
		t.Errorf("fakeProcess.Wait() after exit(nil) = %v, want nil", err)
	}
	if len(fp.signalled()) != 0 || fp.killCount() != 0 {
		t.Errorf("fakeProcess signals = %v, kills = %d, want none", fp.signalled(), fp.killCount())
	}
}
