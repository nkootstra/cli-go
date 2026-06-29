package cli_test

import (
	"strings"
	"testing"

	"github.com/nkootstra/cli/internal/cli"
	"github.com/nkootstra/cli/internal/testutil"
)

func TestLoginHelpIncludesAuthFlags(t *testing.T) {
	cmd := cli.NewRootCommand()

	stdout, stderr, err := testutil.Execute(cmd, "login", "--help")
	if err != nil {
		t.Fatalf("expected help to succeed: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
	for _, want := range []string{
		"Authenticate with an OIDC provider",
		"--oidc-issuer-url",
		"--oidc-client-id",
		"--oidc-scopes",
		"--oidc-redirect-url",
		"--auth-state-path",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected login help to contain %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "--mock-oidc-issuer-url") {
		t.Fatalf("login help should not contain doctor-only mock flag:\n%s", stdout)
	}
}

func TestLoginMissingConfigFailsBeforeBrowser(t *testing.T) {
	cmd := cli.NewRootCommand()

	_, stderr, err := testutil.Execute(cmd, "login")
	if err == nil {
		t.Fatal("expected login to fail without required config")
	}
	if !strings.Contains(err.Error(), "DEVCLI_OIDC_ISSUER_URL") || !strings.Contains(err.Error(), "DEVCLI_OIDC_CLIENT_ID") {
		t.Fatalf("expected missing env vars in error, got %v", err)
	}
	if strings.Contains(stderr, "Open this URL") {
		t.Fatalf("login should not print auth URL before validating config: %q", stderr)
	}
}
