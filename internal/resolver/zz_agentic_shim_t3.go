// agentic:shim
package resolver

// HistoryTarget is a RED compile shim: it concatenates without validation or
// separator so every join and rejection row fails by assertion.
func HistoryTarget(snapshotRoot, treePath string) (string, error) {
	return snapshotRoot + treePath, nil
}
