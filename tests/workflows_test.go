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

// dockerStages splits a Dockerfile into its named build stages.
func dockerStages(body string) (map[string]string, []string) {
	stages := map[string]string{}
	var order []string
	var cur string
	var sb strings.Builder

	from := regexp.MustCompile(`(?i)^FROM\s+.*\sAS\s+(\S+)`)
	flush := func() {
		if cur != "" {
			stages[cur] = sb.String()
		}
	}
	for _, line := range strings.Split(body, "\n") {
		if m := from.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			flush()
			cur = m[1]
			order = append(order, cur)
			sb.Reset()
			continue
		}
		sb.WriteString(line + "\n")
	}
	flush()
	return stages, order
}

// TestDockerfileRuntimeStagesAgree guards the block each runtime stage repeats.
// Docker cannot share instructions between stages built on different bases, so
// ENV/WORKDIR/EXPOSE/HEALTHCHECK/ENTRYPOINT exist three times over and can drift
// apart silently -- a variant that boots but never reports healthy, or listens
// on the wrong port, looks perfectly fine right up until someone deploys it.
//
// The exec-form CMD is the load-bearing one. The distroless variant has no
// /bin/sh, so rewriting it as `HEALTHCHECK CMD ./golauth -healthcheck` would
// leave that image permanently unhealthy while the other two stayed green.
func TestDockerfileRuntimeStagesAgree(t *testing.T) {
	body, err := os.ReadFile("../Dockerfile")
	require.NoError(t, err)

	stages, order := dockerStages(string(body))
	for _, want := range []string{"dist-alpine", "dist-debian", "dist-distroless"} {
		require.Contains(t, stages, want, "runtime stage %s is missing", want)
	}

	required := []string{
		"ENV MIGRATION_SOURCE_URL=./migrations",
		"WORKDIR /app",
		"EXPOSE 8080",
		`CMD ["./golauth", "-healthcheck"]`,
		`ENTRYPOINT ["./golauth"]`,
	}
	for name, stage := range stages {
		if !strings.HasPrefix(name, "dist-") {
			continue
		}
		for _, r := range required {
			require.Contains(t, stage, r, "runtime stage %s is missing %q", name, r)
		}
		require.Regexp(t, `(?m)^USER \S`, stage,
			"runtime stage %s must drop to a non-root user", name)
	}

	// A bare `docker build .` -- which is what docker-compose and
	// `make build-image` do -- produces the last stage. That has to be the
	// variant published as `latest`.
	require.Equal(t, "dist-distroless", order[len(order)-1],
		"the default (last) stage must be the distroless one")
}

// TestDockerWorkflowCoversEveryVariant fails if a runtime stage exists in the
// Dockerfile but nothing builds it, which would publish two variants and
// silently drop the third.
func TestDockerWorkflowCoversEveryVariant(t *testing.T) {
	body, err := os.ReadFile("../.github/workflows/docker.yaml")
	require.NoError(t, err)
	text := string(body)

	dockerfile, err := os.ReadFile("../Dockerfile")
	require.NoError(t, err)
	stages, _ := dockerStages(string(dockerfile))

	for name := range stages {
		if !strings.HasPrefix(name, "dist-") {
			continue
		}
		require.Contains(t, text, "target: "+name,
			"the docker workflow never builds runtime stage %s", name)
	}
	require.Contains(t, text, "linux/arm64",
		"the default variant is meant to be multi-arch")
}

// TestCrossCompileChainIsWiredEndToEnd guards the three links that carry the
// target architecture from buildx down to the compiler. Break any one of them
// and the arm64 build still succeeds -- it just publishes an amd64 binary under
// an arm64 manifest, which nothing downstream notices until it fails to exec on
// a real arm64 host. This exact mistake was made and caught by hand once.
func TestCrossCompileChainIsWiredEndToEnd(t *testing.T) {
	dockerfile, err := os.ReadFile("../Dockerfile")
	require.NoError(t, err)
	df := string(dockerfile)

	require.Contains(t, df, "FROM --platform=$BUILDPLATFORM",
		"the builder must pin to the build platform and cross-compile, not be emulated")
	require.Contains(t, df, "ARG TARGETOS")
	require.Contains(t, df, "ARG TARGETARCH")
	require.Contains(t, df, "GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build",
		"the builder must pass buildx's target platform into the build")

	makefile, err := os.ReadFile("../Makefile")
	require.NoError(t, err)
	mk := string(makefile)

	require.Regexp(t, `(?m)^GOOS \?= linux`, mk,
		"GOOS must default with ?= so the Dockerfile's environment wins")
	require.Regexp(t, `(?m)^GOARCH \?= amd64`, mk,
		"GOARCH must default with ?= so the Dockerfile's environment wins")
	require.Contains(t, mk, "GOOS=$(GOOS) GOARCH=$(GOARCH) go build",
		"the build target must use the variables, not hardcode the arch")
}

// TestZAPSuppressionsAreJustified enforces the house rule from CONTRIBUTING.md
// mechanically: a muted DAST finding must carry a reason.
//
// A suppression file is where a security tool goes to die. One IGNORE with no
// explanation is indistinguishable from a real finding someone silenced to get
// a build green, and six months later nobody can tell which it was. Requiring a
// sentence makes that decision reviewable in the diff, where it belongs.
func TestZAPSuppressionsAreJustified(t *testing.T) {
	body, err := os.ReadFile("../.zap/rules.tsv")
	require.NoError(t, err)

	checked := 0
	for i, line := range strings.Split(string(body), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cols := strings.Split(line, "\t")
		require.Len(t, cols, 3,
			"line %d must be <rule id>\\t<action>\\t<justification>, got %q", i+1, line)
		require.Contains(t, []string{"IGNORE", "WARN", "FAIL"}, cols[1],
			"line %d: %q is not a ZAP action", i+1, cols[1])
		require.Greater(t, len(strings.Fields(cols[2])), 5,
			"rule %s is muted without a real reason; a few words is not a justification", cols[0])
		checked++
	}
	require.Positive(t, checked, "no rules were scanned")
}

// TestDASTWorkflowScansBothPersonas guards the property that makes the scan
// worth its ten minutes. An admin token may legitimately reach every route, so
// a scan that only runs as admin proves nothing about authorization; the plain
// user is the persona that can trip a broken-object-level-authorization bug on
// the ADMIN routes.
func TestDASTWorkflowScansBothPersonas(t *testing.T) {
	body, err := os.ReadFile("../.github/workflows/dast.yaml")
	require.NoError(t, err)
	text := string(body)

	require.Contains(t, text, "persona: [admin, user]",
		"an admin-only scan does not exercise authorization")
	require.Contains(t, text, "/health/ready",
		"the scan must wait on readiness, not sleep")
	require.Contains(t, text, "docs/openapi.yaml",
		"the scan is driven by the spec; without it ZAP cannot navigate a JSON API")
	require.Contains(t, text, "rules.tsv",
		"the justified-suppression file must be passed to ZAP")
}
