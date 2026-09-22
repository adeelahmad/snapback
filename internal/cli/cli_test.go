package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

const testUsage = "usage: snapback version|help|COMMAND"

// newEnv returns an Env over two buffers with Getenv reading vars.
func newEnv(vars map[string]string) (Env, *bytes.Buffer, *bytes.Buffer) {
	var out, errb bytes.Buffer
	env := Env{
		Stdout:     &out,
		Stderr:     &errb,
		Getenv:     func(k string) string { return vars[k] },
		ConfigPath: "/c.yaml",
	}
	return env, &out, &errb
}

func noop(context.Context, Env, []string) int { return 0 }

func TestDispatchRunsNamedCommand(t *testing.T) {
	env, _, _ := newEnv(nil)
	var ran []string
	var gotArgs []string
	var gotEnv Env
	record := func(name string) func(context.Context, Env, []string) int {
		return func(_ context.Context, e Env, args []string) int {
			ran = append(ran, name)
			gotArgs = args
			gotEnv = e
			return 7
		}
	}
	cmds := []Command{
		{Name: "a", Summary: "first", Run: record("a")},
		{Name: "b", Summary: "second", Run: record("b")},
	}

	got := Dispatch(context.Background(), env, testUsage, cmds, []string{"b", "x", "--json"})

	if got != 7 {
		t.Errorf("Dispatch(b x --json) = %d, want 7", got)
	}
	if want := []string{"b"}; !slices.Equal(ran, want) {
		t.Errorf("Dispatch(b x --json) ran %q, want %q", ran, want)
	}
	if want := []string{"x", "--json"}; !slices.Equal(gotArgs, want) {
		t.Errorf("Dispatch(b x --json) passed args %q, want %q", gotArgs, want)
	}
	if gotEnv.Stdout != env.Stdout || gotEnv.Stderr != env.Stderr || gotEnv.ConfigPath != env.ConfigPath || gotEnv.Getenv == nil {
		t.Errorf("Dispatch(b x --json) passed Env %+v, want the caller's Env %+v", gotEnv, env)
	}
}

func TestDispatchUsageErrors(t *testing.T) {
	cmds := []Command{
		{Name: "a", Summary: "first", Run: noop},
		{Name: "b", Summary: "second", Run: noop},
	}
	tests := []struct {
		name string
		args []string
	}{
		{"nil", nil},
		{"empty", []string{}},
		{"unknown", []string{"bogus"}},
		{"wrong case", []string{"B"}},
		{"flag", []string{"--version"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, out, errb := newEnv(nil)

			got := Dispatch(context.Background(), env, testUsage, cmds, tt.args)

			if got != 2 {
				t.Errorf("Dispatch(%q) = %d, want 2", tt.args, got)
			}
			if out.Len() != 0 {
				t.Errorf("Dispatch(%q) stdout = %q, want empty", tt.args, out.String())
			}
			stderr := errb.String()
			if !strings.HasPrefix(stderr, testUsage) {
				t.Errorf("Dispatch(%q) stderr = %q, want prefix %q", tt.args, stderr, testUsage)
			}
			for _, c := range cmds {
				if !strings.Contains(stderr, c.Name) {
					t.Errorf("Dispatch(%q) stderr = %q, want it to list %q", tt.args, stderr, c.Name)
				}
			}
		})
	}
}

func TestDispatchHelpListsCommands(t *testing.T) {
	cmds := []Command{
		{Name: "zed", Summary: "last command", Run: noop},
		{Name: "alpha", Summary: "first command", Run: noop},
	}
	for _, arg := range []string{"help", "-h", "--help"} {
		t.Run(arg, func(t *testing.T) {
			env, out, errb := newEnv(nil)

			got := Dispatch(context.Background(), env, testUsage, cmds, []string{arg})

			if got != 0 {
				t.Errorf("Dispatch(%q) = %d, want 0", arg, got)
			}
			if errb.Len() != 0 {
				t.Errorf("Dispatch(%q) stderr = %q, want empty", arg, errb.String())
			}
			stdout := out.String()
			ia := strings.Index(stdout, "alpha")
			iz := strings.Index(stdout, "zed")
			if ia < 0 || iz < 0 || ia > iz {
				t.Errorf("Dispatch(%q) stdout = %q, want alpha listed before zed", arg, stdout)
			}
			for _, c := range cmds {
				line := lineContaining(stdout, c.Name)
				rest, ok := strings.CutPrefix(strings.TrimSpace(line), c.Name)
				if !ok || !strings.Contains(rest, c.Summary) {
					t.Errorf("Dispatch(%q) stdout line %q, want %q followed by %q", arg, line, c.Name, c.Summary)
				}
			}
		})
	}
}

func lineContaining(s, sub string) string {
	for line := range strings.SplitSeq(s, "\n") {
		if strings.Contains(line, sub) {
			return line
		}
	}
	return ""
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 0},
		{"usage", &UsageError{}, 2},
		{"wrapped usage", fmt.Errorf("parse: %w", &UsageError{Msg: "bad flag"}), 2},
		{"coded", errcode.New(errcode.LinkConflict, "links.ensure", errors.New("exists")), 1},
		{"plain", errors.New("x"), 1},
	}
	for _, tt := range tests {
		if got := ExitCode(tt.err); got != tt.want {
			t.Errorf("ExitCode(%s) = %d, want %d", tt.name, got, tt.want)
		}
	}
}
