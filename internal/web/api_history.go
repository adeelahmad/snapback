package web

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
)

const likelyIdentical = "likely identical"

type rootJSON struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	State string `json:"state"`
}

type entryJSON struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mtime"`
	Dir     bool      `json:"dir"`
	Absent  bool      `json:"absent"`
}

type versionJSON struct {
	Snapshot provider.SnapshotID `json:"snapshot"`
	Time     time.Time           `json:"time"`
	Size     int64               `json:"size"`
	ModTime  time.Time           `json:"mtime"`
	Label    string              `json:"label,omitempty"`
}

// resolve checks that root is a configured root and that p is a local,
// slash-separated path inside it. It returns the cleaned path, "." for the
// root itself.
func (s *Server) resolve(root, p string) (string, error) {
	known := false
	for _, r := range s.opts.History.Roots() {
		if r.ID == root {
			known = true
			break
		}
	}
	if !known {
		return "", fmt.Errorf("unknown root %q", root)
	}
	if p == "" {
		return ".", nil
	}
	if strings.ContainsRune(p, 0) || !filepath.IsLocal(p) {
		return "", fmt.Errorf("path %q is not inside the root", p)
	}
	return path.Clean(p), nil
}

func (s *Server) handleAPIRoots(w http.ResponseWriter, r *http.Request) {
	roots := s.opts.History.Roots()
	out := make([]rootJSON, 0, len(roots))
	for _, rt := range roots {
		out = append(out, rootJSON(rt))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleAPIHistory(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	root := q.Get("root")
	dir, err := s.resolve(root, q.Get("path"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, err)
		return
	}
	entries, err := s.opts.History.List(r.Context(), root, dir, provider.SnapshotID(q.Get("snapshot")))
	if errcode.Of(err) == errcode.MappingAbsent {
		writeError(w, http.StatusNotFound, errcode.MappingAbsent, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, errcode.Of(err), err)
		return
	}
	out := make([]entryJSON, 0, len(entries))
	for _, e := range entries {
		out = append(out, entryJSON(e))
	}
	writeJSON(w, http.StatusOK, out)
}

// handleAPIVersions keeps the order History returns and marks versions that
// share a group with another as likely identical.
func (s *Server) handleAPIVersions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	root := q.Get("root")
	file, err := s.resolve(root, q.Get("path"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, err)
		return
	}
	versions, err := s.opts.History.Versions(r.Context(), root, file)
	if errcode.Of(err) == errcode.MappingAbsent {
		writeError(w, http.StatusNotFound, errcode.MappingAbsent, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, errcode.Of(err), err)
		return
	}
	groups := make(map[int]int)
	for _, v := range versions {
		groups[v.Group]++
	}
	out := make([]versionJSON, 0, len(versions))
	for _, v := range versions {
		vj := versionJSON{Snapshot: v.Snapshot, Time: v.Time, Size: v.Size, ModTime: v.ModTime}
		if groups[v.Group] > 1 {
			vj.Label = likelyIdentical
		}
		out = append(out, vj)
	}
	writeJSON(w, http.StatusOK, out)
}

// handleAPIDownload streams one regular file from a snapshot. The file is
// opened through os.OpenRoot so ".." and symlinks cannot leave the snapshot.
func (s *Server) handleAPIDownload(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	root := q.Get("root")
	file, err := s.resolve(root, q.Get("path"))
	if err != nil {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, err)
		return
	}
	snapDir, _, err := s.opts.History.SnapshotDir(root, provider.SnapshotID(q.Get("snapshot")))
	if err != nil {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, err)
		return
	}
	f, err := openInSnapshot(snapDir, file)
	if err != nil {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, err)
		return
	}
	defer func() { _ = f.Close() }()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", path.Base(file)))
	w.Header().Set("Cache-Control", "no-store")
	_, _ = io.Copy(w, f)
}

// openInSnapshot opens the regular file name under dir without following
// anything that leaves dir.
func openInSnapshot(dir, name string) (*os.File, error) {
	rt, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rt.Close() }()
	f, err := rt.Open(name)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if !fi.Mode().IsRegular() {
		_ = f.Close()
		return nil, errors.New("not a regular file")
	}
	return f, nil
}
