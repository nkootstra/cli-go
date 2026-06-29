package doctor_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/nkootstra/cli/internal/auth"
	"github.com/nkootstra/cli/internal/config"
	"github.com/nkootstra/cli/internal/doctor"
)

func TestRunReportsMissingAuthConfig(t *testing.T) {
	results := doctor.Run(t.Context(), doctor.Input{
		Config: config.Config{},
		Store:  auth.NewFileStore(filepath.Join(t.TempDir(), "auth.json")),
	})

	result := findResult(t, results, "auth configuration")
	if result.Status != doctor.StatusFail {
		t.Fatalf("status = %s", result.Status)
	}
	if result.Remediation == "" {
		t.Fatal("expected remediation")
	}
}

func TestRunReportsReachableOIDCDiscovery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issuer":"` + r.Host + `"}`))
	}))
	defer server.Close()

	results := doctor.Run(t.Context(), doctor.Input{
		Config: config.Config{OIDCIssuerURL: server.URL, OIDCClientID: "devcli"},
		Store:  auth.NewFileStore(filepath.Join(t.TempDir(), "auth.json")),
	})

	result := findResult(t, results, "oidc discovery")
	if result.Status != doctor.StatusPass {
		t.Fatalf("status = %s detail=%s", result.Status, result.Detail)
	}
}

func TestRunReportsUnreachableOIDCDiscovery(t *testing.T) {
	results := doctor.Run(t.Context(), doctor.Input{
		Config: config.Config{OIDCIssuerURL: "http://127.0.0.1:1", OIDCClientID: "devcli"},
		Store:  auth.NewFileStore(filepath.Join(t.TempDir(), "auth.json")),
	})

	result := findResult(t, results, "oidc discovery")
	if result.Status != doctor.StatusFail {
		t.Fatalf("status = %s", result.Status)
	}
}

func TestRunNamesMockDiscoverySeparately(t *testing.T) {
	results := doctor.Run(t.Context(), doctor.Input{
		Config: config.Config{
			OIDCIssuerURL:     "http://127.0.0.1:1",
			MockOIDCIssuerURL: "http://127.0.0.1:2",
			OIDCClientID:      "devcli",
		},
		Store: auth.NewFileStore(filepath.Join(t.TempDir(), "auth.json")),
	})

	findResult(t, results, "oidc discovery")
	findResult(t, results, "mock oidc discovery")
}

func TestRunReportsMissingStateAsInfo(t *testing.T) {
	results := doctor.Run(t.Context(), doctor.Input{
		Config: config.Config{OIDCIssuerURL: "http://127.0.0.1:1", OIDCClientID: "devcli"},
		Store:  auth.NewFileStore(filepath.Join(t.TempDir(), "auth.json")),
	})

	result := findResult(t, results, "auth state")
	if result.Status != doctor.StatusInfo {
		t.Fatalf("status = %s", result.Status)
	}
}

func TestRunReportsCorruptStateAsFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(path, []byte("{bad-json"), 0o600); err != nil {
		t.Fatalf("write corrupt state: %v", err)
	}

	results := doctor.Run(t.Context(), doctor.Input{
		Config: config.Config{OIDCIssuerURL: "http://127.0.0.1:1", OIDCClientID: "devcli"},
		Store:  auth.NewFileStore(path),
	})

	result := findResult(t, results, "auth state")
	if result.Status != doctor.StatusFail {
		t.Fatalf("status = %s", result.Status)
	}
}

func findResult(t *testing.T, results []doctor.Result, name string) doctor.Result {
	t.Helper()
	for _, result := range results {
		if result.Name == name {
			return result
		}
	}
	t.Fatalf("missing result %q in %#v", name, results)
	return doctor.Result{}
}
