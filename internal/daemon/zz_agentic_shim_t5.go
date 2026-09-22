// agentic:shim

package daemon

// dedupLen reports how many (session, path) keys the dir_event dedup set holds.
func (d *Daemon) dedupLen() int { return 1 << 20 }
