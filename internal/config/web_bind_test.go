package config

import (
	"path/filepath"
	"reflect"
	"testing"
)

// webYAML returns a minimal config whose web section is the given lines.
func webYAML(t *testing.T, tmp, web string) []byte {
	t.Helper()
	return minimalYAML(tmp, web, "", "")
}

func TestWebBindRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))

	data := webYAML(t, tmp, "web:\n  bind: 0.0.0.0:7373\n  allowed_origins:\n    - https://x.example\n")

	c, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(web bind) = %v, want nil error", err)
	}
	if got, want := c.Web.Bind, "0.0.0.0:7373"; got != want {
		t.Errorf("Web.Bind = %q, want %q", got, want)
	}
	if got, want := c.Web.AllowedOrigins, []string{"https://x.example"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Web.AllowedOrigins = %#v, want %#v", got, want)
	}

	out, err := Marshal(c)
	if err != nil {
		t.Fatalf("Marshal(c) = %v, want nil error", err)
	}
	back, err := Parse(out)
	if err != nil {
		t.Fatalf("Parse(Marshal(c)) = %v, want nil error", err)
	}
	if !reflect.DeepEqual(back.Web, c.Web) {
		t.Errorf("Parse(Marshal(c)).Web = %#v, want %#v", back.Web, c.Web)
	}
}

// TestWebBindDefaultsToListen pins the defaulting rule: an absent bind takes
// the listen value, applied by ApplyDefaults like the other derived fields.
func TestWebBindDefaultsToListen(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))

	c, err := Parse(webYAML(t, tmp, "web:\n  listen: 127.0.0.1:7373\n"))
	if err != nil {
		t.Fatalf("Parse(web listen only) = %v, want nil error", err)
	}
	if got, want := c.Web.Bind, "127.0.0.1:7373"; got != want {
		t.Errorf("Web.Bind = %q, want the listen value %q", got, want)
	}

	cases := []struct {
		name string
		web  Web
		want string
	}{
		{"empty bind takes listen", Web{Listen: "127.0.0.1:9999"}, "127.0.0.1:9999"},
		{"explicit bind is kept", Web{Listen: "127.0.0.1:7373", Bind: "0.0.0.0:7373"}, "0.0.0.0:7373"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Web: tt.web}
			ApplyDefaults(c)
			if got := c.Web.Bind; got != tt.want {
				t.Errorf("ApplyDefaults: Web.Bind = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestWebBindAllowedOriginsValidation pins that a bad origin is a field-level
// error on web.allowed_origins[0] and that no origins at all stays valid.
func TestWebBindAllowedOriginsValidation(t *testing.T) {
	tests := []struct {
		name    string
		origins []string
		want    string // the field path, or "" when the config must stay valid
	}{
		{"absent", nil, ""},
		{"empty list", []string{}, ""},
		{"https origin", []string{"https://x.example"}, ""},
		{"no scheme", []string{"x.example"}, "web.allowed_origins[0]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			c.Web.AllowedOrigins = tt.origins

			err := Validate(c)
			if tt.want == "" {
				if err != nil {
					t.Errorf("Validate(%s) = %v, want nil", tt.name, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate(%s) = nil, want a *ValidationError", tt.name)
			}
			fields := validationFields(t, err)
			if findField(fields, tt.want) == nil {
				t.Errorf("Validate(%s) fields = %+v, want one on %s", tt.name, fields, tt.want)
			}
		})
	}
}
