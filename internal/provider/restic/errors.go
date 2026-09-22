package restic

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// classify maps a restic failure to an errcode error with every secret redacted.
// A lock failure and every other restic failure map to RepoUnavailable carrying
// the restic message; restic is never retried.
func classify(op string, err error, stderr []byte, secrets []string) error {
	if errors.Is(err, exec.ErrNotFound) {
		return errcode.New(errcode.PrereqMissing, op, err)
	}
	msg := strings.TrimSpace(string(stderr))
	if msg == "" {
		msg = err.Error()
	}
	for _, s := range secrets {
		if s != "" {
			msg = strings.ReplaceAll(msg, s, "<redacted>")
		}
	}
	return errcode.New(errcode.RepoUnavailable, op, errors.New(msg))
}
