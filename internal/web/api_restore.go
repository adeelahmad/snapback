package web

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
)

// handleAPIRestore copies one file out of a snapshot into the live tree under
// a dated name, never overwriting an existing file.
func (s *Server) handleAPIRestore(w http.ResponseWriter, r *http.Request) {
	root := r.FormValue("root")
	file, err := s.resolve(root, r.FormValue("file"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, err)
		return
	}
	id := provider.SnapshotID(r.FormValue("snapshot"))
	if !id.Valid() {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, fmt.Errorf("snapshot id %q is not 64 hex characters", id))
		return
	}
	snapDir, at, err := s.opts.History.SnapshotDir(root, id)
	if err != nil {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, err)
		return
	}
	src, err := openInSnapshot(snapDir, file)
	if err != nil {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, err)
		return
	}
	defer func() { _ = src.Close() }()

	live, err := os.OpenRoot(s.livePath(root))
	if err != nil {
		writeError(w, http.StatusInternalServerError, errcode.Of(err), err)
		return
	}
	defer func() { _ = live.Close() }()

	ext := path.Ext(file)
	stem := strings.TrimSuffix(file, ext) + " (" + at.Local().Format("2006-01-02") + ")"
	var dst *os.File
	var name string
	for n := 1; ; n++ {
		name = stem + ext
		if n > 1 {
			name = fmt.Sprintf("%s %d%s", stem, n, ext)
		}
		dst, err = live.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if !errors.Is(err, fs.ErrExist) {
			break
		}
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, errcode.Of(err), err)
		return
	}
	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		writeError(w, http.StatusInternalServerError, errcode.Of(err), err)
		return
	}
	if err := dst.Close(); err != nil {
		writeError(w, http.StatusInternalServerError, errcode.Of(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"path": filepath.Join(s.livePath(root), filepath.FromSlash(name))})
}

// handleAPIOpen launches the host file manager on the directory containing
// one live-tree path, via s.opts.Opener.
func (s *Server) handleAPIOpen(w http.ResponseWriter, r *http.Request) {
	root := r.FormValue("root")
	p, err := s.resolve(root, r.FormValue("path"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, err)
		return
	}
	if s.opts.Opener == nil {
		writeError(w, http.StatusServiceUnavailable, errcode.PrereqMissing, errors.New("no file manager opener"))
		return
	}
	dir := filepath.Join(s.livePath(root), filepath.FromSlash(path.Dir(p)))
	if err := s.opts.Opener(r.Context(), dir); err != nil {
		writeError(w, http.StatusInternalServerError, errcode.Of(err), err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// livePath returns the live-tree path of the configured root id.
func (s *Server) livePath(root string) string {
	for _, rt := range s.opts.History.Roots() {
		if rt.ID == root {
			return rt.Path
		}
	}
	return ""
}
