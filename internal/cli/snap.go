package cli

import "time"

// snapPollInterval is how often snap --wait polls the daemon.
const snapPollInterval = 2 * time.Second

// SnapCommand returns the snap command.
func SnapCommand(d Deps) Command {
	panic("SUB-AGENT-TODO: resolve PATH (default cwd) with resolver.SelectRoot; build SnapRequest{Path, Host, Tags: [\"snapback:adhoc\", user tags...], Excludes: [LinkName, StateDir, HistoryMount, BackendMountDir]}; call NewSnapper(cfg, repoID).Snap once (no lock retry); Daemon.SnapSubmitted; report pending/ready; --wait polls Visible with Sleep/Now")
}
