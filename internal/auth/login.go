package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nkootstra/cli/internal/config"
	"golang.org/x/oauth2"
)

var (
	ErrStateMismatch  = errors.New("oidc callback state mismatch")
	ErrMissingCode    = errors.New("oidc callback is missing code")
	ErrMissingIDToken = errors.New("oidc token response is missing id_token")
)

type URLOpener func(string) error

type LoginOptions struct {
	Config      config.Config
	Store       Store
	OpenURL     URLOpener
	StateValue  string
	NonceValue  string
	HTTPClient  *http.Client
	WaitTimeout time.Duration
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

	oidcClient, err := NewOIDCClient(ctx, opts.Config)
	if err != nil {
		return State{}, err
	}

	callbackServer, err := StartCallbackServer(opts.Config.OIDCCallbackListenAddr, stateValue)
	if err != nil {
		return State{}, err
	}
	defer callbackServer.Shutdown(context.Background())

	authURL := oidcClient.AuthCodeURL(stateValue, verifier, nonceValue)
	if err := opts.OpenURL(authURL); err != nil {
		return State{}, fmt.Errorf("open auth url: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, opts.WaitTimeout)
	defer cancel()

	code, err := callbackServer.Wait(waitCtx)
	if err != nil {
		return State{}, err
	}

	session, err := oidcClient.ExchangeAndVerify(ctx, code, verifier, nonceValue)
	if err != nil {
		return State{}, err
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
