package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

func TestConfigValidate(t *testing.T) {
	invalid := &config.ValidationError{Fields: []config.FieldError{{Path: "roots[0].local_path", Msg: "must be absolute"}}}
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{name: "valid", wantCode: 0},
		{name: "invalid", err: invalid, wantCode: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPaths []string
			d := Deps{LoadConfig: func(path string) (config.Config, error) {
				gotPaths = append(gotPaths, path)
				if tt.err != nil {
					return config.Config{}, tt.err
				}
				return fixtureCfg("/r", "/s/state", "/s/history", "/s/backend"), nil
			}}
			env, out, _ := newEnv(nil)

			code := ConfigCommand(d, nil).Run(context.Background(), env, []string{"validate", "--json"})

			if code != tt.wantCode {
				t.Errorf("config validate --json = %d, want %d", code, tt.wantCode)
			}
			if want := []string{"/c.yaml"}; !slices.Equal(gotPaths, want) {
				t.Errorf("config validate LoadConfig paths = %q, want %q", gotPaths, want)
			}
			if tt.err == nil {
				if !strings.Contains(out.String(), "configuration valid") {
					t.Errorf("config validate stdout = %q, want it to contain %q", out.String(), "configuration valid")
				}
				return
			}
			e := decodeEnvelope(t, out.Bytes())
			if e.Code != "invalid_configuration" {
				t.Errorf("config validate code = %q, want %q", e.Code, "invalid_configuration")
			}
			if !strings.Contains(e.Error, "roots[0].local_path") {
				t.Errorf("config validate error = %q, want it to name %q", e.Error, "roots[0].local_path")
			}
		})
	}
}

func TestConfigShowRedacted(t *testing.T) {
	cfg := fixtureCfg("/r", "/s/state", "/s/history", "/s/backend")
	cfg.Repositories[0].Environment = map[string]string{"RESTIC_PASSWORD": "s3cr3t"}
	d := Deps{LoadConfig: func(string) (config.Config, error) { return cfg, nil }}
	env, out, errb := newEnv(nil)

	code := ConfigCommand(d, nil).Run(context.Background(), env, []string{"show"})

	if code != 0 {
		t.Fatalf("config show = %d, want 0 (stderr %q)", code, errb.String())
	}
	got := out.String()
	if got == "" {
		t.Fatal("config show stdout is empty, want the configuration")
	}
	if !strings.Contains(got, "RESTIC_PASSWORD") {
		t.Errorf("config show stdout = %q, want it to keep the key RESTIC_PASSWORD", got)
	}
	if strings.Contains(got, "s3cr3t") {
		t.Errorf("config show stdout = %q, want the value s3cr3t redacted", got)
	}
	want, err := config.Marshal(config.Redact(&cfg))
	if err != nil {
		t.Fatalf("config.Marshal: %v", err)
	}
	if got != string(want) {
		t.Errorf("config show stdout = %q, want %q", got, want)
	}
}

func TestConfigDelegatesToFallback(t *testing.T) {
	for _, args := range [][]string{{}, {"--file", "x.yaml"}} {
		t.Run("fallback "+strings.Join(args, " "), func(t *testing.T) {
			loads := 0
			d := Deps{LoadConfig: func(string) (config.Config, error) { loads++; return config.Config{}, nil }}
			var ran bool
			var gotArgs []string
			fb := func(_ context.Context, _ Env, a []string) int {
				ran = true
				gotArgs = a
				return 0
			}
			env, _, _ := newEnv(nil)

			code := ConfigCommand(d, fb).Run(context.Background(), env, args)

			if code != 0 {
				t.Errorf("config %q = %d, want 0", args, code)
			}
			if !ran || !slices.Equal(gotArgs, args) {
				t.Errorf("config %q fallback ran=%v with %q, want it run with %q", args, ran, gotArgs, args)
			}
			if loads != 0 {
				t.Errorf("config %q called LoadConfig %d times, want 0", args, loads)
			}
		})
		t.Run("nil fallback "+strings.Join(args, " "), func(t *testing.T) {
			d := Deps{LoadConfig: func(string) (config.Config, error) { return config.Config{}, nil }}
			env, _, errb := newEnv(nil)

			code := ConfigCommand(d, nil).Run(context.Background(), env, args)

			if code != 2 {
				t.Errorf("config %q with nil fallback = %d, want 2", args, code)
			}
			if !strings.Contains(errb.String(), "validate|show") {
				t.Errorf("config %q with nil fallback stderr = %q, want it to name %q", args, errb.String(), "validate|show")
			}
		})
	}
}

// configPathJSON is the shape of `config path --json`: a bare {"path":…}, not
// the ok/data envelope the other config subcommands write.
type configPathJSON struct {
	Path string `json:"path"`
}

func decodeConfigPathJSON(t *testing.T, b []byte) configPathJSON {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var v configPathJSON
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("decode config path json %q: %v", b, err)
	}
	return v
}

func failIfLoadConfigCalled(t *testing.T) func(string) (config.Config, error) {
	t.Helper()
	return func(path string) (config.Config, error) {
		t.Fatalf("LoadConfig(%q) called, want config path to never read the config file", path)
		return config.Config{}, nil
	}
}

