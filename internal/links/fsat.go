package links

import "errors"

// ErrEscape reports a rel component that is a symlink or not a directory.
var ErrEscape = errors.New("links: path escapes through a symlink or non-directory")

func openDirChain(root, rel string) (int, error) {
	panic("SUB-AGENT-TODO: T3 open root then each rel component with unix.Openat(O_RDONLY|O_DIRECTORY|O_NOFOLLOW|O_CLOEXEC), closing intermediates; a symlinked or non-directory component -> ErrEscape")
}

func symlinkAt(fd int, name, target string) error {
	panic("SUB-AGENT-TODO: T3 unix.Symlinkat(target, fd, name)")
}

func readlinkAt(fd int, name string) (string, error) {
	panic("SUB-AGENT-TODO: T3 unix.Readlinkat(fd, name, buf) returning the raw target bytes")
}

func caseFoldSibling(fd int, name string) (string, bool, error) {
	panic("SUB-AGENT-TODO: T3 list fd once; return an entry that is strings.EqualFold to name but not equal to it")
}
