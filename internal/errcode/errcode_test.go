package errcode

import (
	"errors"
	"fmt"
	"io/fs"
	"testing"
)

func TestCodeValues(t *testing.T) {
	tests := []struct {
		name string
		code Code
		want string
	}{
		{"InvalidConfig", InvalidConfig, "invalid_configuration"},
		{"PrereqMissing", PrereqMissing, "prerequisite_missing"},
		{"PermissionDenied", PermissionDenied, "permission_denied"},
		{"LinkConflict", LinkConflict, "link_conflict"},
		{"RepoUnavailable", RepoUnavailable, "repository_unavailable"},
		{"MappingAbsent", MappingAbsent, "mapping_absent"},
		{"MountFailure", MountFailure, "mount_failure"},
		{"UnsupportedServiceManager", UnsupportedServiceManager, "unsupported_service_manager"},
		{"InodeBudgetExceeded", InodeBudgetExceeded, "inode_budget_exceeded"},
		{"OnAccessUnavailable", OnAccessUnavailable, "on_access_unavailable"},
		{"StaleState", StaleState, "stale_state"},
	}
	seen := make(map[Code]string)
	for _, tt := range tests {
		if got := string(tt.code); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, got, tt.want)
		}
		if prev, ok := seen[tt.code]; ok {
			t.Errorf("%s and %s share the value %q, want distinct codes", prev, tt.name, tt.code)
		}
		seen[tt.code] = tt.name
	}
	if got, want := len(seen), 11; got != want {
		t.Errorf("distinct codes = %d, want %d", got, want)
	}
}

// TestConstantsAreTypedCode is a compile-time guard: it builds only while
// every constant is assignable to Code.
func TestConstantsAreTypedCode(t *testing.T) {
	codes := []Code{
		InvalidConfig, PrereqMissing, PermissionDenied, LinkConflict,
		RepoUnavailable, MappingAbsent, MountFailure, UnsupportedServiceManager,
		InodeBudgetExceeded, OnAccessUnavailable, StaleState,
	}
	if got, want := len(codes), 11; got != want {
		t.Errorf("len(codes) = %d, want %d", got, want)
	}
}

func TestErrorString(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "with cause",
			err:  New(LinkConflict, "links.ensure", errors.New("exists")),
			want: "links.ensure: link_conflict: exists",
		},
		{
			name: "nil cause",
			err:  New(StaleState, "config.save", nil),
			want: "config.save: stale_state",
		},
	}
	for _, tt := range tests {
		if got := tt.err.Error(); got != tt.want {
			t.Errorf("%s: Error() = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestUnwrapReachesCause(t *testing.T) {
	err := New(PermissionDenied, "op", fs.ErrPermission)
	if !errors.Is(err, fs.ErrPermission) {
		t.Errorf("errors.Is(New(PermissionDenied, %q, fs.ErrPermission), fs.ErrPermission) = false, want true", "op")
	}
}

func TestOf(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Code
	}{
		{"nil", nil, ""},
		{"plain error", errors.New("x"), ""},
		{"direct", New(MountFailure, "mount.start", errors.New("boom")), "mount_failure"},
		{"wrapped", fmt.Errorf("ctx: %w", New(RepoUnavailable, "restic.list", errors.New("down"))), "repository_unavailable"},
		{
			"outermost wins",
			New(StaleState, "config.save", New(LinkConflict, "links.ensure", errors.New("exists"))),
			"stale_state",
		},
	}
	for _, tt := range tests {
		if got := Of(tt.err); got != tt.want {
			t.Errorf("Of(%s) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