func TestConfigPathPlainAndJSON(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "plain", args: []string{"path"}},
		{name: "json", args: []string{"path", "--json"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Deps{LoadConfig: failIfLoadConfigCalled(t)}
			env, out, errb := newEnv(nil)

			code := ConfigCommand(d, nil).Run(context.Background(), env, tt.args)

			if code != 0 {
				t.Fatalf("config %q = %d, want 0 (stderr %q)", tt.args, code, errb.String())
			}
			if errb.Len() != 0 {
				t.Errorf("config %q stderr = %q, want empty", tt.args, errb.String())
			}
			if tt.name == "json" {
				got := decodeConfigPathJSON(t, out.Bytes())
				if got.Path != env.ConfigPath {
					t.Errorf("config %q path = %q, want %q", tt.args, got.Path, env.ConfigPath)
				}
				return
			}
			if want := env.ConfigPath + "\n"; out.String() != want {
				t.Errorf("config %q stdout = %q, want %q", tt.args, out.String(), want)
			}
		})
	}
}

func TestConfigPathUsesConfigOverride(t *testing.T) {
	env, out, errb := newEnv(nil)
	env.ConfigPath = "/srv/override/config.yaml"
	d := Deps{LoadConfig: failIfLoadConfigCalled(t)}

	code := ConfigCommand(d, nil).Run(context.Background(), env, []string{"path"})

	if code != 0 {
		t.Fatalf("config path (--config override) = %d, want 0 (stderr %q)", code, errb.String())
	}
	if want := env.ConfigPath + "\n"; out.String() != want {
		t.Errorf("config path (--config override) stdout = %q, want %q", out.String(), want)
	}
}

func TestConfigPathDefaultFromXDGAndHome(t *testing.T) {
	tests := []struct {
		name string
		want func(xdg, home string) string
	}{
		{
			name: "XDG_CONFIG_HOME set",
			want: func(xdg, home string) string { return filepath.Join(xdg, "snapback", "config.yaml") },
		},
		{
			name: "HOME fallback",
			want: func(xdg, home string) string { return filepath.Join(home, ".config", "snapback", "config.yaml") },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xdg := t.TempDir()
			home := t.TempDir()
			t.Setenv("HOME", home)
			if tt.name == "XDG_CONFIG_HOME set" {
				t.Setenv("XDG_CONFIG_HOME", xdg)
			} else {
				t.Setenv("XDG_CONFIG_HOME", "")
			}
			env, out, errb := newEnv(nil)
			env.ConfigPath = ""
			d := Deps{LoadConfig: failIfLoadConfigCalled(t)}

			code := ConfigCommand(d, nil).Run(context.Background(), env, []string{"path"})

			if code != 0 {
				t.Fatalf("config path (%s) = %d, want 0 (stderr %q)", tt.name, code, errb.String())
			}
			if want := tt.want(xdg, home) + "\n"; out.String() != want {
				t.Errorf("config path (%s) stdout = %q, want %q", tt.name, out.String(), want)
			}
		})
	}
}

func TestConfigPathNeverReadsOrCreatesFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "nested", "config.yaml")
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		t.Fatalf("stat(%q) = %v before run, want it to not exist", cfgPath, err)
	}
	env, out, errb := newEnv(nil)
	env.ConfigPath = cfgPath
	d := Deps{LoadConfig: failIfLoadConfigCalled(t)}

	code := ConfigCommand(d, nil).Run(context.Background(), env, []string{"path"})

	if code != 0 {
		t.Fatalf("config path (missing file) = %d, want 0 (stderr %q)", code, errb.String())
	}
	if want := cfgPath + "\n"; out.String() != want {
		t.Errorf("config path (missing file) stdout = %q, want %q", out.String(), want)
	}
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		t.Errorf("stat(%q) = %v after run, want it to still not exist (never created)", cfgPath, err)
	}
}

func TestConfigHelpPrintsUsage(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "-h", args: []string{"-h"}},
		{name: "--help", args: []string{"--help"}},
		{name: "validate -h", args: []string{"validate", "-h"}},
		{name: "show --help", args: []string{"show", "--help"}},
		{name: "path -h", args: []string{"path", "-h"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Deps{LoadConfig: failIfLoadConfigCalled(t)}
			env, out, errb := newEnv(nil)

			code := ConfigCommand(d, nil).Run(context.Background(), env, tt.args)

			if code != 0 {
				t.Fatalf("config %q = %d, want 0 (stderr %q)", tt.args, code, errb.String())
			}
			if out.Len() != 0 {
				t.Errorf("config %q stdout = %q, want empty", tt.args, out.String())
			}
			got := errb.String()
			for _, want := range []string{"Usage: snapback config", "Args:", "Example:"} {
				if !strings.Contains(got, want) {
					t.Errorf("config %q stderr = %q, want it to contain %q", tt.args, got, want)
				}
			}
		})
	}
}

func TestConfigUnknownFlagPrintsUsage(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "validate", args: []string{"validate", "--bogus"}},
		{name: "path", args: []string{"path", "--bogus"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Deps{LoadConfig: failIfLoadConfigCalled(t)}
			env, _, errb := newEnv(nil)

			code := ConfigCommand(d, nil).Run(context.Background(), env, tt.args)

			if code != 2 {
				t.Fatalf("config %q = %d, want 2", tt.args, code)
			}
			got := errb.String()
			for _, want := range []string{"Usage: snapback config", "bogus"} {
				if !strings.Contains(got, want) {
					t.Errorf("config %q stderr = %q, want it to contain %q", tt.args, got, want)
				}
			}
		})
	}
}
