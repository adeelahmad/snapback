package latency

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Runner executes an external program with an argument array and returns
// its combined output. It mirrors resticfx.Runner.
type Runner interface {
	Run(ctx context.Context, name string, args []string) ([]byte, error)
}

// VerifyDeleted reports whether an rclone lsf result shows the remote path
// was removed.
func VerifyDeleted(lsfOut []byte, lsfErr error) bool {
	if len(lsfOut) != 0 {
		return false
	}
	return lsfErr == nil || isDirNotFound(lsfErr)
}

func isDirNotFound(err error) bool {
	return strings.Contains(err.Error(), "directory not found")
}

func purgeAndVerify(ctx context.Context, r Runner) bool {
	// A purge failure is not fatal: the lsf check below is the source of truth.
	_, _ = r.Run(ctx, "rclone", []string{"purge", AllowedRemote})
	out, err := r.Run(ctx, "rclone", []string{"lsf", AllowedRemote})
	return VerifyDeleted(out, err)
}

func checkEmpty(ctx context.Context, r Runner) error {
	out, err := r.Run(ctx, "rclone", []string{"lsf", AllowedRemote})
	if err != nil && !isDirNotFound(err) {
		return fmt.Errorf("list %s: %w", AllowedRemote, err)
	}
	if len(out) != 0 {
		return errors.New("remote " + AllowedRemote + " is non-empty; refusing to write")
	}
	return nil
}
