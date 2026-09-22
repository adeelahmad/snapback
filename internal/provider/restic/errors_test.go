package restic

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

func TestClassifyLockError(t *testing.T) {
	stderr := []byte("unable to create lock in backend: repository is already locked by PID 1234 on host by user\n")

	err := classify("snapshots", errors.New("exit status 1"), stderr, []string{repo})

	if got, want := errcode.Of(err), errcode.RepoUnavailable; got != want {
		t.Errorf("errcode.Of(classify(lock)) = %q, want %q", got, want)
	}
	if err == nil {
		t.Fatal("classify(lock) = nil, want an error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "already locked by PID 1234") {
		t.Errorf("classify(lock).Error() = %q, want it to contain %q", msg, "already locked by PID 1234")
	}
	if !strings.Contains(msg, "snapshots") {
		t.Errorf("classify(lock).Error() = %q, want it to contain the op %q", msg, "snapshots")
	}
}

func TestClassifyNotFound(t *testing.T) {
	cause := fmt.Errorf("exec: %q: %w", bin, exec.ErrNotFound)

	err := classify("cat config", cause, nil, []string{repo})

	if got, want := errcode.Of(err), errcode.PrereqMissing; got != want {
		t.Errorf("errcode.Of(classify(ErrNotFound)) = %q, want %q", got, want)
	}
}

func TestClassifyOtherFailure(t *testing.T) {
	stderr := []byte("Fatal: wrong password or no key found\n")

	err := classify("cat config", errors.New("exit status 1"), stderr, []string{repo})

	if got, want := errcode.Of(err), errcode.RepoUnavailable; got != want {
		t.Errorf("errcode.Of(classify(wrong password)) = %q, want %q", got, want)
	}
	if err == nil {
		t.Fatal("classify(wrong password) = nil, want an error")
	}
	if msg := err.Error(); !strings.Contains(msg, "wrong password") {
		t.Errorf("classify(wrong password).Error() = %q, want it to contain %q", msg, "wrong password")
	}
}

func TestClassifyRedactsSecrets(t *testing.T) {
	stderr := []byte("Fatal: unable to open config file at " + repo + ": key topsecret rejected\n")
	secrets := []string{repo, "topsecret", ""}

	err := classify("cat config", errors.New("exit status 1"), stderr, secrets)

	if err == nil {
		t.Fatal("classify(secret stderr) = nil, want an error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "<redacted>") {
		t.Fatalf("classify(secret stderr).Error() = %q, want it to contain %q", msg, "<redacted>")
	}
	if !strings.Contains(msg, "Fatal: unable to open config file") {
		t.Errorf("classify(secret stderr).Error() = %q, want the restic message kept (an empty secret must not blank it)", msg)
	}
	for _, leak := range []string{repo, "topsecret", "user:secret"} {
		if strings.Contains(msg, leak) {
			t.Errorf("classify(secret stderr).Error() = %q, want no %q", msg, leak)
		}
	}
}
