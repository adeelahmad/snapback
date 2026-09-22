package webui

import (
	"html"
	"regexp"
	"strings"
	"testing"
)

var (
	// RE2 has no backreferences: the pill holds plain text, so the first
	// closing tag after it is its own.
	daemonPillRE   = regexp.MustCompile(`(?is)<[a-z][a-z0-9]*\b[^>]*\bdata-js="daemon-status"[^>]*>([^<]*)</[a-z][a-z0-9]*>`)
	daemonFormRE   = regexp.MustCompile(`(?is)<form\b([^>]*)>(.*?)</form>`)
	daemonCSRFRE   = regexp.MustCompile(`(?is)<input\b[^>]*\bname="csrf_token"[^>]*>`)
	daemonButtonRE = regexp.MustCompile(`(?is)<button\b[^>]*>`)
	daemonScriptRE = regexp.MustCompile(`(?i)<script\b`)
)

// daemonAttrRE matches the quoted value of the attribute name in a tag.
func daemonAttrRE(name string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(name) + `="([^"]*)"`)
}

// renderDaemonControl renders v or fails the test.
func renderDaemonControl(t *testing.T, v DaemonControlView) string {
	t.Helper()
	out, err := RenderDaemonControl(v)
	if err != nil {
		t.Fatalf("RenderDaemonControl(%+v) error = %v", v, err)
	}
	return string(out)
}

// daemonForms returns the rendered <form> bodies keyed by their action.
func daemonForms(t *testing.T, out string) map[string]string {
	t.Helper()
	forms := daemonFormRE.FindAllStringSubmatch(out, -1)
	if got, want := len(forms), 2; got != want {
		t.Fatalf("RenderDaemonControl() <form> count = %d, want %d (start and stop)\nrendered: %s", got, want, out)
	}
	byAction := make(map[string]string, len(forms))
	for i, f := range forms {
		attrs, body := f[1], f[2]
		m := daemonAttrRE("method").FindStringSubmatch(attrs)
		if m == nil {
			t.Fatalf("RenderDaemonControl() <form>[%d] attrs = %q, want method=\"post\"", i, attrs)
		}
		if got, want := strings.ToLower(m[1]), "post"; got != want {
			t.Errorf("RenderDaemonControl() <form>[%d] method = %q, want %q (JS-off fallback)", i, m[1], want)
		}
		a := daemonAttrRE("action").FindStringSubmatch(attrs)
		if a == nil {
			t.Fatalf("RenderDaemonControl() <form>[%d] attrs = %q, want an action", i, attrs)
		}
		byAction[a[1]] = body
	}
	for _, want := range []string{"/api/daemon/start", "/api/daemon/stop"} {
		if _, ok := byAction[want]; !ok {
			t.Fatalf("RenderDaemonControl() has no <form action=%q>, want it; got actions %v\nrendered: %s", want, actionsOf(byAction), out)
		}
	}
	return byAction
}

// actionsOf lists the keys of byAction for failure messages.
func actionsOf(byAction map[string]string) []string {
	out := make([]string, 0, len(byAction))
	for a := range byAction {
		out = append(out, a)
	}
	return out
}

// daemonButton returns the single <button> tag inside a form body.
func daemonButton(t *testing.T, action, body string) string {
	t.Helper()
	buttons := daemonButtonRE.FindAllString(body, -1)
	if got, want := len(buttons), 1; got != want {
		t.Fatalf("RenderDaemonControl() form %s <button> count = %d, want %d\nform body: %s", action, got, want, body)
	}
	return buttons[0]
}

func TestRenderDaemonControlStatusPill(t *testing.T) {
	tests := []struct {
		name    string
		running bool
		want    string
	}{
		{"running", true, "running"},
		{"stopped", false, "stopped"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := renderDaemonControl(t, DaemonControlView{Running: tt.running, CSRFToken: "tok-daemon"})
			pills := daemonPillRE.FindAllStringSubmatch(out, -1)
			if got, want := len(pills), 1; got != want {
				t.Fatalf(`RenderDaemonControl(Running=%t) elements with data-js="daemon-status" = %d, want %d`+"\nrendered: %s", tt.running, got, want, out)
			}
			text := strings.TrimSpace(html.UnescapeString(pills[0][1]))
			if text != tt.want {
				t.Errorf("RenderDaemonControl(Running=%t) status pill text = %q, want %q", tt.running, text, tt.want)
			}
		})
	}
}

