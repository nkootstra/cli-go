package auth

import (
	"errors"
	"time"
)

var (
	ErrNotLoggedIn = errors.New("no user is logged in")
	ErrSessionExpired = errors.New("session is expired")
)

func CurrentSession(state State, now time.Time) (State, error) {
	if !state.Authenticated {
		return State{}, ErrNotLoggedIn
	}
	if state.Expired(now) {
		return State{}, ErrSessionExpired
	}
	return state, nil
}
