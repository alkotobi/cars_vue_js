// Package log provides a thin wrapper around Go's structured logging
// facilities (log/slog) so that the rest of the codebase depends on a small,
// testable interface instead of concrete global loggers.
//
// Concurrency model: loggers are safe for concurrent use and expose only
// immutable configuration to callers.
package log

