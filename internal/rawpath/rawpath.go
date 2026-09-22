// Package rawpath encodes raw path bytes reversibly as JSON.
package rawpath

// Path holds raw path bytes, which need not be valid UTF-8.
type Path []byte

// MarshalJSON encodes valid UTF-8 as a JSON string and anything else as
// {"b64":"<standard base64>"}.
func (p Path) MarshalJSON() ([]byte, error) {
	panic("SUB-AGENT-TODO: T1 valid UTF-8 (incl. empty/nil) -> json.Marshal(string(p)); else {\"b64\":base64.StdEncoding}")
}

// UnmarshalJSON decodes a JSON string or a {"b64":...} object back to the
// exact bytes; null leaves the value unchanged.
func (p *Path) UnmarshalJSON(data []byte) error {
	panic("SUB-AGENT-TODO: T1 string -> bytes; object with only key b64 -> StdEncoding decode; null unchanged; other shape/unknown key/bad base64 -> error prefixed rawpath:")
}
