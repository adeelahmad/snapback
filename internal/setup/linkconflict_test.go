package setup

import (
	"errors"
	"fmt"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

func TestClassifyLinkError(t *testing.T) {
	conflict := errcode.New(errcode.LinkConflict, "links.ensure", errors.New(`"/work/.snapshot" is not an owned link`))

	tests := []struct {
		name     string
		root     string
		linkName string
		err      error
		want     Conflict
		wantOK   bool
	}{
		{
			name:     "conflict error is classified",
			root:     "/work",
			linkName: ".snapshot",
			err:      conflict,
			want: Conflict{
				Path: "/work/.snapshot",
				Kind: "foreign .snapshot entry",
				Fix:  "move it aside, then run: snapback link /work",
			},
			wantOK: true,
		},
		{
			name:     "wrapped conflict error is classified",
			root:     "/data/photos",
			linkName: ".snapshot",
			err:      fmt.Errorf("ensure link: %w", conflict),
			want: Conflict{
				Path: "/data/photos/.snapshot",
				Kind: "foreign .snapshot entry",
				Fix:  "move it aside, then run: snapback link /data/photos",
			},
			wantOK: true,
		},
		{
			name:     "unrelated error is not classified",
			root:     "/work",
			linkName: ".snapshot",
			err:      errcode.New(errcode.PermissionDenied, "links.ensure", errors.New("denied")),
		},
		{
			name:     "plain error is not classified",
			root:     "/work",
			linkName: ".snapshot",
			err:      errors.New("boom"),
		},
		{
			name:     "no error is not classified",
			root:     "/work",
			linkName: ".snapshot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOK := ClassifyLinkError(tt.root, tt.linkName, tt.err)
			if got != tt.want || gotOK != tt.wantOK {
				t.Errorf("ClassifyLinkError(%q, %q, %v) = %+v, %t, want %+v, %t",
					tt.root, tt.linkName, tt.err, got, gotOK, tt.want, tt.wantOK)
			}
		})
	}
}
