package logger

import (
	"log"
	"os"
)

// Logger wraps standard logger with structured logging capabilities
type Logger struct {
	*log.Logger
}

// New creates a new logger instance
func New() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

// Info logs informational messages
func (l *Logger) Info(format string, v ...any) {
	l.Printf("[INFO] "+format, v...)
}

// Error logs error messages
func (l *Logger) Error(format string, v ...any) {
	l.Printf("[ERROR] "+format, v...)
}

// Warn logs warning messages
func (l *Logger) Warn(format string, v ...any) {
	l.Printf("[WARN] "+format, v...)
}

// Debug logs debug messages
func (l *Logger) Debug(format string, v ...any) {
	l.Printf("[DEBUG] "+format, v...)
}

// Fatal logs fatal messages and exits
func (l *Logger) Fatal(format string, v ...any) {
	l.Fatalf("[FATAL] "+format, v...)
}
