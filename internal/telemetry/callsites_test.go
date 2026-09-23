package telemetry_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// telemetryImportPath is the import path a file must carry for its
// "telemetry" identifier to count as this package.
const telemetryImportPath = "github.com/adeelahmad/snapback/internal/telemetry"

// eventConstructors is the closed set of telemetry package functions that
// build an [Event] for delivery. A call to one of these, qualified as
// telemetry.<name>, is what this test treats as an emission call site: every
// real emission path (setup, daemon, mount, doctor) builds its event this way
// and hands it to a Client.Emit, so pinning the constructor calls pins every
// place an event can originate.
var eventConstructors = map[string]bool{
	"SetupCompleted": true,
	"DaemonStarted":  true,
	"MountReady":     true,
	"DoctorFailed":   true,
	"ErrorEvent":     true,
}

// callSite identifies one call to a telemetry event constructor by the
// calling package's directory (relative to the repo root) and the enclosing
// function name, not by line number, so an unrelated edit near the call does
// not break this test.
type callSite struct {
	dir  string
	fn   string
	ctor string
}

func (c callSite) key() string { return c.dir + "|" + c.fn + "|" + c.ctor }

// wantCallSites is the closed, hardcoded set of call sites to telemetry event
// constructors across internal/ and cmd/, pinned by walking the real tree for
// S6-09/T6. The plan row that created this test guessed six call sites before
// the rest of sprint 6 landed; the actual tree has eleven, because
// internal/cli/telemetry_show.go builds a sample of all five events (for the
// `telemetry show` preview, which never calls Emit) in addition to the six
// real emission sites in setup, daemon, mount and doctor. A twelfth call site
// anywhere in the tree fails this test until this list is amended.
var wantCallSites = []callSite{
	{"internal/cli", "emitSetupOutcome", "SetupCompleted"},
	{"internal/cli", "emitSetupOutcome", "ErrorEvent"},
	{"internal/cli", "showSampleBatch", "SetupCompleted"},
	{"internal/cli", "showSampleBatch", "DaemonStarted"},
	{"internal/cli", "showSampleBatch", "MountReady"},
	{"internal/cli", "showSampleBatch", "DoctorFailed"},
	{"internal/cli", "showSampleBatch", "ErrorEvent"},
	{"internal/daemon", "emitStarted", "DaemonStarted"},
	{"internal/doctor", "emitDoctorFailed", "DoctorFailed"},
	{"internal/mount", "mountReadyEvent", "MountReady"},
	{"internal/mount", "mountFailureEvent", "ErrorEvent"},
}

// TestCallSitesMatchGoldenSet walks internal/ and cmd/ with go/ast and asserts
// that every call to a telemetry event constructor is exactly the pinned set
// in wantCallSites. See plan.md row S6-09/T6.
func TestCallSitesMatchGoldenSet(t *testing.T) {
	repoRoot := findRepoRoot(t)

	got := collectCallSites(t, repoRoot)

	gotKeys := callSiteKeys(got)
	wantKeys := callSiteKeys(wantCallSites)

	if !slices.Equal(gotKeys, wantKeys) {
		t.Fatalf("telemetry event constructor call sites changed\ngot (%d):\n%s\nwant (%d):\n%s",
			len(gotKeys), strings.Join(gotKeys, "\n"), len(wantKeys), strings.Join(wantKeys, "\n"))
	}
}

func callSiteKeys(sites []callSite) []string {
	keys := make([]string, len(sites))
	for i, s := range sites {
		keys[i] = s.key()
	}
	sort.Strings(keys)
	return keys
}

// findRepoRoot returns the repository root, derived from this file's own
// path: internal/telemetry/callsites_test.go, three directories below root.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// collectCallSites walks internal/ and cmd/ under repoRoot and returns every
// call site matching a telemetry event constructor, skipping _test.go files,
// vendor directories and internal/telemetry itself (the constructors are
// defined there, not called there).
func collectCallSites(t *testing.T, repoRoot string) []callSite {
	t.Helper()
	var sites []callSite
	for _, root := range []string{"internal", "cmd"} {
		walkRoot := filepath.Join(repoRoot, root)
		err := filepath.WalkDir(walkRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			rel, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return err
			}
			relDir := filepath.ToSlash(filepath.Dir(rel))
			if relDir == "internal/telemetry" {
				return nil
			}
			sites = append(sites, findCallSitesInFile(t, path, relDir)...)
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", walkRoot, err)
		}
	}
	return sites
}

// findCallSitesInFile parses path and returns one callSite per call to a
// telemetry event constructor found in a top-level function or method body,
// keyed to relDir and the enclosing function's name.
func findCallSitesInFile(t *testing.T, path, relDir string) []callSite {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	localName := telemetryLocalName(f)
	if localName == "" {
		return nil
	}

	var sites []callSite
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			id, ok := sel.X.(*ast.Ident)
			if !ok || id.Name != localName {
				return true
			}
			if !eventConstructors[sel.Sel.Name] {
				return true
			}
			sites = append(sites, callSite{dir: relDir, fn: fd.Name.Name, ctor: sel.Sel.Name})
			return true
		})
	}
	return sites
}

// telemetryLocalName returns the identifier f uses for the telemetry
// package's import, honoring a rename, or "" if f does not import it.
func telemetryLocalName(f *ast.File) string {
	for _, imp := range f.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil || p != telemetryImportPath {
			continue
		}
		if imp.Name != nil {
			return imp.Name.Name
		}
		return "telemetry"
	}
	return ""
}
