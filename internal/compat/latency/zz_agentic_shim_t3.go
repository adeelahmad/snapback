// agentic:shim
package latency

import "errors"

type scratch struct {
	root         string
	passwordFile string
	cacheDir     string
	dataDir      string
	mountDir     string
}

func newScratch() (scratch, error) {
	return scratch{}, errors.New("agentic shim: newScratch not implemented")
}

func (s scratch) Close() error {
	return errors.New("agentic shim: scratch.Close not implemented")
}
