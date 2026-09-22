package webui

import (
	"regexp"
	"strings"
	"testing"
)

// formSectionsFixture is a split form with two basic controls, three advanced
// ones across two sections, and one error anchored to an advanced control.
func formSectionsFixture() FormSections {
	return FormSections{
		Basic: []ConfigSection{{
			Title: "Repository",
			Controls: []Control{
				{
					Kind:     KindText,
					Path:     "repositories[0].uri",
					Label:    "Repository URI",
					Value:    "rclone:b2:bucket",
					Help:     "Where the restic repository lives.",
					Required: true,
				},
				{
					Kind:  KindPassword,
					Path:  "repositories[0].credential_file",
					Label: "Password file",
				},
			},
		}},
		Advanced: []ConfigSection{
			{
				Title: "Mount",
				Controls: []Control{
					{
						Kind:  KindDuration,
						Path:  "mount.refresh_interval",
						Label: "Refresh interval",
						Value: "5m",
					},
					{
						Kind:  KindSelect,
						Path:  "repositories[0].lock_mode",
						Label: "Lock mode",
						Value: "none",
						Options: []Option{
							{Value: "normal", Label: "Normal"},
							{Value: "none", Label: "None"},
						},
						Error: "lock mode is not supported here",
					},
				},
			},
			{
				Title: "Service",
				Controls: []Control{{
					Kind:  KindToggle,
					Path:  "service.enabled",
					Label: "Run as a service",
					Value: "true",
				}},
			},
		},
		AdvancedCount: 3,
	}
}

// renderFormSections renders the fixture and fails the test on any error.
func renderFormSections(t *testing.T, v FormSections) string {
	t.Helper()
	out, err := RenderFormSections(v)
	if err != nil {
		t.Fatalf("RenderFormSections() error = %v, want the split form markup", err)
	}
	return string(out)
}

// advancedDetails returns the substring of html from the advanced <details>
// open tag to its closing </details>.
func advancedDetails(t *testing.T, html string) string {
	t.Helper()
	i := strings.Index(html, `<details data-js="form-advanced"`)
	if i < 0 {
		t.Fatalf("RenderFormSections() = %q, want a <details data-js=\"form-advanced\"> element", html)
	}
	rest := html[i:]
	j := strings.Index(rest, "</details>")
	if j < 0 {
		t.Fatalf("advanced disclosure = %q, want a closing </details>", rest)
	}
	return rest[:j+len("</details>")]
}

func TestRenderFormSectionsBasicBlockIsOpen(t *testing.T) {
	html := renderFormSections(t, formSectionsFixture())

	if got := strings.Count(html, `data-js="form-basic"`); got != 1 {
		t.Errorf("count of data-js=%q = %d, want exactly 1 in %q", "form-basic", got, html)
	}
	if strings.Contains(html, `<details data-js="form-basic"`) {
		t.Errorf("basic block is a <details> in %q, want a plain always-visible block", html)
	}
	for _, name := range []string{
		`name="repositories[0].uri"`,
		`name="repositories[0].credential_file"`,
	} {
		if !strings.Contains(html, name) {
			t.Errorf("RenderFormSections() missing basic control %s in %q", name, html)
		}
	}
	if !strings.Contains(html, "Repository") {
		t.Errorf("RenderFormSections() missing the basic section title %q in %q", "Repository", html)
	}
}

func TestRenderFormSectionsAdvancedIsCollapsed(t *testing.T) {
	html := renderFormSections(t, formSectionsFixture())

	if got := strings.Count(html, `<details data-js="form-advanced"`); got != 1 {
		t.Errorf("count of <details data-js=%q = %d, want exactly 1 in %q", "form-advanced", got, html)
	}
	details := advancedDetails(t, html)
	openTag := details[:strings.Index(details, ">")+1]
	if regexp.MustCompile(`(\s|")open(\s|>|=)`).MatchString(openTag) {
		t.Errorf("advanced <details> open tag = %q, want no open attribute", openTag)
	}
	for _, name := range []string{
		`name="mount.refresh_interval"`,
		`name="repositories[0].lock_mode"`,
		`name="service.enabled"`,
	} {
		if !strings.Contains(details, name) {
			t.Errorf("advanced disclosure missing control %s in %q", name, details)
		}
	}
}

func TestRenderFormSectionsSummaryCarriesCountBadge(t *testing.T) {
	v := formSectionsFixture()
	html := renderFormSections(t, v)
	details := advancedDetails(t, html)

	i := strings.Index(details, "<summary")
	j := strings.Index(details, "</summary>")
	if i < 0 || j < i {
		t.Fatalf("advanced disclosure = %q, want a <summary> element", details)
	}
	summary := details[i : j+len("</summary>")]
	if want := `<span class="badge">3</span>`; !strings.Contains(summary, want) {
		t.Errorf("advanced <summary> = %q, want it to contain %s", summary, want)
	}
	if got := strings.Count(details, "<summary"); got != 1 {
		t.Errorf("count of <summary in the advanced disclosure = %d, want exactly 1", got)
	}
}

func TestRenderFormSectionsUsesControlPartialsOnly(t *testing.T) {
	html := renderFormSections(t, formSectionsFixture())

	if strings.Contains(html, "<textarea") {
		t.Errorf("RenderFormSections() contains a <textarea in %q, want the T4 controls only", html)
	}
	if strings.Contains(html, "<script") {
		t.Errorf("RenderFormSections() contains a <script in %q, want markup only", html)
	}
	if !strings.Contains(html, `class="field__control"`) {
		t.Errorf("RenderFormSections() = %q, want the control partials' field__control markup", html)
	}
	if !strings.Contains(html, `<select class="field__control" id="f-repositories[0].lock_mode"`) {
		t.Errorf("RenderFormSections() = %q, want the select control dispatched by Kind", html)
	}
}

func TestRenderFormSectionsRendersControlError(t *testing.T) {
	html := renderFormSections(t, formSectionsFixture())

	want := `<p class="field__error" id="f-repositories[0].lock_mode-error">lock mode is not supported here</p>`
	if !strings.Contains(html, want) {
		t.Errorf("RenderFormSections() = %q, want the anchored error %s", html, want)
	}
}

func TestRenderFormSectionsEmptyAdvancedHidesDisclosure(t *testing.T) {
	v := formSectionsFixture()
	v.Advanced = nil
	v.AdvancedCount = 0

	html := renderFormSections(t, v)

	if strings.Contains(html, "form-advanced") {
		t.Errorf("RenderFormSections() with no advanced fields = %q, want no advanced disclosure", html)
	}
	if !strings.Contains(html, `data-js="form-basic"`) {
		t.Errorf("RenderFormSections() with no advanced fields = %q, want the basic block", html)
	}
}
