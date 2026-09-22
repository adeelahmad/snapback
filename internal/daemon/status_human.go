package daemon

import "github.com/adeelahmad/snapback/internal/status"

// RenderHuman renders s as byte-stable, human-readable multi-line text.
func RenderHuman(s status.Snapshot) string {
	_ = s
	return ""
}
