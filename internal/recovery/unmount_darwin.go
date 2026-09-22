package recovery

import "context"

func defaultUnmount(ctx context.Context, point string) error {
	panic("SUB-AGENT-TODO: T3a: exec.CommandContext(ctx, \"umount\", \"-f\", point) via argv; wrap the error with the point and combined output")
}
