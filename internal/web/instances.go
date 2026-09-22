package web

import (
	"fmt"
	"strings"

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
func InstanceCards(cfg *config.Config) []InstanceCard {
	if len(cfg.Repositories) == 0 {
		return nil
	}
	all := Fields(cfg)
	cards := make([]InstanceCard, len(cfg.Repositories))
	index := make(map[string]int, len(cfg.Repositories))
	for i, repo := range cfg.Repositories {
		index[repo.ID] = i
		cards[i] = InstanceCard{
			ID:         repo.ID,
			Type:       "restic",
			Fields:     repositoryFields(all, fmt.Sprintf("repositories[%d].", i)),
			MountPoint: repo.MountPoint,
		}
	}
	var unassigned []config.Root
	for _, root := range cfg.Roots {
		if i, ok := index[root.RepositoryID]; ok {
			cards[i].Roots = append(cards[i].Roots, root)
			continue
		}
		unassigned = append(unassigned, root)
	}
	if len(unassigned) > 0 {
		cards = append(cards, InstanceCard{ID: unassignedCardID, Roots: unassigned})
	}
	return cards
}

// repositoryFields picks the leaf fields of one repositories[i] entry, keeping
// the order Fields established; the row marker has no trailing dot and so is
// left out by the prefix itself.
func repositoryFields(all []webui.Field, prefix string) []webui.Field {
	var out []webui.Field
	for _, f := range all {
		if strings.HasPrefix(f.Path, prefix) {
			out = append(out, f)
		}
	}
	return out
}
