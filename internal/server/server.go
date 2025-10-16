package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dipjyotimetia/pubsub-emulator/internal/dashboard"
	"github.com/dipjyotimetia/pubsub-emulator/pkg/logger"
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

// Start starts the HTTP server and blocks until shutdown signal is received
func (s *Server) Start() error {
	// Create HTTP server mux
	mux := http.NewServeMux()

	// Register dashboard routes
	s.dashboard.RegisterRoutes(mux)

	// Apply middlewares (order matters: logging -> CORS -> routes)
	handler := dashboard.HTTPLoggingMiddleware(s.log)(mux)
	handler = dashboard.CORSMiddleware(handler)

	// Create HTTP server
	s.srv = &http.Server{
		Addr:         ":" + s.port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		s.log.Info("Starting HTTP server on port %s", s.port)
		s.log.Info("Dashboard available at http://localhost:%s", s.port)
		serverErrors <- s.srv.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		s.log.Info("Received shutdown signal: %v", sig)

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Asking listener to shutdown and shed load
		if err := s.srv.Shutdown(ctx); err != nil {
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

// ListenAndServe starts the server without graceful shutdown handling
// This is useful for testing or when graceful shutdown is not needed
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	s.dashboard.RegisterRoutes(mux)

	handler := dashboard.HTTPLoggingMiddleware(s.log)(mux)
	handler = dashboard.CORSMiddleware(handler)

	s.srv = &http.Server{
		Addr:         ":" + s.port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.log.Info("Starting HTTP server on port %s", s.port)
	s.log.Info("Dashboard available at http://localhost:%s", s.port)
	return s.srv.ListenAndServe()
}
