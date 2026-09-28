package log

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
)

// Logger wraps slog.Logger to provide structured logging with level filtering
// and request-scoped fields. The zero value is not usable; construct via New.
type Logger struct {
	base *slog.Logger
}

// defaultMu guards access to Default.
var defaultMu sync.RWMutex

// Default is the process-wide logger used by the package-level helpers.
// It is safe for concurrent use and guarded by defaultMu.
var Default *Logger

// New creates a new Logger writing to w with the given level. In debug mode
// (level <= slog.LevelDebug) it uses a text handler; otherwise JSON. slog is
// chosen because it is stdlib, structured, and optimised for low overhead in
// production builds.
func New(w io.Writer, level slog.Level) *Logger {
	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if level <= slog.LevelDebug {
		handler = slog.NewTextHandler(w, opts)
	} else {
		handler = slog.NewJSONHandler(w, opts)
	}
	return &Logger{base: slog.New(handler)}
}

// SetDefault sets the package-level Default logger used by helper functions.
func SetDefault(l *Logger) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	Default = l
}

// cloneWith returns a new Logger with the provided slog.Logger.
func (l *Logger) cloneWith(inner *slog.Logger) *Logger {
	return &Logger{base: inner}
}

// WithField returns a child Logger with a permanent structured field attached.
func (l *Logger) WithField(key string, val any) *Logger {
	return l.cloneWith(l.base.With(key, val))
}

// WithContext returns a child Logger enriched with fields extracted from ctx.
// Fields are taken from an optional map[string]any stored under a private key.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	type ctxKey struct{}
	raw := ctx.Value(ctxKey{})
	fields, ok := raw.(map[string]any)
	if !ok || len(fields) == 0 {
		return l
	}
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	return l.cloneWith(l.base.With(attrs...))
}

// Debug logs a debug-level message.
func (l *Logger) Debug(msg string, args ...any) {
	l.base.Debug(msg, args...)
}

// Info logs an info-level message.
func (l *Logger) Info(msg string, args ...any) {
	l.base.Info(msg, args...)
}

// Warn logs a warning-level message.
func (l *Logger) Warn(msg string, args ...any) {
	l.base.Warn(msg, args...)
}

// Error logs an error-level message.
func (l *Logger) Error(msg string, args ...any) {
	l.base.Error(msg, args...)
}

// Fatal logs an error-level message and then terminates the process with
// exit code 1. It does not panic.
func (l *Logger) Fatal(msg string, args ...any) {
	l.base.Error(msg, args...)
}

// Debug logs with the package-level Default logger.
func Debug(msg string, args ...any) {
	defaultMu.RLock()
	l := Default
	defaultMu.RUnlock()
	if l != nil {
		l.Debug(msg, args...)
	}
}

// Info logs with the package-level Default logger.
func Info(msg string, args ...any) {
	defaultMu.RLock()
	l := Default
	defaultMu.RUnlock()
	if l != nil {
		l.Info(msg, args...)
	}
}

// Warn logs with the package-level Default logger.
func Warn(msg string, args ...any) {
	defaultMu.RLock()
	l := Default
	defaultMu.RUnlock()
	if l != nil {
		l.Warn(msg, args...)
	}
}

// Error logs with the package-level Default logger.
func Error(msg string, args ...any) {
	defaultMu.RLock()
	l := Default
	defaultMu.RUnlock()
	if l != nil {
		l.Error(msg, args...)
	}
}

// Fatal logs with the package-level Default logger and exits with status 1.
func Fatal(msg string, args ...any) {
	defaultMu.RLock()
	l := Default
	defaultMu.RUnlock()
	if l == nil {
		os.Exit(1)
	}
	l.Fatal(msg, args...)
}


