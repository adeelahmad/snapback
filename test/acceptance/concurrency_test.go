//go:build integration

package acceptance

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

// countDirRecords counts the JSON objects anywhere in v whose "dir" equals dir.
func countDirRecords(v any, dir string) int {
	n := 0
	switch x := v.(type) {
	case map[string]any:
		if d, ok := x["dir"].(string); ok && d == dir {
			n++
		}
		for _, c := range x {
			n += countDirRecords(c, dir)
		}
	case []any:
		for _, c := range x {
			n += countDirRecords(c, dir)
		}
	}
	return n
}

func TestAcc09ConcurrentIdempotentMetachars(t *testing.T) {
	recordEvidence(t, "acc-09")
	requireLinkPrereqs(t)
	e := newEnv(t)
	root := filepath.Join(e.Root, "work")
	name := `a;$(touch pwned)'"`
	mkdirs(t, root, filepath.Join("proj", name))
	hist := writeLinkConfig(t, e, linkConfig{Root: root, SeedPath: "proj", MaxDepth: 1})
	startDaemon(t, e)
	dir := filepath.Join(root, "proj", name)

	const workers = 32
	codes := make([]int, workers)
	stderrs := make([]string, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if i%2 == 0 {
				_, stderrs[i], codes[i] = runSnapback(t, e, "link", dir)
				return
			}
			_, stderrs[i], codes[i] = runSnapback(t, e, "notify", "--session", "s"+strconv.Itoa(i), "--", dir)
		}()
	}
	wg.Wait()

	for i, code := range codes {
		if code != 0 {
			t.Errorf("call %d exit = %d, want 0 (stderr %q)", i, code, stderrs[i])
		}
	}
	rel := filepath.ToSlash(filepath.Join("proj", name))
	if got := ownedLinks(t, root, hist); len(got) != 1 || got[0] != rel {
		t.Errorf("owned links = %q, want exactly [%q]", got, rel)
	}
	stdout, stderr, code := runSnapback(t, e, "links", "list", "--managed", "--json")
	if code != 0 {
		t.Fatalf("snapback links list exit = %d, want 0 (stderr %q)", code, stderr)
	}
	var v any
	if err := json.Unmarshal([]byte(stdout), &v); err != nil {
		t.Fatalf("links list --json is not JSON: %v (%q)", err, stdout)
	}
	if got := countDirRecords(v, dir); got != 1 {
		t.Errorf("registry records for %q = %d, want 1", dir, got)
	}
	err := filepath.WalkDir(e.Root, func(p string, d fs.DirEntry, err error) error {
		if err == nil && d.Name() == "pwned" {
			t.Errorf("found injected file %s, want none", p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", e.Root, err)
	}
}
