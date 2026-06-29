package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	EnvOIDCIssuerURL          = "DEVCLI_OIDC_ISSUER_URL"
	EnvOIDCClientID           = "DEVCLI_OIDC_CLIENT_ID"
	EnvOIDCScopes             = "DEVCLI_OIDC_SCOPES"
	EnvOIDCRedirectURL        = "DEVCLI_OIDC_REDIRECT_URL"
	EnvOIDCCallbackListenAddr = "DEVCLI_OIDC_CALLBACK_LISTEN_ADDR"
	EnvAuthStatePath          = "DEVCLI_AUTH_STATE_PATH"
	EnvMockOIDCIssuerURL      = "DEVCLI_MOCK_OIDC_ISSUER_URL"
)

type Config struct {
	OIDCIssuerURL          string
	OIDCClientID           string
	OIDCScopes             []string
	OIDCRedirectURL        string
	OIDCCallbackListenAddr string
	AuthStatePath          string
	MockOIDCIssuerURL      string
}

type Overrides struct {
	OIDCIssuerURL          string
	OIDCClientID           string
	OIDCScopes             []string
	OIDCRedirectURL        string
	OIDCCallbackListenAddr string
	AuthStatePath          string
	MockOIDCIssuerURL      string
}

type Input struct {
	Env       map[string]string
	Overrides Overrides
}

func Resolve(input Input) Config {
	env := input.Env
	if env == nil {
		env = environMap()
	}

	cfg := Config{
		OIDCIssuerURL:          env[EnvOIDCIssuerURL],
		OIDCClientID:           env[EnvOIDCClientID],
		OIDCScopes:             ParseScopes(env[EnvOIDCScopes]),
		OIDCRedirectURL:        env[EnvOIDCRedirectURL],
		OIDCCallbackListenAddr: env[EnvOIDCCallbackListenAddr],
		AuthStatePath:          env[EnvAuthStatePath],
		MockOIDCIssuerURL:      env[EnvMockOIDCIssuerURL],
	}

	applyDefaults(&cfg)
	applyOverrides(&cfg, input.Overrides)
	return cfg
}

func applyDefaults(cfg *Config) {
	if len(cfg.OIDCScopes) == 0 {
		cfg.OIDCScopes = []string{"openid", "profile", "email"}
	}
	if cfg.OIDCCallbackListenAddr == "" {
		cfg.OIDCCallbackListenAddr = "127.0.0.1:8765"
	}
	if cfg.OIDCRedirectURL == "" {
		cfg.OIDCRedirectURL = "http://" + cfg.OIDCCallbackListenAddr + "/callback"
	}
	if cfg.AuthStatePath == "" {
		cfg.AuthStatePath = defaultAuthStatePath()
	}
}

func applyOverrides(cfg *Config, overrides Overrides) {
	if overrides.OIDCIssuerURL != "" {
		cfg.OIDCIssuerURL = overrides.OIDCIssuerURL
	}
	if overrides.OIDCClientID != "" {
		cfg.OIDCClientID = overrides.OIDCClientID
	}
	if len(overrides.OIDCScopes) > 0 {
		cfg.OIDCScopes = overrides.OIDCScopes
	}
	if overrides.OIDCRedirectURL != "" {
		cfg.OIDCRedirectURL = overrides.OIDCRedirectURL
	}
	if overrides.OIDCCallbackListenAddr != "" {
		cfg.OIDCCallbackListenAddr = overrides.OIDCCallbackListenAddr
	}
	if overrides.AuthStatePath != "" {
		cfg.AuthStatePath = overrides.AuthStatePath
	}
	if overrides.MockOIDCIssuerURL != "" {
		cfg.MockOIDCIssuerURL = overrides.MockOIDCIssuerURL
	}
}

type MissingSettingsError struct {
	Names []string
}

func (e MissingSettingsError) Error() string {
	return fmt.Sprintf("missing required auth settings: %s", strings.Join(e.Names, ", "))
}

func ValidateAuth(cfg Config) error {
	var missing []string
	if cfg.OIDCIssuerURL == "" {
		missing = append(missing, EnvOIDCIssuerURL)
	}
	if cfg.OIDCClientID == "" {
		missing = append(missing, EnvOIDCClientID)
	}
	if len(missing) > 0 {
		return MissingSettingsError{Names: missing}
	}
	return nil
}

func IsMissingSetting(err error, name string) bool {
	var missing MissingSettingsError
	if !errors.As(err, &missing) {
		return false
	}
	for _, candidate := range missing.Names {
		if candidate == name {
			return true
		}
	}
	return false
}

func ParseScopes(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	scopes := make([]string, 0, len(fields))
	for _, field := range fields {
		if field != "" {
			scopes = append(scopes, field)
		}
	}
	return scopes
}

func environMap() map[string]string {
	env := map[string]string{}
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			env[key] = value
		}
	}
	return env
}

func defaultAuthStatePath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return filepath.Join(".devcli", "auth.json")
	}
	return filepath.Join(dir, "devcli", "auth.json")
}
