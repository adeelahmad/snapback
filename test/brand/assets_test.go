package brand

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

const (
	socialCard       = brandDir + "/github-social-dark.png"
	socialCardWidth  = 1280
	socialCardHeight = 640
)

var brandAssets = []string{
	"mark.svg",
	"favicon.svg",
	"favicon-16.png",
	"favicon-32.png",
	"app-icon-180.png",
	"og-card-dark.png",
	"github-social-dark.png",
	"readme-banner-light.png",
	"readme-banner-dark.png",
}

func TestBrandAssetsExist(t *testing.T) {
	for _, name := range brandAssets {
		t.Run(name, func(t *testing.T) {
			rel := brandDir + "/" + name
			data := readAsset(t, rel)
			switch filepath.Ext(name) {
			case ".png":
				if !bytes.HasPrefix(data, pngSignature) {
					t.Errorf("%s does not start with the PNG signature", rel)
				}
			case ".svg":
				if !bytes.Contains(data, []byte("<svg")) {
					t.Errorf("%s has no <svg element", rel)
				}
			}
		})
	}
}

func TestGithubSocialCardIs1280x640(t *testing.T) {
	w, h, err := pngSize(readAsset(t, socialCard))
	if err != nil {
		t.Fatalf("pngSize(%s): %v", socialCard, err)
	}
	if w != socialCardWidth || h != socialCardHeight {
		t.Errorf("pngSize(%s) = %dx%d, want %dx%d", socialCard, w, h, socialCardWidth, socialCardHeight)
	}
}

func TestBrandSVGsHaveNoScript(t *testing.T) {
	svgs, err := filepath.Glob(filepath.Join(repoRoot(t), brandDir, "*.svg"))
	if err != nil {
		t.Fatalf("glob %s/*.svg: %v", brandDir, err)
	}
	if len(svgs) == 0 {
		t.Fatalf("%s has no SVG files, want at least mark.svg and favicon.svg", brandDir)
	}
	for _, path := range svgs {
		rel := brandDir + "/" + filepath.Base(path)
		if strings.Contains(strings.ToLower(string(readAsset(t, rel))), "<script") {
			t.Errorf("%s contains a <script element", rel)
		}
	}
}
