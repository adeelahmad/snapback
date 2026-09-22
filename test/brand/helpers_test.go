// Package brand pins the brand assets under docs/brand/ that the README,
// the docs site and the GitHub social preview rely on.
package brand

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const brandDir = "docs/brand"

var pngSignature = []byte("\x89PNG\r\n\x1a\n")

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no go.mod found above the test working directory")
		}
		dir = parent
	}
}

// readAsset returns the bytes of a repo-relative file, failing the test when
// it is missing or empty.
func readAsset(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	if len(data) == 0 {
		t.Fatalf("%s is empty", rel)
	}
	return data
}

// pngSize returns the width and height recorded in a PNG's IHDR chunk.
func pngSize(data []byte) (width, height uint32, err error) {
	if !bytes.HasPrefix(data, pngSignature) {
		return 0, 0, fmt.Errorf("missing PNG signature")
	}
	if len(data) < 24 || string(data[12:16]) != "IHDR" {
		return 0, 0, fmt.Errorf("first chunk is not IHDR")
	}
	return binary.BigEndian.Uint32(data[16:20]), binary.BigEndian.Uint32(data[20:24]), nil
}

func TestPNGSize(t *testing.T) {
	ihdr := append(append([]byte{}, pngSignature...), 0, 0, 0, 13, 'I', 'H', 'D', 'R', 0, 0, 5, 0, 0, 0, 2, 128)
	tests := []struct {
		name    string
		in      []byte
		wantW   uint32
		wantH   uint32
		wantErr bool
	}{
		{name: "valid", in: ihdr, wantW: 1280, wantH: 640},
		{name: "not png", in: []byte("<svg></svg>"), wantErr: true},
		{name: "truncated", in: pngSignature, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w, h, err := pngSize(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("pngSize(%s) error = %v, want error %v", tc.name, err, tc.wantErr)
			}
			if w != tc.wantW || h != tc.wantH {
				t.Errorf("pngSize(%s) = %dx%d, want %dx%d", tc.name, w, h, tc.wantW, tc.wantH)
			}
		})
	}
}
