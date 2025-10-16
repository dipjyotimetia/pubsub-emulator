package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with backward-compatible format string methods
type Logger struct {
	*slog.Logger
}

// New creates a new structured logger instance with JSON output
func New() *Logger {
	// Create JSON handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true, // Adds source file and line number
	})

	return &Logger{
		Logger: slog.New(handler),
	}
}

// NewWithLevel creates a new logger with custom level
func NewWithLevel(level slog.Level) *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})

	return &Logger{
		Logger: slog.New(handler),
	}
}

// NewTextLogger creates a human-readable text logger (for development)
func NewTextLogger() *Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})

	return &Logger{
		Logger: slog.New(handler),
	}
}

// Info logs informational messages (backward compatible with format strings)
func (l *Logger) Info(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.Logger.Info(msg)
}

// InfoContext logs with structured context
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.Logger.InfoContext(ctx, msg, args...)
}

// Error logs error messages (backward compatible with format strings)
func (l *Logger) Error(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.Logger.Error(msg)
}

// ErrorContext logs errors with structured context
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.Logger.ErrorContext(ctx, msg, args...)
}

// Warn logs warning messages (backward compatible with format strings)
func (l *Logger) Warn(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.Logger.Warn(msg)
}

// WarnContext logs warnings with structured context
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.Logger.WarnContext(ctx, msg, args...)
}

// Debug logs debug messages (backward compatible with format strings)
func (l *Logger) Debug(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.Logger.Debug(msg)
}

// DebugContext logs debug with structured context
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.Logger.DebugContext(ctx, msg, args...)
}

// Fatal logs fatal messages and exits
func (l *Logger) Fatal(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.Logger.Error(msg, slog.String("level", "FATAL"))
	os.Exit(1)
}

// With returns a new logger with additional context attributes
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

// WithGroup returns a new logger with a named group
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		Logger: l.Logger.WithGroup(name),
	}
}
