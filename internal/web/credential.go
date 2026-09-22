package web

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
)

// repoIDRE is the repository id a credential file may be named after. It keeps
// the form's id inside the credentials directory: no separators, no "..".
var repoIDRE = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// storeCredential writes the password typed into the web form to
// <stateDir>/credentials/<repoID>.pass (directory 0700, file 0600) and returns
// that path, so only the path ever reaches config.yaml.
//
// The secret is written verbatim — restic reads the whole password file as the
// password, so no trailing newline is added. The file is written to a
// temporary name in the same directory and renamed over the old one, so a
// failed write never leaves a truncated password behind.
//
// An empty secret means "keep what is configured": it returns the existing
// file's path if there is one, leaving its bytes and mtime untouched, and
// otherwise the empty string.
func storeCredential(stateDir, repoID, secret string) (string, error) {
	if !repoIDRE.MatchString(repoID) {
		return "", fmt.Errorf("repository id %q: want only letters, digits, dot, dash or underscore", repoID)
	}
	path := filepath.Join(stateDir, "credentials", repoID+".pass")
	if secret == "" {
		if _, err := os.Stat(path); err != nil {
			return "", nil
		}
		return path, nil
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create credentials dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, repoID+".pass.*")
	if err != nil {
		return "", fmt.Errorf("create credential file: %w", err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return "", fmt.Errorf("chmod credential file: %w", err)
	}
	if _, err := tmp.WriteString(secret); err != nil {
		_ = tmp.Close()
		return "", fmt.Errorf("write credential file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close credential file: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return "", fmt.Errorf("install credential file: %w", err)
	}
	return path, nil
}

// credentialMode reports how repository i of the submitted form supplies its
// password: "typed" for a secret entered in the form, "file" for a path the
// user already has, which is also the default.
func credentialMode(v url.Values, i int) string {
	if v.Get(fmt.Sprintf("repositories[%d].password_mode", i)) == "typed" {
		return "typed"
	}
	return "file"
}
