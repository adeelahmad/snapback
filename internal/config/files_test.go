package config

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/fsmode"
)

// filesErrors returns the field errors Validate reports on the given files.*
// path.
func filesErrors(t *testing.T, err error, path string) []FieldError {
	t.Helper()
	if err == nil {
		return nil
	}
	var out []FieldError
	for _, f := range validationFields(t, err) {
		if f.Path == path {
			out = append(out, f)
		}
	}
	return out
}

// TestFilesRoundTrip pins that an explicit files section survives
// Parse -> Marshal -> Parse with both modes intact.
func TestFilesRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))

	data := minimalYAML(tmp, "files:\n  dir_mode: \"0750\"\n  file_mode: \"0640\"\n", "", "")

	c, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(files) = %v, want nil error", err)
	}
	want := Files{DirMode: "0750", FileMode: "0640"}
	if got := c.Files; !reflect.DeepEqual(got, want) {
		t.Errorf("Parse(files).Files = %#v, want %#v", got, want)
	}

	out, err := Marshal(c)
	if err != nil {
		t.Fatalf("Marshal(c) = %v, want nil error", err)
	}
	for _, line := range []string{"files:", "dir_mode:", "file_mode:"} {
		if !bytes.Contains(out, []byte(line)) {
			t.Errorf("Marshal(c) = %s, want it to carry %q", out, line)
		}
	}
	back, err := Parse(out)
	if err != nil {
		t.Fatalf("Parse(Marshal(c)) = %v, want nil error", err)
	}
	if got := back.Files; !reflect.DeepEqual(got, want) {
		t.Errorf("Parse(Marshal(c)).Files = %#v, want %#v", got, want)
	}
}

// TestFilesUmaskRoundTrip pins that the umask spelling survives the same round
// trip and is not rewritten into the two explicit keys on the way out.
func TestFilesUmaskRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))

	c, err := Parse(minimalYAML(tmp, "files:\n  umask: \"027\"\n", "", ""))
	if err != nil {
		t.Fatalf("Parse(files.umask) = %v, want nil error", err)
	}
	want := Files{Umask: "027"}
	if got := c.Files; !reflect.DeepEqual(got, want) {
		t.Errorf("Parse(files.umask).Files = %#v, want %#v", got, want)
	}

	out, err := Marshal(c)
	if err != nil {
		t.Fatalf("Marshal(c) = %v, want nil error", err)
	}
	if bytes.Contains(out, []byte("dir_mode")) || bytes.Contains(out, []byte("file_mode")) {
		t.Errorf("Marshal(c) = %s, want only the umask key the user wrote", out)
	}
}

// TestFilesAbsentSectionIsZero pins that a config without a files section stays
// valid, yields the zero Files and marshals back without the key, so the
// example golden and TestMarshalRoundTrip keep passing.
func TestFilesAbsentSectionIsZero(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))

	c, err := Parse(minimalYAML(tmp, "", "", ""))
	if err != nil {
		t.Fatalf("Parse(no files) = %v, want nil error", err)
	}
	if got := c.Files; got != (Files{}) {
		t.Errorf("Parse(no files).Files = %#v, want the zero Files", got)
	}

	out, err := Marshal(c)
	if err != nil {
		t.Fatalf("Marshal(c) = %v, want nil error", err)
	}
	if bytes.Contains(out, []byte("files")) {
		t.Errorf("Marshal(c) = %s, want no files key", out)
	}
}

// TestFilesModes pins the resolver: explicit modes win, the umask spelling
// yields the same pair, and an absent section resolves to the fsmode defaults.
func TestFilesModes(t *testing.T) {
	tests := []struct {
		name  string
		files Files
		want  fsmode.Modes
	}{
		{"absent", Files{}, fsmode.Modes{}.OrDefault()},
		{"explicit", Files{DirMode: "0750", FileMode: "0640"}, fsmode.Modes{Dir: 0o750, File: 0o640}},
		{"umask", Files{Umask: "027"}, fsmode.Modes{Dir: 0o750, File: 0o640}},
		{"umask 022", Files{Umask: "022"}, fsmode.Modes{Dir: 0o755, File: 0o644}},
		{"dir only", Files{DirMode: "0750"}, fsmode.Modes{Dir: 0o750, File: 0o600}},
		{"file only", Files{FileMode: "0640"}, fsmode.Modes{Dir: 0o700, File: 0o640}},
		{"bare octal", Files{DirMode: "750", FileMode: "640"}, fsmode.Modes{Dir: 0o750, File: 0o640}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.files.Modes()
			if err != nil {
				t.Fatalf("Files%#v.Modes() = %v, want nil error", tt.files, err)
			}
			if got != tt.want {
				t.Errorf("Files%#v.Modes() = {%04o %04o}, want {%04o %04o}",
					tt.files, got.Dir, got.File, tt.want.Dir, tt.want.File)
			}
		})
	}
}

