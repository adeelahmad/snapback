package links

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

// ErrEscape reports a rel component that is a symlink or not a directory.
var ErrEscape = errors.New("links: path escapes through a symlink or non-directory")

const dirFlags = unix.O_RDONLY | unix.O_DIRECTORY | unix.O_NOFOLLOW | unix.O_CLOEXEC

func openDirChain(root, rel string) (int, error) {
	fd, err := unix.Open(root, dirFlags, 0)
	if err != nil {
		return -1, openErr(root, err)
	}
	for _, comp := range strings.Split(rel, "/") {
		if comp == "" || comp == "." {
			continue
		}
		if comp == ".." {
			_ = unix.Close(fd)
			return -1, fmt.Errorf("open %q: %w", comp, ErrEscape)
		}
		next, err := unix.Openat(fd, comp, dirFlags, 0)
		_ = unix.Close(fd)
		if err != nil {
			return -1, openErr(comp, err)
		}
		fd = next
	}
	return fd, nil
}

// openErr maps the errors O_NOFOLLOW|O_DIRECTORY produce for a symlink or
// non-directory component to ErrEscape.
func openErr(name string, err error) error {
	if errors.Is(err, unix.ELOOP) || errors.Is(err, unix.ENOTDIR) {
		return fmt.Errorf("open %q: %w: %w", name, ErrEscape, err)
	}
	return fmt.Errorf("open %q: %w", name, err)
}

func lstatAt(fd int, name string) (unix.Stat_t, error) {
	var st unix.Stat_t
	err := unix.Fstatat(fd, name, &st, unix.AT_SYMLINK_NOFOLLOW)
	return st, err
}

func symlinkAt(fd int, name, target string) error {
	return unix.Symlinkat(target, fd, name)
}

func readlinkAt(fd int, name string) (string, error) {
	for size := 256; ; size *= 2 {
		buf := make([]byte, size)
		n, err := unix.Readlinkat(fd, name, buf)
		if err != nil {
			return "", err
		}
		if n < size {
			return string(buf[:n]), nil
		}
	}
}

// unlinkAt removes a non-directory entry; it never removes directories.
func unlinkAt(fd int, name string) error {
	return unix.Unlinkat(fd, name, 0)
}

func caseFoldSibling(fd int, name string) (string, bool, error) {
	// A fresh open file description keeps fd's directory offset untouched.
	dfd, err := unix.Openat(fd, ".", dirFlags, 0)
	if err != nil {
		return "", false, err
	}
	f := os.NewFile(uintptr(dfd), ".")
	names, err := f.Readdirnames(-1)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return "", false, err
	}
	for _, n := range names {
		if n != name && strings.EqualFold(n, name) {
			return n, true, nil
		}
	}
	return "", false, nil
}
