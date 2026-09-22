package aliases

import (
	"os/exec"
	"strings"
	"testing"
)

const (
	pkgPath      = "github.com/adeelahmad/snapback/internal/aliases"
	providerPath = "github.com/adeelahmad/snapback/internal/provider"
)

func TestDepsStdlibAndProviderOnly(t *testing.T) {
	out, err := exec.Command("go", "list", "-f", `{{join .Imports "\n"}}`, pkgPath).CombinedOutput()
	if err != nil {
		t.Fatalf("go list %s: %v\n%s", pkgPath, err, out)
	}
	imports := strings.Fields(string(out))
	if len(imports) == 0 {
		t.Fatalf("go list %s returned no imports, want at least one", pkgPath)
	}
	banned := map[string]bool{"os": true, "os/exec": true, "io/fs": true, "path/filepath": true}
	for _, imp := range imports {
		if banned[imp] || strings.Contains(imp, "go-fuse") {
			t.Errorf("import %q is banned in %s", imp, pkgPath)
			continue
		}
		first, _, _ := strings.Cut(imp, "/")
		if strings.Contains(first, ".") && imp != providerPath {
			t.Errorf("import %q is neither standard library nor %s", imp, providerPath)
		}
	}
}
