package web

import (
	"strconv"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/webui"
)

// instancesConfig is the fixture every instance card test shares: two
// repositories and three roots, two of them bound to repository A and one to
// repository B.
func instancesConfig() *config.Config {
	return &config.Config{
		Version: 1,
		Repositories: []config.Repository{
			{ID: "a", Repository: "/srv/a", MountPoint: "/mnt/a"},
			{ID: "b", Repository: "/srv/b"},
		},
		Roots: []config.Root{
			{ID: "home", LocalPath: "/home/me", RepositoryID: "a"},
			{ID: "work", LocalPath: "/srv/work", RepositoryID: "b"},
			{ID: "docs", LocalPath: "/home/me/docs", RepositoryID: "a"},
		},
	}
}

func rootIDs(roots []config.Root) []string {
	out := make([]string, len(roots))
	for i, r := range roots {
		out[i] = r.ID
	}
	return out
}

func fieldPaths(fields []webui.Field) []string {
	out := make([]string, len(fields))
	for i, f := range fields {
		out[i] = f.Path
	}
	return out
}

// TestInstanceCardsOnePerRepository pins that every repositories[i] yields one
// card, in configuration order, carrying the repository id as ID and "restic"
// as Type.
func TestInstanceCardsOnePerRepository(t *testing.T) {
	cards := InstanceCards(instancesConfig())
	if len(cards) != 2 {
		t.Fatalf("len(InstanceCards(cfg)) = %d, want 2", len(cards))
	}
	for i, want := range []string{"a", "b"} {
		if got := cards[i].ID; got != want {
			t.Errorf("cards[%d].ID = %q, want %q", i, got, want)
		}
		if got := cards[i].Type; got != "restic" {
			t.Errorf("cards[%d].Type = %q, want %q", i, got, "restic")
		}
	}
}

// TestInstanceCardsCarryBoundRoots pins that each card holds exactly the roots
// whose repository_id names it, in configuration order.
func TestInstanceCardsCarryBoundRoots(t *testing.T) {
	cards := InstanceCards(instancesConfig())
	if len(cards) != 2 {
		t.Fatalf("len(InstanceCards(cfg)) = %d, want 2", len(cards))
	}
	want := [][]string{{"home", "docs"}, {"work"}}
	for i, w := range want {
		got := rootIDs(cards[i].Roots)
		if strings.Join(got, ",") != strings.Join(w, ",") {
			t.Errorf("root ids of cards[%d] = %v, want %v", i, got, w)
		}
	}
}

// TestInstanceCardsCopyMountPoint pins that repositories[i].mount_point, added
// by S5-38, reaches the card.
func TestInstanceCardsCopyMountPoint(t *testing.T) {
	cards := InstanceCards(instancesConfig())
	if len(cards) != 2 {
		t.Fatalf("len(InstanceCards(cfg)) = %d, want 2", len(cards))
	}
	if got, want := cards[0].MountPoint, "/mnt/a"; got != want {
		t.Errorf("cards[0].MountPoint = %q, want %q", got, want)
	}
	if got := cards[1].MountPoint; got != "" {
		t.Errorf("cards[1].MountPoint = %q, want %q", got, "")
	}
}

// TestInstanceCardsFieldsAreRepositoryFields pins that a card carries exactly
// the repositories[i].* fields Fields(cfg) renders, in the same order, so the
// card never invents its own form model.
func TestInstanceCardsFieldsAreRepositoryFields(t *testing.T) {
	cfg := instancesConfig()
	cards := InstanceCards(cfg)
	if len(cards) != 2 {
		t.Fatalf("len(InstanceCards(cfg)) = %d, want 2", len(cards))
	}
	all := Fields(cfg)
	for i := range cards {
		prefix := "repositories[" + strconv.Itoa(i) + "]."
		var want []string
		for _, f := range all {
			if strings.HasPrefix(f.Path, prefix) {
				want = append(want, f.Path)
			}
		}
		if len(want) == 0 {
			t.Fatalf("Fields(cfg) has no field under %q", prefix)
		}
		got := fieldPaths(cards[i].Fields)
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("field paths of cards[%d] = %v, want %v", i, got, want)
		}
	}
}

// TestInstanceCardsUnassignedRoot pins that a root naming an unconfigured
// repository lands in a last card with ID "unassigned", instead of being
// dropped or attached to an unrelated repository.
func TestInstanceCardsUnassignedRoot(t *testing.T) {
	cfg := instancesConfig()
	cfg.Roots = append(cfg.Roots, config.Root{ID: "stray", LocalPath: "/stray", RepositoryID: "ghost"})
	cards := InstanceCards(cfg)
	if len(cards) != 3 {
		t.Fatalf("len(InstanceCards(cfg)) = %d, want 3", len(cards))
	}
	last := cards[len(cards)-1]
	if got, want := last.ID, "unassigned"; got != want {
		t.Errorf("last card ID = %q, want %q", got, want)
	}
	if got := rootIDs(last.Roots); len(got) != 1 || got[0] != "stray" {
		t.Errorf("root ids of the last card = %v, want [stray]", got)
	}
	for i, w := range [][]string{{"home", "docs"}, {"work"}} {
		if got := rootIDs(cards[i].Roots); strings.Join(got, ",") != strings.Join(w, ",") {
			t.Errorf("root ids of cards[%d] = %v, want %v", i, got, w)
		}
	}
}

// TestInstanceCardsNoRepositories pins that a configuration without
// repositories yields no cards at all.
func TestInstanceCardsNoRepositories(t *testing.T) {
	cards := InstanceCards(&config.Config{Version: 1})
	if len(cards) != 0 {
		t.Errorf("len(InstanceCards(empty)) = %d, want 0", len(cards))
	}
}

// TestInstanceCardSchemaV2Slot is the schema-v2 slot assertion: every case
// constructs an InstanceCard with KEYED fields only, so a later schema-v2
// field (a user-chosen instance name, say) can be added to the struct without
// touching any caller. If this test ever needs an extra value to compile, the
// struct stopped being additive and the change is a breaking one.
func TestInstanceCardSchemaV2Slot(t *testing.T) {
	cases := []struct {
		name string
		card InstanceCard
		want string
	}{
		{name: "restic repository", card: InstanceCard{ID: "a", Type: "restic"}, want: "restic"},
		{name: "unassigned bucket", card: InstanceCard{ID: unassignedCardID, Type: ""}, want: ""},
		{name: "future schema v2 type", card: InstanceCard{ID: "nas", Type: "borg"}, want: "borg"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.card.Type; got != tc.want {
				t.Errorf("card.Type = %q, want %q", got, tc.want)
			}
			if tc.card.ID == "" {
				t.Error("card.ID = \"\", want a non-empty id")
			}
			cards := InstanceCards(&config.Config{
				Version:      1,
				Repositories: []config.Repository{{ID: tc.card.ID, Repository: "/srv/" + tc.card.ID}},
			})
			if len(cards) != 1 {
				t.Fatalf("len(InstanceCards(one repository)) = %d, want 1", len(cards))
			}
			if got := cards[0].ID; got != tc.card.ID {
				t.Errorf("cards[0].ID = %q, want %q", got, tc.card.ID)
			}
		})
	}
}
