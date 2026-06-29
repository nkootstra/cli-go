package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

type CallbackServer struct {
	server *http.Server
	codeCh chan string
	errCh  chan error
}

func StartCallbackServer(addr, wantState string) (*CallbackServer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen for oidc callback: %w", err)
	}

	callback := &CallbackServer{
		codeCh: make(chan string, 1),
		errCh:  make(chan error, 1),
	}

	mux := http.NewServeMux()
	callback.server = &http.Server{Handler: mux}
	mux.HandleFunc("/callback", callback.handleCallback(wantState))

	go callback.serve(listener)

	return callback, nil
}

func (s *CallbackServer) Wait(ctx context.Context) (string, error) {
	select {
	case code := <-s.codeCh:
		return code, nil
	case err := <-s.errCh:
		return "", err
	case <-ctx.Done():
		return "", fmt.Errorf("wait for oidc callback: %w", ctx.Err())
	}
}

func (s *CallbackServer) Shutdown(ctx context.Context) {
	shutdownCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	_ = s.server.Shutdown(shutdownCtx)
}

func (s *CallbackServer) handleCallback(wantState string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code, err := ValidateCallback(wantState, r.URL.Query().Get("state"), r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			s.sendError(err)
			return
		}
		_, _ = w.Write([]byte("Login complete. You can close this browser window.\n"))
		s.sendCode(code)
	}
}

func (s *CallbackServer) serve(listener net.Listener) {
	if err := s.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.sendError(err)
	}
}

func (s *CallbackServer) sendCode(code string) {
	select {
	case s.codeCh <- code:
	default:
	}
}

func (s *CallbackServer) sendError(err error) {
	select {
	case s.errCh <- err:
	default:
	}
}
