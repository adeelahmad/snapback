package webui

// HistoryView is the History page model.
type HistoryView struct {
	Chrome
	// Root and Path are the selected root and path; snapshot forms keep them.
	Root, Path      string
	Roots           []RootItem
	Timeline        []SnapshotTick
	LinkedDirs      []LinkedDir
	Entries         []Entry
	Versions        *VersionsPanel
	FileManagerPath string
}

// RootItem is one backup root in the roots panel.
type RootItem struct {
	Name, Path            string
	RepoState, MountState string
}

// LinkedDir is one linked directory of the selected root; URL opens its
// history.
type LinkedDir struct {
	Path, URL string
}

// SnapshotTick is one snapshot on the timeline; URL selects it.
type SnapshotTick struct {
	ID, Time    string
	Alias, Host string
	URL         string
	Warm        bool
	Selected    bool
}

// Entry is one row of the file list; State is ok, absent or failed.
type Entry struct {
	Name, Size, Modified string
	State                string
}

// VersionsPanel lists every occurrence of one file across snapshots.
type VersionsPanel struct {
	File   string
	Groups []VersionGroup
}

// VersionGroup is a run of occurrences, likely identical when flagged.
type VersionGroup struct {
	LikelyIdentical bool
	Occurrences     []Occurrence
}

// Occurrence is one snapshot's copy of the file.
type Occurrence struct {
	SnapshotID, Time, Size, Modified, DownloadURL, RestoreName string
}
