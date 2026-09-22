package projectdocs

import (
	"strings"
	"testing"
)

// specAddendumBPointer is the SPEC.md pointer line for Addendum B; it mirrors
// specPointerLine, the Addendum A pointer that precedes it.
const specAddendumBPointer = "See SPEC-ADDENDUM-B.md for opt-in telemetry, crash reports and diagnostic bundles."

// addendumBProhibitions are quoted verbatim from Addendum B's "MUST NEVER"
// list: file names, paths, repository URIs, hostnames, usernames, snapshot IDs.
var addendumBProhibitions = []string{
	"File names",
	"file paths, directory names or path fragments",
	"Repository URIs, bucket names, host names or IP addresses of any backend",
	"The machine's hostname, domain",
	"usernames or home directory name",
	"Snapshot IDs, tree IDs, blob IDs or backup tags",
}

func TestSpecPointsAtAddendumB(t *testing.T) {
	spec := readDoc(t, "SPEC.md")
	head := strings.SplitN(spec, "\n", 4)
	var found bool
	for _, line := range head {
		if line == specAddendumBPointer {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("SPEC.md first lines %q carry no Addendum B pointer %q", head, specAddendumBPointer)
	}
	if !strings.Contains(spec, "SPEC-ADDENDUM-B.md") {
		t.Errorf("SPEC.md never names SPEC-ADDENDUM-B.md")
	}
}

func TestAddendumBForbidsIdentifyingFields(t *testing.T) {
	doc := readDoc(t, "SPEC-ADDENDUM-B.md")
	if !strings.Contains(doc, "MUST NEVER be sent") {
		t.Fatalf("SPEC-ADDENDUM-B.md has no %q list", "MUST NEVER be sent")
	}
	for _, want := range addendumBProhibitions {
		if !strings.Contains(doc, want) {
			t.Errorf("SPEC-ADDENDUM-B.md does not forbid %q verbatim", want)
		}
	}
	if !strings.Contains(doc, "off by default") {
		t.Errorf("SPEC-ADDENDUM-B.md does not state that telemetry is %q", "off by default")
	}
}
