package crawler

import (
	"errors"
	"io/fs"
	"reflect"
	"slices"
	"strings"
	"testing"
)

const (
	testRoot = "/tmp/x"
	testDest = "/tmp/y"
)

var wantToolNames = []string{"rg", "rg -L", "fd", "fd -L", "find", "find -L", "rsync -a"}

// allTools returns Tools() and fails the test unless it has the seven planned
// rows, so per-row checks can never pass on an empty table.
func allTools(t *testing.T) []Tool {
	t.Helper()
	tools := Tools()
	if len(tools) != len(wantToolNames) {
		t.Fatalf("len(Tools()) = %d, want %d", len(tools), len(wantToolNames))
	}
	return tools
}

func toolByName(t *testing.T, name string) Tool {
	t.Helper()
	for _, tool := range Tools() {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("Tools() has no row %q", name)
	return Tool{}
}

// fakeLookPath finds only the named binaries, at /fake/bin/<name>.
func fakeLookPath(found ...string) func(string) (string, error) {
	return func(name string) (string, error) {
		if slices.Contains(found, name) {
			return "/fake/bin/" + name, nil
		}
		return "", fs.ErrNotExist
	}
}

func TestToolTableExactRows(t *testing.T) {
	var got []string
	for _, tool := range Tools() {
		got = append(got, tool.Name)
	}
	if !reflect.DeepEqual(got, wantToolNames) {
		t.Errorf("Tools() names = %q, want %q", got, wantToolNames)
	}
}

func TestToolTableFollowFlags(t *testing.T) {
	wantFollows := map[string]bool{"rg -L": true, "fd -L": true, "find -L": true}
	for _, tool := range allTools(t) {
		want := wantFollows[tool.Name]
		if tool.Follows != want {
			t.Errorf("Tool %q Follows = %t, want %t", tool.Name, tool.Follows, want)
		}
		argv := tool.Argv(testRoot, testDest)
		if got := slices.Contains(argv, "-L"); got != want {
			t.Errorf("Tool %q Argv(%q, %q) = %q, contains -L = %t, want %t", tool.Name, testRoot, testDest, argv, got, want)
		}
	}
}

func TestToolArgsAreArraysWithRoot(t *testing.T) {
	banned := []string{";", "|", "&&", "$(", "`"}
	for _, tool := range allTools(t) {
		argv := tool.Argv(testRoot, testDest)
		if len(argv) == 0 {
			t.Errorf("Tool %q Argv(%q, %q) is empty", tool.Name, testRoot, testDest)
			continue
		}
		if !slices.Contains(argv, testRoot) {
			t.Errorf("Tool %q Argv(%q, %q) = %q, want root %q as its own element", tool.Name, testRoot, testDest, argv, testRoot)
		}
		for _, arg := range argv {
			if arg == "sh" || arg == "bash" || arg == "-c" {
				t.Errorf("Tool %q Argv(%q, %q) = %q, contains shell element %q", tool.Name, testRoot, testDest, argv, arg)
			}
			for _, b := range banned {
				if strings.Contains(arg, b) {
					t.Errorf("Tool %q Argv(%q, %q) element %q contains shell metacharacter %q", tool.Name, testRoot, testDest, arg, b)
				}
			}
		}
	}
}

func TestSearchToolsVisitHiddenEntries(t *testing.T) {
	tests := []struct {
		name      string
		wantFlags []string
	}{
		{"rg", []string{"--hidden", "--no-ignore"}},
		{"rg -L", []string{"--hidden", "--no-ignore"}},
		{"fd", []string{"-H", "-I"}},
		{"fd -L", []string{"-H", "-I"}},
	}
	for _, tt := range tests {
		argv := toolByName(t, tt.name).Argv(testRoot, testDest)
		for _, flag := range tt.wantFlags {
			if !slices.Contains(argv, flag) {
				t.Errorf("Tool %q Argv(%q, %q) = %q, want it to contain %q", tt.name, testRoot, testDest, argv, flag)
			}
		}
	}
}

func TestRsyncArgsUseDestination(t *testing.T) {
	argv := toolByName(t, "rsync -a").Argv(testRoot, testDest)
	if len(argv) == 0 {
		t.Fatalf("rsync -a Argv(%q, %q) is empty", testRoot, testDest)
	}
	if !slices.Contains(argv, "-a") {
		t.Errorf("rsync -a Argv(%q, %q) = %q, want it to contain -a", testRoot, testDest, argv)
	}
	for _, bad := range []string{"-L", "--copy-links"} {
		if slices.Contains(argv, bad) {
			t.Errorf("rsync -a Argv(%q, %q) = %q, must not contain %q", testRoot, testDest, argv, bad)
		}
	}
	if got, want := argv[len(argv)-1], testDest+"/"; got != want {
		t.Errorf("rsync -a Argv(%q, %q) last element = %q, want %q", testRoot, testDest, got, want)
	}
}

func TestResolvePrefersFd(t *testing.T) {
	got, err := Resolve(toolByName(t, "fd"), fakeLookPath("fd", "fdfind"))
	if want := "/fake/bin/fd"; got != want || err != nil {
		t.Errorf("Resolve(fd) = %q, %v, want %q, nil", got, err, want)
	}
}

func TestResolveFallsBackToFdfind(t *testing.T) {
	got, err := Resolve(toolByName(t, "fd -L"), fakeLookPath("fdfind"))
	if want := "/fake/bin/fdfind"; got != want || err != nil {
		t.Errorf("Resolve(fd -L) = %q, %v, want %q, nil", got, err, want)
	}
}

func TestResolveMissingNamesBinaries(t *testing.T) {
	tests := []struct {
		name      string
		wantInErr []string
	}{
		{"fd", []string{"fd", "fdfind"}},
		{"rsync -a", []string{"rsync"}},
	}
	for _, tt := range tests {
		_, err := Resolve(toolByName(t, tt.name), fakeLookPath())
		if !errors.Is(err, ErrToolMissing) {
			t.Errorf("Resolve(%q) error = %v, want errors.Is ErrToolMissing", tt.name, err)
			continue
		}
		for _, bin := range tt.wantInErr {
			if !strings.Contains(err.Error(), bin) {
				t.Errorf("Resolve(%q) error = %q, want it to name %q", tt.name, err, bin)
			}
		}
	}

	row := VSCodeRow()
	if row.Status != "not-tested-here" {
		t.Errorf("VSCodeRow().Status = %q, want %q", row.Status, "not-tested-here")
	}
	if !strings.Contains(row.Reason, "search.followSymlinks") {
		t.Errorf("VSCodeRow().Reason = %q, want it to mention search.followSymlinks", row.Reason)
	}
	if len(row.Argv) != 0 {
		t.Errorf("VSCodeRow().Argv = %q, want none", row.Argv)
	}
}
