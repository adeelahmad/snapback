package doctor

// applyPlatform adjusts checks that the running platform cannot pass.
func applyPlatform(results []Check, goos string, strict bool) []Check {
	// SUB-AGENT-TODO: skip inapplicable checks on darwin unless strict.
	return results
}

// exitCode reports the doctor exit status for results.
func exitCode(results []Check) int {
	// SUB-AGENT-TODO: 0 when no check failed, else 1.
	return 0
}
