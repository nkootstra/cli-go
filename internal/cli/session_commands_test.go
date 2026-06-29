package cli_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nkootstra/cli/internal/auth"
	"github.com/nkootstra/cli/internal/cli"
	"github.com/nkootstra/cli/internal/testutil"
)

func TestLogoutWithoutStateSucceeds(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "auth.json")
	cmd := cli.NewRootCommand()

	stdout, _, err := testutil.Execute(cmd, "logout", "--auth-state-path", statePath)
	if err != nil {
		t.Fatalf("expected idempotent logout: %v", err)
	}
	if !strings.Contains(stdout, "No active session") {
		t.Fatalf("expected no active session message, got %q", stdout)
	}
}

func TestLogoutRemovesExistingState(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "auth.json")
	store := auth.NewFileStore(statePath)
	if err := store.Write(t.Context(), auth.State{Authenticated: true, Subject: "user-123"}); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	cmd := cli.NewRootCommand()
	stdout, _, err := testutil.Execute(cmd, "logout", "--auth-state-path", statePath)
	if err != nil {
		t.Fatalf("expected logout: %v", err)
	}
	if !strings.Contains(stdout, "Logged out") {
		t.Fatalf("expected logged out message, got %q", stdout)
	}

	got, err := store.Read(t.Context())
	if err != nil {
		t.Fatalf("read after logout: %v", err)
	}
	if got.Authenticated {
		t.Fatalf("expected state removed, got %#v", got)
	}
}

func TestWhoamiWithNoStateFailsWithoutIdentity(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "auth.json")
	cmd := cli.NewRootCommand()

	stdout, _, err := testutil.Execute(cmd, "whoami", "--auth-state-path", statePath)
	if err == nil {
		t.Fatal("expected whoami to fail without login")
	}
	if !strings.Contains(err.Error(), "no user is logged in") {
		t.Fatalf("expected unauthenticated error, got %v", err)
	}
	if strings.Contains(stdout, "Subject") {
		t.Fatalf("expected no identity output, got %q", stdout)
	}
}

func TestWhoamiPrintsIdentityWithoutTokens(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "auth.json")
	store := auth.NewFileStore(statePath)
	if err := store.Write(t.Context(), auth.State{
		Authenticated: true,
		Subject:       "user-123",
		Issuer:        "http://issuer.example",
		Audience:      "devcli",
		Email:         "person@example.com",
		Name:          "Test Person",
		ExpiresAt:     time.Now().Add(time.Hour),
		AccessToken:   "secret-access",
		IDToken:       "secret-id",
	}); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	cmd := cli.NewRootCommand()
	stdout, _, err := testutil.Execute(cmd, "whoami", "--auth-state-path", statePath)
	if err != nil {
		t.Fatalf("expected whoami: %v", err)
	}
	for _, want := range []string{"user-123", "person@example.com", "Test Person", "http://issuer.example"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected whoami output to contain %q:\n%s", want, stdout)
		}
	}
	for _, secret := range []string{"secret-access", "secret-id"} {
		if strings.Contains(stdout, secret) {
			t.Fatalf("whoami leaked token %q:\n%s", secret, stdout)
		}
	}
}

func TestWhoamiExpiredStateFails(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "auth.json")
	store := auth.NewFileStore(statePath)
	if err := store.Write(t.Context(), auth.State{
		Authenticated: true,
		Subject:       "user-123",
		ExpiresAt:     time.Now().Add(-time.Minute),
	}); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	cmd := cli.NewRootCommand()
	_, _, err := testutil.Execute(cmd, "whoami", "--auth-state-path", statePath)
	if err == nil || !strings.Contains(err.Error(), "session is expired") {
		t.Fatalf("expected expired session error, got %v", err)
	}
}
