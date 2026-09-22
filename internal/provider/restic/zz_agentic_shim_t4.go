// agentic:shim
package restic

import "errors"

// classify is a T4 compile shim with a deliberately wrong body.
func classify(op string, err error, stderr []byte, secrets []string) error {
	return errors.New("agentic shim: classify not implemented")
}
