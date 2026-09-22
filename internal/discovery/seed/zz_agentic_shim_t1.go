// agentic:shim
package seed

// Plan is a compile shim for S3-08 T1.
type Plan struct {
	Dirs  []string
	Count int
}

// PlanPath is a compile shim with a deliberately wrong body.
func PlanPath(root, seedPath string, maxDepth int, excludes []string) (Plan, error) {
	return Plan{Dirs: []string{"agentic-shim"}, Count: 99}, nil
}
