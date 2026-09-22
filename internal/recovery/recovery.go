package recovery

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/adeelahmad/snapback/internal/links"
)

// Mount is one parsed mountinfo line.
type Mount struct {
	Point, FSType, Source string
}

// Input is what Scan needs.
type Input struct {
	Mountinfo io.Reader
	Owned     []string
	PIDFile   string
	Alive     func(pid int) bool
	Unmount   func(ctx context.Context, point string) error
	Repair    func(ctx context.Context) (links.RepairReport, error)
}

// Report is what Scan did.
type Report struct {
	Unmounted, Foreign, SkippedLive []string
	Repair                          links.RepairReport
}

// ParseMountinfo parses /proc/self/mountinfo.
func ParseMountinfo(r io.Reader) ([]Mount, error) {
	var mounts []Mount
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		pre, post, ok := strings.Cut(line, " - ")
		fields := strings.Fields(pre)
		tail := strings.Fields(post)
		if !ok || len(fields) < 5 || len(tail) < 2 {
			return nil, fmt.Errorf("malformed mountinfo line %q", line)
		}
		point, err := unescape(fields[4])
		if err != nil {
			return nil, err
		}
		source, err := unescape(tail[1])
		if err != nil {
			return nil, err
		}
		mounts = append(mounts, Mount{Point: point, FSType: tail[0], Source: source})
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read mountinfo: %w", err)
	}
	return mounts, nil
}

// unescape decodes the octal escapes (such as \040) the kernel uses in mountinfo.
func unescape(s string) (string, error) {
	if !strings.Contains(s, `\`) {
		return s, nil
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			continue
		}
		if i+4 > len(s) {
			return "", fmt.Errorf("bad escape in mountinfo field %q", s)
		}
		n, err := strconv.ParseUint(s[i+1:i+4], 8, 8)
		if err != nil {
			return "", fmt.Errorf("bad escape in mountinfo field %q: %w", s, err)
		}
		b.WriteByte(byte(n))
		i += 3
	}
	return b.String(), nil
}

// Scan unmounts stale Snapback-owned FUSE mounts.
func Scan(ctx context.Context, in Input) (Report, error) {
	mounts, err := ParseMountinfo(in.Mountinfo)
	if err != nil {
		return Report{}, err
	}
	var rep Report
	var owned []string
	for _, root := range in.Owned {
		for _, m := range mounts {
			if isFuse(m.FSType) && under(m.Point, root) {
				owned = append(owned, m.Point)
			}
		}
	}
	for _, m := range mounts {
		if isFuse(m.FSType) && !underAny(m.Point, in.Owned) {
			rep.Foreign = append(rep.Foreign, m.Point)
		}
	}
	if ownerAlive(in) {
		rep.SkippedLive = owned
		return rep, nil
	}
	unmount := in.Unmount
	if unmount == nil {
		unmount = defaultUnmount
	}
	for _, p := range owned {
		if err := unmount(ctx, p); err != nil {
			return rep, err
		}
		rep.Unmounted = append(rep.Unmounted, p)
	}
	if in.Repair != nil {
		r, err := in.Repair(ctx)
		if err != nil {
			return rep, fmt.Errorf("repair links: %w", err)
		}
		rep.Repair = r
	}
	return rep, nil
}

// ownerAlive reports whether PIDFile names a live process other than this one.
// A missing or unreadable pidfile means no live owner.
func ownerAlive(in Input) bool {
	if in.PIDFile == "" || in.Alive == nil {
		return false
	}
	data, err := os.ReadFile(in.PIDFile)
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 || pid == os.Getpid() {
		return false
	}
	return in.Alive(pid)
}

func isFuse(fstype string) bool {
	return fstype == "fuse" || strings.HasPrefix(fstype, "fuse.")
}

// under reports whether point equals root or lies below it by path component.
func under(point, root string) bool {
	root = strings.TrimSuffix(root, "/")
	return point == root || strings.HasPrefix(point, root+"/")
}

func underAny(point string, roots []string) bool {
	for _, r := range roots {
		if under(point, r) {
			return true
		}
	}
	return false
}
