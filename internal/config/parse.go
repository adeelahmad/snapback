package config

// Parse strictly decodes YAML config data, applies defaults and validates it.
func Parse(data []byte) (*Config, error) {
	panic("SUB-AGENT-TODO: strict yaml decode (KnownFields) into a defaults-filled Config, apply post-decode defaults, then return Validate's error or the config")
}

// Marshal encodes c as YAML that Parse reads back to an equal Config.
func Marshal(c *Config) ([]byte, error) {
	panic("SUB-AGENT-TODO: yaml encode c deterministically so Parse(Marshal(c)) round-trips and re-marshal is byte-identical")
}
