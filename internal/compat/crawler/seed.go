package crawler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Shape describes the seeded tree: Depth levels below the root, Fanout
// subdirectories per directory.
type Shape struct {
	Depth  int
	Fanout int
}

// DefaultShape returns the shape the crawler test seeds.
func DefaultShape() Shape {
	return Shape{Depth: 3, Fanout: 3}
}

// Seed builds a tree under root with a .snapshot symlink to linkTarget in
// every directory and returns the directories and links it created.
func Seed(root, linkTarget string, shape Shape) (dirs, links []string, err error) {
	if err := requireUnderTemp(root); err != nil {
		return nil, nil, err
	}
	if err := seedDir(root, linkTarget, shape.Depth, shape.Fanout, &dirs, &links); err != nil {
		return nil, nil, err
	}
	return dirs, links, nil
}

func requireUnderTemp(root string) error {
	abs, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve seed root %q: %w", root, err)
	}
	tmp, err := filepath.Abs(os.TempDir())
	if err != nil {
		return fmt.Errorf("resolve temp dir: %w", err)
	}
	rel, err := filepath.Rel(tmp, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("seed root %q is outside %q", root, tmp)
	}
	return nil
}

func seedDir(dir, linkTarget string, depth, fanout int, dirs, links *[]string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("seed\n"), 0o644); err != nil {
		return err
	}
	link := filepath.Join(dir, ".snapshot")
	if err := os.Symlink(linkTarget, link); err != nil {
		return err
	}
	*dirs = append(*dirs, dir)
	*links = append(*links, link)
	if depth == 0 {
		return nil
	}
	for i := range fanout {
		if err := seedDir(filepath.Join(dir, "d"+strconv.Itoa(i)), linkTarget, depth-1, fanout, dirs, links); err != nil {
			return err
		}
	}
	return nil
}
