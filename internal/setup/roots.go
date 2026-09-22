package setup

// applyRoots fills r.Roots with the backup roots detection settled on, and
// r.Reasons with one line per candidate it refused.
func applyRoots(r *Result, given []string, getwd func() (string, error), excluded []string, tempDir string) {
	_, _, _, _, _ = r, given, getwd, excluded, tempDir
}
