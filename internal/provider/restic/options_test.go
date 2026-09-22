package restic

import (
	"slices"
	"strings"
	"testing"
)

const (
	id1  = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	id2  = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	repo = "s3:https://user:secret@host/bucket"
	pw   = "/etc/snapback/pw"
	bin  = "/usr/bin/restic"
)

func validOptions() Options {
	return Options{Binary: bin, Repository: repo, PasswordFile: pw}
}

func mustNew(t *testing.T, opts Options) *Provider {
	t.Helper()
	p, err := New(opts)
	if err != nil {
		t.Fatalf("New(%+v) error = %v, want nil", opts, err)
	}
	if p == nil {
		t.Fatalf("New(%+v) = nil provider, want non-nil", opts)
	}
	return p
}

func TestNewValidatesOptions(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(o *Options)
		wantErr bool
	}{
		{name: "valid", mutate: func(*Options) {}},
		{name: "relative binary", mutate: func(o *Options) { o.Binary = "restic" }, wantErr: true},
		{name: "empty repository", mutate: func(o *Options) { o.Repository = "" }, wantErr: true},
		{name: "relative password file", mutate: func(o *Options) { o.PasswordFile = "pw" }, wantErr: true},
		{name: "relative rclone binary", mutate: func(o *Options) { o.RcloneBinary = "bin/rclone" }, wantErr: true},
		{name: "relative cache dir", mutate: func(o *Options) { o.CacheDir = "cache" }, wantErr: true},
		{name: "cache dir with no cache", mutate: func(o *Options) { o.CacheDir = "/c"; o.NoCache = true }, wantErr: true},
		{name: "env sets PATH", mutate: func(o *Options) { o.Env = map[string]string{"PATH": "x"} }, wantErr: true},
		{name: "env sets RESTIC_PASSWORD", mutate: func(o *Options) { o.Env = map[string]string{"RESTIC_PASSWORD": "x"} }, wantErr: true},
		{name: "env sets RESTIC_REPOSITORY", mutate: func(o *Options) { o.Env = map[string]string{"RESTIC_REPOSITORY": "x"} }, wantErr: true},
		{name: "env sets RESTIC_PASSWORD_COMMAND", mutate: func(o *Options) { o.Env = map[string]string{"RESTIC_PASSWORD_COMMAND": "x"} }, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := validOptions()
			tt.mutate(&opts)
			p, err := New(opts)
			if !tt.wantErr {
				if err != nil || p == nil {
					t.Fatalf("New(%+v) = %v, %v, want non-nil provider, nil error", opts, p, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("New(%+v) error = nil, want error", opts)
			}
			if msg := err.Error(); strings.Contains(msg, "secret") || strings.Contains(msg, repo) {
				t.Errorf("New(%+v) error = %q, want no repository or secret in message", opts, msg)
			}
		})
	}
}

func TestNewDefaultsRunner(t *testing.T) {
	p := mustNew(t, validOptions())
	if _, ok := p.runner.(ExecRunner); !ok {
		t.Errorf("New(nil Runner).runner = %T, want ExecRunner", p.runner)
	}
}

func TestChildEnvIsControlled(t *testing.T) {
	t.Run("rclone and home", func(t *testing.T) {
		t.Setenv("HOME", "/home/u")
		t.Setenv("SECRET_FROM_PARENT", "leak")
		opts := validOptions()
		opts.RcloneBinary = "/opt/rclone/bin/rclone"
		opts.Env = map[string]string{"B": "2", "A": "1"}
		p := mustNew(t, opts)

		got := p.childEnv()
		want := []string{"A=1", "B=2", "HOME=/home/u", "PATH=/opt/rclone/bin:/usr/bin:/bin", "RESTIC_REPOSITORY=" + repo}
		if !slices.Equal(got, want) {
			t.Errorf("childEnv() = %q, want %q", got, want)
		}
		for _, kv := range got {
			if strings.Contains(kv, "SECRET_FROM_PARENT") {
				t.Errorf("childEnv() contains %q, want no inherited parent variables", kv)
			}
		}
	})
	t.Run("no rclone, empty home", func(t *testing.T) {
		t.Setenv("HOME", "")
		p := mustNew(t, validOptions())

		got := p.childEnv()
		want := []string{"PATH=/usr/bin:/bin", "RESTIC_REPOSITORY=" + repo}
		if !slices.Equal(got, want) {
			t.Errorf("childEnv() = %q, want %q", got, want)
		}
		for _, kv := range got {
			if strings.HasPrefix(kv, "HOME=") {
				t.Errorf("childEnv() contains %q, want no HOME entry", kv)
			}
		}
	})
}
