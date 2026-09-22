package site_test

import (
	"encoding/json"
	"encoding/xml"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

const (
	robotsTxt   = "web/public/robots.txt"
	sitemapXML  = "web/public/sitemap.xml"
	mkdocsYML   = "mkdocs.yml"
	sitemapURL  = "https://snapback.run/sitemap.xml"
	docsURL     = "https://snapback.run/docs/"
	sitemapNS   = "http://www.sitemaps.org/schemas/sitemap/0.9"
	lastmodForm = "2006-01-02"
)

var (
	scriptRE     = regexp.MustCompile(`(?is)<script\b([^>]*)>(.*?)</script>`)
	lastmodRE    = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	yamlScalarRE = regexp.MustCompile(`^([a-z_]+):\s*(.*)$`)
)

func TestRobotsTxt(t *testing.T) {
	robots := readRepoFile(t, robotsTxt)

	if !hasLine(robots, "User-agent: *") {
		t.Errorf("%s has no %q line, want every crawler addressed", robotsTxt, "User-agent: *")
	}
	for _, line := range strings.Split(robots, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if ok && strings.EqualFold(key, "Disallow") && strings.TrimSpace(value) != "" {
			t.Errorf("%s line %q disallows a path, want everything allowed", robotsTxt, strings.TrimSpace(line))
		}
	}
	if want := "Sitemap: " + sitemapURL; !hasLine(robots, want) {
		t.Errorf("%s has no %q line", robotsTxt, want)
	}
}

// urlset is the sitemaps.org <urlset> document.
type urlset struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []struct {
		Loc     string `xml:"loc"`
		Lastmod string `xml:"lastmod"`
	} `xml:"url"`
}

func TestSitemap(t *testing.T) {
	data := readRepoFile(t, sitemapXML)

	var set urlset
	if err := xml.Unmarshal([]byte(data), &set); err != nil {
		t.Fatalf("xml.Unmarshal(%s) = %v, want a valid <urlset>", sitemapXML, err)
	}
	if got := set.XMLName.Space; got != sitemapNS {
		t.Errorf("%s urlset namespace = %q, want %q", sitemapXML, got, sitemapNS)
	}

	var locs []string
	for _, u := range set.URLs {
		loc := strings.TrimSpace(u.Loc)
		locs = append(locs, loc)
		lastmod := strings.TrimSpace(u.Lastmod)
		if !lastmodRE.MatchString(lastmod) {
			t.Errorf("%s lastmod for %q = %q, want YYYY-MM-DD", sitemapXML, loc, lastmod)
			continue
		}
		if _, err := time.Parse(lastmodForm, lastmod); err != nil {
			t.Errorf("%s lastmod for %q = %q is not a real date: %v", sitemapXML, loc, lastmod, err)
		}
	}
	for _, want := range []string{siteOrigin, docsURL} {
		if !slices.Contains(locs, want) {
			t.Errorf("%s locs = %q, want %q listed", sitemapXML, locs, want)
		}
	}
}

// ldJSONBlocks returns the bodies of every <script type="application/ld+json">
// element in html.
func ldJSONBlocks(html string) []string {
	var blocks []string
	for _, m := range scriptRE.FindAllStringSubmatch(html, -1) {
		for _, a := range attrRE.FindAllStringSubmatch(m[1], -1) {
			if strings.EqualFold(a[1], "type") && strings.EqualFold(a[2], "application/ld+json") {
				blocks = append(blocks, m[2])
			}
		}
	}
	return blocks
}

// hasAnyKey reports whether v, or any object nested inside it, has one of keys.
func hasAnyKey(v any, keys []string) (string, bool) {
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			if slices.Contains(keys, k) {
				return k, true
			}
			if found, ok := hasAnyKey(child, keys); ok {
				return found, true
			}
		}
	case []any:
		for _, child := range x {
			if found, ok := hasAnyKey(child, keys); ok {
				return found, true
			}
		}
	}
	return "", false
}

func TestStructuredData(t *testing.T) {
	blocks := ldJSONBlocks(readRepoFile(t, indexHTML))
	if got, want := len(blocks), 1; got != want {
		t.Fatalf("%s application/ld+json script count = %d, want %d", indexHTML, got, want)
	}

	var ld map[string]any
	if err := json.Unmarshal([]byte(blocks[0]), &ld); err != nil {
		t.Fatalf("json.Unmarshal(ld+json) = %v, want a JSON object", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{"@type", "SoftwareApplication"},
		{"name", "snapback"},
		{"applicationCategory", "UtilitiesApplication"},
		{"url", siteOrigin},
	}
	for _, tc := range tests {
		if got, _ := ld[tc.key].(string); got != tc.want {
			t.Errorf("ld+json %s = %q, want %q", tc.key, got, tc.want)
		}
	}

	if got, _ := ld["operatingSystem"].(string); !strings.Contains(got, "Linux") {
		t.Errorf("ld+json operatingSystem = %q, want it to mention %q", got, "Linux")
	}
	if got, _ := ld["license"].(string); !strings.Contains(got, "MIT") {
		t.Errorf("ld+json license = %q, want MIT", got)
	}
	offers, _ := ld["offers"].(map[string]any)
	if got, _ := offers["price"].(string); got != "0" {
		t.Errorf("ld+json offers.price = %v, want %q", offers["price"], "0")
	}

	banned := []string{"aggregateRating", "review", "reviews"}
	if key, ok := hasAnyKey(ld, banned); ok {
		t.Errorf("ld+json has %q, want no rating or review claims", key)
	}
}

func TestIndexSEOMeta(t *testing.T) {
	tags := headTags(readRepoFile(t, indexHTML))

	if tg, ok := findTag(tags, "html", nil); !ok || tg.attrs["lang"] != "en" {
		t.Errorf("%s html lang = %q, want %q", indexHTML, tg.attrs["lang"], "en")
	}
	if got, want := metaContent(tags, "robots"), "index,follow"; got != want {
		t.Errorf("%s meta robots = %q, want %q", indexHTML, got, want)
	}
	if got := metaContent(tags, "description"); got == "" {
		t.Errorf("%s meta description is empty or missing, want it kept", indexHTML)
	}
}

// topLevelScalars returns the top-level "key: value" scalars of a YAML file,
// with surrounding quotes removed from each value.
func topLevelScalars(yaml string) map[string]string {
	scalars := map[string]string{}
	for _, line := range strings.Split(yaml, "\n") {
		m := yamlScalarRE.FindStringSubmatch(strings.TrimRight(line, "\r"))
		if m == nil {
			continue
		}
		value := strings.TrimSpace(m[2])
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			value = value[1 : len(value)-1]
		}
		scalars[m[1]] = value
	}
	return scalars
}

func TestDocsMeta(t *testing.T) {
	scalars := topLevelScalars(readRepoFile(t, mkdocsYML))
	if scalars["site_name"] == "" {
		t.Fatalf("%s has no site_name, want the docs config parsed", mkdocsYML)
	}

	desc := scalars["site_description"]
	if n := utf8.RuneCountInString(desc); n < minDescLen || n > maxDescLen {
		t.Errorf("%s site_description = %q (length %d), want %d-%d characters", mkdocsYML, desc, n, minDescLen, maxDescLen)
	}
	if got := scalars["site_author"]; got == "" {
		t.Errorf("%s site_author = %q, want non-empty", mkdocsYML, got)
	}
}
