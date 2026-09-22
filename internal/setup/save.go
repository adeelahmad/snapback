package setup

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"go.yaml.in/yaml/v3"

	"github.com/adeelahmad/snapback/internal/config"
)

// Save writes cfg to path through the configuration layer, so setup and the
// web form share one serialiser, one validator and one atomic 0600 write.
// Top-level keys of an existing file that Config does not model are kept.
func Save(cfg *config.Config, path string) error {
	config.ApplyDefaults(cfg)

	var rev config.Revision
	old, err := os.ReadFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		old = nil
	case err != nil:
		return fmt.Errorf("setup.save: %w", err)
	default:
		// Load reports the file's revision even when it does not parse, which
		// is what lets a file carrying unknown keys be replaced safely.
		_, rev, _ = config.Load(path)
	}

	if _, err := config.Save(path, cfg, rev); err != nil {
		return err
	}
	if len(old) == 0 {
		return nil
	}
	return mergeUnknownKeys(path, old)
}

// mergeUnknownKeys rewrites path with the top-level keys of old that the
// freshly written configuration does not carry.
func mergeUnknownKeys(path string, old []byte) error {
	var oldDoc, newDoc yaml.Node
	if err := yaml.Unmarshal(old, &oldDoc); err != nil {
		return fmt.Errorf("setup.save: previous config: %w", err)
	}
	cur, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("setup.save: %w", err)
	}
	if err := yaml.Unmarshal(cur, &newDoc); err != nil {
		return fmt.Errorf("setup.save: saved config: %w", err)
	}
	oldMap, newMap := mappingOf(&oldDoc), mappingOf(&newDoc)
	if oldMap == nil || newMap == nil {
		return nil
	}

	known := make(map[string]bool, len(newMap.Content)/2)
	for i := 0; i+1 < len(newMap.Content); i += 2 {
		known[newMap.Content[i].Value] = true
	}
	added := false
	for i := 0; i+1 < len(oldMap.Content); i += 2 {
		if known[oldMap.Content[i].Value] {
			continue
		}
		newMap.Content = append(newMap.Content, oldMap.Content[i], oldMap.Content[i+1])
		added = true
	}
	if !added {
		return nil
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(newMap); err != nil {
		return fmt.Errorf("setup.save: merge: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("setup.save: merge: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("setup.save: %w", err)
	}
	return nil
}

// mappingOf returns n's mapping node, or nil if n does not hold one.
func mappingOf(n *yaml.Node) *yaml.Node {
	if n.Kind == yaml.DocumentNode && len(n.Content) == 1 {
		n = n.Content[0]
	}
	if n.Kind != yaml.MappingNode {
		return nil
	}
	return n
}
