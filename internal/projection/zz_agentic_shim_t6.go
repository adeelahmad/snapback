// agentic:shim

package projection

// Readlink is a RED compile shim that reports a wrong target for every inode.
func (*Generation) Readlink(_ uint64) (target string, found bool) {
	return "shim: not the stored target", true
}
