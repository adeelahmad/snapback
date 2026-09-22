package fidelity

import (
	"io/fs"
	"time"
)

// Fixtures returns the fixed fixture set handed to the resticfx generator.
func Fixtures() []Meta {
	return []Meta{
		{Path: "empty.txt", Size: 0, Mode: 0o644, MTime: time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC)},
		{Path: "small.txt", Size: 1024, Mode: 0o644, MTime: time.Date(2022, 1, 2, 3, 4, 7, 123456789, time.UTC)},
		{Path: "large.bin", Size: 3 << 20, Mode: 0o600, MTime: time.Date(1998, 3, 14, 9, 26, 53, 0, time.UTC)},
		{Path: "run.sh", Size: 64, Mode: 0o755, MTime: time.Date(2019, 11, 5, 8, 30, 0, 0, time.UTC)},
		{Path: "readonly.txt", Size: 128, Mode: 0o444, MTime: time.Date(2020, 2, 29, 23, 59, 59, 0, time.UTC)},
		{Path: "with space.txt", Size: 32, Mode: 0o644, MTime: time.Date(2023, 7, 15, 16, 45, 0, 0, time.UTC)},
		{Path: "café-日本.txt", Size: 48, Mode: 0o644, MTime: time.Date(2024, 4, 1, 10, 0, 0, 0, time.UTC)},
		{Path: "sub/nested.txt", Size: 16, Mode: 0o644, MTime: time.Date(2025, 9, 9, 9, 9, 9, 0, time.UTC)},
		{Path: "link-to-small", Mode: fs.ModeSymlink | 0o777, LinkTarget: "small.txt"},
	}
}
