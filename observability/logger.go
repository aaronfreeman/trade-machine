package observability

import (
	"context"
	"log/slog"
	"os"
)

// Logger is the global logger instance
var Logger *slog.Logger

// InitLogger initializes the global logger with the appropriate handler
// For production, use JSON format; for development, use text format
func InitLogger(production bool) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if production {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// InitLoggerWithLevel initializes the logger with a specific log level
func InitLoggerWithLevel(production bool, level slog.Level) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: level,
	}

	if production {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// WithContext returns a logger with context fields
func WithContext(ctx context.Context) *slog.Logger {
	if Logger == nil {
		InitLogger(false)
	}
	return Logger
}

// Info logs an info message
func Info(msg string, args ...any) {
	if Logger == nil {
		InitLogger(false)
	}
	Logger.Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	if Logger == nil {
		InitLogger(false)
	}
	Logger.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	if Logger == nil {
		InitLogger(false)
	}
	Logger.Error(msg, args...)
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	if Logger == nil {
		InitLogger(false)
	}
	Logger.Debug(msg, args...)
}

// Fatal logs an error message and exits
func Fatal(msg string, args ...any) {
	if Logger == nil {
		InitLogger(false)
	}
	Logger.Error(msg, args...)
	os.Exit(1)
}

// WithSymbol returns a logger with symbol field
func WithSymbol(symbol string) *slog.Logger {
	if Logger == nil {
		InitLogger(false)
	}
	return Logger.With("symbol", symbol)
}

// WithAgent returns a logger with agent type field
func WithAgent(agentType string) *slog.Logger {
	if Logger == nil {
		InitLogger(false)
	}
	return Logger.With("agent_type", agentType)
}

// WithError returns a logger with error field
func WithError(err error) *slog.Logger {
	if Logger == nil {
		InitLogger(false)
	}
	return Logger.With("error", err)
}
