package webui

// SetupView is the Setup page model.
type SetupView struct {
	Chrome
	ResticPaths    []string
	RclonePaths    []string
	RepoURI        string
	CredentialFile string
	Roots          []string
	Errors         []string

	// DetectedHost and DetectedPrefix carry what machine detection inferred:
	// the repository's snapshot hostname and the prefix map it derived.
	DetectedHost   Control
	DetectedPrefix Control
}

// Control is one form control on a page, shaped like the controls.html
// partials: a field path, its label, its value and its help text.
type Control struct {
	Path    string
	Label   string
	Value   string
	Help    string
	Options []ControlOption
}

// ControlOption is one option of a select, chips or row control.
type ControlOption struct {
	Value string
	Label string
}

// ConfigView is the Configuration page model.
type ConfigView struct {
	Chrome
	Revision        string
	Roots           []string
	Filters         []string
	Exclusions      []string
	SeedPaths       []string
	DiscoveryMode   string
	CacheDir        string
	RefreshInterval string
	Errors          []string
	Saved           bool
}
