package webui

// StatusView is the Status page model; nil pointers mean not measured.
type StatusView struct {
	Chrome
	Mounts            []MountStatus
	LastRefresh       *string
	Errors            []string
	EligibleSnapshots *int
	ManagedLinks      *int
	Prewarm           string
	DiscoveryMode     string
	ThrottleEvents    *int
	Integrations      []IntegrationState
}

// MountStatus is one mount and its state.
type MountStatus struct {
	Name  string
	State string
}

// IntegrationState is one integration and its state.
type IntegrationState struct {
	Name  string
	State string
}

// IntegrationsView is the Integrations page model.
type IntegrationsView struct {
	Chrome
	ShellHook      map[string]string
	ServiceState   string
	ServiceInstall string
}
