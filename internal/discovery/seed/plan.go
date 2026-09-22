package seed

// Plan is the ordered set of directories a seed run would create links in.
type Plan struct {
	Dirs  []string
	Count int
}

// PlanPath walks root (or the seedPath subtree of it) down to maxDepth and
// returns the directories that are not excluded.
func PlanPath(root, seedPath string, maxDepth int, excludes []string) (Plan, error) {
	panic("SUB-AGENT-TODO: T1 walk root (or root/seedPath, rejecting seed paths outside root with errcode.InvalidConfig) to maxDepth without following symlinks, skip DefaultExcludes plus excludes, return Dirs and Count")
}
