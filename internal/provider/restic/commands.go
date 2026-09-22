package restic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// metadataTimeout bounds each metadata command (Validate, List, Probe).
const metadataTimeout = 2 * time.Minute

// Validate checks that the repository is reachable and returns its identity.
func (p *Provider) Validate(ctx context.Context) (provider.Identity, error) {
	ctx, cancel := context.WithTimeout(ctx, metadataTimeout)
	defer cancel()
	stdout, stderr, err := p.runner.Run(ctx, p.opts.Binary, p.validateArgs(), p.childEnv())
	if err != nil {
		return provider.Identity{}, classify("validate", err, stderr, p.secrets())
	}
	var cfg struct {
		ID      string `json:"id"`
		Version int    `json:"version"`
	}
	if err := json.Unmarshal(stdout, &cfg); err != nil {
		return provider.Identity{}, fmt.Errorf("restic: parse cat config: %w", err)
	}
	if !provider.SnapshotID(cfg.ID).Valid() {
		return provider.Identity{}, fmt.Errorf("restic: repository id %q is not 64 lowercase hex", cfg.ID)
	}
	return provider.Identity{RepoID: cfg.ID, Version: cfg.Version}, nil
}

// List returns every snapshot in the repository.
func (p *Provider) List(ctx context.Context) ([]provider.Snapshot, error) {
	ctx, cancel := context.WithTimeout(ctx, metadataTimeout)
	defer cancel()
	stdout, stderr, err := p.runner.Run(ctx, p.opts.Binary, p.listArgs(), p.childEnv())
	if err != nil {
		return nil, classify("list", err, stderr, p.secrets())
	}
	var raw []struct {
		ID       provider.SnapshotID `json:"id"`
		Time     time.Time           `json:"time"`
		Hostname string              `json:"hostname"`
		Tags     []string            `json:"tags"`
		Paths    []string            `json:"paths"`
	}
	if err := json.Unmarshal(stdout, &raw); err != nil {
		return nil, fmt.Errorf("restic: parse snapshots: %w", err)
	}
	snaps := make([]provider.Snapshot, 0, len(raw))
	for _, r := range raw {
		if !r.ID.Valid() {
			return nil, fmt.Errorf("restic: snapshot id %q is not 64 lowercase hex", r.ID)
		}
		snaps = append(snaps, provider.Snapshot{ID: r.ID, Time: r.Time, Hostname: r.Hostname, Tags: r.Tags, Paths: r.Paths})
	}
	return snaps, nil
}

// Snap backs up req.Path and returns the new snapshot's ID.
func (p *Provider) Snap(ctx context.Context, req provider.SnapRequest) (provider.SnapshotID, error) {
	if !filepath.IsAbs(req.Path) {
		return "", errors.New("restic: snap path must be absolute")
	}
	if req.Host == "" {
		return "", errors.New("restic: snap host is required")
	}
	stdout, stderr, err := p.runner.Run(ctx, p.opts.Binary, p.snapArgs(req), p.childEnv())
	if err != nil {
		return "", classify("snap", err, stderr, p.secrets())
	}
	sc := bufio.NewScanner(bytes.NewReader(stdout))
	for sc.Scan() {
		var msg struct {
			MessageType string              `json:"message_type"`
			SnapshotID  provider.SnapshotID `json:"snapshot_id"`
		}
		if json.Unmarshal(sc.Bytes(), &msg) != nil || msg.MessageType != "summary" {
			continue
		}
		if !msg.SnapshotID.Valid() {
			return "", fmt.Errorf("restic: backup snapshot id %q is not 64 lowercase hex", msg.SnapshotID)
		}
		return msg.SnapshotID, nil
	}
	return "", errors.New("restic: backup output has no summary")
}
