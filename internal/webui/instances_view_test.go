package webui

import (
	"regexp"
	"strings"
	"testing"
)

// instanceNote is the honesty note the card partial must carry verbatim: named
// instances are a schema v2 feature that does not exist yet.
const instanceNote = "Named instances arrive with config schema v2 (SPEC-ADDENDUM-A §2) and are not implemented yet."

var (
	instanceArticleRE = regexp.MustCompile(`(?is)<article class="instance" data-instance="([^"]*)">`)
	instanceHeadingRE = regexp.MustCompile(`(?is)<h[23]\b[^>]*>(.*?)</h[23]>`)
	instanceBadgeRE   = regexp.MustCompile(`(?is)<span class="badge[^"]*"[^>]*>(.*?)</span>`)
	instanceRowsRE    = regexp.MustCompile(`(?is)<[a-z]+\b[^>]*\bdata-js="rows"[^>]*>`)
	instanceRowAddRE  = regexp.MustCompile(`(?is)<button\b[^>]*\bdata-js="row-add"[^>]*>`)
	instanceRowDelRE  = regexp.MustCompile(`(?is)<button\b[^>]*\bdata-js="row-remove"[^>]*>`)
)

// instanceCardsFixture is a two-card view: a mounted repository with roots and
// a second one with no mount point.
func instanceCardsFixture() []InstanceCardView {
	return []InstanceCardView{
		{
			ID:         "default",
			Type:       "restic",
			MountPoint: "/mnt/home-nas",
			Controls: []Control{
				{Kind: KindText, Path: "repositories[0].uri", Label: "Repository", Value: "/srv/restic"},
			},
			Roots: []string{"/home/ada", "/srv/work"},
		},
		{
			ID:   "laptop",
			Type: "restic",
			Controls: []Control{
				{Kind: KindText, Path: "repositories[1].uri", Label: "Repository", Value: "s3:example/bucket"},
			},
			Roots: []string{"/home/ada/laptop"},
		},
	}
}

// renderInstanceCards renders cards or fails the test.
func renderInstanceCards(t *testing.T, cards []InstanceCardView) string {
	t.Helper()
	out, err := RenderInstanceCards(cards)
	if err != nil {
		t.Fatalf("RenderInstanceCards(%d cards) error = %v", len(cards), err)
	}
	return string(out)
}

// instanceCardBodies splits a render into the markup of each card, in order.
func instanceCardBodies(t *testing.T, out string) []string {
	t.Helper()
	locs := instanceArticleRE.FindAllStringIndex(out, -1)
	bodies := make([]string, len(locs))
	for i, loc := range locs {
		end := len(out)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		bodies[i] = out[loc[0]:end]
	}
	return bodies
}

func TestRenderInstanceCardsOneArticlePerCardInOrder(t *testing.T) {
	cards := instanceCardsFixture()
	out := renderInstanceCards(t, cards)
	ms := instanceArticleRE.FindAllStringSubmatch(out, -1)
	if got, want := len(ms), len(cards); got != want {
		t.Fatalf("RenderInstanceCards() has %d <article class=\"instance\" data-instance>, want %d\nrender = %q", got, want, out)
	}
	for i, m := range ms {
		if got, want := m[1], cards[i].ID; got != want {
			t.Errorf("RenderInstanceCards() article[%d] data-instance = %q, want %q", i, got, want)
		}
	}
}

func TestRenderInstanceCardsShowsIDAndTypeBadge(t *testing.T) {
	cards := instanceCardsFixture()
	bodies := instanceCardBodies(t, renderInstanceCards(t, cards))
	if got, want := len(bodies), len(cards); got != want {
		t.Fatalf("RenderInstanceCards() cards = %d, want %d", got, want)
	}
	for i, body := range bodies {
		heads := instanceHeadingRE.FindAllStringSubmatch(body, -1)
		if len(heads) == 0 {
			t.Errorf("RenderInstanceCards() card[%d] has no <h2> or <h3>, want one naming %q", i, cards[i].ID)
		}
		named := false
		for _, h := range heads {
			if strings.Contains(h[1], cards[i].ID) {
				named = true
			}
		}
		if len(heads) > 0 && !named {
			t.Errorf("RenderInstanceCards() card[%d] heading = %q, want it to name %q", i, heads[0][1], cards[i].ID)
		}
		badged := false
		for _, b := range instanceBadgeRE.FindAllStringSubmatch(body, -1) {
			if strings.Contains(b[1], cards[i].Type) {
				badged = true
			}
		}
		if !badged {
			t.Errorf("RenderInstanceCards() card[%d] has no badge carrying type %q\ncard = %q", i, cards[i].Type, body)
		}
	}
}

