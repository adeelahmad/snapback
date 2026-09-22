package cli

import (
	"bytes"
	"testing"
)

func TestNext(t *testing.T) {
	const command = "ls /p/.snapshot"
	got := Next(command)
	want := "next: ls /p/.snapshot\n"
	if got != want {
		t.Errorf("Next(%q) = %q, want %q", command, got, want)
	}
}

func TestNextPanicsOnUnrunnableCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{name: "empty", command: ""},
		{name: "whitespace", command: "  \t "},
		{name: "newline", command: "a\nb"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("Next(%q) returned normally, want panic", test.command)
				}
			}()
			_ = Next(test.command)
		})
	}
}

func TestWriteNext(t *testing.T) {
	const command = "ls /p/.snapshot"
	var buf bytes.Buffer
	if err := WriteNext(&buf, command); err != nil {
		t.Fatalf("WriteNext(buf, %q) = %v, want nil", command, err)
	}
	got := buf.String()
	want := "next: ls /p/.snapshot\n"
	if got != want {
		t.Errorf("WriteNext(buf, %q) wrote %q, want %q", command, got, want)
	}
}
