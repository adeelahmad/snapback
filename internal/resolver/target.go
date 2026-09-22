package resolver

// HistoryTarget joins a snapshot root and a tree path, refusing any tree
// path that could escape the root.
func HistoryTarget(snapshotRoot, treePath string) (string, error) {
	panic("SUB-AGENT-TODO: T3 root absolute non-empty, filepath.Clean; strip one leading /; reject leading /, empty/./.. component, NUL -> \"\", error prefixed resolver:; empty tree -> root else root+\"/\"+treePath")
}
