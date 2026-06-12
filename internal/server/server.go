package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dipjyotimetia/pubsub-emulator/internal/dashboard"
	"github.com/dipjyotimetia/pubsub-emulator/pkg/logger"
)

const (
	readTimeout             = 15 * time.Second
	writeTimeout            = 15 * time.Second
	idleTimeout             = 60 * time.Second
	gracefulShutdownTimeout = 30 * time.Second
)

// Server represents the HTTP server with graceful shutdown capability
type Server struct {
	dashboard *dashboard.Dashboard
	port      string
	log       *logger.Logger
	srv       *http.Server
}

// Config holds the server configuration
type Config struct {
	Port      string
	Dashboard *dashboard.Dashboard
	Logger    *logger.Logger
}

// New creates a new Server instance
func New(cfg *Config) *Server {
	return &Server{
		dashboard: cfg.Dashboard,
		port:      cfg.Port,
		log:       cfg.Logger,
	}
}

// Start starts the HTTP server and blocks until ctx is cancelled (e.g. on a
// shutdown signal) or the server fails, then shuts down gracefully.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	s.dashboard.RegisterRoutes(mux)

	handler := dashboard.HTTPLoggingMiddleware(s.log)(mux)
	handler = dashboard.CORSMiddleware(handler)

	s.srv = &http.Server{
		Addr:         ":" + s.port,
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	serverErrors := make(chan error, 1)

	go func() {
		s.log.Info("Starting HTTP server on port %s", s.port)
		s.log.Info("Dashboard available at http://localhost:%s", s.port)
		serverErrors <- s.srv.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("server error: %w", err)

	case <-ctx.Done():
		s.log.Info("Shutdown signal received, stopping HTTP server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()

		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			s.log.Error("Graceful shutdown failed: %v", err)
			if err := s.srv.Close(); err != nil {
				return fmt.Errorf("could not stop server gracefully: %w", err)
			}
		}

		s.log.Info("Server stopped gracefully")
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}
	return nil
}
