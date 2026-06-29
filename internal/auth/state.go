package auth

import "time"

type State struct {
	Authenticated bool      `json:"authenticated"`
	Subject       string    `json:"subject,omitempty"`
	Issuer        string    `json:"issuer,omitempty"`
	Audience      string    `json:"audience,omitempty"`
	Email         string    `json:"email,omitempty"`
	Name          string    `json:"name,omitempty"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
	AccessToken   string    `json:"access_token,omitempty"`
	RefreshToken  string    `json:"refresh_token,omitempty"`
	IDToken       string    `json:"id_token,omitempty"`
}

func (s State) Expired(now time.Time) bool {
	if !s.Authenticated || s.ExpiresAt.IsZero() {
		return false
	}
	return !s.ExpiresAt.After(now)
}

func (s State) Usable(now time.Time) bool {
	return s.Authenticated && !s.Expired(now)
}
