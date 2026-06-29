package auth_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/nkootstra/cli/internal/auth"
)

func TestFileStoreMissingStateReturnsUnauthenticated(t *testing.T) {
	store := auth.NewFileStore(filepath.Join(t.TempDir(), "auth.json"))

	state, err := store.Read(t.Context())
	if err != nil {
		t.Fatalf("expected no low-level error, got %v", err)
	}
	if state.Authenticated {
		t.Fatalf("expected unauthenticated state, got %#v", state)
	}
}

func TestFileStoreWriteReadAndDelete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "auth.json")
	store := auth.NewFileStore(path)
	expires := time.Now().Add(time.Hour).UTC().Truncate(time.Second)

	want := auth.State{
		Authenticated: true,
		Subject:       "user-123",
		Issuer:        "http://issuer.example",
		Audience:      "devcli",
		Email:         "person@example.com",
		Name:          "Test Person",
		ExpiresAt:     expires,
		AccessToken:   "access-token",
		IDToken:       "id-token",
	}

	if err := store.Write(t.Context(), want); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	got, err := store.Read(t.Context())
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if got.Subject != want.Subject || got.Issuer != want.Issuer || got.Email != want.Email || !got.ExpiresAt.Equal(expires) {
		t.Fatalf("state mismatch: got %#v want %#v", got, want)
	}

	if err := store.Delete(t.Context()); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	got, err = store.Read(t.Context())
	if err != nil {
		t.Fatalf("read after delete failed: %v", err)
	}
	if got.Authenticated {
		t.Fatalf("expected unauthenticated after delete, got %#v", got)
	}
}

func TestFileStoreDeleteAbsentStateSucceeds(t *testing.T) {
	store := auth.NewFileStore(filepath.Join(t.TempDir(), "auth.json"))

	if err := store.Delete(t.Context()); err != nil {
		t.Fatalf("expected idempotent delete, got %v", err)
	}
}

func TestFileStoreCorruptStateReturnsClearError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(path, []byte("{not-json"), 0o600); err != nil {
		t.Fatalf("write corrupt fixture: %v", err)
	}

	_, err := auth.NewFileStore(path).Read(t.Context())
	if !errors.Is(err, auth.ErrCorruptState) {
		t.Fatalf("expected ErrCorruptState, got %v", err)
	}
}

func TestStateExpired(t *testing.T) {
	if !(auth.State{Authenticated: true, ExpiresAt: time.Now().Add(-time.Minute)}).Expired(time.Now()) {
		t.Fatal("expected past expiry to be expired")
	}
	if (auth.State{Authenticated: true, ExpiresAt: time.Now().Add(time.Minute)}).Expired(time.Now()) {
		t.Fatal("expected future expiry to be usable")
	}
	if (auth.State{}).Expired(time.Now()) {
		t.Fatal("expected unauthenticated state not to be expired")
	}
}

func TestFileStoreUsesRestrictivePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix mode bits are not reliable on Windows")
	}

	path := filepath.Join(t.TempDir(), "auth.json")
	store := auth.NewFileStore(path)
	if err := store.Write(t.Context(), auth.State{Authenticated: true, Subject: "user"}); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("permissions = %o, want 600", got)
	}
}
