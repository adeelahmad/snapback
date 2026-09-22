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
