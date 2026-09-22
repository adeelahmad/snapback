package projectdocs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// packagesPages are the two front doors that have to carry the same GitHub
// Packages install block, each with the link it uses for other doc pages.
var packagesPages = []struct {
	name        string // repo-relative page
	dockerLink  string // how this page links the container page
	otherPage   string // a doc link the page already makes, to pin the style
	linkComment string
}{
	{
		name:        "README.md",
		dockerLink:  "docs-site/docker.md",
		otherPage:   "docs-site/usage.md",
		linkComment: "README links doc pages by their repo-relative docs-site path",
	},
	{
		name:        "docs-site/index.md",
		dockerLink:  "docker.md",
		otherPage:   "usage.md",
		linkComment: "index.md links sibling pages by bare file name",
	},
}

// packagesLead matches the heading or bold lead that opens the install block.
var packagesLead = regexp.MustCompile(`(?i)^\s*(#{2,4}\s+|\*\*)[^\n]*GitHub Packages`)

// packagesBlock returns the lines of doc from the "GitHub Packages" heading or
// bold lead up to the next markdown heading, or "" when no such lead exists.
func packagesBlock(doc string) string {
	lines := strings.Split(doc, "\n")
	start := -1
	for i, line := range lines {
		if packagesLead.MatchString(line) {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	end := len(lines)
	for j := start + 1; j < len(lines); j++ {
		if strings.HasPrefix(lines[j], "#") {
			end = j
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}

func TestInstallPackagesBlockHasALead(t *testing.T) {
	for _, page := range packagesPages {
		t.Run(page.name, func(t *testing.T) {
			if packagesBlock(readDoc(t, page.name)) == "" {
				t.Errorf("%s has no heading or bold lead naming %q, so readers never "+
					"learn the image exists", page.name, "GitHub Packages")
			}
		})
	}
}

func TestInstallPackagesBlockShowsTheGhcrPullLine(t *testing.T) {
	const pull = "docker pull ghcr.io/adeelahmad/snapback"

	for _, page := range packagesPages {
		t.Run(page.name, func(t *testing.T) {
			block := packagesBlock(readDoc(t, page.name))
			if block == "" {
				t.Fatalf("%s has no GitHub Packages block to hold %q", page.name, pull)
			}
			if !strings.Contains(block, pull) {
				t.Errorf("%s GitHub Packages block %q has no %q line", page.name, block, pull)
			}
		})
	}
}

func TestInstallPackagesBlockLinksTheDockerPage(t *testing.T) {
	for _, page := range packagesPages {
		t.Run(page.name, func(t *testing.T) {
			doc := readDoc(t, page.name)
			if !strings.Contains(doc, page.otherPage) {
				t.Fatalf("%s no longer links %q, so %s no longer holds",
					page.name, page.otherPage, page.linkComment)
			}
			block := packagesBlock(doc)
			if block == "" {
				t.Fatalf("%s has no GitHub Packages block to link %q", page.name, page.dockerLink)
			}
			if !strings.Contains(block, "]("+page.dockerLink+")") {
				t.Errorf("%s GitHub Packages block %q has no link to %q (%s)",
					page.name, block, page.dockerLink, page.linkComment)
			}
		})
	}
}

func TestInstallPackagesBlockSaysTheImageShipsWithTheNextRelease(t *testing.T) {
	for _, page := range packagesPages {
		t.Run(page.name, func(t *testing.T) {
			block := packagesBlock(readDoc(t, page.name))
			if block == "" {
				t.Fatalf("%s has no GitHub Packages block to state when the image ships",
					page.name)
			}
			if !lineWithAll(block, "release") {
				t.Errorf("%s GitHub Packages block %q has no sentence saying the image "+
					"ships with the next tagged release, so it claims an image that is "+
					"not pushed yet", page.name, block)
			}
		})
	}
}

// foreignRegistries are the registries Snapback does not publish to, so no page
// may name them.
var foreignRegistries = []string{"docker.io/", "quay.io", "hub.docker.com"}

func TestInstallPackagesClaimsNoForeignRegistry(t *testing.T) {
	for _, page := range packagesPages {
		t.Run(page.name, func(t *testing.T) {
			doc := readDoc(t, page.name)
			for i, line := range strings.Split(doc, "\n") {
				for _, reg := range foreignRegistries {
					if strings.Contains(strings.ToLower(line), reg) {
						t.Errorf("%s:%d: names %q but the release workflow pushes to "+
							"ghcr.io only", page.name, i+1, reg)
					}
				}
			}
		})
	}
}

// packageManagerRe matches the package channels the repo does not build, whose
// goreleaser keys packagerKeys names.
var packageManagerRe = regexp.MustCompile(`(?i)(homebrew|brew install|brew tap|apt-get install|\.deb\b|\.rpm\b)`)

// packagerKeys are the goreleaser sections whose presence would make a
// Homebrew tap or a deb/rpm package real.
var packagerKeys = []string{"brews:", "nfpms:", "aurs:"}

func TestInstallPackagesClaimsNoPackageManagerWeDoNotBuild(t *testing.T) {
	cfg, err := os.ReadFile(filepath.Join(repoRoot(t), ".goreleaser.yaml"))
	if err != nil {
		t.Fatalf("read .goreleaser.yaml: %v", err)
	}
	for _, key := range packagerKeys {
		if strings.Contains(string(cfg), key) {
			t.Skipf(".goreleaser.yaml has %q, so the pages may document that channel", key)
		}
	}

	for _, page := range packagesPages {
		t.Run(page.name, func(t *testing.T) {
			doc := readDoc(t, page.name)
			for i, line := range strings.Split(doc, "\n") {
				for _, m := range packageManagerRe.FindAllString(line, -1) {
					t.Errorf("%s:%d: names %q but the release builds no such package",
						page.name, i+1, m)
				}
			}
		})
	}
}
