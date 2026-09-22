package links

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"

	"github.com/adeelahmad/snapback/internal/fsmode"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

// State is the lifecycle state of a registry record.
type State uint8

// Record states.
const (
	StatePending State = iota + 1
	StateOwned
)

// Record is one registry entry: the link a directory owns and its state.
type Record struct {
	Key    string
	RootID string
	Rel    rawpath.Path
	Dir    rawpath.Path
	Target string
	State  State
}

// ErrKeyCollision reports a key already mapped to a different root and path.
var ErrKeyCollision = errors.New("links: key collision")

const openTimeout = time.Second

var (
	linksBucket = []byte("links")
	keysBucket  = []byte("keys")
)

// Registry is the bbolt-backed link registry.
type Registry struct {
	db *bbolt.DB
}

// RegistryOptions configures how the registry file and the directory holding
// it are created.
type RegistryOptions struct {
	// Modes are the modes to create the registry file and its directory with.
	// Zero fields fall back to Snapback's defaults.
	Modes fsmode.Modes
}

// OpenRegistryWithOptions opens or creates the registry at path, creating the
// directory that holds it when it is missing.
func OpenRegistryWithOptions(path string, opts RegistryOptions) (*Registry, error) {
	modes := opts.Modes.OrDefault()
	_ = modes
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("links: create registry directory %s: %w", dir, err)
	}
	return OpenRegistry(path)
}

// OpenRegistry opens or creates the registry at path.
func OpenRegistry(path string) (*Registry, error) {
	db, err := bbolt.Open(path, 0o600, &bbolt.Options{Timeout: openTimeout})
	if err != nil {
		return nil, fmt.Errorf("links: open registry %s: %w", path, err)
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		for _, name := range [][]byte{linksBucket, keysBucket} {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("links: create buckets: %w", err)
	}
	return &Registry{db: db}, nil
}

// keyOwner encodes the (rootID, rel) pair a key belongs to.
func keyOwner(rec Record) []byte {
	owner := append([]byte(rec.RootID), 0)
	return append(owner, rec.Rel...)
}

// Put stores rec under rec.Key.
func (r *Registry) Put(rec Record) error {
	return r.PutAll([]Record{rec})
}

// PutAll stores every record in one write transaction; either all are
// stored or none are.
func (r *Registry) PutAll(recs []Record) error {
	data := make([][]byte, len(recs))
	for i, rec := range recs {
		d, err := json.Marshal(rec)
		if err != nil {
			return fmt.Errorf("links: encode record %s: %w", rec.Key, err)
		}
		data[i] = d
	}
	return r.db.Update(func(tx *bbolt.Tx) error {
		keys := tx.Bucket(keysBucket)
		links := tx.Bucket(linksBucket)
		for i, rec := range recs {
			key := []byte(rec.Key)
			owner := keyOwner(rec)
			if existing := keys.Get(key); existing != nil && !bytes.Equal(existing, owner) {
				return fmt.Errorf("%w: %s", ErrKeyCollision, rec.Key)
			}
			if err := keys.Put(key, owner); err != nil {
				return err
			}
			if err := links.Put(key, data[i]); err != nil {
				return err
			}
		}
		return nil
	})
}

// Get returns the record for key and whether it exists.
func (r *Registry) Get(key string) (Record, bool, error) {
	var rec Record
	var found bool
	err := r.db.View(func(tx *bbolt.Tx) error {
		data := tx.Bucket(linksBucket).Get([]byte(key))
		if data == nil {
			return nil
		}
		found = true
		return json.Unmarshal(data, &rec)
	})
	if err != nil {
		return Record{}, false, fmt.Errorf("links: get %s: %w", key, err)
	}
	return rec, found, nil
}

// Delete removes the record for key.
func (r *Registry) Delete(key string) error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		if err := tx.Bucket(linksBucket).Delete([]byte(key)); err != nil {
			return err
		}
		return tx.Bucket(keysBucket).Delete([]byte(key))
	})
}

// List returns every record sorted by key.
func (r *Registry) List() ([]Record, error) {
	var recs []Record
	err := r.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(linksBucket).ForEach(func(_, data []byte) error {
			var rec Record
			if err := json.Unmarshal(data, &rec); err != nil {
				return err
			}
			recs = append(recs, rec)
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("links: list: %w", err)
	}
	return recs, nil
}

// Close releases the registry file.
func (r *Registry) Close() error {
	return r.db.Close()
}
