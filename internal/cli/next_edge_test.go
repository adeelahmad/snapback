package cli

import (
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// callNext runs f and turns a panic into a test failure, so a panicking
// next-step line fails by assertion instead of taking the whole test binary
// down with it.
func callNext(t *testing.T, what string, f func() string) (out string, ok bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s panicked with %v, want a next-step line", what, r)
			out, ok = "", false
		}
	}()
	return f(), true
}

// assertOneLine checks that got is a single "next: " line with no raw control
// character in it, which is the property every next-step line must hold
// whatever the user-controlled text inside it looks like.
func assertOneLine(t *testing.T, what, got string) {
	t.Helper()
	if !strings.HasPrefix(got, "next: ") {
		t.Errorf("%s = %q, want a line starting with %q", what, got, "next: ")
	}
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("%s = %q, want a line ending in a newline", what, got)
	}
	if n := strings.Count(got, "\n"); n != 1 {
		t.Errorf("%s = %q, contains %d newlines, want exactly 1", what, got, n)
	}
	body := strings.TrimSuffix(got, "\n")
	for _, r := range body {
		if unicode.IsControl(r) {
			t.Errorf("%s = %q, contains the raw control character %q", what, got, r)
			break
		}
	}
	if strings.TrimSpace(strings.TrimPrefix(body, "next: ")) == "" {
		t.Errorf("%s = %q, want a non-empty next step after the prefix", what, got)
	}
}

// TestNextEdgeNeverPanics pins that Next never panics on user-controlled text.
// A path can legally hold a newline or a tab, and printing a hint must not take
// the command down.
func TestNextEdgeNeverPanics(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{name: "newline", command: "ls /a\nb/.snapshot"},
		{name: "tab", command: "ls /a\tb/.snapshot"},
		{name: "empty", command: ""},
		{name: "whitespace", command: "  \t "},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			what := "Next(" + strconv.Quote(test.command) + ")"
			got, ok := callNext(t, what, func() string { return Next(test.command) })
			if !ok {
				return
			}
			assertOneLine(t, what, got)
		})
	}
}

// TestNextEdgePlainCommandUnchanged pins that the common hint keeps its exact
// wording: quoting is for odd text only, never for the everyday line.
func TestNextEdgePlainCommandUnchanged(t *testing.T) {
	const command = "ls /home/u/work/.snapshot"
	what := "Next(" + strconv.Quote(command) + ")"
	got, ok := callNext(t, what, func() string { return Next(command) })
	if !ok {
		return
	}
	if want := "next: " + command + "\n"; got != want {
		t.Errorf("%s = %q, want %q", what, got, want)
	}
}

// TestNextEdgePathQuoting pins the quoting rule for the path argument: %q-style
// quoting when, and only when, the path holds a control character or a space.
func TestNextEdgePathQuoting(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "plain",
			path: "/home/u/work/.snapshot",
			want: "next: ls /home/u/work/.snapshot\n",
		},
		{
			name: "newline",
			path: "/home/u/a\nb/.snapshot",
			want: "next: ls " + strconv.Quote("/home/u/a\nb/.snapshot") + "\n",
		},
		{
			name: "tab",
			path: "/home/u/a\tb/.snapshot",
			want: "next: ls " + strconv.Quote("/home/u/a\tb/.snapshot") + "\n",
		},
		{
			name: "space",
			path: "/home/u/My Work/.snapshot",
			want: "next: ls " + strconv.Quote("/home/u/My Work/.snapshot") + "\n",
		},
		{
			name: "leading dash",
			path: "-rf/.snapshot",
			want: "next: ls -rf/.snapshot\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			what := "NextPath(\"ls\", " + strconv.Quote(test.path) + ")"
			got, ok := callNext(t, what, func() string { return NextPath("ls", test.path) })
			if !ok {
				return
			}
			if got != test.want {
				t.Errorf("%s = %q, want %q", what, got, test.want)
			}
			assertOneLine(t, what, got)
		})
	}
}

// TestNextEdgeEmptyPath pins that an empty path still yields one usable line
// rather than a panic or a dangling "next: ls".
func TestNextEdgeEmptyPath(t *testing.T) {
	what := "NextPath(\"ls\", \"\")"
	got, ok := callNext(t, what, func() string { return NextPath("ls", "") })
	if !ok {
		return
	}
	assertOneLine(t, what, got)
	if strings.HasSuffix(strings.TrimSuffix(got, "\n"), " ") {
		t.Errorf("%s = %q, want no dangling separator", what, got)
	}
	if want := "next: ls " + strconv.Quote("") + "\n"; got != want {
		t.Errorf("%s = %q, want %q", what, got, want)
	}
}
