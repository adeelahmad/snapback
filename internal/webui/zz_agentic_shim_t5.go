// agentic:shim
package webui

// HistoryView is the History page view model (shim for S3-14 T5 RED).
type HistoryView struct {
	Chrome
	Roots           []RootItem
	Timeline        []SnapshotTick
	Entries         []Entry
	Versions        *VersionsPanel
	FileManagerPath string
}

// RootItem is one backup root in the roots panel.
type RootItem struct {
	Name, Path            string
	RepoState, MountState string
}

// SnapshotTick is one snapshot on the timeline.
type SnapshotTick struct {
	ID, Time string
	Warm     bool
	Selected bool
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
