package setup

import "strings"

// applyEnv fills the Restic fields of r from the environment read through
// getenv. A variable that is unset, empty or whitespace-only leaves its field
// untouched, and a nil getenv leaves r unchanged.
func applyEnv(r *Result, getenv func(string) string) {
	if getenv == nil {
		return
	}
	if repo := strings.TrimSpace(getenv("RESTIC_REPOSITORY")); repo != "" {
		r.RepoURI = repo
	}
	if file := strings.TrimSpace(getenv("RESTIC_PASSWORD_FILE")); file != "" {
		r.CredentialFile = file
	}
}
