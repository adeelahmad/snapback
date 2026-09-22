package fsmode_test

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/fsmode"
)

func TestParseAcceptsOctalSpellings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want fs.FileMode
	}{
		{name: "leading zero", in: "0750", want: 0o750},
		{name: "bare octal", in: "750", want: 0o750},
		{name: "go prefix", in: "0o750", want: 0o750},
		{name: "file mode", in: "0640", want: 0o640},
		{name: "group readable dir", in: "0755", want: 0o755},
		{name: "world readable file", in: "0644", want: 0o644},
		{name: "owner only dir", in: "0700", want: 0o700},
		{name: "owner only file", in: "0600", want: 0o600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := fsmode.Parse(tt.in)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %#o, want %#o", tt.in, uint32(got), uint32(tt.want))
			}
		})
	}
}

func TestParseRejectsUnsafeOrMalformedModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
	}{
		{name: "set-uid", in: "4750"},
		{name: "set-gid", in: "2750"},
		{name: "sticky", in: "1777"},
		{name: "world writable dir", in: "0757"},
		{name: "world writable everything", in: "0777"},
		{name: "not octal", in: "abc"},
		{name: "digit out of range", in: "0999"},
		{name: "owner read only", in: "0400"},
		{name: "owner write only", in: "0200"},
		{name: "group only", in: "0050"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := fsmode.Parse(tt.in)
			if err == nil {
				t.Fatalf("Parse(%q) = %#o, want an error", tt.in, uint32(got))
			}
			if !strings.Contains(err.Error(), tt.in) {
				t.Errorf("Parse(%q) error %q does not name the offending value", tt.in, err)
			}
		})
	}
}

func TestParseRejectsEmptyString(t *testing.T) {
	t.Parallel()

	if _, err := fsmode.Parse(""); err == nil {
		t.Fatal(`Parse("") returned no error, want an error`)
	}
}

func TestFromUmaskComputesBothModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want fsmode.Modes
	}{
		{name: "group readable", in: "027", want: fsmode.Modes{Dir: 0o750, File: 0o640}},
		{name: "owner only", in: "077", want: fsmode.Modes{Dir: 0o700, File: 0o600}},
		{name: "world readable", in: "022", want: fsmode.Modes{Dir: 0o755, File: 0o644}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := fsmode.FromUmask(tt.in)
			if err != nil {
				t.Fatalf("FromUmask(%q) returned error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("FromUmask(%q) = {Dir:%#o, File:%#o}, want {Dir:%#o, File:%#o}",
					tt.in, uint32(got.Dir), uint32(got.File), uint32(tt.want.Dir), uint32(tt.want.File))
			}
		})
	}
}

func TestFromUmaskRejectsMalformedValues(t *testing.T) {
	t.Parallel()

	const in = "999"

	got, err := fsmode.FromUmask(in)
	if err == nil {
		t.Fatalf("FromUmask(%q) = {Dir:%#o, File:%#o}, want an error",
			in, uint32(got.Dir), uint32(got.File))
	}
	if !strings.Contains(err.Error(), in) {
		t.Errorf("FromUmask(%q) error %q does not name the offending value", in, err)
	}
}

func TestModesOrDefaultFillsOnlyZeroFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   fsmode.Modes
		want fsmode.Modes
	}{
		{name: "empty", in: fsmode.Modes{}, want: fsmode.Modes{Dir: 0o700, File: 0o600}},
		{name: "dir set", in: fsmode.Modes{Dir: 0o750}, want: fsmode.Modes{Dir: 0o750, File: 0o600}},
		{name: "file set", in: fsmode.Modes{File: 0o640}, want: fsmode.Modes{Dir: 0o700, File: 0o640}},
		{
			name: "both set",
			in:   fsmode.Modes{Dir: 0o755, File: 0o644},
			want: fsmode.Modes{Dir: 0o755, File: 0o644},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.in.OrDefault()
			if got != tt.want {
				t.Errorf("Modes{Dir:%#o, File:%#o}.OrDefault() = {Dir:%#o, File:%#o}, want {Dir:%#o, File:%#o}",
					uint32(tt.in.Dir), uint32(tt.in.File),
					uint32(got.Dir), uint32(got.File),
					uint32(tt.want.Dir), uint32(tt.want.File))
			}
		})
	}
}
