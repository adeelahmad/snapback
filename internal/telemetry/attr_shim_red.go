package telemetry

// Attr is a RED compile shim for S6-01/T3, which owns attr.go and the real
// type, its closed key set and its validation. Delete this file when attr.go
// lands; nothing here is production behaviour.
type Attr struct {
	Key   string
	Value string
}
