package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestWorkflowActionsPinnedBySHA fails if any GitHub Actions `uses:` reference
// in .github/workflows is not pinned to a full 40-hex commit sha. A tag can be
// moved to point at new code; a sha cannot. This is the guard behind the
// "every action is pinned by sha" rule from the supply-chain hardening.
func TestWorkflowActionsPinnedBySHA(t *testing.T) {
	files, err := filepath.Glob("../.github/workflows/*.y*ml")
	require.NoError(t, err)
	require.NotEmpty(t, files, "no workflow files found")

	uses := regexp.MustCompile(`(?m)^\s*(?:-\s*)?uses:\s*["']?([^"'\s#]+)`)
	pinned := regexp.MustCompile(`^[\w.\-/]+@[0-9a-f]{40}$`)

	checked := 0
	for _, f := range files {
		body, err := os.ReadFile(f)
		require.NoError(t, err)

		for _, m := range uses.FindAllStringSubmatch(string(body), -1) {
			ref := m[1]
			// Local composite actions and container refs are not tag-pinnable.
			if strings.HasPrefix(ref, "./") || strings.HasPrefix(ref, "docker://") {
				continue
			}
			require.Regexp(t, pinned, ref,
				"%s: action %q must be pinned by a 40-char commit sha", filepath.Base(f), ref)
			checked++
		}
	}
	require.Positive(t, checked, "no external actions were scanned")
}

// TestQualityWorkflowHasSupplyChainGates is a crude presence check: the plan's
// two non-negotiable gates -- govulncheck and a coverage floor -- must stay in
// the quality workflow.
func TestQualityWorkflowHasSupplyChainGates(t *testing.T) {
	body, err := os.ReadFile("../.github/workflows/quality.yaml")
	require.NoError(t, err)
	text := string(body)
	require.Contains(t, text, "govulncheck", "the vulnerability scan step is gone")
	require.Contains(t, text, "COVERAGE_FLOOR", "the coverage gate is gone")
}

// TestSASTWorkflowKeepsItsTeeth guards the two properties of the CodeQL job that
// are easy to "simplify" away and silent when they break.
//
// build-mode: manual is load-bearing. This tree does not compile from a clean
// checkout -- the mocks are gitignored and generated -- so autobuild produces a
// green run that analysed nothing. Nothing else in CI would notice.
//
// security-extended is the difference between the queries every repository gets
// and the ones worth paying triage for on an auth server.
func TestSASTWorkflowKeepsItsTeeth(t *testing.T) {
	body, err := os.ReadFile("../.github/workflows/sast.yaml")
	require.NoError(t, err)
	text := string(body)
	require.Contains(t, text, "build-mode: manual",
		"CodeQL autobuild cannot build this tree; the mocks are generated")
	require.Contains(t, text, "queries: security-extended",
		"the extended query suite is the point of running CodeQL here")
	require.Contains(t, text, "go generate -v ./...",
		"the manual build must generate mocks or it analyses an empty tree")
	require.Contains(t, text, "gosec", "the gosec SARIF job is gone")
}
