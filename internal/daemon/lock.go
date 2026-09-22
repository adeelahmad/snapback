package daemon

// Lock takes the single-instance lock in stateDir.
func Lock(stateDir string) (unlock func(), err error) {
	panic("SUB-AGENT-TODO: create stateDir/daemon.lock and take a non-blocking exclusive flock; on EWOULDBLOCK return errcode stale_state naming another running daemon; write os.Getpid() to stateDir/daemon.pid; unlock releases the flock and closes the file")
}
