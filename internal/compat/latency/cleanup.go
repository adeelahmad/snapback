package latency

import "context"

// Runner executes an external program with an argument array and returns
// its combined output. It mirrors resticfx.Runner.
type Runner interface {
	Run(ctx context.Context, name string, args []string) ([]byte, error)
}

// VerifyDeleted reports whether an rclone lsf result shows the remote path
// was removed.
func VerifyDeleted(lsfOut []byte, lsfErr error) bool {
	panic("SUB-AGENT-TODO: true only for empty lsfOut with nil lsfErr, or an lsfErr whose text contains `directory not found` with empty lsfOut; any other error or any listed entry is false")
}

func purgeAndVerify(ctx context.Context, r Runner) bool {
	panic("SUB-AGENT-TODO: run `rclone purge gdrive:snapback-stage1` then `rclone lsf gdrive:snapback-stage1` as arg arrays (no shell); a purge error still runs lsf; return VerifyDeleted(lsf output, lsf error)")
}

func checkEmpty(ctx context.Context, r Runner) error {
	panic("SUB-AGENT-TODO: run `rclone lsf gdrive:snapback-stage1` before any write; nil when lsf errors with `directory not found` or lists nothing; error mentioning `non-empty` when entries are listed; error on any other lsf error")
}
