package site_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	bannedIcon = regexp.MustCompile(`(?i)\b(hat|cap|shield|lock|padlock|cloud|hard-?drive|harddrive|clock|refresh|rotate-?c?c?w|history-icon|camera|shutter|aperture)\b`)

	// Sources of names in site code: import specifiers, declared identifiers,
	// JSX component tags, className values and CSS class selectors.
	importSpec  = regexp.MustCompile(`(?:from|import)\s*\(?\s*['"]([^'"]+)['"]`)
	declared    = regexp.MustCompile(`\b(?:const|let|var|function|class|interface|type|enum)\s+([A-Za-z_$][\w$]*)`)
	jsxTag      = regexp.MustCompile(`<([A-Z][\w.]*)`)
	classAttr   = regexp.MustCompile(`class(?:Name)?\s*=\s*\{?\s*["'` + "`" + `]([^"'` + "`" + `]*)["'` + "`" + `]`)
	cssSelector = regexp.MustCompile(`\.([A-Za-z_-][\w-]*)`)
	camelHump   = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

// walkFiles returns the repo-relative paths of the regular files under rel.
func walkFiles(t *testing.T, rel string) []string {
	t.Helper()
	root := repoRoot(t)
	var files []string
	err := filepath.WalkDir(filepath.Join(root, rel), func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		r, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(r))
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", rel, err)
	}
	return files
}

// words splits camelCase names into dash-separated words so that \b sees
// the boundary in names like ClockIcon.
func words(name string) string {
	return camelHump.ReplaceAllString(name, "$1-$2")
}

// sourceNames returns every identifier, import specifier and class name in src.
func sourceNames(src string) []string {
	var names []string
	for _, re := range []*regexp.Regexp{importSpec, declared, jsxTag, classAttr} {
		for _, m := range re.FindAllStringSubmatch(src, -1) {
			names = append(names, m[1])
		}
	}
	return names
}

func TestNoBannedIconNames(t *testing.T) {
	var sources []string
	for _, rel := range walkFiles(t, "web/src") {
		switch filepath.Ext(rel) {
		case ".ts", ".tsx", ".css":
			sources = append(sources, rel)
		}
	}
	if len(sources) == 0 {
		t.Fatalf("web/src has no .ts, .tsx or .css files, want at least one")
	}
	public := walkFiles(t, "web/public")
	if len(public) == 0 {
		t.Fatalf("web/public has no files, want the self-hosted fonts and brand assets")
	}

	for _, rel := range append(append([]string{}, sources...), public...) {
		base := filepath.Base(rel)
		if bannedIcon.MatchString(words(base)) {
			t.Errorf("file name %s matches a banned icon name", rel)
		}
		if strings.Contains(strings.ToLower(rel), "lucide") {
			t.Errorf("file name %s mentions lucide", rel)
		}
	}

	for _, rel := range sources {
		src := readRepoFile(t, rel)
		if strings.Contains(strings.ToLower(src), "lucide") {
			t.Errorf("%s mentions lucide, want no icon library", rel)
		}
		names := sourceNames(src)
		if filepath.Ext(rel) == ".css" {
			for _, m := range cssSelector.FindAllStringSubmatch(src, -1) {
				names = append(names, m[1])
			}
		}
		for _, name := range names {
			if m := bannedIcon.FindString(words(name)); m != "" {
				t.Errorf("%s: name %q contains banned icon word %q", rel, name, m)
			}
		}
	}
}

func TestNoRuntimeCDNInSource(t *testing.T) {
	files := append([]string{"web/index.html"}, walkFiles(t, "web/src")...)
	banned := []string{"fonts.googleapis", "fonts.gstatic", "unpkg", "jsdelivr", "cdnjs", `<script src="http`}
	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if len(data) == 0 {
			t.Errorf("%s is empty, want content to scan", rel)
			continue
		}
		for _, b := range banned {
			if strings.Contains(string(data), b) {
				t.Errorf("%s contains %q, want no runtime CDN", rel, b)
			}
		}
	}
}
