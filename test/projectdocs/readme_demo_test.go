package projectdocs

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

const demoGIF = "docs-site/img/quickstart-restore.gif"

func TestReadmeReferencesDemoGIF(t *testing.T) {
	readme := readDoc(t, "README.md")
	if !bytes.Contains([]byte(readme), []byte(demoGIF)) {
		t.Errorf("README does not reference %s", demoGIF)
	}
}

func TestDemoGIFExistsAndIsAGIF(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot(t), demoGIF))
	if err != nil {
		t.Fatalf("read %s: %v", demoGIF, err)
	}
	if len(data) == 0 {
		t.Fatalf("%s is empty", demoGIF)
	}
	if !bytes.HasPrefix(data, []byte("GIF89a")) && !bytes.HasPrefix(data, []byte("GIF87a")) {
		t.Errorf("%s does not start with a GIF signature", demoGIF)
	}
}
