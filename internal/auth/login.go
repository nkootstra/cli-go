package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/nkootstra/cli/internal/config"
	"golang.org/x/oauth2"
)

var (
	ErrStateMismatch = errors.New("oidc callback state mismatch")
	ErrMissingCode = errors.New("oidc callback is missing code")
	ErrMissingIDToken = errors.New("oidc token response is missing id_token")
)

type URLOpener func(string) error

type LoginOptions struct {
	Config config.Config
	Store Store
	OpenURL URLOpener
	StateValue string
	NonceValue string
	HTTPClient *http.Client
	WaitTimeout time.Duration
}

type tokenClaims struct {
	Subject string `json:"sub"`
	Email string `json:"email"`
	Name string `json:"name"`
	Nonce string `json:"nonce"`
}

func Login(ctx context.Context, opts LoginOptions) (State, error) {
	if opts.Store == nil {
		return State{}, errors.New("auth store is required")
	}
	if err := config.ValidateAuth(opts.Config); err != nil {
		return State{}, err
	}
	if opts.OpenURL == nil {
		return State{}, errors.New("url opener is required")
	}
	if opts.WaitTimeout == 0 {
		opts.WaitTimeout = 2 * time.Minute
	}

	stateValue := opts.StateValue
	if stateValue == "" {
		stateValue = oauth2.GenerateVerifier()
	}
	nonceValue := opts.NonceValue
	if nonceValue == "" {
		nonceValue = oauth2.GenerateVerifier()
	}
	verifier := oauth2.GenerateVerifier()

	if opts.HTTPClient != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, opts.HTTPClient)
	}

	provider, err := oidc.NewProvider(ctx, opts.Config.OIDCIssuerURL)
	if err != nil {
		return State{}, fmt.Errorf("discover oidc issuer: %w", err)
	}

	oauthConfig := oauth2.Config{
		ClientID: opts.Config.OIDCClientID,
		Endpoint: provider.Endpoint(),
		RedirectURL: opts.Config.OIDCRedirectURL,
		Scopes: opts.Config.OIDCScopes,
	}

	codeCh, errCh, shutdown, err := listenForCallback(opts.Config.OIDCCallbackListenAddr, stateValue)
	if err != nil {
		return State{}, err
	}
	defer shutdown()

	authURL := oauthConfig.AuthCodeURL(
		stateValue,
		oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("nonce", nonceValue),
	)
	if err := opts.OpenURL(authURL); err != nil {
		return State{}, fmt.Errorf("open auth url: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, opts.WaitTimeout)
	defer cancel()

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return State{}, err
	case <-waitCtx.Done():
		return State{}, fmt.Errorf("wait for oidc callback: %w", waitCtx.Err())
	}

	token, err := oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return State{}, fmt.Errorf("exchange oidc code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return State{}, ErrMissingIDToken
	}

	idToken, err := provider.Verifier(&oidc.Config{ClientID: opts.Config.OIDCClientID}).Verify(ctx, rawIDToken)
	if err != nil {
		return State{}, fmt.Errorf("verify id token: %w", err)
	}

	var claims tokenClaims
	if err := idToken.Claims(&claims); err != nil {
		return State{}, fmt.Errorf("read id token claims: %w", err)
	}
	if claims.Nonce != "" && claims.Nonce != nonceValue {
		return State{}, errors.New("oidc id token nonce mismatch")
	}

	session := State{
		Authenticated: true,
		Subject: claims.Subject,
		Issuer: opts.Config.OIDCIssuerURL,
		Audience: opts.Config.OIDCClientID,
		Email: claims.Email,
		Name: claims.Name,
		ExpiresAt: idToken.Expiry,
		AccessToken: token.AccessToken,
		RefreshToken: token.RefreshToken,
		IDToken: rawIDToken,
	}
	if err := opts.Store.Write(ctx, session); err != nil {
		return State{}, err
	}
	return session, nil
}

func ValidateCallback(wantState, gotState, code string) (string, error) {
	if gotState != wantState {
		return "", ErrStateMismatch
	}
	if strings.TrimSpace(code) == "" {
		return "", ErrMissingCode
	}
	return code, nil
}

func listenForCallback(addr, wantState string) (<-chan string, <-chan error, func(), error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("listen for oidc callback: %w", err)
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code, err := ValidateCallback(wantState, r.URL.Query().Get("state"), r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			select {
			case errCh <- err:
			default:
			}
			return
		}
		_, _ = w.Write([]byte("Login complete. You can close this browser window.\n"))
		select {
		case codeCh <- code:
		default:
		}
	})

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	return codeCh, errCh, func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}, nil
}
