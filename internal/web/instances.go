package web

import (
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/webui"
)

// unassignedCardID is the ID of the card holding roots whose repository_id
// matches no configured repository.
const unassignedCardID = "unassigned"

// InstanceCard is one backup instance as the instances page renders it. Under
// config schema v1 an instance is exactly one repositories[i] entry plus the
// roots bound to it, so ID is the repository id and Type is always "restic".
// The struct is schema-agnostic on purpose: schema v2 adds user-named
// instances by filling ID and Type differently, without changing callers.
type InstanceCard struct {
	ID         string
	Type       string
	Fields     []webui.Field
	Roots      []config.Root
	MountPoint string
}

// InstanceCards groups cfg into one card per repository, in configuration
// order, followed by an "unassigned" card when a root references a repository
// id that is not configured.
func InstanceCards(cfg *config.Config) []InstanceCard { return nil }
