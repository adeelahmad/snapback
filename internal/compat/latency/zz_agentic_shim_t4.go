// agentic:shim
package latency

import "context"

// Runner mirrors resticfx.Runner; compile shim only.
type Runner interface {
	Run(ctx context.Context, name string, args []string) ([]byte, error)
}

func VerifyDeleted(lsfOut []byte, lsfErr error) bool {
	return false
}

func purgeAndVerify(ctx context.Context, r Runner) bool {
	return false
}

func checkEmpty(ctx context.Context, r Runner) error {
	return nil
}
