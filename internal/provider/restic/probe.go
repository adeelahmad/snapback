package restic

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/adeelahmad/snapback/internal/provider"
)

// Probe reports whether treePath exists inside snapshot id under mountDir.
func (p *Provider) Probe(ctx context.Context, mountDir string, id provider.SnapshotID, treePath string) (provider.ProbeResult, error) {
	if !id.Valid() {
		return 0, errors.New("restic: probe: invalid snapshot id")
	}
	rel := strings.TrimPrefix(treePath, "/")
	if rel != "" && !filepath.IsLocal(rel) {
		return 0, errors.New("restic: probe: path escapes the snapshot root")
	}
	name := filepath.Join(p.SnapshotRoot(mountDir, id), rel)

	type result struct {
		fi  os.FileInfo
		err error
	}
	// Buffered so a blocked lstat that finishes after ctx is done does not leak the goroutine.
	resc := make(chan result, 1)
	go func() {
		fi, err := p.lstat(name)
		resc <- result{fi, err}
	}()
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case r := <-resc:
		switch {
		case r.err == nil && r.fi.IsDir():
			return provider.ProbeDir, nil
		case r.err == nil:
			return provider.ProbeNotDir, nil
		case errors.Is(r.err, fs.ErrNotExist), errors.Is(r.err, syscall.ENOTDIR):
			return provider.ProbeAbsent, nil
		default:
			return 0, r.err
		}
	}
}
