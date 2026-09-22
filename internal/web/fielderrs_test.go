package web

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

func TestFieldErrors(t *testing.T) {
	validation := &config.ValidationError{Fields: []config.FieldError{
		{Path: "repositories[0].lock_mode", Msg: "must be normal or none"},
		{Path: "roots[1].prefix_map[0].hostname", Msg: "required"},
		{Msg: "at least one repository is required"},
	}}
	duplicate := &config.ValidationError{Fields: []config.FieldError{
		{Path: "web.listen", Msg: "must be a loopback address"},
		{Path: "web.listen", Msg: "must include a port"},
	}}

	tests := []struct {
		name       string
		in         error
		wantByPath map[string]string
		wantBanner []string
	}{
		{
			name:       "nil",
			in:         nil,
			wantByPath: map[string]string{},
			wantBanner: nil,
		},
		{
			name: "fields and a pathless error",
			in:   validation,
			wantByPath: map[string]string{
				"repositories[0].lock_mode":       "must be normal or none",
				"roots[1].prefix_map[0].hostname": "required",
			},
			wantBanner: []string{"at least one repository is required"},
		},
		{
			name: "wrapped validation error",
			in:   fmt.Errorf("load config: %w", validation),
			wantByPath: map[string]string{
				"repositories[0].lock_mode":       "must be normal or none",
				"roots[1].prefix_map[0].hostname": "required",
			},
			wantBanner: []string{"at least one repository is required"},
		},
		{
			name:       "second message for one path goes to the banner",
			in:         duplicate,
			wantByPath: map[string]string{"web.listen": "must be a loopback address"},
			wantBanner: []string{"web.listen: must include a port"},
		},
		{
			name:       "plain error",
			in:         errors.New("repository is unreachable"),
			wantByPath: map[string]string{},
			wantBanner: []string{"repository is unreachable"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotByPath, gotBanner := fieldErrors(tc.in)
			if !maps.Equal(gotByPath, tc.wantByPath) {
				t.Errorf("fieldErrors(%v) byPath = %v, want %v", tc.in, gotByPath, tc.wantByPath)
			}
			if !slices.Equal(gotBanner, tc.wantBanner) {
				t.Errorf("fieldErrors(%v) banner = %v, want %v", tc.in, gotBanner, tc.wantBanner)
			}
		})
	}
}

func TestFieldErrorsDropsNothing(t *testing.T) {
	in := &config.ValidationError{Fields: []config.FieldError{
		{Path: "repositories[0].lock_mode", Msg: "must be normal or none"},
		{Path: "repositories[0].lock_mode", Msg: "must not be empty"},
		{Path: "roots[1].prefix_map[0].hostname", Msg: "required"},
		{Msg: "at least one repository is required"},
	}}

	byPath, banner := fieldErrors(in)

	if got, want := len(byPath)+len(banner), len(in.Fields); got != want {
		t.Errorf("fieldErrors(%v) reported %d messages, want %d", in, got, want)
	}
}
