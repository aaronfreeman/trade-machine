package observability

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestInitLogger_Development(t *testing.T) {
	InitLogger(false)

	if Logger == nil {
		t.Error("Logger should not be nil after initialization")
	}
}

func TestInitLogger_Production(t *testing.T) {
	InitLogger(true)

	if Logger == nil {
		t.Error("Logger should not be nil after initialization")
	}
}

func TestInitLoggerWithLevel(t *testing.T) {
	InitLoggerWithLevel(false, slog.LevelDebug)

	if Logger == nil {
		t.Error("Logger should not be nil after initialization")
	}
}

func TestWithContext(t *testing.T) {
	Logger = nil // Reset
	ctx := context.Background()
	logger := WithContext(ctx)

	if logger == nil {
		t.Error("WithContext should not return nil")
	}
}

func TestLoggingFunctions(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	Logger = slog.New(handler)

	t.Run("Info", func(t *testing.T) {
		buf.Reset()
		Info("test info message", "key", "value")
		if !strings.Contains(buf.String(), "test info message") {
			t.Error("Info should log the message")
		}
		if !strings.Contains(buf.String(), "key=value") {
			t.Error("Info should log the key-value pair")
		}
	})

	t.Run("Warn", func(t *testing.T) {
		buf.Reset()
		Warn("test warn message", "warning_key", "warning_value")
		if !strings.Contains(buf.String(), "test warn message") {
			t.Error("Warn should log the message")
		}
		if !strings.Contains(buf.String(), "WARN") {
			t.Error("Warn should log at WARN level")
		}
	})

	t.Run("Error", func(t *testing.T) {
		buf.Reset()
		Error("test error message", "error_key", "error_value")
		if !strings.Contains(buf.String(), "test error message") {
			t.Error("Error should log the message")
		}
		if !strings.Contains(buf.String(), "ERROR") {
			t.Error("Error should log at ERROR level")
		}
	})

	t.Run("Debug", func(t *testing.T) {
		buf.Reset()
		Debug("test debug message", "debug_key", "debug_value")
		if !strings.Contains(buf.String(), "test debug message") {
			t.Error("Debug should log the message")
		}
		if !strings.Contains(buf.String(), "DEBUG") {
			t.Error("Debug should log at DEBUG level")
		}
	})
}

func TestWithSymbol(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	Logger = slog.New(handler)

	logger := WithSymbol("AAPL")
	logger.Info("test message")

	if !strings.Contains(buf.String(), "symbol=AAPL") {
		t.Error("WithSymbol should add symbol field to logger")
	}
}

func TestWithAgent(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	Logger = slog.New(handler)

	logger := WithAgent("fundamental")
	logger.Info("test message")

	if !strings.Contains(buf.String(), "agent_type=fundamental") {
		t.Error("WithAgent should add agent_type field to logger")
	}
}

func TestWithError(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	Logger = slog.New(handler)

	testErr := errors.New("test error")
	logger := WithError(testErr)
	logger.Info("test message")

	if !strings.Contains(buf.String(), "error=") {
		t.Error("WithError should add error field to logger")
	}
}

func TestLoggingWithNilLogger(t *testing.T) {
	// Test that functions handle nil Logger by initializing
	Logger = nil
	Info("test message") // Should not panic

	Logger = nil
	Warn("test message") // Should not panic

	Logger = nil
	Error("test message") // Should not panic

	Logger = nil
	Debug("test message") // Should not panic

	Logger = nil
	_ = WithSymbol("AAPL") // Should not panic

	Logger = nil
	_ = WithAgent("fundamental") // Should not panic

	Logger = nil
	_ = WithError(errors.New("test")) // Should not panic

	Logger = nil
	_ = WithContext(context.Background()) // Should not panic
}

func TestJSONFormat_Production(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	Logger = slog.New(handler)

	Info("test json message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, `"msg":"test json message"`) {
		t.Error("JSON handler should output JSON format")
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Error("JSON handler should include key-value pairs in JSON")
	}
}
