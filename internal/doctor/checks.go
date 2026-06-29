package doctor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nkootstra/cli/internal/auth"
	"github.com/nkootstra/cli/internal/config"
)

type Status string

const (
	StatusPass Status = "PASS"
	StatusFail Status = "FAIL"
	StatusInfo Status = "INFO"
)

type Result struct {
	Name        string
	Status      Status
	Detail      string
	Remediation string
}

type Input struct {
	Config config.Config
	Store  auth.Store
	Client *http.Client
}

func Run(ctx context.Context, input Input) []Result {
	client := input.Client
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Second}
	}

	results := []Result{
		checkAuthConfig(input.Config),
		checkAuthState(ctx, input.Store),
	}
	if input.Config.OIDCIssuerURL != "" {
		results = append(results, checkDiscovery(ctx, client, input.Config.OIDCIssuerURL))
	}
	if input.Config.MockOIDCIssuerURL != "" && input.Config.MockOIDCIssuerURL != input.Config.OIDCIssuerURL {
		results = append(results, checkDiscovery(ctx, client, input.Config.MockOIDCIssuerURL))
	}
	return results
}

func HasFailure(results []Result) bool {
	for _, result := range results {
		if result.Status == StatusFail {
			return true
		}
	}
	return false
}

func checkAuthConfig(cfg config.Config) Result {
	if err := config.ValidateAuth(cfg); err != nil {
		return Result{
			Name:        "auth configuration",
			Status:      StatusFail,
			Detail:      err.Error(),
			Remediation: "Set --oidc-issuer-url/--oidc-client-id or " + config.EnvOIDCIssuerURL + "/" + config.EnvOIDCClientID + ".",
		}
	}
	return Result{Name: "auth configuration", Status: StatusPass, Detail: "required OIDC settings are present"}
}

func checkAuthState(ctx context.Context, store auth.Store) Result {
	if store == nil {
		return Result{Name: "auth state", Status: StatusInfo, Detail: "auth state store is not configured"}
	}

	state, err := store.Read(ctx)
	if err != nil {
		status := StatusFail
		detail := err.Error()
		if errors.Is(err, auth.ErrCorruptState) {
			detail = "auth state is corrupt"
		}
		return Result{Name: "auth state", Status: status, Detail: detail, Remediation: "Run devcli logout, then devcli login."}
	}
	if !state.Authenticated {
		return Result{Name: "auth state", Status: StatusInfo, Detail: "no user is logged in"}
	}
	if state.Expired(time.Now()) {
		return Result{Name: "auth state", Status: StatusFail, Detail: "session is expired", Remediation: "Run devcli login again."}
	}
	return Result{Name: "auth state", Status: StatusPass, Detail: "session state is readable"}
}

func checkDiscovery(ctx context.Context, client *http.Client, issuerURL string) Result {
	url := strings.TrimRight(issuerURL, "/") + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{Name: "oidc discovery", Status: StatusFail, Detail: err.Error(), Remediation: "Check the issuer URL."}
	}
	resp, err := client.Do(req)
	if err != nil {
		return Result{Name: "oidc discovery", Status: StatusFail, Detail: fmt.Sprintf("unable to reach %s: %v", url, err), Remediation: "Start the provider or update the issuer URL."}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Result{Name: "oidc discovery", Status: StatusFail, Detail: fmt.Sprintf("%s returned %s", url, resp.Status), Remediation: "Check the issuer URL and provider health."}
	}
	return Result{Name: "oidc discovery", Status: StatusPass, Detail: url}
}
