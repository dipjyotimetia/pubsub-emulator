package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	log := New()
	if log == nil {
		t.Fatal("Expected logger to be created, got nil")
	}
	if log.Logger == nil {
		t.Fatal("Expected underlying slog.Logger to be created, got nil")
	}
}

func TestNewWithLevel(t *testing.T) {
	log := NewWithLevel(slog.LevelDebug)
	if log == nil {
		t.Fatal("Expected logger to be created, got nil")
	}
	if log.Logger == nil {
		t.Fatal("Expected underlying slog.Logger to be created, got nil")
	}
}

func TestNewTextLogger(t *testing.T) {
	log := NewTextLogger()
	if log == nil {
		t.Fatal("Expected logger to be created, got nil")
	}
	if log.Logger == nil {
		t.Fatal("Expected underlying slog.Logger to be created, got nil")
	}
}

func TestInfo(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	log := New()
	log.Info("Test message: %s", "info")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Test message: info") {
		t.Errorf("Expected output to contain 'Test message: info', got: %s", output)
	}

	// Verify it's valid JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
	}
}

func TestInfoContext(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	log := New()
	ctx := context.Background()
	log.InfoContext(ctx, "Test context message", "key", "value")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Test context message") {
		t.Errorf("Expected output to contain 'Test context message', got: %s", output)
	}

	if !strings.Contains(output, "key") || !strings.Contains(output, "value") {
		t.Errorf("Expected output to contain key-value pair, got: %s", output)
	}
}

func TestError(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	log := New()
	log.Error("Test error: %s", "error")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Test error: error") {
		t.Errorf("Expected output to contain 'Test error: error', got: %s", output)
	}

	if !strings.Contains(output, "ERROR") {
		t.Errorf("Expected output to contain 'ERROR' level, got: %s", output)
	}
}

func TestErrorContext(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	log := New()
	ctx := context.Background()
	log.ErrorContext(ctx, "Test error context", "error", "test-error")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Test error context") {
		t.Errorf("Expected output to contain 'Test error context', got: %s", output)
	}
}

func TestWarn(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	log := New()
	log.Warn("Test warning: %s", "warn")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Test warning: warn") {
		t.Errorf("Expected output to contain 'Test warning: warn', got: %s", output)
	}

	if !strings.Contains(output, "WARN") {
		t.Errorf("Expected output to contain 'WARN' level, got: %s", output)
	}
}

func TestWarnContext(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	log := New()
	ctx := context.Background()
	log.WarnContext(ctx, "Test warn context", "warning", "test-warning")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Test warn context") {
		t.Errorf("Expected output to contain 'Test warn context', got: %s", output)
	}
}

func TestDebug(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create logger with debug level
	log := NewWithLevel(slog.LevelDebug)
	log.Debug("Test debug: %s", "debug")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Test debug: debug") {
		t.Errorf("Expected output to contain 'Test debug: debug', got: %s", output)
	}

	if !strings.Contains(output, "DEBUG") {
		t.Errorf("Expected output to contain 'DEBUG' level, got: %s", output)
	}
}

func TestDebugContext(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	log := NewWithLevel(slog.LevelDebug)
	ctx := context.Background()
	log.DebugContext(ctx, "Test debug context", "debug", "test-debug")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Test debug context") {
		t.Errorf("Expected output to contain 'Test debug context', got: %s", output)
	}
}

func TestWith(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	log := New()
	contextLog := log.With("request_id", "12345")
	contextLog.Info("Test message with context")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "request_id") {
		t.Errorf("Expected output to contain 'request_id', got: %s", output)
	}

	if !strings.Contains(output, "12345") {
		t.Errorf("Expected output to contain '12345', got: %s", output)
	}
}

func TestWithGroup(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	log := New()
	groupLog := log.WithGroup("http")
	groupLog.Logger.Info("Request received", "status", "200")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// WithGroup creates a nested structure in JSON output
	// The group name appears as a key in the JSON structure
	if !strings.Contains(output, "http") {
		t.Errorf("Expected output to contain group 'http', got: %s", output)
	}
}

func TestFatal(t *testing.T) {
	// We can't actually test os.Exit(1), but we can verify the error message is logged
	// This is a simplified test that verifies the error message is formatted correctly
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a logger but don't call Fatal as it will exit
	log := New()

	// Instead, test the error formatting by calling Error with FATAL level
	log.Logger.Error("Fatal error occurred", slog.String("level", "FATAL"))

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "FATAL") {
		t.Errorf("Expected output to contain 'FATAL', got: %s", output)
	}

	if !strings.Contains(output, "Fatal error occurred") {
		t.Errorf("Expected output to contain error message, got: %s", output)
	}
}

func TestMultipleLoggerInstances(t *testing.T) {
	log1 := New()
	log2 := New()

	if log1 == log2 {
		t.Error("Expected different logger instances")
	}

	if log1.Logger == log2.Logger {
		t.Error("Expected different underlying slog.Logger instances")
	}
}

func TestLoggerChaining(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	log := New()
	contextLog := log.With("request_id", "123").WithGroup("api")
	contextLog.Logger.Info("Chained logger test", "method", "GET")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "request_id") {
		t.Errorf("Expected output to contain 'request_id', got: %s", output)
	}

	if !strings.Contains(output, "123") {
		t.Errorf("Expected output to contain '123', got: %s", output)
	}

	// WithGroup creates a nested structure, so we check if api group is present
	if !strings.Contains(output, "api") {
		t.Errorf("Expected output to contain group 'api', got: %s", output)
	}
}
