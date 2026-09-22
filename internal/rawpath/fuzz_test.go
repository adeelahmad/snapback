package rawpath

import (
	"bytes"
	"encoding/json"
	"testing"
	"unicode/utf8"
)

func FuzzRoundTrip(f *testing.F) {
	all := make([]byte, 256)
	for i := range all {
		all[i] = byte(i)
	}
	seeds := [][]byte{
		[]byte(""),
		[]byte("docs"),
		[]byte("\xff"),
		[]byte("Ünïcode"),
		[]byte("a\x00b"),
		[]byte(`{"b64":"x"}`),
		all,
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		data, err := json.Marshal(Path(b))
		if err != nil {
			t.Fatalf("json.Marshal(Path(%q)) error = %v, want nil", b, err)
		}
		isString := bytes.HasPrefix(data, []byte(`"`))
		if want := utf8.Valid(b); isString != want {
			t.Errorf("json.Marshal(Path(%q)) = %s, JSON string = %t, want %t", b, data, isString, want)
		}
		var got Path
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("json.Unmarshal(%s) error = %v, want nil", data, err)
		}
		if !bytes.Equal(got, b) {
			t.Errorf("round trip of %q via %s = %q, want %q", b, data, got, b)
		}
	})
}
