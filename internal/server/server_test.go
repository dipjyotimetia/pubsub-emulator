package server

import (
	"context"
	"testing"
	"time"

	"github.com/dipjyotimetia/pubsub-emulator/internal/dashboard"
	"github.com/dipjyotimetia/pubsub-emulator/pkg/logger"
)

func TestNew(t *testing.T) {
	log := logger.New()
	// Create a minimal dashboard (client can be nil for testing)
	dash := dashboard.New(nil, "test-project", log)

	cfg := &Config{
		Port:      "8080",
		Dashboard: dash,
		Logger:    log,
	}

	srv := New(cfg)

	if srv == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	if srv.port != "8080" {
		t.Errorf("Expected port '8080', got '%s'", srv.port)
	}

	if srv.log == nil {
		t.Error("Expected logger to be set")
	}

	if srv.dashboard == nil {
		t.Error("Expected dashboard to be set")
	}
}

func TestNew_WithEmptyPort(t *testing.T) {
	log := logger.New()
	dash := dashboard.New(nil, "test-project", log)

	cfg := &Config{
		Port:      "",
		Dashboard: dash,
		Logger:    log,
	}

	srv := New(cfg)

	if srv.port != "" {
		t.Errorf("Expected empty port, got '%s'", srv.port)
	}
}

func TestShutdown_NilServer(t *testing.T) {
	log := logger.New()
	dash := dashboard.New(nil, "test-project", log)

	cfg := &Config{
		Port:      "8080",
		Dashboard: dash,
		Logger:    log,
	}

	srv := New(cfg)

	// Test shutdown when srv.srv is nil
	ctx := context.Background()
	err := srv.Shutdown(ctx)
	if err != nil {
		t.Errorf("Expected no error when shutting down nil server, got %v", err)
	}
}

func TestConfig_AllFieldsSet(t *testing.T) {
	log := logger.New()
	dash := dashboard.New(nil, "test-project", log)

	cfg := &Config{
		Port:      "9090",
		Dashboard: dash,
		Logger:    log,
	}

	if cfg.Port != "9090" {
		t.Errorf("Expected Port '9090', got '%s'", cfg.Port)
	}

	if cfg.Dashboard == nil {
		t.Error("Expected Dashboard to be set")
	}

	if cfg.Logger == nil {
		t.Error("Expected Logger to be set")
	}
}

func TestNew_WithNilDashboard(t *testing.T) {
	log := logger.New()

	cfg := &Config{
		Port:      "8080",
		Dashboard: nil,
		Logger:    log,
	}

	srv := New(cfg)

	if srv == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	if srv.dashboard != nil {
		t.Error("Expected dashboard to be nil")
	}
}

func TestNew_WithNilLogger(t *testing.T) {
	dash := dashboard.New(nil, "test-project", logger.New())

	cfg := &Config{
		Port:      "8080",
		Dashboard: dash,
		Logger:    nil,
	}

	srv := New(cfg)

	if srv == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	if srv.log != nil {
		t.Error("Expected logger to be nil")
	}
}

func TestConfig_EmptyConfig(t *testing.T) {
	cfg := &Config{}

	if cfg.Port != "" {
		t.Errorf("Expected empty Port, got '%s'", cfg.Port)
	}

	if cfg.Dashboard != nil {
		t.Error("Expected Dashboard to be nil")
	}

	if cfg.Logger != nil {
		t.Error("Expected Logger to be nil")
	}
}

func TestShutdown_ContextCanceled(t *testing.T) {
	log := logger.New()
	dash := dashboard.New(nil, "test-project", log)

	cfg := &Config{
		Port:      "8080",
		Dashboard: dash,
		Logger:    log,
	}

	srv := New(cfg)

	// Create an already canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := srv.Shutdown(ctx)
	if err != nil {
		t.Errorf("Expected no error when shutting down with canceled context, got %v", err)
	}
}

func TestShutdown_WithTimeout(t *testing.T) {
	log := logger.New()
	dash := dashboard.New(nil, "test-project", log)

	cfg := &Config{
		Port:      "8080",
		Dashboard: dash,
		Logger:    log,
	}

	srv := New(cfg)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := srv.Shutdown(ctx)
	if err != nil {
		t.Errorf("Expected no error when shutting down with timeout context, got %v", err)
	}
}
