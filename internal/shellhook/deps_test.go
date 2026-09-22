package shellhook

import (
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var guardShells = []string{"bash", "zsh", "fish"}

func TestNotifyPathImportsNoConfig(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("os.ReadDir(.) = %v", err)
	}
	imports := map[string]bool{}
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parser.ParseFile(%q) = %v", name, err)
		}
		for _, imp := range f.Imports {
			p, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				t.Fatalf("strconv.Unquote(%s) = %v", imp.Path.Value, err)
			}
			imports[p] = true
		}
	}
	if len(imports) == 0 {
		t.Fatal("no imports found in package sources, want at least internal/ipc")
	}
	if !imports["github.com/adeelahmad/snapback/internal/ipc"] {
		t.Errorf("package imports = %v, want internal/ipc among them", imports)
	}
	for _, banned := range []string{"github.com/adeelahmad/snapback/internal/config", "gopkg.in/yaml.v3"} {
		if imports[banned] {
			t.Errorf("package imports %q, want no config parse on the notify path", banned)
		}
	}
}

func TestSnippetsDetachAndSilenceNotify(t *testing.T) {
	detach := map[string]string{"bash": "&)", "zsh": "&!", "fish": "& disown"}
	for _, shell := range guardShells {
		src, err := Script(shell)
		if err != nil {
			t.Fatalf("Script(%q) error = %v", shell, err)
		}
		var lines []string
		for _, l := range strings.Split(src, "\n") {
			if strings.Contains(l, "snapback notify") {
				lines = append(lines, l)
			}
		}
		if len(lines) == 0 {
			t.Errorf("Script(%q) has no `snapback notify` line, want at least one", shell)
			continue
		}
		for _, l := range lines {
			for _, want := range []string{">/dev/null 2>&1", "</dev/null", "--timeout 200ms", detach[shell]} {
				if !strings.Contains(l, want) {
					t.Errorf("Script(%q) notify line %q lacks %q", shell, l, want)
				}
			}
			if !regexp.MustCompile(` -- "?\$\{?dir`).MatchString(l) {
				t.Errorf("Script(%q) notify line %q lacks `--` before the dir argument", shell, l)
			}
		}
	}
}

// outputCmd matches a command word that can print to or read from the
// terminal when it stands as its own word.
var outputCmd = regexp.MustCompile(`(^|[\s;&|(!{])(echo|printf|print|read)(\s|$)`)

// sentinelCmd matches the uses that cannot reach the prompt: the `printf x`
// newline sentinel and `printf -v`, which writes into a variable.
var sentinelCmd = regexp.MustCompile(`\bprintf (x|-v )`)

func TestSnippetsHaveNoBarePromptOutput(t *testing.T) {
	for _, shell := range guardShells {
		src, err := Script(shell)
		if err != nil {
			t.Fatalf("Script(%q) error = %v", shell, err)
		}
		if strings.TrimSpace(src) == "" {
			t.Errorf("Script(%q) is empty, want a hook snippet", shell)
			continue
		}
		for i, l := range strings.Split(src, "\n") {
			code := strings.TrimSpace(l)
			if strings.HasPrefix(code, "#") {
				continue
			}
			if !outputCmd.MatchString(sentinelCmd.ReplaceAllString(code, "")) {
				continue
			}
			if !strings.Contains(code, "/dev/null") {
				t.Errorf("Script(%q) line %d %q prints or reads outside a /dev/null redirect", shell, i+1, code)
			}
		}
	}
}
