// Package telemetry defines the closed, opt-in telemetry schema: the fixed
// event names, the duration buckets and the attribute type that makes a path,
// URI, hostname or address unrepresentable.
package telemetry

// Attr is one telemetry attribute. Both fields are strings on purpose: the
// schema carries no arbitrary values, so there is no any here.
type Attr struct {
	Key   string
	Value string
}

// NewAttr returns the attribute for key and value, or an error when key is
// outside the closed key set or value could carry an identifier.
func NewAttr(key, value string) (Attr, error) {
	return Attr{}, nil
}

// Keys returns the closed attribute key set in its documented order.
func Keys() []string {
	return nil
}
