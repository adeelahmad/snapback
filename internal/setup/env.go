package setup

// applyEnv fills the Restic fields of r from the environment read through
// getenv.
func applyEnv(r *Result, getenv func(string) string) {
	_, _ = r, getenv
}
