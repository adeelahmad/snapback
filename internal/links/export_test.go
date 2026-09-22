package links

import (
	"testing"

	"go.etcd.io/bbolt"
)

// writeTxns returns the number of write transactions committed to r so far.
// bbolt advances the meta txid once per committed write transaction.
func writeTxns(t *testing.T, r *Registry) int {
	t.Helper()
	var id int
	if err := r.db.View(func(tx *bbolt.Tx) error {
		id = tx.ID()
		return nil
	}); err != nil {
		t.Fatalf("View() = %v", err)
	}
	return id
}
