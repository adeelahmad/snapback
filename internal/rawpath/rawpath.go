// Package rawpath encodes raw path bytes reversibly as JSON.
package rawpath

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"unicode/utf8"
)

// Path holds raw path bytes, which need not be valid UTF-8.
type Path []byte

type b64Object struct {
	B64 string `json:"b64"`
}

// MarshalJSON encodes valid UTF-8 as a JSON string and anything else as
// {"b64":"<standard base64>"}.
func (p Path) MarshalJSON() ([]byte, error) {
	if utf8.Valid(p) {
		return json.Marshal(string(p))
	}
	return json.Marshal(b64Object{B64: base64.StdEncoding.EncodeToString(p)})
}

// UnmarshalJSON decodes a JSON string or a {"b64":...} object back to the
// exact bytes; null leaves the value unchanged.
func (p *Path) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("rawpath: %w", err)
		}
		*p = Path(s)
		return nil
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("rawpath: %w", err)
	}
	raw, ok := obj["b64"]
	if !ok || len(obj) != 1 {
		return errors.New(`rawpath: object must have exactly one key "b64"`)
	}
	var enc string
	if err := json.Unmarshal(raw, &enc); err != nil {
		return fmt.Errorf("rawpath: b64: %w", err)
	}
	b, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return fmt.Errorf("rawpath: b64: %w", err)
	}
	*p = Path(b)
	return nil
}
