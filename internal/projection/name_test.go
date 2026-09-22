package projection

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestValidateName(t *testing.T) {
	snapshotID := strings.Repeat("0123456789abcdef", 4)
	accepted := []string{
		"a", "latest", ".hidden", "a b", "...", "UPPER", "Ünïcode",
		"2026-09-22T10-00-00Z", snapshotID,
	}
	rejected := []string{"", ".", "..", "a/b", "/", "a\x00b"}

	for _, name := range accepted {
		t.Run(fmt.Sprintf("accept %q", name), func(t *testing.T) {
			if err := validateName(name); err != nil {
				t.Fatalf("validateName(%q) = %v, want nil", name, err)
			}
		})
	}
	for _, name := range rejected {
		t.Run(fmt.Sprintf("reject %q", name), func(t *testing.T) {
			err := validateName(name)
			if err == nil {
				t.Fatalf("validateName(%q) = nil, want error wrapping ErrInvalidName", name)
			}
			if !errors.Is(err, ErrInvalidName) {
				t.Errorf("validateName(%q) = %v, want errors.Is(err, ErrInvalidName)", name, err)
			}
			if want := fmt.Sprintf("%q", name); !strings.Contains(err.Error(), want) {
				t.Errorf("validateName(%q) error %q does not contain %s", name, err.Error(), want)
			}
		})
	}
}
