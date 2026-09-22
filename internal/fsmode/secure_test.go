package fsmode

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecureConstantsAreFixed(t *testing.T) {
	for _, tc := range []struct {
		name string
		got  fs.FileMode
		want fs.FileMode
	}{
		{name: "SecureFile", got: SecureFile, want: 0o600},
		{name: "SecureDir", got: SecureDir, want: 0o700},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("%s = %04o, want %04o", tc.name, tc.got, tc.want)
			}
		})
	}
}

// TestSecureIgnoresConfiguredModes pins the ruling that Secure is a constant, not
// a setting: no files.* value can widen it.
func TestSecureIgnoresConfiguredModes(t *testing.T) {
	want := Modes{Dir: 0o700, File: 0o600}
	for _, configured := range []Modes{
		{},
		{Dir: 0o750, File: 0o640},
		{Dir: 0o755, File: 0o644},
		{Dir: 0o777, File: 0o666},
	} {
		t.Run(fmt.Sprintf("dir_%04o_file_%04o", configured.Dir, configured.File), func(t *testing.T) {
			if got := Secure(); got != want {
				t.Errorf("Secure() with %s configured = %s, want %s",
					formatModes(configured.OrDefault()), formatModes(got), formatModes(want))
			}
			if got := Secure().OrDefault(); got != want {
				t.Errorf("Secure().OrDefault() = %s, want %s", formatModes(got), formatModes(want))
			}
		})
	}
}

// secureMode is one mode a call site must use: either by naming the fsmode
// constant or by spelling the same octal literal.
type secureMode struct {
	ident string
	mode  fs.FileMode
}

// TestSecureCallSitesUseFixedModes reads the source of every call site that the
// ruling exempts from files.* and asserts it creates its state with the secure
// modes, not with a configured one.
func TestSecureCallSitesUseFixedModes(t *testing.T) {
	file := secureMode{ident: "SecureFile", mode: SecureFile}
	dir := secureMode{ident: "SecureDir", mode: SecureDir}

	for _, tc := range []struct {
		name  string
		file  string
		fn    string
		modes []secureMode
	}{
		{name: "credential store", file: "internal/web/credential.go", fn: "storeCredential", modes: []secureMode{dir, file}},
		{name: "password file", file: "internal/compat/resticfx/password.go", fn: "NewPasswordFile", modes: []secureMode{file}},
		{name: "daemon lock and pid", file: "internal/daemon/lock.go", fn: "lockWith", modes: []secureMode{file}},
		{name: "ipc socket dir", file: "internal/ipc/server.go", fn: "Listen", modes: []secureMode{dir, file}},
		{name: "saved config dir", file: "internal/config/store.go", fn: "Save", modes: []secureMode{dir, file}},
		{name: "saved config write", file: "internal/config/store.go", fn: "writeSync", modes: []secureMode{file}},
		{name: "setup save", file: "internal/setup/save.go", fn: "Save", modes: []secureMode{file}},
		{name: "doctor bundle", file: "internal/doctor/bundle.go", fn: "WriteBundle", modes: []secureMode{file}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := funcBody(t, filepath.Join("..", "..", filepath.FromSlash(tc.file)), tc.fn)
			for _, m := range tc.modes {
				literal := fmt.Sprintf("0o%o", uint32(m.mode))
				if strings.Contains(body, m.ident) || strings.Contains(body, literal) {
					continue
				}
				t.Errorf("%s: %s does not create its state with fsmode.%s: want %s or the literal %s",
					tc.file, tc.fn, m.ident, m.ident, literal)
			}
		})
	}
}

// funcBody returns the printed body of the named top-level function in path.
func funcBody(t *testing.T, path, name string) string {
	t.Helper()
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	for _, decl := range parsed.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != name || fn.Body == nil {
			continue
		}
		var b strings.Builder
		if err := printer.Fprint(&b, fset, fn.Body); err != nil {
			t.Fatalf("print %s.%s: %v", path, name, err)
		}
		return b.String()
	}
	t.Fatalf("%s: no func %s", path, name)
	return ""
}

func formatModes(m Modes) string {
	return fmt.Sprintf("{Dir:%04o File:%04o}", m.Dir, m.File)
}
