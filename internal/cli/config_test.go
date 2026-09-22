package cli

import (
	"context"
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
