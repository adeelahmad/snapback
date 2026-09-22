package rawpath

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestMarshalUTF8AsString(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
	}{
		{name: "empty", in: []byte("")},
		{name: "nil", in: nil},
		{name: "ascii", in: []byte("docs")},
		{name: "space", in: []byte("a b")},
		{name: "unicode", in: []byte("Ünïcode/ß")},
		{name: "decomposed", in: []byte("é")},
		{name: "html", in: []byte("<&>")},
		{name: "quoted", in: []byte(`"quoted"`)},
		{name: "newline", in: []byte("line\nbreak")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(Path(tt.in))
			if err != nil {
				t.Fatalf("json.Marshal(Path(%q)) error = %v, want nil", tt.in, err)
			}
			if !bytes.HasPrefix(got, []byte(`"`)) {
				t.Fatalf("json.Marshal(Path(%q)) = %s, want a JSON string", tt.in, got)
			}
			var s string
			if err := json.Unmarshal(got, &s); err != nil {
				t.Fatalf("json.Unmarshal(%s) into string error = %v, want nil", got, err)
			}
			if s != string(tt.in) {
				t.Errorf("json.Unmarshal(%s) into string = %q, want %q", got, s, tt.in)
			}
		})
	}
}

func TestMarshalInvalidUTF8AsB64(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
	}{
		{name: "single", in: []byte("\xff")},
		{name: "suffix", in: []byte("docs/\xff\xfe")},
		{name: "continuation", in: []byte("a\x80b")},
		{name: "truncated", in: []byte("\xc3")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(Path(tt.in))
			if err != nil {
				t.Fatalf("json.Marshal(Path(%q)) error = %v, want nil", tt.in, err)
			}
			want := `{"b64":"` + base64.StdEncoding.EncodeToString(tt.in) + `"}`
			if string(got) != want {
				t.Errorf("json.Marshal(Path(%q)) = %s, want %s", tt.in, got, want)
			}
		})
	}
}

func TestUnmarshalShapes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		start Path
		want  []byte
	}{
		{name: "string", input: `"docs"`, want: []byte("docs")},
		{name: "b64 invalid utf8", input: `{"b64":"/w=="}`, want: []byte("\xff")},
		{name: "b64 valid utf8", input: `{"b64":"ZG9jcw=="}`, want: []byte("docs")},
		{name: "null keeps value", input: `null`, start: Path("keep"), want: []byte("keep")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.start
			if err := json.Unmarshal([]byte(tt.input), &got); err != nil {
				t.Fatalf("json.Unmarshal(%s) error = %v, want nil", tt.input, err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("json.Unmarshal(%s) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestUnmarshalRejects(t *testing.T) {
	inputs := []string{
		`42`,
		`[]`,
		`true`,
		`{}`,
		`{"b64":"***"}`,
		`{"b64":"/w==","x":1}`,
		`{"B64":"/w=="}`,
		`{"b64":7}`,
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			var p Path
			err := json.Unmarshal([]byte(input), &p)
			if err == nil {
				t.Fatalf("json.Unmarshal(%s) error = nil, want an error containing %q", input, "rawpath:")
			}
			if !strings.Contains(err.Error(), "rawpath:") {
				t.Errorf("json.Unmarshal(%s) error = %q, want it to contain %q", input, err, "rawpath:")
			}
		})
	}
}

func TestRoundTripInStruct(t *testing.T) {
	type record struct {
		P Path `json:"path"`
	}
	tests := []struct {
		name     string
		in       Path
		wantJSON string
	}{
		{name: "utf8", in: Path("docs")},
		{name: "invalid utf8", in: Path("\xff\xfe"), wantJSON: `"path":{"b64":`},
		{name: "empty", in: Path("")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(record{P: tt.in})
			if err != nil {
				t.Fatalf("json.Marshal(record{P: %q}) error = %v, want nil", tt.in, err)
			}
			if tt.wantJSON != "" && !strings.Contains(string(data), tt.wantJSON) {
				t.Errorf("json.Marshal(record{P: %q}) = %s, want it to contain %s", tt.in, data, tt.wantJSON)
			}
			var got record
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("json.Unmarshal(%s) error = %v, want nil", data, err)
			}
			if !bytes.Equal(got.P, tt.in) {
				t.Errorf("round trip of %q via %s = %q, want %q", tt.in, data, got.P, tt.in)
			}
		})
	}
}
