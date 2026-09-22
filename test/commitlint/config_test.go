package commitlint

import (
	"encoding/json"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

const configPath = ".commitlintrc.json"

var allowedTypes = []string{"feat", "fix", "refactor", "docs", "test", "chore", "perf", "ci"}

func loadConfig(t *testing.T) map[string]any {
	t.Helper()
	raw := readRepoFile(t, configPath)
	var cfg map[string]any
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		t.Fatalf("unmarshal %s: %v", configPath, err)
	}
	return cfg
}

func typeEnum(t *testing.T) []any {
	t.Helper()
	cfg := loadConfig(t)
	rules, ok := cfg["rules"].(map[string]any)
	if !ok {
		t.Fatalf("rules is %T, want object", cfg["rules"])
	}
	rule, ok := rules["type-enum"].([]any)
	if !ok || len(rule) != 3 {
		t.Fatalf("rules[type-enum] = %#v, want [level, applicability, types]", rules["type-enum"])
	}
	return rule
}

func toStrings(t *testing.T, v any, what string) []string {
	t.Helper()
	items, ok := v.([]any)
	if !ok {
		t.Fatalf("%s is %T, want array", what, v)
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		s, ok := item.(string)
		if !ok {
			t.Fatalf("%s element %#v is not a string", what, item)
		}
		out = append(out, s)
	}
	return out
}

func TestConfigIsValidJSON(t *testing.T) {
	raw := readRepoFile(t, configPath)
	if strings.TrimSpace(raw) == "" {
		t.Fatalf("%s is empty", configPath)
	}
	var cfg map[string]any
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		t.Fatalf("unmarshal %s: %v", configPath, err)
	}
}

func TestConfigExtendsConventional(t *testing.T) {
	cfg := loadConfig(t)
	got := toStrings(t, cfg["extends"], "extends")
	want := []string{"@commitlint/config-conventional"}
	if !slices.Equal(got, want) {
		t.Fatalf("extends = %v, want %v", got, want)
	}
}

func TestConfigTypeEnumExactlyEightTypes(t *testing.T) {
	got := toStrings(t, typeEnum(t)[2], "type-enum[2]")
	sorted := slices.Sorted(slices.Values(got))
	if len(slices.Compact(slices.Clone(sorted))) != len(sorted) {
		t.Fatalf("type-enum has duplicates: %v", got)
	}
	want := slices.Sorted(slices.Values(allowedTypes))
	if !slices.Equal(sorted, want) {
		t.Fatalf("type-enum types = %v, want %v", sorted, want)
	}
}

func TestConfigTypeEnumIsErrorAlways(t *testing.T) {
	rule := typeEnum(t)
	level, ok := rule[0].(float64)
	if !ok || level != 2 {
		t.Fatalf("type-enum level = %#v, want 2 (error)", rule[0])
	}
	if rule[1] != "always" {
		t.Fatalf("type-enum applicability = %#v, want \"always\"", rule[1])
	}
}

func TestCommitlintCLIVerdicts(t *testing.T) {
	bin, err := exec.LookPath("commitlint")
	if err != nil {
		t.Skip("commitlint CLI not on PATH")
	}
	cases := []struct {
		msg  string
		pass bool
	}{
		{"feat: add x", true},
		{"ci: pin node", true},
		{"Initial commit", false},
		{"Update README.md", false},
		{"build: bump", false},
	}
	for _, tc := range cases {
		t.Run(tc.msg, func(t *testing.T) {
			cmd := exec.Command(bin, "--config", repoRoot+"/"+configPath)
			cmd.Stdin = strings.NewReader(tc.msg + "\n")
			out, err := cmd.CombinedOutput()
			if tc.pass && err != nil {
				t.Fatalf("commitlint rejected %q: %v\n%s", tc.msg, err, out)
			}
			if !tc.pass && err == nil {
				t.Fatalf("commitlint accepted %q, want rejection\n%s", tc.msg, out)
			}
		})
	}
}
