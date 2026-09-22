package latency

import (
	"strings"
	"testing"
)

const wantRemote = "gdrive:snapback-stage1"

func TestCheckRemoteAllowsOnlyExactRemote(t *testing.T) {
	if err := CheckRemote(wantRemote); err != nil {
		t.Errorf("CheckRemote(%q) = %v, want nil", wantRemote, err)
	}

	refused := []string{
		"",
		"gdrive:",
		"gdrive:snapback-stage1/x",
		"gdrive:snapback-stage1 ",
		"GDRIVE:snapback-stage1",
		"gdrive:other",
		"s3:x",
		"rclone:gdrive:snapback-stage1",
		"/tmp/repo",
	}
	for _, remote := range refused {
		t.Run(remote, func(t *testing.T) {
			err := CheckRemote(remote)
			if err == nil {
				t.Fatalf("CheckRemote(%q) = nil, want refusal", remote)
			}
			if !strings.Contains(err.Error(), wantRemote) {
				t.Errorf("CheckRemote(%q) error %q does not name %q", remote, err, wantRemote)
			}
		})
	}
}

func TestRemoteFromEnv(t *testing.T) {
	const envVar = "SNAPBACK_RCLONE_REMOTE"
	getenvWith := func(value string) func(string) string {
		return func(key string) string {
			if key == envVar {
				return value
			}
			return ""
		}
	}

	t.Run("unset", func(t *testing.T) {
		got, err := RemoteFromEnv(getenvWith(""))
		if err == nil {
			t.Fatalf("RemoteFromEnv(unset) = %q, nil; want error", got)
		}
		for _, want := range []string{envVar, wantRemote} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("RemoteFromEnv(unset) error %q does not name %q", err, want)
			}
		}
	})

	t.Run("other remote", func(t *testing.T) {
		got, err := RemoteFromEnv(getenvWith("gdrive:other"))
		if err == nil {
			t.Fatalf("RemoteFromEnv(gdrive:other) = %q, nil; want refusal", got)
		}
		if !strings.Contains(err.Error(), wantRemote) {
			t.Errorf("RemoteFromEnv(gdrive:other) error %q does not name %q", err, wantRemote)
		}
	})

	t.Run("allowed", func(t *testing.T) {
		got, err := RemoteFromEnv(getenvWith(wantRemote))
		if err != nil {
			t.Fatalf("RemoteFromEnv(allowed) error = %v, want nil", err)
		}
		if got != wantRemote {
			t.Errorf("RemoteFromEnv(allowed) = %q, want %q", got, wantRemote)
		}
	})
}

func TestRepoSpecIsConstant(t *testing.T) {
	if got, want := RepoSpec(), "rclone:gdrive:snapback-stage1"; got != want {
		t.Errorf("RepoSpec() = %q, want %q", got, want)
	}
	if got := AllowedRemote; got != wantRemote {
		t.Errorf("AllowedRemote = %q, want %q", got, wantRemote)
	}
}
