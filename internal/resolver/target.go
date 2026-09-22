package resolver

import (
	"fmt"
	"path/filepath"
	"strings"
)

// HistoryTarget joins a snapshot root and a tree path, refusing any tree
// path that could escape the root.
func HistoryTarget(snapshotRoot, treePath string) (string, error) {
	if snapshotRoot == "" || !filepath.IsAbs(snapshotRoot) {
		return "", fmt.Errorf("resolver: snapshot root %q is not absolute", snapshotRoot)
	}
	root := filepath.Clean(snapshotRoot)
	tree := strings.TrimPrefix(treePath, "/")
	if tree == "" {
		return root, nil
	}
	if strings.ContainsRune(tree, 0) {
		return "", fmt.Errorf("resolver: tree path %q contains NUL", treePath)
	}
	for _, c := range strings.Split(tree, "/") {
		if c == "" || c == "." || c == ".." {
			return "", fmt.Errorf("resolver: tree path %q has invalid component %q", treePath, c)
		}
	}
	return root + "/" + tree, nil
}
