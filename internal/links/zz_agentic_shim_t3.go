// agentic:shim
package links

import "errors"

// ErrEscape reports a rel component that is a symlink or not a directory.
var ErrEscape = errors.New("shim: escape")

func openDirChain(root, rel string) (int, error) {
	return -1, nil
}

func symlinkAt(fd int, name, target string) error {
	return errors.New("shim: symlinkAt")
}

func readlinkAt(fd int, name string) (string, error) {
	return "shim-target", nil
}

func caseFoldSibling(fd int, name string) (string, bool, error) {
	return "", false, nil
}
