package setup

import (
	"errors"
	"fmt"
	"io/fs"
	"reflect"

	"github.com/adeelahmad/snapback/internal/config"
)

// Report describes the configuration file a second setup run found.
type Report struct {
	Exists bool
	Same   bool
	Lines  []string
}

// Existing reports what the configuration file at path already holds compared
// with want, over the fields setup writes: the repository URI, its password
// file and restic binary, and each root's local path and prefix map. It never
// writes, so a caller that gets Same skips Save and leaves the file untouched.
func Existing(path string, want *config.Config) (Report, error) {
	have, _, err := config.Load(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Report{}, nil
	}
	if err != nil {
		return Report{Exists: true}, fmt.Errorf("setup.existing: %w", err)
	}

	diffs := diffConfig(have, want)
	if len(diffs) == 0 {
		return Report{Exists: true, Same: true, Lines: sameLines(path, have)}, nil
	}
	return Report{Exists: true, Lines: append([]string{"config: " + path + " (differs)"}, diffs...)}, nil
}

// sameLines describes the configuration that is already in place.
func sameLines(path string, have *config.Config) []string {
	lines := []string{"config: " + path + " (unchanged)"}
	for _, repo := range have.Repositories {
		lines = append(lines, "repository: "+repo.Repository)
	}
	for _, root := range have.Roots {
		lines = append(lines, "root: "+root.LocalPath)
	}
	return lines
}

// diffConfig returns one line per field of have that want would change.
func diffConfig(have, want *config.Config) []string {
	var lines []string
	lines = append(lines, diffRepositories(have.Repositories, want.Repositories)...)
	return append(lines, diffRoots(have.Roots, want.Roots)...)
}

// diffRepositories compares the repository fields setup writes.
func diffRepositories(have, want []config.Repository) []string {
	var lines []string
	if len(have) != len(want) {
		return []string{fmt.Sprintf("repositories: %d → %d", len(have), len(want))}
	}
	for i, w := range want {
		lines = append(lines, changed("repository", have[i].Repository, w.Repository)...)
		lines = append(lines, changed("password_file", have[i].PasswordFile, w.PasswordFile)...)
		lines = append(lines, changed("restic_binary", have[i].ResticBinary, w.ResticBinary)...)
	}
	return lines
}

// diffRoots compares each root's local path and prefix map.
func diffRoots(have, want []config.Root) []string {
	if len(have) != len(want) {
		return []string{fmt.Sprintf("roots: %d → %d", len(have), len(want))}
	}
	var lines []string
	for i, w := range want {
		lines = append(lines, changed("root", have[i].LocalPath, w.LocalPath)...)
		if !reflect.DeepEqual(have[i].PrefixMap, w.PrefixMap) {
			lines = append(lines, "prefix_map: "+w.LocalPath+" differs")
		}
	}
	return lines
}

// changed returns a single "field: old → new" line, or nothing if old equals new.
func changed(field, old, updated string) []string {
	if old == updated {
		return nil
	}
	return []string{field + ": " + old + " → " + updated}
}
