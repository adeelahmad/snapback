// agentic:shim

package daemon

// lockFunc is the seam the run command and Daemon.Run take the
// single-instance lock through. Shim: nothing calls it yet.
var lockFunc = Lock
