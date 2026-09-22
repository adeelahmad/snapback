package acceptance

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"testing"
)

// evidenceNamePattern is the recordEvidence name set that test/reports/v01_matrix_test.go
// and docs/reports/v0.1-acceptance.md expect: acc-NN (two digits), fixtures, and the
// perf-* and platform-* rows.
var evidenceNamePattern = regexp.MustCompile(`^(acc-[0-9]{2}|fixtures|perf-[a-z0-9-]+|platform-[a-z0-9-]+)$`)

var accTestPattern = regexp.MustCompile(`^TestAcc([0-9]{2})`)

// evidenceCalls returns, per top-level function, the literal names passed to recordEvidence.
func evidenceCalls(t *testing.T) map[string][]string {
	t.Helper()
	files, err := filepath.Glob("*_test.go")
	if err != nil {
		t.Fatalf("glob test files: %v", err)
	}
	fset := token.NewFileSet()
	calls := map[string][]string{}
	for _, name := range files {
		f, err := parser.ParseFile(fset, name, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if _, ok := calls[fn.Name.Name]; !ok {
				calls[fn.Name.Name] = nil
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok || len(call.Args) < 2 {
					return true
				}
				if id, ok := call.Fun.(*ast.Ident); !ok || id.Name != "recordEvidence" {
					return true
				}
				lit, ok := call.Args[1].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					t.Errorf("%s: %s passes a non-literal evidence name", fset.Position(call.Pos()), fn.Name.Name)
					return true
				}
				v, err := strconv.Unquote(lit.Value)
				if err != nil {
					t.Errorf("%s: unquote %s: %v", fset.Position(lit.Pos()), lit.Value, err)
					return true
				}
				calls[fn.Name.Name] = append(calls[fn.Name.Name], v)
				return true
			})
		}
	}
	return calls
}

func TestEvidenceNamesMatchMatrix(t *testing.T) {
	for fn, names := range evidenceCalls(t) {
		for _, name := range names {
			if !evidenceNamePattern.MatchString(name) {
				t.Errorf("%s: recordEvidence(t, %q) does not match %s", fn, name, evidenceNamePattern)
			}
		}
	}
}

func TestEveryAccTestRecordsEvidence(t *testing.T) {
	for fn, names := range evidenceCalls(t) {
		m := accTestPattern.FindStringSubmatch(fn)
		if m == nil {
			continue
		}
		want := "acc-" + m[1]
		if !slices.Contains(names, want) {
			t.Errorf("%s records evidence %q, want a recordEvidence(t, %q) call", fn, names, want)
		}
	}
}
