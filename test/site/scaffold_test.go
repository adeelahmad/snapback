package site_test

import (
	"maps"
	"regexp"
	"slices"
	"strings"
	"testing"
)

var exactVersion = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

const wantNode = "26.0.0"

func TestPackageJSONPinsExactVersions(t *testing.T) {
	pkg := loadPackageJSON(t)
	for _, section := range []struct {
		name string
		deps map[string]string
	}{
		{"dependencies", pkg.Dependencies},
		{"devDependencies", pkg.DevDependencies},
	} {
		if len(section.deps) == 0 {
			t.Errorf("web/package.json %s is empty, want pinned entries", section.name)
			continue
		}
		for name, version := range section.deps {
			if !exactVersion.MatchString(version) {
				t.Errorf("web/package.json %s[%q] = %q, want exact x.y.z", section.name, name, version)
			}
		}
	}
}

func TestPackageJSONAllowedDependencies(t *testing.T) {
	pkg := loadPackageJSON(t)

	got := slices.Sorted(maps.Keys(pkg.Dependencies))
	want := []string{"react", "react-dom"}
	if !slices.Equal(got, want) {
		t.Errorf("web/package.json dependencies = %v, want %v", got, want)
	}

	allowedDev := map[string]bool{
		"vite":                 true,
		"@vitejs/plugin-react": true,
		"typescript":           true,
		"vitest":               true,
		"@types/react":         true,
		"@types/react-dom":     true,
		"@types/node":          true,
	}
	if len(pkg.DevDependencies) == 0 {
		t.Errorf("web/package.json devDependencies is empty, want the Vite toolchain")
	}
	for name := range pkg.DevDependencies {
		if !allowedDev[name] {
			t.Errorf("web/package.json devDependencies has %q, want only %v", name, slices.Sorted(maps.Keys(allowedDev)))
		}
	}

	for _, banned := range []string{"tailwindcss", "next", "lucide-react"} {
		if _, ok := pkg.Dependencies[banned]; ok {
			t.Errorf("web/package.json dependencies has banned %q", banned)
		}
		if _, ok := pkg.DevDependencies[banned]; ok {
			t.Errorf("web/package.json devDependencies has banned %q", banned)
		}
	}
}

func TestPackageJSONScripts(t *testing.T) {
	pkg := loadPackageJSON(t)
	for _, tc := range []struct {
		script string
		want   string
	}{
		{"build", "vite build"},
		{"test", "vitest run"},
		{"lint", "tsc --noEmit"},
	} {
		if got := pkg.Scripts[tc.script]; got != tc.want {
			t.Errorf("web/package.json scripts.%s = %q, want %q", tc.script, got, tc.want)
		}
	}
	if got := pkg.Scripts["prebuild"]; !strings.Contains(got, "gen:tokens") {
		t.Errorf("web/package.json scripts.prebuild = %q, want it to contain %q", got, "gen:tokens")
	}
}

func TestNodePinned(t *testing.T) {
	nvmrc := strings.TrimSpace(readRepoFile(t, "web/.nvmrc"))
	if !exactVersion.MatchString(nvmrc) {
		t.Errorf("web/.nvmrc = %q, want exact x.y.z", nvmrc)
	}
	if nvmrc != wantNode {
		t.Errorf("web/.nvmrc = %q, want %q", nvmrc, wantNode)
	}

	pkg := loadPackageJSON(t)
	if got := pkg.Engines["node"]; got != nvmrc {
		t.Errorf("web/package.json engines.node = %q, want %q (web/.nvmrc)", got, nvmrc)
	}

	npmrc := readRepoFile(t, "web/.npmrc")
	for _, want := range []string{"save-exact=true", "engine-strict=true"} {
		if !hasLine(npmrc, want) {
			t.Errorf("web/.npmrc missing line %q", want)
		}
	}
}

func TestLockfileCommittedAndConsistent(t *testing.T) {
	var lock struct {
		LockfileVersion int `json:"lockfileVersion"`
		Packages        map[string]struct {
			Dependencies map[string]string `json:"dependencies"`
		} `json:"packages"`
	}
	loadJSON(t, "web/package-lock.json", &lock)

	if lock.LockfileVersion < 3 {
		t.Errorf("web/package-lock.json lockfileVersion = %d, want >= 3", lock.LockfileVersion)
	}

	pkg := loadPackageJSON(t)
	root, ok := lock.Packages[""]
	if !ok {
		t.Fatalf(`web/package-lock.json has no packages[""] entry`)
	}
	if len(pkg.Dependencies) == 0 {
		t.Fatalf("web/package.json dependencies is empty, want react and react-dom")
	}
	if !maps.Equal(root.Dependencies, pkg.Dependencies) {
		t.Errorf(`web/package-lock.json packages[""].dependencies = %v, want %v (web/package.json)`, root.Dependencies, pkg.Dependencies)
	}
}

func TestGitignoreCoversSiteOutputs(t *testing.T) {
	gitignore := readRepoFile(t, ".gitignore")
	for _, want := range []string{"web/node_modules/", "_site/", "dist/"} {
		if !hasLine(gitignore, want) {
			t.Errorf(".gitignore missing line %q", want)
		}
	}
}

func TestNoticeListsRuntimeDeps(t *testing.T) {
	notice := readRepoFile(t, "NOTICE")
	for _, dep := range []string{"react", "react-dom"} {
		// The name must stand alone so "react-dom" cannot satisfy "react".
		re := regexp.MustCompile(`(?m)^.*(?:^|[^\w-])` + regexp.QuoteMeta(dep) + `(?:[^\w-]|$).*\bMIT\b`)
		if !re.MatchString(notice) {
			t.Errorf("NOTICE has no line naming %q with MIT", dep)
		}
	}
}
