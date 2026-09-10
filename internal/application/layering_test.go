package application_test

import (
	"os/exec"
	"strings"
	"testing"
)

// TestApplicationDoesNotImportInfra is the architecture guard for the dependency
// rule: the use-case layer must not depend on the delivery layer. It asks the
// toolchain for the full (non-test) dependency closure of internal/application
// and fails, naming the offenders, if any of them live under internal/infra.
//
// This is cheaper than a convention nobody enforces: it caught the historical
// internal/application/token -> internal/infra/api/controller/model import.
func TestApplicationDoesNotImportInfra(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps",
		"github.com/golauth/golauth/internal/application/...").CombinedOutput()
	if err != nil {
		t.Fatalf("go list failed: %v\n%s", err, out)
	}

	const forbidden = "github.com/golauth/golauth/internal/infra/"
	var offenders []string
	for _, dep := range strings.Fields(string(out)) {
		if strings.HasPrefix(dep, forbidden) {
			offenders = append(offenders, dep)
		}
	}

	if len(offenders) > 0 {
		t.Fatalf("internal/application must not import internal/infra, but its dependency closure contains:\n  %s",
			strings.Join(offenders, "\n  "))
	}
}
