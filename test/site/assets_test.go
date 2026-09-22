package site_test

import (
	"bytes"
	"encoding/xml"
	"errors"
	"image/png"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

// fontTokens is the subset of web/tokens.json that lists the font files.
type fontTokens struct {
	Type struct {
		Fonts []struct {
			Family string `json:"family"`
			File   string `json:"file"`
		} `json:"fonts"`
	} `json:"type"`
}

// readRepoBytes returns the raw bytes of rel (relative to the repo root).
func readRepoBytes(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return data
}

func TestFontsSelfHosted(t *testing.T) {
	var tokens fontTokens
	loadJSON(t, "web/tokens.json", &tokens)
	if len(tokens.Type.Fonts) == 0 {
		t.Fatalf("web/tokens.json type.fonts is empty, want at least one font")
	}
	for _, font := range tokens.Type.Fonts {
		rel := "web/public/fonts/" + path.Base(font.File)
		t.Run(rel, func(t *testing.T) {
			data := readRepoBytes(t, rel)
			if got, want := string(data[:min(4, len(data))]), "wOF2"; got != want {
				t.Errorf("%s magic = %q, want %q", rel, got, want)
			}
		})
	}
}

func TestBrandSVGs(t *testing.T) {
	files := []string{
		"web/public/brand/mark.svg",
		"web/public/brand/mark-inverse.svg",
		"web/public/brand/lockup-horizontal.svg",
		"web/public/brand/lockup-horizontal-inverse.svg",
		"web/public/brand/wordmark.svg",
		"web/public/brand/wordmark-inverse.svg",
		"web/public/favicon.svg",
	}
	for _, rel := range files {
		t.Run(rel, func(t *testing.T) {
			data := readRepoBytes(t, rel)
			root, err := svgRoot(data)
			if err != nil {
				t.Fatalf("parse %s: %v", rel, err)
			}
			if root != "svg" {
				t.Errorf("%s root element = %q, want %q", rel, root, "svg")
			}
			for _, banned := range []string{"<linearGradient", "<radialGradient"} {
				if bytes.Contains(data, []byte(banned)) {
					t.Errorf("%s contains %s, want a flat mark", rel, banned)
				}
			}
		})
	}
}

// svgRoot parses the whole document and returns the local name of its root element.
func svgRoot(data []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	root := ""
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		if start, ok := tok.(xml.StartElement); ok && root == "" {
			root = start.Name.Local
		}
	}
	if root == "" {
		return "", errors.New("no root element")
	}
	return root, nil
}

func TestBrandPNGs(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
	}{
		{"favicon-16.png", 16, 16},
		{"favicon-32.png", 32, 32},
		{"app-icon-180.png", 180, 180},
		{"app-icon-512.png", 512, 512},
	}
	for _, tt := range tests {
		rel := "web/public/" + tt.name
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := png.DecodeConfig(bytes.NewReader(readRepoBytes(t, rel)))
			if err != nil {
				t.Fatalf("png.DecodeConfig(%s): %v", rel, err)
			}
			if cfg.Width != tt.width || cfg.Height != tt.height {
				t.Errorf("%s size = %dx%d, want %dx%d", rel, cfg.Width, cfg.Height, tt.width, tt.height)
			}
		})
	}
}

func TestOGCardsAreWide(t *testing.T) {
	const (
		minWidth  = 1200
		wantRatio = 1.91
		tolerance = 0.02
	)
	for _, name := range []string{"og-card-light.png", "og-card-dark.png"} {
		rel := "web/public/" + name
		t.Run(name, func(t *testing.T) {
			cfg, err := png.DecodeConfig(bytes.NewReader(readRepoBytes(t, rel)))
			if err != nil {
				t.Fatalf("png.DecodeConfig(%s): %v", rel, err)
			}
			if cfg.Width < minWidth {
				t.Errorf("%s width = %d, want >= %d", rel, cfg.Width, minWidth)
			}
			if cfg.Height == 0 {
				t.Fatalf("%s height = 0, want > 0", rel)
			}
			ratio := float64(cfg.Width) / float64(cfg.Height)
			if math.Abs(ratio-wantRatio) > tolerance {
				t.Errorf("%s aspect ratio = %.3f, want %.2f +/- %.2f", rel, ratio, wantRatio, tolerance)
			}
		})
	}
}

func TestNoticeListsFonts(t *testing.T) {
	notice := readRepoFile(t, "NOTICE")
	for _, want := range []string{"IBM Plex Sans", "IBM Plex Mono", "JetBrains Mono", "SIL Open Font License"} {
		if !strings.Contains(notice, want) {
			t.Errorf("NOTICE does not mention %q", want)
		}
	}
}
