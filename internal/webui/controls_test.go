package webui

import (
	"html/template"
	"strings"
	"testing"
)

// controlOption is one choice of a select, one chip or one repeatable row.
type controlOption struct {
	Value string
	Label string
}

// controlView is the field view every control partial takes. It mirrors the
// shape the form model will render with (Path, Label, Help, Value, Options,
// Error, Required) and is declared here until that model lands.
type controlView struct {
	Path     string
	Label    string
	Help     string
	Value    string
	Options  []controlOption
	Error    string
	Required bool
}

// controlTemplates parses templates/controls.html on its own, the way a page
// template set will include the partials.
func controlTemplates(t *testing.T) *template.Template {
	t.Helper()
	tmpl, err := template.ParseFS(embedded, "templates/controls.html")
	if err != nil {
		t.Fatalf("template.ParseFS(embedded, templates/controls.html) error = %v, want the control partials", err)
	}
	return tmpl
}

// renderControl executes one partial with v and returns its output.
func renderControl(t *testing.T, name string, v controlView) string {
	t.Helper()
	var b strings.Builder
	if err := controlTemplates(t).ExecuteTemplate(&b, name, v); err != nil {
		t.Fatalf("ExecuteTemplate(%q) error = %v", name, err)
	}
	return b.String()
}

