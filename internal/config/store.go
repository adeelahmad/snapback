package config

import (
	"errors"
	"os"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// ErrRevisionConflict reports that the stored configuration changed since the
// caller read it, so Save refused to overwrite it.
var ErrRevisionConflict = errcode.New(errcode.StaleState, "config.save", errors.New("config revision conflict"))

// rename is the atomic-replace step of Save; tests swap it to inject failures.
var rename = os.Rename

// Load reads and validates the configuration file at path and returns its revision.
func Load(path string) (*Config, Revision, error) {
	panic("SUB-AGENT-TODO: read path (missing -> error satisfying errors.Is(err, fs.ErrNotExist)); rev = lowercase hex SHA-256 of the bytes; Parse + Validate; on invalid return (nil, rev, err)")
}

// Save validates c and atomically replaces the file at path if its current
// revision equals expected ("" means absent), returning the new revision.
func Save(path string, c *Config, expected Revision) (Revision, error) {
	panic("SUB-AGENT-TODO: Validate (invalid -> no write); Marshal; mkdir parent 0700; exclusive unix.Flock on <path>.lock; compare current rev with expected (mismatch -> ErrRevisionConflict, file untouched); os.CreateTemp in same dir, chmod 0600, write, Sync, close; rename(tmp, path); fsync dir; remove temp on any failure; return new rev")
}
