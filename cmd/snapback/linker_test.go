package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeLinkerConfig writes a one-repository, one-root config file under tmp
// and returns its path, the state dir and the root dir. extra is appended
// verbatim, so a caller can make the config invalid.
func writeLinkerConfig(t *testing.T, tmp, extra string) (cfgPath, stateDir, root string) {
	t.Helper()
	restic := filepath.Join(tmp, "restic")
	if err := os.WriteFile(restic, []byte(fakeRestic), 0o755); err != nil {
		t.Fatalf("os.WriteFile(%s) = %v", restic, err)
	}
	pw := filepath.Join(tmp, "password")
	if err := os.WriteFile(pw, []byte("secret\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%s) = %v", pw, err)
	}
	root = filepath.Join(tmp, "work")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", root, err)
	}
	stateDir = filepath.Join(tmp, "state")
	var b strings.Builder
	b.WriteString("version: 1\n")
	b.WriteString("state_dir: " + stateDir + "\n")
	b.WriteString("repositories:\n")
	b.WriteString("  - id: personal\n")
	b.WriteString("    repository: " + filepath.Join(tmp, "repo") + "\n")
	b.WriteString("    restic_binary: " + restic + "\n")
	b.WriteString("    password_file: " + pw + "\n")
	b.WriteString("roots:\n")
	b.WriteString("  - id: work\n")
	b.WriteString("    local_path: " + root + "\n")
	b.WriteString("    repository_id: personal\n")
	b.WriteString(extra)
	cfgPath = filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(b.String()), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%s) = %v", cfgPath, err)
	}
	return cfgPath, stateDir, root
}

type linksListEnvelope struct {
	OK   bool `json:"ok"`
	Data []struct {
		Key string `json:"key"`
		Dir string `json:"dir"`
	} `json:"data"`
}

// runLinksList runs "links list --json" against cfgPath and decodes the result.
func runLinksList(t *testing.T, cfgPath string) linksListEnvelope {
	t.Helper()
	var stdout, stderr bytes.Buffer
	args := []string{"--config", cfgPath, "links", "list", "--json"}
	if code := run(args, &stdout, &stderr); code != 0 {
		t.Fatalf("run(%q) = %d, want 0 (stdout %q, stderr %q)", args, code, stdout.String(), stderr.String())
	}
	var env linksListEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("run(%q) stdout = %q, not JSON: %v", args, stdout.String(), err)
	}
	if !env.OK {
		t.Errorf("run(%q) ok = false, want true", args)
	}
	return env
}

func TestRealDepsLinkerWired(t *testing.T) {
	tmp := shortTempDir(t)
	cfgPath, _, root := writeLinkerConfig(t, tmp, "")
	dir := filepath.Join(root, "docs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", dir, err)
	}

	if realDeps(cfgPath).Linker == nil {
		t.Fatalf("realDeps(%s).Linker = nil, want a lazy production Linker", cfgPath)
	}

	if got := runLinksList(t, cfgPath); len(got.Data) != 0 {
		t.Errorf("links list --json on an empty state dir = %+v, want an empty list", got.Data)
	}

	var stdout, stderr bytes.Buffer
	args := []string{"--config", cfgPath, "link", "--json", dir}
	if code := run(args, &stdout, &stderr); code != 0 {
		t.Fatalf("run(%q) = %d, want 0 (stdout %q, stderr %q)", args, code, stdout.String(), stderr.String())
	}

	got := runLinksList(t, cfgPath)
	if len(got.Data) != 1 || got.Data[0].Dir != dir {
		t.Errorf("links list --json after link %s = %+v, want one record for %s", dir, got.Data, dir)
	}
}

func TestRealDepsLinkerBadConfig(t *testing.T) {
	tmp := shortTempDir(t)
	cfgPath, stateDir, _ := writeLinkerConfig(t, tmp, "bogus_field: true\n")

	if realDeps(cfgPath).Linker == nil {
		t.Fatalf("realDeps(%s).Linker = nil, want a lazy production Linker", cfgPath)
	}

	var stdout, stderr bytes.Buffer
	args := []string{"--config", cfgPath, "links", "list", "--json"}
	code := run(args, &stdout, &stderr)

	if code != 1 {
		t.Errorf("run(%q) = %d, want 1 (stdout %q, stderr %q)", args, code, stdout.String(), stderr.String())
	}
	var env struct {
		OK   bool   `json:"ok"`
		Code string `json:"code"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("run(%q) stdout = %q, not JSON: %v", args, stdout.String(), err)
	}
	if got, want := env.Code, "invalid_configuration"; got != want {
		t.Errorf("run(%q) code = %q, want %q", args, got, want)
	}
	if entries, err := os.ReadDir(stateDir); err == nil && len(entries) != 0 {
		t.Errorf("state dir %s after an invalid config has %d entries, want none (no registry file)", stateDir, len(entries))
	}
}
