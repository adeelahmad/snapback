package shellhook

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite testdata/golden files")

func TestScriptUnknownShell(t *testing.T) {
	for _, shell := range []string{"", "tcsh", "BASH", "sh"} {
		got, err := Script(shell)
		if got != "" {
			t.Errorf("Script(%q) = %q, want empty", shell, got)
		}
		if err == nil {
			t.Errorf("Script(%q) error = nil, want an error naming bash|zsh|fish", shell)
			continue
		}
		if !strings.Contains(err.Error(), "bash|zsh|fish") {
			t.Errorf("Script(%q) error = %q, want it to contain %q", shell, err, "bash|zsh|fish")
		}
	}
}

func TestScriptBashGolden(t *testing.T) {
	got, err := Script("bash")
	if err != nil {
		t.Fatalf(`Script("bash") error = %v, want nil`, err)
	}
	golden := filepath.Join("testdata", "golden", "bash.golden")
	if *update {
		if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("golden %s missing (run with -update): %v", golden, err)
	}
	if got != string(want) {
		t.Errorf(`Script("bash") = %q, want golden %q`, got, want)
	}
	for _, s := range []string{"pwd -P", "notify --timeout 200ms", "/dev/null", "PROMPT_COMMAND"} {
		if !strings.Contains(got, s) {
			t.Errorf(`Script("bash") lacks %q`, s)
		}
	}
	// Assigning PROMPT_COMMAND to the hook alone would drop the user's value.
	clobber := regexp.MustCompile(`(?m)^\s*PROMPT_COMMAND=["']?__snapback`)
	if clobber.MatchString(got) {
		t.Errorf(`Script("bash") assigns PROMPT_COMMAND to the hook without keeping its prior value`)
	}
}
