package community

import (
	"strings"
	"testing"
)

const (
	conductPath          = "CODE_OF_CONDUCT.md"
	headingPledge        = "## Our Pledge"
	headingEnforcement   = "## Enforcement"
	headingAttribution   = "## Attribution"
	covenantName         = "Contributor Covenant"
	covenantVersion      = "version 2.1"
	conductRepoContact   = "github.com/adeelahmad/snapback"
	conductHandleContact = "@adeelahmad"
	upstreamPlaceholder  = "[INSERT CONTACT METHOD]"
)

func TestConductIsContributorCovenant21(t *testing.T) {
	body := readOwned(t, conductPath)

	got := map[string]bool{}
	for _, line := range strings.Split(body, "\n") {
		got[strings.TrimRight(line, " \t\r")] = true
	}
	for _, want := range []string{headingPledge, headingEnforcement, headingAttribution} {
		if !got[want] {
			t.Errorf("%s: missing heading %q", conductPath, want)
		}
	}
	for _, want := range []string{covenantName, covenantVersion} {
		if !strings.Contains(body, want) {
			t.Errorf("%s: missing text %q", conductPath, want)
		}
	}
}

func TestConductHasGitHubContactMethod(t *testing.T) {
	// The existing section helpers are bound to their own files, so cut the
	// Enforcement section directly instead of adding a third line-loop copy.
	_, section, found := strings.Cut(readOwned(t, conductPath), "\n"+headingEnforcement+"\n")
	if !found {
		t.Fatalf("%s: heading %q not found", conductPath, headingEnforcement)
	}
	section, _, _ = strings.Cut(section, "\n## ")

	if !strings.Contains(section, conductRepoContact) && !strings.Contains(section, conductHandleContact) {
		t.Errorf("%s %q section: missing GitHub contact (%q or %q)", conductPath, headingEnforcement, conductRepoContact, conductHandleContact)
	}
	if strings.Contains(section, upstreamPlaceholder) {
		t.Errorf("%s %q section: upstream placeholder %q still present", conductPath, headingEnforcement, upstreamPlaceholder)
	}
}
