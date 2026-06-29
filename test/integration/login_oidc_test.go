package integration_test

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nkootstra/cli/internal/auth"
	"github.com/nkootstra/cli/internal/config"
)

func TestOIDCLoginAgainstMockServer(t *testing.T) {
	if os.Getenv("DEVCLI_RUN_INTEGRATION") != "1" {
		t.Skip("set DEVCLI_RUN_INTEGRATION=1 and start oauth2-mock-server to run this test")
	}

	cfg := config.Resolve(config.Input{})
	cfg.AuthStatePath = filepath.Join(t.TempDir(), "auth.json")

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	store := auth.NewFileStore(cfg.AuthStatePath)
	state, err := auth.Login(ctx, auth.LoginOptions{
		Config:      cfg,
		Store:       store,
		StateValue:  "integration-state",
		NonceValue:  "integration-nonce",
		WaitTimeout: 20 * time.Second,
		OpenURL: func(authURL string) error {
			return followAuthorizeRedirect(authURL)
		},
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if !state.Authenticated {
		t.Fatalf("expected authenticated state, got %#v", state)
	}
	if state.Subject == "" {
		t.Fatalf("expected subject in state: %#v", state)
	}
}

func followAuthorizeRedirect(authURL string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if strings.Contains(req.URL.String(), "127.0.0.1:8765/callback") {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	resp, err := client.Get(authURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location, err := resp.Location()
		if err != nil {
			return err
		}
		if _, err := url.Parse(location.String()); err != nil {
			return err
		}
		callbackResp, err := http.Get(location.String())
		if err != nil {
			return err
		}
		defer callbackResp.Body.Close()
		return nil
	}
	return nil
}
