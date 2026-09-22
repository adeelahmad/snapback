package ci_test

import (
	"regexp"
	"strings"
	"testing"
)

const siteJobName = "site"

var (
	setupNodeRe       = regexp.MustCompile(`uses:\s*['"]?actions/setup-node@v\d+['"]?\s*$`)
	nodeVersionFileRe = regexp.MustCompile(`node-version-file:\s*['"]?web/\.nvmrc['"]?\s*$`)
	npmCachePathRe    = regexp.MustCompile(`cache-dependency-path:\s*['"]?web/package-lock\.json['"]?\s*$`)
	floatingNodeRe    = regexp.MustCompile(`node-version:\s*['"]?(latest|lts/\*|current)['"]?`)
	workingDirWebRe   = regexp.MustCompile(`working-directory:\s*['"]?web['"]?\s*$`)
	npmInstallRe      = regexp.MustCompile(`\bnpm\s+(install|i|add)\b`)
)

// siteJobBlock returns the site job text, failing the test when the job is absent.
func siteJobBlock(t *testing.T) string {
	t.Helper()
	block := jobBlock(readCI(t), siteJobName)
	if strings.TrimSpace(block) == "" {
		t.Fatalf("ci.yml has no %q job", siteJobName)
	}
	return block
}

// jobSteps returns the text of each `- ` list item directly under the job's steps: key.
func jobSteps(block string) []string {
	lines := strings.Split(block, "\n")
	stepsAt := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "steps:" {
			stepsAt = i
			break
		}
	}
	if stepsAt < 0 {
		return nil
	}
	base := indentOf(lines[stepsAt])
	var steps []string
	var cur []string
	itemIndent := -1
	for _, line := range lines[stepsAt+1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && indentOf(line) <= base {
			break
		}
		if strings.HasPrefix(trimmed, "- ") && (itemIndent < 0 || indentOf(line) == itemIndent) {
			itemIndent = indentOf(line)
			if cur != nil {
				steps = append(steps, strings.Join(cur, "\n"))
			}
			cur = []string{line}
			continue
		}
		if cur != nil {
			cur = append(cur, line)
		}
	}
	if cur != nil {
		steps = append(steps, strings.Join(cur, "\n"))
	}
	return steps
}

// runCommands returns the trimmed shell lines of a step's run: key, single-line or block scalar.
func runCommands(step string) []string {
	lines := strings.Split(step, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- "))
		if !strings.HasPrefix(trimmed, "run:") {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "run:"))
		if rest != "" && rest != "|" && rest != ">" && rest != "|-" && rest != ">-" {
			return []string{strings.Trim(rest, `"'`)}
		}
		keyIndent := indentOf(line)
		var cmds []string
		for _, l := range lines[i+1:] {
			if strings.TrimSpace(l) == "" {
				continue
			}
			if indentOf(l) <= keyIndent {
				break
			}
			cmds = append(cmds, strings.TrimSpace(l))
		}
		return cmds
	}
	return nil
}

// stepIndex returns the index of the first step at or after from with a run line satisfying match, or -1.
func stepIndex(steps []string, from int, match func(string) bool) int {
	for i := from; i < len(steps); i++ {
		for _, cmd := range runCommands(steps[i]) {
			if match(cmd) {
				return i
			}
		}
	}
	return -1
}

func TestCISiteJobExists(t *testing.T) {
	text := readCI(t)
	for _, name := range []string{siteJobName, "test", crossCompileJob, fuseJobName} {
		if strings.TrimSpace(jobBlock(text, name)) == "" {
			t.Errorf("jobBlock(ci.yml, %q) = \"\", want a non-empty job", name)
		}
	}
}

func TestCISiteJobPinsNode(t *testing.T) {
	block := siteJobBlock(t)
	step := stepContaining(block, "actions/setup-node@")
	if step == "" {
		t.Fatalf("site job has no actions/setup-node step")
	}
	checks := []struct {
		name string
		re   *regexp.Regexp
	}{
		{"actions/setup-node@v<major>", setupNodeRe},
		{"node-version-file: web/.nvmrc", nodeVersionFileRe},
		{"cache-dependency-path: web/package-lock.json", npmCachePathRe},
	}
	for _, c := range checks {
		if !matchesAnyLine(c.re, step) {
			t.Errorf("setup-node step lacks %s; step:\n%s", c.name, step)
		}
	}
	if m := floatingNodeRe.FindString(block); m != "" {
		t.Errorf("site job pins a floating Node version %q, want node-version-file only", m)
	}
}

// matchesAnyLine reports whether re matches any single line of text.
func matchesAnyLine(re *regexp.Regexp, text string) bool {
	for _, line := range strings.Split(text, "\n") {
		if re.MatchString(strings.TrimRight(line, " \r")) {
			return true
		}
	}
	return false
}

func TestCISiteJobSteps(t *testing.T) {
	block := siteJobBlock(t)
	steps := jobSteps(block)
	if len(steps) == 0 {
		t.Fatalf("site job has no steps")
	}

	defaults := strings.Contains(block, "defaults:") && matchesAnyLine(workingDirWebRe, block)
	wantOrder := []string{"npm ci", "npm run lint", "npm test", "npm run build"}
	from := 0
	for _, want := range wantOrder {
		i := stepIndex(steps, from, func(cmd string) bool { return cmd == want })
		if i < 0 {
			t.Fatalf("site job has no %q step at or after step %d, want order %v", want, from, wantOrder)
		}
		if !defaults && !matchesAnyLine(workingDirWebRe, steps[i]) {
			t.Errorf("step %q has no working-directory: web and the job has no defaults.run.working-directory: web", want)
		}
		from = i + 1
	}

	grepIdx := stepIndex(steps, from, func(cmd string) bool {
		plain := strings.ReplaceAll(cmd, `\`, "")
		return strings.Contains(cmd, "grep") && strings.Contains(cmd, "dist") && strings.Contains(plain, "fonts.googleapis")
	})
	if grepIdx < 0 {
		t.Errorf("site job has no step after %q that greps dist for fonts.googleapis", wantOrder[len(wantOrder)-1])
	}
}

func TestCISiteJobNoInstallFallback(t *testing.T) {
	block := siteJobBlock(t)
	steps := jobSteps(block)
	if stepIndex(steps, 0, func(cmd string) bool { return cmd == "npm ci" }) < 0 {
		t.Fatalf("site job has no npm ci step, want the lockfile install")
	}
	for _, step := range steps {
		for _, cmd := range runCommands(step) {
			if npmInstallRe.MatchString(cmd) {
				t.Errorf("site job runs %q, want npm ci only", cmd)
			}
		}
	}
}
