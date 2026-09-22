package latency

// scratch is a disposable temp tree holding the run's password file and working dirs.
type scratch struct {
	root         string
	passwordFile string
	cacheDir     string
	dataDir      string
	mountDir     string
}

func newScratch() (scratch, error) {
	panic("SUB-AGENT-TODO: T3 os.MkdirTemp root under os.TempDir() (refuse anything outside it); write password file mode 0600 with hex of >= 32 crypto/rand bytes; create empty cache, data and mount dirs inside root")
}

// Close removes the scratch root and everything under it.
func (s scratch) Close() error {
	panic("SUB-AGENT-TODO: T3 os.RemoveAll(s.root)")
}
