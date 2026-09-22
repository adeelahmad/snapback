package projectdocs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// webUIPage is the published page that documents the local web UI.
const webUIPage = "docs-site/web-ui.md"

// webUIRemoteWarning pins internal/web.remoteBindWarning verbatim, with the
// bind address left as %s. The page has to carry this line word for word, so
// an operator recognises what the server printed.
const webUIRemoteWarning = "warning: snapback web is reachable from other machines on %s; " +
	"the session token is the only protection"

// webUIDoc returns webUIPage, or "" when it is missing or unreadable. It
// reports the read failure instead of aborting, so every content assertion
// below still runs and names what the page has to say.
func webUIDoc(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), webUIPage))
	if err != nil {
		t.Errorf("read %s: %v", webUIPage, err)
		return ""
	}
	return string(data)
}

// securitySection returns the body of the first heading whose text contains
// "Security", or "" when the page has no such heading.
func securitySection(doc string) string {
	for _, line := range strings.Split(doc, "\n") {
		trimmed := strings.TrimRight(line, " \t")
		n := len(trimmed) - len(strings.TrimLeft(trimmed, "#"))
		if n == 0 || !strings.HasPrefix(trimmed[n:], " ") {
			continue
		}
		if strings.Contains(trimmed[n+1:], "Security") {
			return section(doc, trimmed)
		}
	}
	return ""
}

func TestWebUISecuritySectionExists(t *testing.T) {
	doc := webUIDoc(t)

	if strings.TrimSpace(securitySection(doc)) == "" {
		t.Errorf("%s has no heading containing %q with a body, want a section stating how the "+
			"web UI is protected", webUIPage, "Security")
	}
}

func TestWebUISecurityStatesTheLoopbackDefault(t *testing.T) {
	sec := securitySection(webUIDoc(t))

	if !strings.Contains(sec, "127.0.0.1") {
		t.Errorf("%s Security section does not name the loopback default %q",
			webUIPage, "127.0.0.1")
	}
}

func TestWebUISecurityStatesTheSessionToken(t *testing.T) {
	sec := securitySection(webUIDoc(t))

	if !lineWithAll(sec, "token") {
		t.Errorf("%s Security section has no sentence about the session token", webUIPage)
	}
}

func TestWebUISecurityStatesTheOriginRule(t *testing.T) {
	sec := securitySection(webUIDoc(t))

	if !strings.Contains(sec, "Origin") {
		t.Errorf("%s Security section does not state the %s rule", webUIPage, "Origin")
	}
}

func TestWebUISecurityNamesEveryBindFlag(t *testing.T) {
	sec := securitySection(webUIDoc(t))

	for _, flag := range []string{"--bind", "--allow-origin", "--allow-remote"} {
		if !strings.Contains(sec, flag) {
			t.Errorf("%s Security section does not mention the %s flag", webUIPage, flag)
		}
	}
}

func TestWebUISecurityNamesEveryConfigKey(t *testing.T) {
	sec := securitySection(webUIDoc(t))

	for _, key := range []string{"web.bind", "web.allowed_origins"} {
		if !strings.Contains(sec, key) {
			t.Errorf("%s Security section does not mention the %s config key", webUIPage, key)
		}
	}
}

func TestWebUISecurityQuotesTheRemoteBindWarning(t *testing.T) {
	sec := securitySection(webUIDoc(t))

	prefix, suffix, ok := strings.Cut(webUIRemoteWarning, "%s")
	if !ok {
		t.Fatalf("webUIRemoteWarning %q has no %%s placeholder", webUIRemoteWarning)
	}
	for _, line := range strings.Split(sec, "\n") {
		i := strings.Index(line, prefix)
		if i < 0 {
			continue
		}
		rest := line[i+len(prefix):]
		j := strings.Index(rest, suffix)
		if j > 0 {
			return
		}
	}
	t.Errorf("%s Security section has no single line carrying the remote-bind warning %q with a "+
		"bind address in place of the placeholder", webUIPage, webUIRemoteWarning)
}

// webUITransportRe names transport security Snapback does not have, so the page
// may only mention it to deny it.
var webUITransportRe = regexp.MustCompile(`(?i)\b(TLS|HTTPS)\b`)

func TestWebUISecurityClaimsNoTransportSecurity(t *testing.T) {
	doc := webUIDoc(t)

	for i, line := range strings.Split(doc, "\n") {
		matches := webUITransportRe.FindAllString(line, -1)
		if len(matches) == 0 || strings.Contains(strings.ToLower(line), "not") {
			continue
		}
		t.Errorf("%s:%d: names %v without denying it, want no claim of TLS or auth beyond the "+
			"session token; line = %q", webUIPage, i+1, matches, line)
	}
}
