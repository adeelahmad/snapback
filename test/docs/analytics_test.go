package docs

import (
	"strings"
	"testing"
)

const (
	gtagID          = "G-BFWW49ZP0E"
	analyticsNotice = "Google Analytics"
)

// analyticsSettings returns the key/value children of extra.analytics in mkdocs text.
func analyticsSettings(text string) map[string]string {
	block, _ := topLevelBlock(text, "extra")
	settings := map[string]string{}
	indent := -1
	for _, line := range block {
		trimmed := strings.TrimSpace(line)
		depth := len(line) - len(strings.TrimLeft(line, " "))
		if indent < 0 {
			if trimmed == "analytics:" {
				indent = depth
			}
			continue
		}
		if depth <= indent {
			break
		}
		if key, value, ok := strings.Cut(trimmed, ":"); ok {
			settings[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
		}
	}
	return settings
}

func TestMaterialGoogleAnalytics(t *testing.T) {
	settings := analyticsSettings(readRepoFile(t, mkdocsFile))

	want := map[string]string{
		"provider": "google",
		"property": gtagID,
	}
	for key, value := range want {
		if got := settings[key]; got != value {
			t.Errorf("%s extra.analytics.%s = %q, want %q", mkdocsFile, key, got, value)
		}
	}
}

func TestLandingMentionsGoogleAnalytics(t *testing.T) {
	text := readRepoFile(t, landingFile)

	if !strings.Contains(text, analyticsNotice) {
		t.Errorf("%s has no privacy note mentioning %q", landingFile, analyticsNotice)
	}
}
