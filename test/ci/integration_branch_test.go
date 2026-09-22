package ci_test

import (
	"slices"
	"strings"
	"testing"
)

const commitlintWorkflowPath = ".github/workflows/commitlint.yml"

// integrationBranches are the branches whose pushes and pull requests must run CI.
var integrationBranches = []string{"master", "stage-2", "stage-5"}

func TestCITriggerBranchesCoverIntegrationBranches(t *testing.T) {
	text := readCI(t)
	for _, event := range []string{"push", "pull_request"} {
		branches := triggerBranches(text, event)
		if len(branches) == 0 {
			t.Errorf("%s: on.%s has no branches filter", ciWorkflowPath, event)
			continue
		}
		for _, want := range integrationBranches {
			if !slices.Contains(branches, want) {
				t.Errorf("%s: on.%s.branches %v is missing %q", ciWorkflowPath, event, branches, want)
			}
		}
	}
}

func TestCommitlintTriggerBranchesCoverIntegrationBranches(t *testing.T) {
	text := readRepoFile(t, commitlintWorkflowPath)
	for _, event := range []string{"push", "pull_request"} {
		if !strings.Contains(text, event+":") {
			continue
		}
		branches := triggerBranches(text, event)
		if len(branches) == 0 {
			continue // no branch filter: the event already covers every branch
		}
		for _, want := range integrationBranches {
			if !slices.Contains(branches, want) {
				t.Errorf("%s: on.%s.branches %v is missing %q", commitlintWorkflowPath, event, branches, want)
			}
		}
	}
}
