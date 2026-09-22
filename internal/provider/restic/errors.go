package restic

// classify maps a restic failure to an errcode error with every secret redacted.
func classify(op string, err error, stderr []byte, secrets []string) error {
	panic("SUB-AGENT-TODO: T4: exec.ErrNotFound -> errcode.PrereqMissing; stderr with 'unable to create lock' or 'already locked' -> errcode.RepoUnavailable carrying the restic message; any other failure -> errcode.RepoUnavailable; replace every secret with <redacted>; never retry")
}
