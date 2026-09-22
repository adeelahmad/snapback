package links

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	bolt "go.etcd.io/bbolt"
)

const (
	keyA = "0123456789abcdef0123456789abcdef"
	keyB = "11111111111111111111111111111111"
	keyC = "22222222222222222222222222222222"
)

func openTestRegistry(t *testing.T, path string) *Registry {
	t.Helper()
	reg, err := OpenRegistry(path)
	if err != nil {
		t.Fatalf("OpenRegistry(%q) = %v, want nil error", path, err)
	}
	if reg == nil {
		t.Fatalf("OpenRegistry(%q) = nil registry, want non-nil", path)
	}
	return reg
}

func testRecord(key, rel string, state State) Record {
	return Record{
		Key:    key,
		RootID: "home",
		Rel:    []byte(rel),
		Dir:    []byte("/r/" + rel),
		Target: "/h/roots/home/dirs/" + key,
		State:  state,
	}
}

func TestRegistryPutGetRoundTrip(t *testing.T) {
	reg := openTestRegistry(t, filepath.Join(t.TempDir(), "links.db"))
	defer func() { _ = reg.Close() }()

	want := testRecord(keyA, "a/\xff b", StatePending)
	if err := reg.Put(want); err != nil {
		t.Fatalf("Put(%q) = %v, want nil", want.Key, err)
	}
	got, found, err := reg.Get(keyA)
	if err != nil || !found {
		t.Fatalf("Get(%q) = _, %t, %v, want found with nil error", keyA, found, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Get(%q) = %+v, want %+v", keyA, got, want)
	}
}

func TestRegistryPersistsAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "links.db")
	reg := openTestRegistry(t, path)
	rec := testRecord(keyA, "docs", StateOwned)
	if err := reg.Put(rec); err != nil {
		t.Fatalf("Put(%q) = %v, want nil", rec.Key, err)
	}
	if err := reg.Close(); err != nil {
		t.Fatalf("Close() = %v, want nil", err)
	}

	reg = openTestRegistry(t, path)
	defer func() { _ = reg.Close() }()
	got, found, err := reg.Get(keyA)
	if err != nil || !found {
		t.Fatalf("Get(%q) after reopen = _, %t, %v, want found with nil error", keyA, found, err)
	}
	if got.State != StateOwned {
		t.Errorf("Get(%q).State after reopen = %v, want %v", keyA, got.State, StateOwned)
	}
}

func TestRegistryKeyCollisionRejected(t *testing.T) {
	reg := openTestRegistry(t, filepath.Join(t.TempDir(), "links.db"))
	defer func() { _ = reg.Close() }()

	first := testRecord(keyA, "a", StateOwned)
	if err := reg.Put(first); err != nil {
		t.Fatalf("Put(%q, rel a) = %v, want nil", keyA, err)
	}
	clash := testRecord(keyA, "b", StateOwned)
	if err := reg.Put(clash); !errors.Is(err, ErrKeyCollision) {
		t.Errorf("Put(%q, rel b) = %v, want %v", keyA, err, ErrKeyCollision)
	}
	got, found, err := reg.Get(keyA)
	if err != nil || !found {
		t.Fatalf("Get(%q) = _, %t, %v, want found with nil error", keyA, found, err)
	}
	if !reflect.DeepEqual(got, first) {
		t.Errorf("Get(%q) = %+v, want original %+v", keyA, got, first)
	}
}

func TestRegistryDeleteAndList(t *testing.T) {
	reg := openTestRegistry(t, filepath.Join(t.TempDir(), "links.db"))
	defer func() { _ = reg.Close() }()

	for _, rec := range []Record{
		testRecord(keyC, "c", StateOwned),
		testRecord(keyA, "a", StateOwned),
		testRecord(keyB, "b", StateOwned),
	} {
		if err := reg.Put(rec); err != nil {
			t.Fatalf("Put(%q) = %v, want nil", rec.Key, err)
		}
	}

	recs, err := reg.List()
	if err != nil {
		t.Fatalf("List() = %v, want nil error", err)
	}
	if got, want := recordKeys(recs), []string{keyA, keyB, keyC}; !reflect.DeepEqual(got, want) {
		t.Errorf("List() keys = %v, want %v", got, want)
	}

	if err := reg.Delete(keyB); err != nil {
		t.Fatalf("Delete(%q) = %v, want nil", keyB, err)
	}
	recs, err = reg.List()
	if err != nil {
		t.Fatalf("List() after Delete = %v, want nil error", err)
	}
	if got, want := recordKeys(recs), []string{keyA, keyC}; !reflect.DeepEqual(got, want) {
		t.Errorf("List() keys after Delete(%q) = %v, want %v", keyB, got, want)
	}
	if _, found, err := reg.Get(keyB); err != nil || found {
		t.Errorf("Get(%q) after Delete = _, %t, %v, want not found with nil error", keyB, found, err)
	}
}

func recordKeys(recs []Record) []string {
	keys := make([]string, 0, len(recs))
	for _, r := range recs {
		keys = append(keys, r.Key)
	}
	return keys
}

func TestRegistryBucketsCreated(t *testing.T) {
	path := filepath.Join(t.TempDir(), "links.db")
	reg := openTestRegistry(t, path)
	if err := reg.Close(); err != nil {
		t.Fatalf("Close() = %v, want nil", err)
	}

	db, err := bolt.Open(path, 0o600, &bolt.Options{ReadOnly: true})
	if err != nil {
		t.Fatalf("bolt.Open(%q, read-only) = %v, want nil error", path, err)
	}
	defer func() { _ = db.Close() }()
	err = db.View(func(tx *bolt.Tx) error {
		for _, name := range []string{"links", "keys"} {
			if tx.Bucket([]byte(name)) == nil {
				t.Errorf("bucket %q missing, want present", name)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("db.View() = %v, want nil", err)
	}
}
