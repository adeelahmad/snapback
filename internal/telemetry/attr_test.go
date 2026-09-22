package telemetry

import (
	"reflect"
	"strings"
	"testing"
)

// wantKeys is the closed attribute key set, in the order Keys must return it.
var wantKeys = []string{"version", "os", "arch", "check", "code", "duration", "outcome"}

func TestAttrFieldsAreStrings(t *testing.T) {
	typ := reflect.TypeOf(Attr{})
	if got, want := typ.NumField(), 2; got != want {
		t.Fatalf("Attr has %d fields, want %d", got, want)
	}
	for _, name := range []string{"Key", "Value"} {
		field, ok := typ.FieldByName(name)
		if !ok {
			t.Fatalf("Attr has no field %s", name)
		}
		if field.Type.Kind() != reflect.String {
			t.Errorf("Attr.%s is %s, want string", name, field.Type)
		}
	}
}

func TestAttrKeysIsTheClosedSet(t *testing.T) {
	got := Keys()
	if !reflect.DeepEqual(got, wantKeys) {
		t.Fatalf("Keys() = %q, want %q", got, wantKeys)
	}
}

func TestAttrKeysCopyIsNotShared(t *testing.T) {
	first := Keys()
	if len(first) == 0 {
		t.Fatalf("Keys() = %q, want the closed set", first)
	}
	first[0] = "mutated"
	if got := Keys()[0]; got != wantKeys[0] {
		t.Errorf("Keys()[0] after mutating a returned slice = %q, want %q", got, wantKeys[0])
	}
}

func TestNewAttrAcceptsEveryAllowedKey(t *testing.T) {
	values := map[string]string{
		"version":  "1.4.1",
		"os":       "linux",
		"arch":     "arm64",
		"check":    "repository",
		"code":     "mount_failure",
		"duration": "<100ms",
		"outcome":  "ok",
	}
	for _, key := range wantKeys {
		value := values[key]
		attr, err := NewAttr(key, value)
		if err != nil {
			t.Errorf("NewAttr(%q, %q) error = %v, want nil", key, value, err)
			continue
		}
		if attr.Key != key || attr.Value != value {
			t.Errorf("NewAttr(%q, %q) = %+v, want {Key:%q Value:%q}", key, value, attr, key, value)
		}
	}
}

func TestNewAttrRejectsKeysOutsideTheClosedSet(t *testing.T) {
	for _, key := range []string{"path", "host", "repo", "uri", "user", "", "Version"} {
		attr, err := NewAttr(key, "x")
		if err == nil {
			t.Errorf("NewAttr(%q, \"x\") = %+v, want an error", key, attr)
			continue
		}
		for _, allowed := range wantKeys {
			if !strings.Contains(err.Error(), allowed) {
				t.Errorf("NewAttr(%q, \"x\") error = %q, want it to name the allowed key %q", key, err, allowed)
			}
		}
	}
}

func TestNewAttrRejectsValuesThatCouldCarryAnIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"unix path", "/home/u"},
		{"windows path", `C:\x`},
		{"uri scheme", "sftp:host"},
		{"uri", "sftp://backup.lan/repo"},
		{"email", "a@b"},
		{"space", "two words"},
		{"tab", "two\twords"},
		{"newline", "two\nwords"},
		{"nul", "a\x00b"},
		{"65 bytes", strings.Repeat("a", 65)},
	}
	const key = "outcome"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr, err := NewAttr(key, tt.value)
			if err == nil {
				t.Fatalf("NewAttr(%q, %q) = %+v, want an error", key, tt.value, attr)
			}
			if !strings.Contains(err.Error(), key) {
				t.Errorf("NewAttr(%q, %q) error = %q, want it to name the key %q", key, tt.value, err, key)
			}
		})
	}
}

func TestNewAttrAcceptsSixtyFourBytesAndRejectsSixtyFive(t *testing.T) {
	const key = "version"
	if _, err := NewAttr(key, strings.Repeat("a", 64)); err != nil {
		t.Errorf("NewAttr(%q, 64 bytes) error = %v, want nil", key, err)
	}
	if _, err := NewAttr(key, strings.Repeat("a", 65)); err == nil {
		t.Errorf("NewAttr(%q, 65 bytes) error = nil, want an error", key)
	}
}