func TestRenderInstanceCardsRendersControls(t *testing.T) {
	cards := instanceCardsFixture()
	bodies := instanceCardBodies(t, renderInstanceCards(t, cards))
	if got, want := len(bodies), len(cards); got != want {
		t.Fatalf("RenderInstanceCards() cards = %d, want %d", got, want)
	}
	for i, body := range bodies {
		for _, c := range cards[i].Controls {
			if want := `name="` + c.Path + `"`; !strings.Contains(body, want) {
				t.Errorf("RenderInstanceCards() card[%d] missing %s\ncard = %q", i, want, body)
			}
			if !strings.Contains(body, `value="`+c.Value+`"`) {
				t.Errorf("RenderInstanceCards() card[%d] missing value %q for %s", i, c.Value, c.Path)
			}
		}
	}
}

func TestRenderInstanceCardsMountPointControlByPosition(t *testing.T) {
	cards := instanceCardsFixture()
	bodies := instanceCardBodies(t, renderInstanceCards(t, cards))
	if got, want := len(bodies), len(cards); got != want {
		t.Fatalf("RenderInstanceCards() cards = %d, want %d", got, want)
	}
	if want := `name="repositories[0].mount_point"`; !strings.Contains(bodies[0], want) {
		t.Errorf("RenderInstanceCards() card[0] missing %s, want the mount point as a control\ncard = %q", want, bodies[0])
	}
	if want := `value="` + cards[0].MountPoint + `"`; !strings.Contains(bodies[0], want) {
		t.Errorf("RenderInstanceCards() card[0] missing %s, want the mount point value", want)
	}
	if got := `repositories[1].mount_point`; strings.Contains(bodies[1], got) {
		t.Errorf("RenderInstanceCards() card[1] has %s, want no mount point control when MountPoint is empty", got)
	}
}

func TestRenderInstanceCardsRootRowHooks(t *testing.T) {
	cards := instanceCardsFixture()
	bodies := instanceCardBodies(t, renderInstanceCards(t, cards))
	if got, want := len(bodies), len(cards); got != want {
		t.Fatalf("RenderInstanceCards() cards = %d, want %d", got, want)
	}
	for i, body := range bodies {
		if got := len(instanceRowsRE.FindAllString(body, -1)); got != 1 {
			t.Errorf("RenderInstanceCards() card[%d] has %d data-js=\"rows\" wrappers, want 1\ncard = %q", i, got, body)
		}
		if got := len(instanceRowAddRE.FindAllString(body, -1)); got != 1 {
			t.Errorf("RenderInstanceCards() card[%d] has %d data-js=\"row-add\" buttons, want 1", i, got)
		}
		if got, want := len(instanceRowDelRE.FindAllString(body, -1)), len(cards[i].Roots); got != want {
			t.Errorf("RenderInstanceCards() card[%d] has %d data-js=\"row-remove\" buttons, want %d (one per root)", i, got, want)
		}
		for _, root := range cards[i].Roots {
			if !strings.Contains(body, `value="`+root+`"`) {
				t.Errorf("RenderInstanceCards() card[%d] missing root %q\ncard = %q", i, root, body)
			}
		}
	}
}

func TestRenderInstanceCardsHonestyNoteOnce(t *testing.T) {
	for _, tc := range []struct {
		name  string
		cards []InstanceCardView
	}{
		{name: "two cards", cards: instanceCardsFixture()},
		{name: "one card", cards: instanceCardsFixture()[:1]},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := renderInstanceCards(t, tc.cards)
			if got := strings.Count(out, instanceNote); got != 1 {
				t.Errorf("RenderInstanceCards() carries the note %q %d times, want exactly 1\nrender = %q", instanceNote, got, out)
			}
		})
	}
}

func TestRenderInstanceCardsEscapesHostileValues(t *testing.T) {
	out := renderInstanceCards(t, []InstanceCardView{{
		ID:         hostile,
		Type:       hostile,
		MountPoint: "/mnt/" + hostile,
		Roots:      []string{"/home/" + hostile},
	}})
	if out == "" {
		t.Fatal("RenderInstanceCards(hostile card) = \"\", want the rendered card")
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "<script") {
		t.Errorf("RenderInstanceCards(hostile card) contains a raw <script, want the id and type escaped\nrender = %q", out)
	}
	if strings.Contains(lower, "<textarea") {
		t.Errorf("RenderInstanceCards(hostile card) contains a <textarea, want plain inputs only\nrender = %q", out)
	}
	if !strings.Contains(out, "&lt;script&gt;") {
		t.Errorf("RenderInstanceCards(hostile card) has no escaped &lt;script&gt;, want the hostile id rendered escaped\nrender = %q", out)
	}
}
