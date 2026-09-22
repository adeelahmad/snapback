package webui

// Load parses the templates from override, or from the embedded files when
// override is empty.
func Load(override string) (*Pages, error) {
	if override != "" {
		return loadOverride(override)
	}
	return load(embedded)
}

// loadOverride loads templates and assets from a directory or a dist.zip.
func loadOverride(override string) (*Pages, error) {
	panic("SUB-AGENT-TODO: T7 Load(dir) and Load(file.zip) with validation and zip-slip rejection: stat override; a directory becomes os.DirFS, a .zip opens with zip.OpenReader and any entry that is absolute or escapes the root is an error; require templates/layout.html and assets/; then load(fsys); any other path is an error")
}
