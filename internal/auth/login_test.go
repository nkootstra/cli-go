package auth_test

import (
	"testing"

	"github.com/nkootstra/cli/internal/auth"
)

func TestCallbackResultRejectsMismatchedState(t *testing.T) {
	_, err := auth.ValidateCallback("expected-state", "wrong-state", "code-123")
	if err == nil {
		t.Fatal("expected state mismatch to fail")
	}
}

func TestCallbackResultRequiresCode(t *testing.T) {
	_, err := auth.ValidateCallback("state", "state", "")
	if err == nil {
		t.Fatal("expected missing code to fail")
	}
}

func TestCallbackResultAcceptsMatchingStateAndCode(t *testing.T) {
	code, err := auth.ValidateCallback("state", "state", "code-123")
	if err != nil {
		t.Fatalf("expected callback to validate: %v", err)
	}
	if code != "code-123" {
		t.Fatalf("code = %q", code)
	}
}
