package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	// DefaultAddress keeps the scaffold runnable without extra configuration.
	DefaultAddress  = ":8080"
	shutdownTimeout = 5 * time.Second
)

// Run starts a minimal HTTP server and shuts it down when the context is canceled.
func Run(ctx context.Context, addr string) error {
	if addr == "" {
		addr = DefaultAddress
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           NewHandler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen on %s: %w", addr, err)
			return
		}

		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}

		return <-errCh
	}
}

// NewHandler wires the initial HTTP routes for the scaffolded server.
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	return mux
}
