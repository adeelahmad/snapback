package recovery

import (
	"context"
	"fmt"
	"os/exec"
)

func defaultUnmount(ctx context.Context, point string) error {
	out, err := exec.CommandContext(ctx, "fusermount3", "-u", "-z", point).CombinedOutput()
	if err != nil {
		return fmt.Errorf("unmount %s: %w: %s", point, err, out)
	}
	return nil
}
