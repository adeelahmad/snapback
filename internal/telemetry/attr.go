// Package telemetry defines the closed, opt-in telemetry schema: the fixed
// event names, the duration buckets and the attribute type that makes a path,
// URI, hostname or address unrepresentable.
package telemetry

import (
	"fmt"
	"strings"
	"unicode"
)

// keys is the closed attribute key set, in its documented order. Keys returns a
// copy so a caller cannot reach this backing array.
var keys = [...]string{"version", "os", "arch", "check", "code", "duration", "outcome"}

// maxValueBytes caps an attribute value: long enough for a version or a bucket
// label, too short to smuggle a path or a URI.
const maxValueBytes = 64

// Attr is one telemetry attribute. Both fields are strings on purpose: the
// schema carries no arbitrary values, so there is no any here.
type Attr struct {
	Key   string
	Value string
}

// NewAttr returns the attribute for key and value, or an error when key is
// outside the closed key set or value could carry an identifier.
func NewAttr(key, value string) (Attr, error) {
	if !allowed(key) {
		return Attr{}, fmt.Errorf("telemetry: attribute key %q is not one of %s", key, strings.Join(Keys(), ", "))
	}
	if len(value) > maxValueBytes {
		return Attr{}, fmt.Errorf("telemetry: value for key %q is %d bytes, want at most %d", key, len(value), maxValueBytes)
	}
	for _, r := range value {
		if isIdentifierRune(r) {
			return Attr{}, fmt.Errorf("telemetry: value for key %q contains %q", key, r)
		}
	}
	return Attr{Key: key, Value: value}, nil
}

// Keys returns the closed attribute key set in its documented order.
func Keys() []string {
	out := make([]string, len(keys))
	copy(out, keys[:])
	return out
}

func allowed(key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}

// isIdentifierRune reports whether r could carry part of a path, a URI, a
// hostname or an address, or could break the value across fields.
func isIdentifierRune(r rune) bool {
	switch r {
	case '/', '\\', ':', '@', 0:
		return true
	}
	return unicode.IsSpace(r)
}
