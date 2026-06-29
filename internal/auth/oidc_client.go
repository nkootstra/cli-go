package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/nkootstra/cli/internal/config"
	"golang.org/x/oauth2"
)

type OIDCClient struct {
	config      config.Config
	provider    *oidc.Provider
	oauthConfig oauth2.Config
}

type tokenClaims struct {
	Subject string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Nonce   string `json:"nonce"`
}

func NewOIDCClient(ctx context.Context, cfg config.Config) (*OIDCClient, error) {
	provider, err := oidc.NewProvider(ctx, cfg.OIDCIssuerURL)
	if err != nil {
		return nil, fmt.Errorf("discover oidc issuer: %w", err)
	}

	return &OIDCClient{
		config:   cfg,
		provider: provider,
		oauthConfig: oauth2.Config{
			ClientID:    cfg.OIDCClientID,
			Endpoint:    provider.Endpoint(),
			RedirectURL: cfg.OIDCRedirectURL,
			Scopes:      cfg.OIDCScopes,
		},
	}, nil
}

func (c *OIDCClient) AuthCodeURL(stateValue, verifier, nonceValue string) string {
	return c.oauthConfig.AuthCodeURL(
		stateValue,
		oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("nonce", nonceValue),
	)
}

func (c *OIDCClient) ExchangeAndVerify(ctx context.Context, code, verifier, nonceValue string) (State, error) {
	token, err := c.oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return State{}, fmt.Errorf("exchange oidc code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return State{}, ErrMissingIDToken
	}

	idToken, err := c.provider.Verifier(&oidc.Config{ClientID: c.config.OIDCClientID}).Verify(ctx, rawIDToken)
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

	return State{
		Authenticated: true,
		Subject:       claims.Subject,
		Issuer:        c.config.OIDCIssuerURL,
		Audience:      c.config.OIDCClientID,
		Email:         claims.Email,
		Name:          claims.Name,
		ExpiresAt:     idToken.Expiry,
		AccessToken:   token.AccessToken,
		RefreshToken:  token.RefreshToken,
		IDToken:       rawIDToken,
	}, nil
}
