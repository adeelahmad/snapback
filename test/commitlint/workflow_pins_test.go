package commitlint

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

var exactSemver = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

func TestWorkflowPinsNodeVersion(t *testing.T) {
	text := readRepoFile(t, workflowPath)

	matches := regexp.MustCompile(`(?m)node-version:\s*"?(\S+?)"?\s*$`).FindAllStringSubmatch(text, -1)

	if len(matches) != 1 {
		t.Fatalf("found %d node-version entries, want exactly 1", len(matches))
	}
	if v := matches[0][1]; !exactSemver.MatchString(v) {
		t.Fatalf("node-version = %q, want an exact X.Y.Z", v)
	}
}

func TestWorkflowPinsCommitlintPackages(t *testing.T) {
	text := readRepoFile(t, workflowPath)

	majors := map[string]string{}
	for _, pkg := range []string{"@commitlint/cli", "@commitlint/config-conventional"} {
		m := regexp.MustCompile(regexp.QuoteMeta(pkg) + `@(\S+)`).FindStringSubmatch(text)
		if m == nil {
			t.Errorf("workflow does not install %s@<version>", pkg)
			continue
		}
		v := strings.Trim(m[1], `"'`)
		if !exactSemver.MatchString(v) {
			t.Errorf("%s@%s is not an exact X.Y.Z (no ^, ~, latest)", pkg, v)
			continue
		}
		majors[pkg] = strings.SplitN(v, ".", 2)[0]
	}

	if len(majors) == 2 && majors["@commitlint/cli"] != majors["@commitlint/config-conventional"] {
		t.Errorf("commitlint package majors differ: %v", majors)
	}
}

func TestWorkflowPinsActions(t *testing.T) {
	text := readRepoFile(t, workflowPath)

	uses := regexp.MustCompile(`(?m)^\s*(?:-\s+)?uses:\s*"?([^"\s#]+)"?`).FindAllStringSubmatch(text, -1)
	pinned := regexp.MustCompile(`@(v\d+(\.\d+)*|[0-9a-f]{40})$`)

	if len(uses) == 0 {
		t.Fatal("no uses: entries found")
	}
	for _, m := range uses {
		if !pinned.MatchString(m[1]) {
			t.Errorf("uses: %s is not pinned to a version tag or 40-hex SHA", m[1])
		}
	}
}

func TestWorkflowLeastPrivilegePermissions(t *testing.T) {
	text := readRepoFile(t, workflowPath)

	keys := blockChildKeys(text, "permissions")
	block, _ := subBlock(text, "permissions:")
	m := regexp.MustCompile(`(?m)^\s+contents:\s*(\S+)\s*$`).FindStringSubmatch(block)

	if !slices.Equal(keys, []string{"contents"}) {
		t.Fatalf("permissions: children = %q, want [contents]", keys)
	}
	if m == nil || m[1] != "read" {
		t.Fatalf("permissions.contents = %v, want read", m)
	}
}
