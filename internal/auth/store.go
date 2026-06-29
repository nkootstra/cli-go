package auth

import (
	"context"
	"errors"
)

var ErrCorruptState = errors.New("auth state is corrupt")

type Store interface {
	Read(context.Context) (State, error)
	Write(context.Context, State) error
	Delete(context.Context) error
}