func TestRenderDaemonControlFormsCarryCSRF(t *testing.T) {
	const token = "tok-daemon-42"
	out := renderDaemonControl(t, DaemonControlView{Running: false, CSRFToken: token})
	byAction := daemonForms(t, out)
	wantHook := map[string]string{
		"/api/daemon/start": "daemon-start",
		"/api/daemon/stop":  "daemon-stop",
	}
	for action, body := range byAction {
		inputs := daemonCSRFRE.FindAllString(body, -1)
		if got, want := len(inputs), 1; got != want {
			t.Fatalf(`RenderDaemonControl() form %s <input name="csrf_token"> count = %d, want %d`+"\nform body: %s", action, got, want, body)
		}
		in := inputs[0]
		if m := daemonAttrRE("type").FindStringSubmatch(in); m == nil || strings.ToLower(m[1]) != "hidden" {
			t.Errorf(`RenderDaemonControl() form %s csrf input = %q, want type="hidden"`, action, in)
		}
		m := daemonAttrRE("value").FindStringSubmatch(in)
		if m == nil {
			t.Errorf("RenderDaemonControl() form %s csrf input = %q, want a value attribute", action, in)
		} else if got := html.UnescapeString(m[1]); got != token {
			t.Errorf("RenderDaemonControl() form %s csrf token = %q, want %q", action, got, token)
		}
		btn := daemonButton(t, action, body)
		if m := daemonAttrRE("type").FindStringSubmatch(btn); m == nil || strings.ToLower(m[1]) != "submit" {
			t.Errorf(`RenderDaemonControl() form %s button = %q, want type="submit"`, action, btn)
		}
		h := daemonAttrRE("data-js").FindStringSubmatch(btn)
		if h == nil {
			t.Errorf("RenderDaemonControl() form %s button = %q, want data-js=%q", action, btn, wantHook[action])
		} else if h[1] != wantHook[action] {
			t.Errorf("RenderDaemonControl() form %s button data-js = %q, want %q", action, h[1], wantHook[action])
		}
	}
}

func TestRenderDaemonControlDisablesTheImpossibleAction(t *testing.T) {
	tests := []struct {
		name         string
		running      bool
		wantDisabled string
		wantEnabled  string
	}{
		{"running disables start", true, "/api/daemon/start", "/api/daemon/stop"},
		{"stopped disables stop", false, "/api/daemon/stop", "/api/daemon/start"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := renderDaemonControl(t, DaemonControlView{Running: tt.running, CSRFToken: "tok-daemon"})
			byAction := daemonForms(t, out)
			disabled := regexp.MustCompile(`(?i)\bdisabled\b`)
			if btn := daemonButton(t, tt.wantDisabled, byAction[tt.wantDisabled]); !disabled.MatchString(btn) {
				t.Errorf("RenderDaemonControl(Running=%t) form %s button = %q, want it disabled", tt.running, tt.wantDisabled, btn)
			}
			if btn := daemonButton(t, tt.wantEnabled, byAction[tt.wantEnabled]); disabled.MatchString(btn) {
				t.Errorf("RenderDaemonControl(Running=%t) form %s button = %q, want it enabled", tt.running, tt.wantEnabled, btn)
			}
		})
	}
}

func TestRenderDaemonControlHasNoScript(t *testing.T) {
	for _, running := range []bool{true, false} {
		out := renderDaemonControl(t, DaemonControlView{Running: running, CSRFToken: "tok-daemon"})
		// The controls must work with JS off, so the partial is the two
		// forms and nothing scripted.
		daemonForms(t, out)
		if m := daemonScriptRE.FindString(out); m != "" {
			t.Errorf("RenderDaemonControl(Running=%t) contains %q, want no script in the server-rendered partial", running, m)
		}
	}
}

func TestRenderDaemonControlEscapesToken(t *testing.T) {
	token := `tok"` + hostile
	out := renderDaemonControl(t, DaemonControlView{Running: true, CSRFToken: token})
	for _, raw := range []string{`value="tok"`, "<script>alert(1)</script>"} {
		if strings.Contains(out, raw) {
			t.Errorf("RenderDaemonControl() contains raw %q, want the token escaped in the attribute\nrendered: %s", raw, out)
		}
	}
	byAction := daemonForms(t, out)
	for action, body := range byAction {
		in := daemonCSRFRE.FindString(body)
		m := daemonAttrRE("value").FindStringSubmatch(in)
		if m == nil {
			t.Fatalf("RenderDaemonControl() form %s csrf input = %q, want a value attribute", action, in)
		}
		if got := html.UnescapeString(m[1]); got != token {
			t.Errorf("RenderDaemonControl() form %s csrf token = %q, want %q unescaped back", action, got, token)
		}
	}
}
