package history

import (
	"encoding/json"
	"testing"
)

// wantOfflineCode is the SPEC error code a directory reports while its
// repository is unreachable.
const wantOfflineCode = "repository_unavailable"

// infoFields decodes info.json into raw fields, so tests can tell an absent
// field from an empty one.
func infoFields(t *testing.T, dir map[string]entry) map[string]json.RawMessage {
	t.Helper()
	data, _ := readInfo(t, dir)
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("json.Unmarshal(info.json) error = %v, data %q", err, data)
	}
	return fields
}

// stringField returns the string value of fields[name], or "" when absent.
func stringField(t *testing.T, fields map[string]json.RawMessage, name string) string {
	t.Helper()
	raw, ok := fields[name]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("info.json %s = %s, want a JSON string", name, raw)
	}
	return s
}

func offlineInput(d Dir) Input {
	return Input{
		BackendMountDir: backendDir,
		Dirs:            []Dir{d},
		Repos:           map[string]RepoState{d.RepoID: StateFailed},
	}
}

func TestOfflineRepoInfoJSONStatesRepositoryUnavailable(t *testing.T) {
	d := goldenDir()
	fields := infoFields(t, under(mustBuild(t, offlineInput(d)), dirPrefix(d)))

	if got, want := stringField(t, fields, "state"), stateUnavailable; got != want {
		t.Errorf("Build(repo %s failed) info.json state = %q, want %q", d.RepoID, got, want)
	}
	if got := stringField(t, fields, "code"); got != wantOfflineCode {
		t.Errorf("Build(repo %s failed) info.json code = %q, want %q", d.RepoID, got, wantOfflineCode)
	}
}

func TestOfflineListingNotEmptyReady(t *testing.T) {
	empty := emptyDir()
	emptyFields := infoFields(t, under(mustBuild(t, Input{BackendMountDir: backendDir, Dirs: []Dir{empty}}), dirPrefix(empty)))

	offline := emptyDir()
	offlineFields := infoFields(t, under(mustBuild(t, offlineInput(offline)), dirPrefix(offline)))

	gotState := stringField(t, offlineFields, "state")
	if emptyState := stringField(t, emptyFields, "state"); gotState == emptyState {
		t.Errorf("Build(offline, no snapshots) info.json state = %q, want different from empty-ready %q", gotState, emptyState)
	}
	gotCode := stringField(t, offlineFields, "code")
	if emptyCode := stringField(t, emptyFields, "code"); gotCode == emptyCode {
		t.Errorf("Build(offline, no snapshots) info.json code = %q, want a failure code distinct from empty-ready %q", gotCode, emptyCode)
	}
	if gotCode != wantOfflineCode {
		t.Errorf("Build(offline, no snapshots) info.json code = %q, want %q", gotCode, wantOfflineCode)
	}
}

func TestHealthyEmptyInfoJSONReadyWithoutCode(t *testing.T) {
	d := emptyDir()
	in := Input{BackendMountDir: backendDir, Dirs: []Dir{d}, Repos: map[string]RepoState{d.RepoID: StateReady}}
	fields := infoFields(t, under(mustBuild(t, in), dirPrefix(d)))

	if got, want := stringField(t, fields, "state"), stateOK; got != want {
		t.Errorf("Build(repo ready, no snapshots) info.json state = %q, want %q", got, want)
	}
	if raw, ok := fields["code"]; ok {
		t.Errorf("Build(repo ready, no snapshots) info.json code = %s, want absent", raw)
	}
}
