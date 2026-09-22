// agentic:shim
package webui

// SetupView is a compile shim for the S3-14 T4 Setup page view model.
type SetupView struct {
	Chrome
	ResticPaths    []string
	RclonePaths    []string
	RepoURI        string
	CredentialFile string
	Roots          []string
	Errors         []string
}

// ConfigView is a compile shim for the S3-14 T4 Configuration page view model.
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
