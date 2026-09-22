package web

import (
	"net/url"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/webui"
)

// Decode builds a configuration from form values whose names are YAML key
// paths. Every value it cannot apply is reported as a field error instead of
// being dropped.
func Decode(v url.Values) (*config.Config, []config.FieldError) {
	return nil, nil
}

// Fields lists every configuration key of cfg as a renderable form field.
func Fields(cfg *config.Config) []webui.Field {
	return nil
}
