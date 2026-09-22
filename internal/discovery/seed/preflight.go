package seed

// Statfs is the inode view of a filesystem that Preflight budgets against.
type Statfs struct {
	Files, FreeFiles uint64
}

// StatfsOf reports the inode totals of the filesystem holding path.
func StatfsOf(path string) (Statfs, error) {
	panic("SUB-AGENT-TODO: T2 call unix.Statfs(path) and return Files and FreeFiles; wrap the error with path")
}

// Preflight refuses a plan whose link count would exceed the inode budget or
// maxLinks, unless force is set.
func Preflight(p Plan, fsStat func(path string) (Statfs, error), threshold float64, maxLinks int, force bool) error {
	panic("SUB-AGENT-TODO: T2 return nil when force; reject Count > maxLinks; call fsStat once on the plan root and reject when Count exceeds threshold of FreeFiles; wrap fsStat errors so errors.Is holds")
}
