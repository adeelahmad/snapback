package links

import (
	"errors"

	"go.etcd.io/bbolt"

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

// Registry is the bbolt-backed link registry.
type Registry struct {
	db *bbolt.DB
}

// OpenRegistry opens or creates the registry at path.
func OpenRegistry(path string) (*Registry, error) {
	panic("SUB-AGENT-TODO: T1 open bbolt at path with a timeout; create buckets links (key -> JSON Record) and keys (key -> rootID + raw rel)")
}

// Put stores rec under rec.Key.
func (r *Registry) Put(rec Record) error {
	panic("SUB-AGENT-TODO: T1 write JSON rec to links and rootID+rel to keys in one tx; a key mapped to a different (rootID, rel) returns ErrKeyCollision and writes nothing")
}

// Get returns the record for key and whether it exists.
func (r *Registry) Get(key string) (Record, bool, error) {
	panic("SUB-AGENT-TODO: T1 read and decode the links entry for key; a missing key returns false and nil error")
}

// Delete removes the record for key.
func (r *Registry) Delete(key string) error {
	panic("SUB-AGENT-TODO: T1 delete key from both the links and keys buckets")
}

// List returns every record sorted by key.
func (r *Registry) List() ([]Record, error) {
	panic("SUB-AGENT-TODO: T1 decode every links entry in key order (bbolt cursor order)")
}

// Close releases the registry file.
func (r *Registry) Close() error {
	panic("SUB-AGENT-TODO: T1 close the bbolt db")
}
