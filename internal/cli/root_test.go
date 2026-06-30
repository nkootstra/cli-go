package cli_test

import (
	"strings"
	"testing"

	"github.com/nkootstra/cli/internal/cli"
	"github.com/nkootstra/cli/internal/testutil"
	"github.com/nkootstra/cli/internal/version"
)

func TestRootHelpIncludesCommandDescriptions(t *testing.T) {
	cmd := cli.NewRootCommand()

	stdout, stderr, err := testutil.Execute(cmd, "--help")
	if err != nil {
		t.Fatalf("expected help to succeed: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}

	for _, want := range []string{
		"devcli",
		"login",
		"logout",
		"whoami",
		"doctor",
		"Authenticate with an OIDC provider",
		"Remove the saved login session",
		"Show the current logged-in identity",
		"Check devcli configuration and auth health",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected root help to contain %q:\n%s", want, stdout)
		}
	}
}

func TestUnknownCommandReturnsUsageError(t *testing.T) {
	cmd := cli.NewRootCommand()

	_, stderr, err := testutil.Execute(cmd, "logni")
	if err == nil {
		t.Fatal("expected unknown command to fail")
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Fatalf("expected unknown command error, got %q", stderr)
	}
}

func TestRootVersionFlags(t *testing.T) {
	oldVersion := version.Version
	oldCommit := version.Commit
	oldDate := version.Date
	version.Version = "v1.2.3"
	version.Commit = "abc123"
	version.Date = "2026-06-30T12:00:00Z"
	t.Cleanup(func() {
		version.Version = oldVersion
		version.Commit = oldCommit
		version.Date = oldDate
	})

	for _, flag := range []string{"--version", "-v"} {
		cmd := cli.NewRootCommand()

		stdout, stderr, err := testutil.Execute(cmd, flag)
		if err != nil {
			t.Fatalf("%s failed: %v", flag, err)
		}
		if stderr != "" {
			t.Fatalf("%s wrote stderr: %q", flag, stderr)
		}
		if !strings.Contains(stdout, "devcli v1.2.3 (abc123, 2026-06-30T12:00:00Z)") {
			t.Fatalf("%s version output = %q", flag, stdout)
		}
	}
}