func TestControlPartialsGoldenRender(t *testing.T) {
	scalar := controlView{
		Path:     "repositories[0].uri",
		Label:    "Repository URI",
		Help:     "Where the restic repository lives.",
		Value:    "rclone:b2:bucket",
		Error:    "must not be empty",
		Required: true,
	}
	selectView := controlView{
		Path:  "repositories[0].lock_mode",
		Label: "Lock mode",
		Help:  "How the daemon locks the repository.",
		Value: "none",
		Options: []controlOption{
			{Value: "normal", Label: "Normal"},
			{Value: "none", Label: "None"},
		},
	}
	toggleView := controlView{
		Path:  "service.enabled",
		Label: "Run as a service",
		Help:  "Start the daemon at login.",
		Value: "true",
	}
	chipsView := controlView{
		Path:  "reader.deny_processes",
		Label: "Denied processes",
		Help:  "Readers refused by name.",
		Options: []controlOption{
			{Value: "mds", Label: "mds"},
		},
	}
	rowView := controlView{
		Path:  "roots",
		Label: "Roots",
		Help:  "Directories Snapback serves.",
		Options: []controlOption{
			{Value: "/home/a", Label: "Root 1"},
		},
	}

	cases := []struct {
		name string
		view controlView
		want string
	}{
		{
			name: "control-input",
			view: scalar,
			want: `<div class="field">
<label class="field__label" for="f-repositories[0].uri">Repository URI</label>
<input class="field__control" type="text" id="f-repositories[0].uri" name="repositories[0].uri" value="rclone:b2:bucket" title="Where the restic repository lives." aria-describedby="f-repositories[0].uri-help f-repositories[0].uri-error" required>
<p class="field__help" id="f-repositories[0].uri-help">Where the restic repository lives.</p>
<p class="field__error" id="f-repositories[0].uri-error">must not be empty</p>
</div>`,
		},
		{
			name: "control-password",
			view: controlView{Path: "repositories[0].password", Label: "Password", Help: "Written to a file, never to config.yaml."},
			want: `<div class="field">
<label class="field__label" for="f-repositories[0].password">Password</label>
<input class="field__control" type="password" id="f-repositories[0].password" name="repositories[0].password" autocomplete="new-password" title="Written to a file, never to config.yaml." aria-describedby="f-repositories[0].password-help f-repositories[0].password-error">
<p class="field__help" id="f-repositories[0].password-help">Written to a file, never to config.yaml.</p>
<p class="field__error" id="f-repositories[0].password-error"></p>
</div>`,
		},
		{
			name: "control-number",
			view: controlView{Path: "catalog.max_entries", Label: "Maximum entries", Help: "Entries kept in the catalog.", Value: "500"},
			want: `<div class="field">
<label class="field__label" for="f-catalog.max_entries">Maximum entries</label>
<input class="field__control" type="number" inputmode="numeric" id="f-catalog.max_entries" name="catalog.max_entries" value="500" title="Entries kept in the catalog." aria-describedby="f-catalog.max_entries-help f-catalog.max_entries-error">
<p class="field__help" id="f-catalog.max_entries-help">Entries kept in the catalog.</p>
<p class="field__error" id="f-catalog.max_entries-error"></p>
</div>`,
		},
		{
			name: "control-duration",
			view: controlView{Path: "refresh_interval", Label: "Refresh interval", Help: "How often snapshots are re-read.", Value: "15m"},
			want: `<div class="field">
<label class="field__label" for="f-refresh_interval">Refresh interval</label>
<input class="field__control" type="text" pattern="[0-9]+(ns|us|ms|s|m|h)" id="f-refresh_interval" name="refresh_interval" value="15m" title="How often snapshots are re-read." aria-describedby="f-refresh_interval-help f-refresh_interval-error">
<p class="field__help" id="f-refresh_interval-help">How often snapshots are re-read.</p>
<p class="field__error" id="f-refresh_interval-error"></p>
</div>`,
		},
		{
			name: "control-select",
			view: selectView,
			want: `<div class="field">
<label class="field__label" for="f-repositories[0].lock_mode">Lock mode</label>
<select class="field__control" id="f-repositories[0].lock_mode" name="repositories[0].lock_mode" title="How the daemon locks the repository." aria-describedby="f-repositories[0].lock_mode-help f-repositories[0].lock_mode-error">
<option value="normal">Normal</option>
<option value="none" selected>None</option>
</select>
<p class="field__help" id="f-repositories[0].lock_mode-help">How the daemon locks the repository.</p>
<p class="field__error" id="f-repositories[0].lock_mode-error"></p>
</div>`,
		},
		{
			name: "control-toggle",
			view: toggleView,
			want: `<div class="field field--toggle">
<input type="hidden" name="service.enabled" value="false">
<label class="field__label" for="f-service.enabled">Run as a service</label>
<input class="field__control field__switch" type="checkbox" id="f-service.enabled" name="service.enabled" value="true" checked title="Start the daemon at login." aria-describedby="f-service.enabled-help f-service.enabled-error">
<span class="field__switch-track" aria-hidden="true"></span>
<p class="field__help" id="f-service.enabled-help">Start the daemon at login.</p>
<p class="field__error" id="f-service.enabled-error"></p>
</div>`,
		},
		{
			name: "control-chips",
			view: chipsView,
			want: `<div class="field field--chips">
<label class="field__label" for="f-reader.deny_processes">Denied processes</label>
<ul class="field__chips">
<li class="chip"><input type="hidden" name="reader.deny_processes[0]" value="mds"><span class="chip__text">mds</span><button class="button-secondary chip__remove" type="button" data-js="chip-remove">Remove mds</button></li>
</ul>
<input class="field__control" type="text" id="f-reader.deny_processes" name="reader.deny_processes[1]" value="" data-js="chips" title="Readers refused by name." aria-describedby="f-reader.deny_processes-help f-reader.deny_processes-error">
<p class="field__help" id="f-reader.deny_processes-help">Readers refused by name.</p>
<p class="field__error" id="f-reader.deny_processes-error"></p>
</div>`,
		},
		{
			name: "control-row",
			view: rowView,
			want: `<fieldset class="field field--rows" data-js="rows">
<legend class="field__label">Roots</legend>
<div class="field__row">
<label class="field__label" for="f-roots[0]">Root 1</label>
<input class="field__control" type="text" id="f-roots[0]" name="roots[0]" value="/home/a" title="Directories Snapback serves." aria-describedby="f-roots-help f-roots-error">
<button class="button-secondary" type="button" data-js="row-remove">Remove Root 1</button>
</div>
<button class="button-secondary" type="button" data-js="row-add">Add Roots</button>
<p class="field__help" id="f-roots-help">Directories Snapback serves.</p>
<p class="field__error" id="f-roots-error"></p>
</fieldset>`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := renderControl(t, c.name, c.view)
			if got != c.want {
				t.Errorf("%s render =\n%s\nwant:\n%s", c.name, got, c.want)
			}
		})
	}
}

func TestControlPartialsHaveNoTextarea(t *testing.T) {
	src := readTemplate(t, "controls.html")
	if strings.Contains(src, "<textarea") {
		t.Error("controls.html contains <textarea, want controls only")
	}
}
