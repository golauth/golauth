package application_test

import (
	"os/exec"
	"strings"
	"testing"
)

// TestApplicationDoesNotImportInfra is the architecture guard for the dependency
// rule: the use-case layer must not depend on the delivery layer. It asks the
// toolchain for the full (non-test) dependency closure of pkg/application and
// fails, naming the offenders, if any of them live under pkg/infra.
//
// This is cheaper than a convention nobody enforces: it caught the historical
// pkg/application/token -> pkg/infra/api/controller/model import.
func TestApplicationDoesNotImportInfra(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps",
		"github.com/golauth/golauth/pkg/application/...").CombinedOutput()
	if err != nil {
		t.Fatalf("go list failed: %v\n%s", err, out)
	}

	const forbidden = "github.com/golauth/golauth/pkg/infra/"
	var offenders []string
	for _, dep := range strings.Fields(string(out)) {
		if strings.HasPrefix(dep, forbidden) {
			offenders = append(offenders, dep)
		}
	}

	if len(offenders) > 0 {
		t.Fatalf("pkg/application must not import pkg/infra, but its dependency closure contains:\n  %s",
			strings.Join(offenders, "\n  "))
	}
}
