package cli_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/nkootstra/cli/internal/cli"
	"github.com/nkootstra/cli/internal/testutil"
)

func TestDoctorHelpIncludesDescriptionAndFlags(t *testing.T) {
	cmd := cli.NewRootCommand()

	stdout, _, err := testutil.Execute(cmd, "doctor", "--help")
	if err != nil {
		t.Fatalf("expected help: %v", err)
	}
	for _, want := range []string{
		"Check devcli configuration and auth health",
		"--oidc-issuer-url",
		"--oidc-client-id",
		"--auth-state-path",
		"--mock-oidc-issuer-url",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected doctor help to contain %q:\n%s", want, stdout)
		}
	}
}

func TestDoctorFailsWhenAnyCheckFails(t *testing.T) {
	cmd := cli.NewRootCommand()

	stdout, _, err := testutil.Execute(cmd, "doctor", "--auth-state-path", filepath.Join(t.TempDir(), "auth.json"))
	if err == nil {
		t.Fatal("expected doctor to fail with missing config")
	}
	if !strings.Contains(stdout, "FAIL auth configuration") {
		t.Fatalf("expected failed config check, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "DEVCLI_OIDC_ISSUER_URL") {
		t.Fatalf("expected env var remediation, got:\n%s", stdout)
	}
}
