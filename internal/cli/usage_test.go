package cli

import (
	"strings"
	"testing"
)

func TestUsageString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		usage Usage
		want  string
	}{
		{
			name: "all blocks",
			usage: Usage{
				Synopsis: "seed [flags] DIR",
				Args:     "DIR\tthe directory to seed\nFILE\tthe manifest to read",
				Example:  "snapback seed /srv/data",
			},
			want: "Usage: snapback seed [flags] DIR\n" +
				"\n" +
				"Args:\n" +
				"  DIR\tthe directory to seed\n" +
				"  FILE\tthe manifest to read\n" +
				"\n" +
				"Example:\n" +
				"  snapback seed /srv/data\n",
		},
		{
			name:  "synopsis only",
			usage: Usage{Synopsis: "version"},
			want:  "Usage: snapback version\n",
		},
		{
			name: "no args block",
			usage: Usage{
				Synopsis: "status [flags]",
				Example:  "snapback status --json",
			},
			want: "Usage: snapback status [flags]\n" +
				"\n" +
				"Example:\n" +
				"  snapback status --json\n",
		},
		{
			name: "no example block",
			usage: Usage{
				Synopsis: "link [flags] DIR",
				Args:     "DIR\tthe directory to link",
			},
			want: "Usage: snapback link [flags] DIR\n" +
				"\n" +
				"Args:\n" +
				"  DIR\tthe directory to link\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.usage.String()
			if got != tt.want {
				t.Errorf("Usage.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUsageStringSingleTrailingNewline(t *testing.T) {
	t.Parallel()

	u := Usage{Synopsis: "seed [flags] DIR", Args: "DIR\tthe directory", Example: "snapback seed ."}
	got := u.String()
	if !strings.HasSuffix(got, "\n") || strings.HasSuffix(got, "\n\n") {
		t.Errorf("Usage.String() = %q, want exactly one trailing newline", got)
	}
}

func TestUsageStringDeterministic(t *testing.T) {
	t.Parallel()

	u := Usage{Synopsis: "snap [flags] DIR", Args: "DIR\tthe directory", Example: "snapback snap ."}
	first := u.String()
	for i := range 4 {
		if got := u.String(); got != first {
			t.Errorf("Usage.String() call %d = %q, want %q", i+2, got, first)
		}
	}
}
