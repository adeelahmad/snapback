package otlp

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// modulePath is this repository's module path; telemetry code may import its
// own packages and nothing else outside the standard library.
const modulePath = "github.com/adeelahmad/snapback"

// vendorSDKPrefixes are the telemetry and crash-reporting SDKs that decision D1
// keeps out of go.mod. The exporter is hand-rolled over net/http instead, so a
// direct or indirect require of any of these is a regression, not a choice.
var vendorSDKPrefixes = []string{
	"go.opentelemetry.io/",
	"github.com/getsentry/",
	"github.com/DataDog/",
	"github.com/honeycombio/",
	"github.com/newrelic/",
	"github.com/segmentio/analytics",
	"github.com/posthog/",
}

// telemetryDirs are the package directories the import guard walks, relative to
// the repository root.
var telemetryDirs = []string{
	filepath.Join("internal", "telemetry"),
	filepath.Join("internal", "telemetry", "otlp"),
}

func TestNoDepVendorTelemetrySDKInGoMod(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	b, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}

	for _, mod := range requiredModules(string(b)) {
		for _, prefix := range vendorSDKPrefixes {
			if strings.HasPrefix(mod, prefix) {
				t.Errorf("go.mod requires %q, a vendor telemetry SDK (prefix %q); "+
					"decision D1 keeps telemetry on the standard library", mod, prefix)
			}
		}
	}
}

func TestNoDepTelemetryImportsAreStdlibOrLocal(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	fset := token.NewFileSet()

	for _, dir := range telemetryDirs {
		pkgDir := filepath.Join(root, dir)
		entries, err := os.ReadDir(pkgDir)
		if err != nil {
			t.Fatalf("read dir %s: %v", dir, err)
		}

		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}

			file, err := parser.ParseFile(fset, filepath.Join(pkgDir, name), nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("parse %s: %v", filepath.Join(dir, name), err)
			}

			for _, imp := range file.Imports {
				checkImport(t, filepath.Join(dir, name), imp)
			}
		}
	}
}

func checkImport(t *testing.T, file string, imp *ast.ImportSpec) {
	t.Helper()

	path, err := strconv.Unquote(imp.Path.Value)
	if err != nil {
		t.Fatalf("%s: unquote import %s: %v", file, imp.Path.Value, err)
	}

	if isStdlib(path) || path == modulePath || strings.HasPrefix(path, modulePath+"/internal/") {
		return
	}
	t.Errorf("%s imports %q: telemetry may import only the standard library and %s/internal/...",
		file, path, modulePath)
}

// isStdlib reports whether path names a standard library package. The go
// command's own rule applies: a first path element without a dot is reserved
// for the standard library.
func isStdlib(path string) bool {
	first, _, _ := strings.Cut(path, "/")
	return !strings.Contains(first, ".")
}

// requiredModules returns the module paths named by go.mod's require
// directives, in both the block and the single-line form. golang.org/x/mod is
// not a dependency of this repository, so go.mod is scanned by hand.
func requiredModules(goMod string) []string {
	var mods []string
	inBlock := false

	for _, line := range strings.Split(goMod, "\n") {
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)

		switch {
		case inBlock:
			if line == ")" {
				inBlock = false
				continue
			}
		case line == "require (":
			inBlock = true
			continue
		case strings.HasPrefix(line, "require "):
			line = strings.TrimSpace(strings.TrimPrefix(line, "require "))
		default:
			continue
		}

		if path, _, ok := strings.Cut(line, " "); ok && path != "" {
			mods = append(mods, path)
		}
	}
	return mods
}

// repoRoot walks up from the test's working directory to the directory holding
// go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("no go.mod above %s", dir)
		}
		dir = parent
	}
}
