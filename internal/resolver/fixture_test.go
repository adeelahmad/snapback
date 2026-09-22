package resolver

import (
	"strings"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// The §20 metadata fixture shared by the T4 to T6 tests.

var (
	id1 = provider.SnapshotID(strings.Repeat("1", 64))
	id2 = provider.SnapshotID(strings.Repeat("2", 64))
	id3 = provider.SnapshotID(strings.Repeat("3", 64))
	id4 = provider.SnapshotID(strings.Repeat("4", 64))
	id5 = provider.SnapshotID(strings.Repeat("5", 64))
	id6 = provider.SnapshotID(strings.Repeat("6", 64))
	id7 = provider.SnapshotID(strings.Repeat("7", 64))
)

func fixtureTime(month time.Month, day int) time.Time {
	return time.Date(2026, month, day, 10, 0, 0, 0, time.UTC)
}

// s1 is a full backup of the project on the Linux host.
func s1() provider.Snapshot {
	return provider.Snapshot{
		ID: id1, Time: fixtureTime(time.September, 1), Hostname: "linuxbox",
		Paths: []string{"/home/alex/project"}, Tags: []string{"nightly"},
	}
}

// s2 is a macOS snapshot of the docs subdirectory only.
func s2() provider.Snapshot {
	return provider.Snapshot{
		ID: id2, Time: fixtureTime(time.September, 2), Hostname: "macbook",
		Paths: []string{"/Users/alex/project/docs"}, Tags: []string{"snapback:adhoc"},
	}
}

// s3 is a full backup of the project on the Linux host.
func s3() provider.Snapshot {
	return provider.Snapshot{
		ID: id3, Time: fixtureTime(time.September, 3), Hostname: "linuxbox",
		Paths: []string{"/home/alex/project"}, Tags: []string{"nightly"},
	}
}

// s4 is a backup recorded with a relative path.
func s4() provider.Snapshot {
	return provider.Snapshot{
		ID: id4, Time: fixtureTime(time.August, 31), Hostname: "linuxbox",
		Paths: []string{"project"}, Tags: []string{"nightly"},
	}
}

// s5 is an unrelated snapshot of another host.
func s5() provider.Snapshot {
	return provider.Snapshot{
		ID: id5, Time: fixtureTime(time.September, 4), Hostname: "otherhost",
		Paths: []string{"/srv/data"}, Tags: []string{"nightly"},
	}
}

// s6 is a two-path backup set.
func s6() provider.Snapshot {
	return provider.Snapshot{
		ID: id6, Time: fixtureTime(time.August, 30), Hostname: "linuxbox",
		Paths: []string{"/home/alex/project", "/etc"}, Tags: []string{"nightly", "system"},
	}
}

// s7 shares s3's time and records s6's path set out of order with a duplicate.
func s7() provider.Snapshot {
	return provider.Snapshot{
		ID: id7, Time: fixtureTime(time.September, 3), Hostname: "linuxbox",
		Paths: []string{"/etc", "/home/alex/project", "/etc"}, Tags: []string{"nightly", "system"},
	}
}

// fixtureSnaps returns a fresh, deliberately unsorted fixture slice.
func fixtureSnaps() []provider.Snapshot {
	return []provider.Snapshot{s5(), s1(), s4(), s7(), s2(), s6(), s3()}
}

func fixtureRules() []PrefixRule {
	return []PrefixRule{
		{Hostname: "linuxbox", SourcePath: "/home/alex/project", TreePrefix: "/home/alex/project"},
		{Hostname: "macbook", SourcePath: "/Users/alex/project", TreePrefix: "/Users/alex/project"},
		{Hostname: "linuxbox", SourcePath: "project", TreePrefix: "project"},
	}
}

func fixtureRoots() []RootSpec {
	return []RootSpec{
		{ID: "home", LocalPath: "/home/alex/project"},
		{ID: "nested", LocalPath: "/home/alex/project/docs"},
	}
}
