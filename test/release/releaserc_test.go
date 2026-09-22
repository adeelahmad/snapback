package release_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

const (
	releasercFile = ".releaserc.json"
	changelogFile = "CHANGELOG.md"

	pluginCommitAnalyzer = "@semantic-release/commit-analyzer"
	pluginNotesGenerator = "@semantic-release/release-notes-generator"
	pluginChangelog      = "@semantic-release/changelog"
	pluginGit            = "@semantic-release/git"
	pluginGithub         = "@semantic-release/github"
)

type releasercPlugin struct {
	name    string
	options map[string]any
}

type releaserc struct {
	raw     map[string]any
	plugins []releasercPlugin
}

// loadReleaserc decodes .releaserc.json; each plugin entry is either "name" or ["name", {options}].
func loadReleaserc(t *testing.T) releaserc {
	t.Helper()
	var raw map[string]any
	if err := json.Unmarshal([]byte(readRepoFile(t, releasercFile)), &raw); err != nil {
		t.Fatalf("decode %s: %v", releasercFile, err)
	}
	cfg := releaserc{raw: raw}
	entries, _ := raw["plugins"].([]any)
	for i, entry := range entries {
		switch v := entry.(type) {
		case string:
			cfg.plugins = append(cfg.plugins, releasercPlugin{name: v, options: map[string]any{}})
		case []any:
			if len(v) == 0 {
				t.Fatalf("%s plugins[%d]: empty plugin entry", releasercFile, i)
			}
			name, ok := v[0].(string)
			if !ok {
				t.Fatalf("%s plugins[%d]: first element is not a plugin name", releasercFile, i)
			}
			opts := map[string]any{}
			if len(v) > 1 {
				if m, ok := v[1].(map[string]any); ok {
					opts = m
				}
			}
			cfg.plugins = append(cfg.plugins, releasercPlugin{name: name, options: opts})
		default:
			t.Fatalf("%s plugins[%d]: unexpected entry type %T", releasercFile, i, entry)
		}
	}
	return cfg
}

func (c releaserc) pluginOptions(t *testing.T, name string) map[string]any {
	t.Helper()
	for _, p := range c.plugins {
		if p.name == name {
			return p.options
		}
	}
	t.Fatalf("%s: plugin %s not configured", releasercFile, name)
	return nil
}

func TestReleasercBranchMaster(t *testing.T) {
	cfg := loadReleaserc(t)
	got, ok := cfg.raw["branches"].([]any)
	if !ok || !reflect.DeepEqual(got, []any{"master"}) {
		t.Errorf("branches = %#v, want [\"master\"]", cfg.raw["branches"])
	}
}

func TestReleasercPluginsInOrder(t *testing.T) {
	cfg := loadReleaserc(t)
	var got []string
	for _, p := range cfg.plugins {
		got = append(got, p.name)
	}
	want := []string{pluginCommitAnalyzer, pluginNotesGenerator, pluginChangelog, pluginGit, pluginGithub}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("plugins = %v, want %v", got, want)
	}
}

func TestReleasercConventionalCommitsPreset(t *testing.T) {
	cfg := loadReleaserc(t)
	for _, name := range []string{pluginCommitAnalyzer, pluginNotesGenerator} {
		if got := cfg.pluginOptions(t, name)["preset"]; got != "conventionalcommits" {
			t.Errorf("%s preset = %#v, want \"conventionalcommits\"", name, got)
		}
	}
}

func TestReleasercChangelogAndGitAssets(t *testing.T) {
	cfg := loadReleaserc(t)
	if got := cfg.pluginOptions(t, pluginChangelog)["changelogFile"]; got != changelogFile {
		t.Errorf("%s changelogFile = %#v, want %q", pluginChangelog, got, changelogFile)
	}
	gitOpts := cfg.pluginOptions(t, pluginGit)
	assets, _ := gitOpts["assets"].([]any)
	hasChangelog := false
	for _, a := range assets {
		if a == changelogFile {
			hasChangelog = true
		}
	}
	if !hasChangelog {
		t.Errorf("%s assets = %#v, want it to contain %q", pluginGit, gitOpts["assets"], changelogFile)
	}
	msg, _ := gitOpts["message"].(string)
	if !strings.Contains(msg, "[skip ci]") {
		t.Errorf("%s message = %#v, want it to contain \"[skip ci]\"", pluginGit, gitOpts["message"])
	}
}

func TestReleasercGithubNoIssueWrites(t *testing.T) {
	cfg := loadReleaserc(t)
	opts := cfg.pluginOptions(t, pluginGithub)
	for _, key := range []string{"successComment", "failComment", "releasedLabels"} {
		v, present := opts[key]
		if b, isBool := v.(bool); !present || !isBool || b {
			t.Errorf("%s %s = %#v (present=%v), want JSON false", pluginGithub, key, v, present)
		}
	}
}

// findKeys reports every path in v where a map key equals one of keys.
func findKeys(v any, path string, keys map[string]bool) []string {
	var hits []string
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			p := path + "." + k
			if keys[k] {
				hits = append(hits, p)
			}
			hits = append(hits, findKeys(child, p, keys)...)
		}
	case []any:
		for _, child := range x {
			hits = append(hits, findKeys(child, path+"[]", keys)...)
		}
	}
	return hits
}

func TestReleasercTolerantOfLegacyCommits(t *testing.T) {
	cfg := loadReleaserc(t)
	if len(cfg.plugins) == 0 {
		t.Fatalf("%s: no plugins configured", releasercFile)
	}
	forbidden := map[string]bool{"releaseRules": true, "parserOpts": true}
	for _, p := range cfg.plugins {
		if hits := findKeys(p.options, p.name, forbidden); len(hits) > 0 {
			t.Errorf("%s: forbidden keys found (legacy commits must parse as no-release): %v", releasercFile, hits)
		}
	}
	if hits := findKeys(cfg.raw, "", forbidden); len(hits) > 0 {
		t.Errorf("%s: forbidden keys found at config level: %v", releasercFile, hits)
	}
}

func TestChangelogSeededHeader(t *testing.T) {
	text := readRepoFile(t, changelogFile)
	first := ""
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) != "" {
			first = strings.TrimSpace(line)
			break
		}
	}
	if first != "# Changelog" {
		t.Errorf("%s first non-empty line = %q, want \"# Changelog\"", changelogFile, first)
	}
	lower := strings.ToLower(text)
	for _, word := range []string{"production-ready", "cross-platform", "static", "finder-integrated"} {
		if strings.Contains(lower, word) {
			t.Errorf("%s contains honesty-gate word %q", changelogFile, word)
		}
	}
}
