package site_test

import (
	"regexp"
	"strings"
	"testing"
)

const (
	gtagID        = "G-BFWW49ZP0E"
	gtagOrigin    = "https://www.googletagmanager.com"
	gtagLoaderURL = gtagOrigin + "/gtag/js?id=" + gtagID
	gtagConfig    = "gtag('config', '" + gtagID + "')"
	gtagJS        = "gtag('js', new Date())"
)

var (
	scriptOpenRE    = regexp.MustCompile(`(?is)<script\b[^>]*>`)
	asyncAttrRE     = regexp.MustCompile(`(?i)\sasync(\s|>|=)`)
	measurementIDRE = regexp.MustCompile(`\b(?:G|UA|AW|GTM)-[A-Z0-9]{4,}\b`)
)

// scriptTag is one opening <script> tag: its raw text and quoted attributes.
type scriptTag struct {
	raw   string
	attrs map[string]string
}

// scriptTags returns every opening <script> tag in html in order.
func scriptTags(html string) []scriptTag {
	var tags []scriptTag
	for _, open := range scriptOpenRE.FindAllString(html, -1) {
		attrs := map[string]string{}
		for _, a := range attrRE.FindAllStringSubmatch(open, -1) {
			attrs[strings.ToLower(a[1])] = a[2]
		}
		tags = append(tags, scriptTag{raw: open, attrs: attrs})
	}
	return tags
}

func TestGtagLoader(t *testing.T) {
	html := readRepoFile(t, indexHTML)

	var loaders []scriptTag
	for _, tg := range scriptTags(html) {
		if tg.attrs["src"] == gtagLoaderURL {
			loaders = append(loaders, tg)
		}
	}
	if got, want := len(loaders), 1; got != want {
		t.Fatalf("%s script[src=%q] count = %d, want %d", indexHTML, gtagLoaderURL, got, want)
	}
	if !asyncAttrRE.MatchString(loaders[0].raw) {
		t.Errorf("%s gtag loader = %q, want an async attribute", indexHTML, loaders[0].raw)
	}
}

func TestGtagConfig(t *testing.T) {
	html := readRepoFile(t, indexHTML)

	for _, want := range []string{gtagJS, gtagConfig} {
		if got := strings.Count(html, want); got != 1 {
			t.Errorf("%s occurrences of %q = %d, want 1", indexHTML, want, got)
		}
	}
	if !strings.Contains(html, "window.dataLayer = window.dataLayer || []") {
		t.Errorf("%s has no dataLayer initialisation, want the verbatim gtag snippet", indexHTML)
	}
}

func TestNoOtherTrackers(t *testing.T) {
	html := readRepoFile(t, indexHTML)
	lower := strings.ToLower(html)

	banned := []string{
		"google-analytics.com",
		"googletagmanager.com/gtm.js",
		"analytics.js",
		"plausible",
		"fathom",
		"umami",
		"matomo",
		"piwik",
		"hotjar",
		"clarity.ms",
		"segment.com",
		"connect.facebook.net",
		"posthog",
		"mixpanel",
	}
	for _, b := range banned {
		if strings.Contains(lower, b) {
			t.Errorf("%s contains %q, want no tracker other than gtag %s", indexHTML, b, gtagID)
		}
	}
	for _, id := range measurementIDRE.FindAllString(html, -1) {
		if id != gtagID {
			t.Errorf("%s contains measurement ID %q, want only %q", indexHTML, id, gtagID)
		}
	}
}