// TestFilesModesDefaultsAre0700And0600 spells out the default pair the rest of
// Snapback depends on, so a change to it cannot slip through unnoticed.
func TestFilesModesDefaultsAre0700And0600(t *testing.T) {
	got, err := (Files{}).Modes()
	if err != nil {
		t.Fatalf("Files{}.Modes() = %v, want nil error", err)
	}
	if got.Dir != fs.FileMode(0o700) || got.File != fs.FileMode(0o600) {
		t.Errorf("Files{}.Modes() = {%04o %04o}, want {0700 0600}", got.Dir, got.File)
	}
}

// TestFilesModesRejectsBadValues pins that the resolver refuses what the
// validator refuses, so a caller that skips Validate cannot get silent zeroes.
func TestFilesModesRejectsBadValues(t *testing.T) {
	tests := []struct {
		name  string
		files Files
		value string
	}{
		{"bad dir mode", Files{DirMode: "nine"}, "nine"},
		{"bad file mode", Files{FileMode: "0999"}, "0999"},
		{"world writable dir", Files{DirMode: "0757"}, "0757"},
		{"bad umask", Files{Umask: "zzz"}, "zzz"},
		{"umask with explicit mode", Files{Umask: "027", DirMode: "0750"}, "umask"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.files.Modes()
			if err == nil {
				t.Fatalf("Files%#v.Modes() = {%04o %04o}, nil, want an error", tt.files, got.Dir, got.File)
			}
			if !strings.Contains(err.Error(), tt.value) {
				t.Errorf("Files%#v.Modes() error = %q, want it to name %q", tt.files, err, tt.value)
			}
		})
	}
}

// TestFilesUmaskWithExplicitModeIsAValidationError pins that setting both
// spellings is a field error on files.umask whose message says the explicit
// modes win and that the user must pick one spelling.
func TestFilesUmaskWithExplicitModeIsAValidationError(t *testing.T) {
	tests := []struct {
		name  string
		files Files
	}{
		{"umask and dir mode", Files{Umask: "027", DirMode: "0750"}},
		{"umask and file mode", Files{Umask: "027", FileMode: "0640"}},
		{"umask and both", Files{Umask: "027", DirMode: "0750", FileMode: "0640"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			c.Files = tt.files

			got := filesErrors(t, Validate(c), "files.umask")
			if len(got) == 0 {
				t.Fatalf("Validate(files=%#v) = nil, want a field error on files.umask", tt.files)
			}
			msg := got[0].Msg
			if !strings.Contains(msg, "explicit") || !strings.Contains(msg, "one") {
				t.Errorf("Validate(files=%#v) files.umask msg = %q, want it to say the explicit modes win and to pick one spelling", tt.files, msg)
			}
		})
	}
}

// TestFilesBadModeIsAValidationError pins that an unparsable value is a field
// error on the key that carried it, and that the message names the value.
func TestFilesBadModeIsAValidationError(t *testing.T) {
	tests := []struct {
		name  string
		files Files
		path  string
		value string
	}{
		{"non octal dir mode", Files{DirMode: "nine"}, "files.dir_mode", "nine"},
		{"out of range file mode", Files{FileMode: "0999"}, "files.file_mode", "0999"},
		{"world writable dir mode", Files{DirMode: "0757"}, "files.dir_mode", "0757"},
		{"setuid file mode", Files{FileMode: "4640"}, "files.file_mode", "4640"},
		{"owner cannot read dir", Files{DirMode: "0070"}, "files.dir_mode", "0070"},
		{"non octal umask", Files{Umask: "zzz"}, "files.umask", "zzz"},
		{"out of range umask", Files{Umask: "7777"}, "files.umask", "7777"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			c.Files = tt.files

			got := filesErrors(t, Validate(c), tt.path)
			if len(got) == 0 {
				t.Fatalf("Validate(files=%#v) = nil, want a field error on %s", tt.files, tt.path)
			}
			if !strings.Contains(got[0].Msg, tt.value) {
				t.Errorf("Validate(files=%#v) %s msg = %q, want it to name %q", tt.files, tt.path, got[0].Msg, tt.value)
			}
		})
	}
}

// TestFilesValidSectionsPass pins that every spelling the section accepts is
// free of files.* errors, so the validator cannot reject its own defaults.
func TestFilesValidSectionsPass(t *testing.T) {
	tests := []struct {
		name  string
		files Files
	}{
		{"absent", Files{}},
		{"explicit pair", Files{DirMode: "0750", FileMode: "0640"}},
		{"bare octal", Files{DirMode: "750", FileMode: "640"}},
		{"0o prefixed", Files{DirMode: "0o750", FileMode: "0o640"}},
		{"dir only", Files{DirMode: "0700"}},
		{"file only", Files{FileMode: "0600"}},
		{"umask only", Files{Umask: "027"}},
		{"umask zero", Files{Umask: "0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			c.Files = tt.files

			err := Validate(c)
			for _, path := range []string{"files.dir_mode", "files.file_mode", "files.umask"} {
				if got := filesErrors(t, err, path); len(got) != 0 {
					t.Errorf("Validate(files=%#v) = %+v, want no %s error", tt.files, got, path)
				}
			}
		})
	}
}
