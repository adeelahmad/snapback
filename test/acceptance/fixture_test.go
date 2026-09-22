//go:build integration

package acceptance

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

type resticSnapshot struct {
	ID       string    `json:"id"`
	Time     time.Time `json:"time"`
	Hostname string    `json:"hostname"`
	Tags     []string  `json:"tags"`
	Paths    []string  `json:"paths"`
}

type resticNode struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
}

func restic(t *testing.T, fx fixture, args ...string) []byte {
	t.Helper()
	full := append([]string{"-r", fx.Repo, "--password-file", fx.PasswordFile}, args...)
	out, err := exec.Command("restic", full...).Output()
	if err != nil {
		t.Fatalf("restic %s: %v", strings.Join(args, " "), err)
	}
	return out
}

func TestFixtureCoversSpec20Set(t *testing.T) {
	recordEvidence(t, "fixtures")
	requireFUSE(t)
	fx := buildFixture(t)

	var snaps []resticSnapshot
	if err := json.Unmarshal(restic(t, fx, "snapshots", "--json"), &snaps); err != nil {
		t.Fatalf("parse restic snapshots --json: %v", err)
	}
	if len(snaps) == 0 {
		t.Fatal("restic snapshots = none, want the §20 fixture set")
	}

	hosts := map[string]int{}
	sets := map[string]bool{}
	minutes := map[time.Time]int{}
	var docs, relative bool
	for _, s := range snaps {
		hosts[s.Hostname]++
		tags := slices.Clone(s.Tags)
		slices.Sort(tags)
		paths := slices.Clone(s.Paths)
		slices.Sort(paths)
		sets[strings.Join(tags, ",")+"|"+strings.Join(paths, ",")] = true
		minutes[s.Time.Truncate(time.Minute)]++
		for _, p := range s.Paths {
			if p == "proj/docs" || strings.HasSuffix(p, "/proj/docs") {
				docs = true
			}
			if !filepath.IsAbs(p) {
				relative = true
			}
		}
		if slices.Contains(s.Tags, "relative") {
			relative = true
		}
	}
	if got := hosts["lin"]; got < 3 {
		t.Errorf("snapshots on host lin = %d, want >= 3", got)
	}
	if got := hosts["mac"]; got < 1 {
		t.Errorf("snapshots on host mac = %d, want >= 1", got)
	}
	if got := len(sets); got < 2 {
		t.Errorf("distinct tag/path sets = %d, want >= 2", got)
	}
	shared := false
	for _, n := range minutes {
		if n >= 2 {
			shared = true
		}
	}
	if !shared {
		t.Error("no two snapshots share a minute, want a same-minute pair")
	}
	if !docs {
		t.Error("no snapshot with path proj/docs, want the subdir snap")
	}
	if !relative {
		t.Error("no relative-path snapshot, want one")
	}

	s2 := fx.IDs["S2"]
	if s2 == "" {
		t.Fatalf("fixture IDs[S2] is empty, want the S2 snapshot ID (IDs: %v)", fx.IDs)
	}
	nodes := map[string]string{}
	sc := bufio.NewScanner(bytes.NewReader(restic(t, fx, "ls", "--json", s2)))
	for sc.Scan() {
		var n resticNode
		if err := json.Unmarshal(sc.Bytes(), &n); err != nil || n.Name == "" {
			continue
		}
		nodes[n.Name] = n.Type
	}
	if len(nodes) == 0 {
		t.Fatalf("restic ls %s listed no nodes, want the S2 tree", s2)
	}
	for _, name := range []string{"new.txt", "sub dir", "ünï"} {
		if _, ok := nodes[name]; !ok {
			t.Errorf("restic ls S2 lacks %q, want present", name)
		}
	}
	if got := nodes["ln"]; got != "symlink" {
		t.Errorf("restic ls S2 node ln type = %q, want symlink", got)
	}
	if _, ok := nodes["gone.txt"]; ok {
		t.Error("restic ls S2 contains gone.txt, want deleted")
	}
}
