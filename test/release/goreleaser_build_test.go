package release_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/version"
)

const (
	goreleaserFile = ".goreleaser.yaml"
	versionPkg     = "github.com/adeelahmad/snapback/internal/version"
)

var ldflagXRe = regexp.MustCompile(`-X[ =]+(` + regexp.QuoteMeta(versionPkg) + `\.[A-Za-z]+=(?:\{\{[^}]*\}\}(?:/\{\{[^}]*\}\})?|[^\s"']+))`)

func hasLine(text, want string) bool {
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}

func ldflagXValues(text string) []string {
	var out []string
	for _, m := range ldflagXRe.FindAllStringSubmatch(text, -1) {
		out = append(out, m[1])
	}
	return out
}

func TestGoreleaserConfigHeader(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	if !hasLine(text, "version: 2") {
		t.Errorf("%s: missing line `version: 2`", goreleaserFile)
	}
	if !hasLine(text, "project_name: snapback") {
		t.Errorf("%s: missing line `project_name: snapback`", goreleaserFile)
	}
}

func TestGoreleaserBuildsCmdSnapback(t *testing.T) {
	builds := yamlBlock(readRepoFile(t, goreleaserFile), "builds")
	for _, want := range []string{"main: ./cmd/snapback", "binary: snapback"} {
		if !strings.Contains(builds, want) {
			t.Errorf("builds block missing %q", want)
		}
	}
}

func TestGoreleaserCGODisabled(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	builds := yamlBlock(text, "builds")
	if !regexp.MustCompile(`CGO_ENABLED=0`).MatchString(builds) {
		t.Errorf("builds block missing CGO_ENABLED=0")
	}
	if regexp.MustCompile(`CGO_ENABLED=1`).MatchString(text) {
		t.Errorf("%s must not contain CGO_ENABLED=1", goreleaserFile)
	}
}

func TestGoreleaserTargetsExactlySeven(t *testing.T) {
	builds := yamlBlock(readRepoFile(t, goreleaserFile), "builds")
	targets := yamlBlock(builds, "targets")
	if targets == "" {
		t.Fatalf("builds block has no targets: list")
	}
	var raw []string
	for _, line := range strings.Split(targets, "\n")[1:] {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}
		raw = append(raw, strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")), `"'`))
	}
	seen := map[string]bool{}
	var got []string
	for _, tgt := range raw {
		parts := strings.Split(tgt, "_")
		if len(parts) < 2 {
			t.Errorf("malformed target %q", tgt)
			continue
		}
		osArch := parts[0] + "/" + parts[1]
		if seen[osArch] {
			t.Errorf("duplicate target %s (%q)", osArch, tgt)
		}
		seen[osArch] = true
		got = append(got, osArch)
	}
	want := []string{"darwin/amd64", "darwin/arm64", "linux/amd64", "linux/arm", "linux/arm64", "linux/mips", "linux/mipsle"}
	sort.Strings(got)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("targets = %v, want exactly %v", got, want)
	}
	for _, verbatim := range []string{"linux_arm_7", "linux_mips_softfloat", "linux_mipsle_softfloat"} {
		found := false
		for _, tgt := range raw {
			if tgt == verbatim {
				found = true
			}
		}
		if !found {
			t.Errorf("targets missing verbatim %q", verbatim)
		}
	}
}

func TestGoreleaserLdflagsTargetVersionVars(t *testing.T) {
	got := ldflagXValues(readRepoFile(t, goreleaserFile))
	want := []string{
		versionPkg + ".Version={{ .Version }}",
		versionPkg + ".Commit={{ .ShortCommit }}",
		versionPkg + ".Target={{ .Os }}/{{ .Arch }}",
	}
	if len(got) != len(want) {
		t.Fatalf("found %d -X flags %v, want exactly 3 %v", len(got), got, want)
	}
	sort.Strings(got)
	sort.Strings(want)
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("-X flag %q, want %q", got[i], want[i])
		}
	}
}

func TestGoreleaserVersionVarsCompile(t *testing.T) {
	ptrs := map[string]*string{
		"Version": &version.Version,
		"Commit":  &version.Commit,
		"Target":  &version.Target,
	}
	for name, p := range ptrs {
		if p == nil {
			t.Errorf("version.%s address is nil", name)
		}
	}
	if version.Version != "dev" {
		t.Errorf("version.Version = %q in un-ldflagged build, want %q", version.Version, "dev")
	}
}

func TestGoreleaserLdflagsProduceVersionedBinary(t *testing.T) {
	flags := ldflagXValues(readRepoFile(t, goreleaserFile))
	if len(flags) != 3 {
		t.Fatalf("found %d -X flags in %s, want 3", len(flags), goreleaserFile)
	}
	target := runtime.GOOS + "/" + runtime.GOARCH
	repl := strings.NewReplacer(
		"{{ .Version }}", "9.9.9",
		"{{ .ShortCommit }}", "abc1234",
		"{{ .Os }}/{{ .Arch }}", target,
	)
	var args []string
	for _, f := range flags {
		args = append(args, "-X "+repl.Replace(f))
	}
	bin := filepath.Join(t.TempDir(), "snapback")
	build := exec.Command("go", "build", "-ldflags", strings.Join(args, " "), "-o", bin, "./cmd/snapback")
	build.Dir = repoRoot(t)
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	out, err := exec.Command(bin, "version").Output()
	if err != nil {
		t.Fatalf("snapback version: %v", err)
	}
	want := "snapback 9.9.9 (commit abc1234, target " + target + ")\n"
	if string(out) != want {
		t.Errorf("snapback version = %q, want %q", out, want)
	}
}
