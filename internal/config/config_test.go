package config_test

import (
	"testing"

	"github.com/nkootstra/cli/internal/config"
)

func TestResolveUsesEnvironmentVariables(t *testing.T) {
	env := map[string]string{
		"DEVCLI_OIDC_ISSUER_URL":           "http://localhost:8080",
		"DEVCLI_OIDC_CLIENT_ID":            "devcli",
		"DEVCLI_OIDC_SCOPES":               "openid profile,email",
		"DEVCLI_OIDC_REDIRECT_URL":         "http://127.0.0.1:8765/callback",
		"DEVCLI_OIDC_CALLBACK_LISTEN_ADDR": "127.0.0.1:8765",
		"DEVCLI_AUTH_STATE_PATH":           "/tmp/devcli-auth.json",
		"DEVCLI_MOCK_OIDC_ISSUER_URL":      "http://localhost:8080",
	}

	cfg := config.Resolve(config.Input{Env: env})

	if cfg.OIDCIssuerURL != "http://localhost:8080" {
		t.Fatalf("issuer = %q", cfg.OIDCIssuerURL)
	}
	if cfg.OIDCClientID != "devcli" {
		t.Fatalf("client id = %q", cfg.OIDCClientID)
	}
	if got, want := cfg.OIDCScopes, []string{"openid", "profile", "email"}; !equalStrings(got, want) {
		t.Fatalf("scopes = %#v, want %#v", got, want)
	}
	if cfg.OIDCRedirectURL != "http://127.0.0.1:8765/callback" {
		t.Fatalf("redirect = %q", cfg.OIDCRedirectURL)
	}
	if cfg.OIDCCallbackListenAddr != "127.0.0.1:8765" {
		t.Fatalf("listen addr = %q", cfg.OIDCCallbackListenAddr)
	}
	if cfg.AuthStatePath != "/tmp/devcli-auth.json" {
		t.Fatalf("state path = %q", cfg.AuthStatePath)
	}
	if cfg.MockOIDCIssuerURL != "http://localhost:8080" {
		t.Fatalf("mock issuer = %q", cfg.MockOIDCIssuerURL)
	}
}

func TestResolveFlagsOverrideEnvironment(t *testing.T) {
	cfg := config.Resolve(config.Input{
		Env: map[string]string{
			"DEVCLI_OIDC_ISSUER_URL": "http://env.example",
			"DEVCLI_OIDC_CLIENT_ID":  "env-client",
		},
		Overrides: config.Overrides{
			OIDCIssuerURL: "http://flag.example",
			OIDCClientID:  "flag-client",
		},
	})

	if cfg.OIDCIssuerURL != "http://flag.example" {
		t.Fatalf("issuer = %q", cfg.OIDCIssuerURL)
	}
	if cfg.OIDCClientID != "flag-client" {
		t.Fatalf("client id = %q", cfg.OIDCClientID)
	}
}

func TestValidateAuthReportsMissingSettings(t *testing.T) {
	err := config.ValidateAuth(config.Config{})
	if err == nil {
		t.Fatal("expected missing settings error")
	}
	if !config.IsMissingSetting(err, "DEVCLI_OIDC_ISSUER_URL") {
		t.Fatalf("expected issuer env var in error, got %v", err)
	}
	if !config.IsMissingSetting(err, "DEVCLI_OIDC_CLIENT_ID") {
		t.Fatalf("expected client env var in error, got %v", err)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
