package client

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// TestNoOpLogger tests that NoOpLogger does nothing
func TestNoOpLogger(t *testing.T) {
	logger := NewNoOpLogger()

	// Should not panic or do anything
	logger.Debug("test", "key", "value")
	logger.Info("test", "key", "value")
	logger.Warn("test", "key", "value")
	logger.Error("test", "key", "value")

	// Test implements Logger interface
	var _ Logger = logger
}

// TestSlogLogger tests the SlogLogger implementation
func TestSlogLogger(t *testing.T) {
	tests := []struct {
		name     string
		level    slog.Level
		logFunc  func(logger Logger)
		contains string
	}{
		{
			name:  "Debug message",
			level: slog.LevelDebug,
			logFunc: func(logger Logger) {
				logger.Debug("debug message", "key", "value")
			},
			contains: "debug message",
		},
		{
			name:  "Info message",
			level: slog.LevelInfo,
			logFunc: func(logger Logger) {
				logger.Info("info message", "key", "value")
			},
			contains: "info message",
		},
		{
			name:  "Warn message",
			level: slog.LevelWarn,
			logFunc: func(logger Logger) {
				logger.Warn("warn message", "key", "value")
			},
			contains: "warn message",
		},
		{
			name:  "Error message",
			level: slog.LevelError,
			logFunc: func(logger Logger) {
				logger.Error("error message", "key", "value")
			},
			contains: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewSlogLogger(&buf, tt.level)

			tt.logFunc(logger)

			output := buf.String()
			if !strings.Contains(output, tt.contains) {
				t.Errorf("Output %q does not contain %q", output, tt.contains)
			}
		})
	}
}

// TestSlogLoggerLevels tests that log levels are respected
func TestSlogLoggerLevels(t *testing.T) {
	tests := []struct {
		name         string
		level        slog.Level
		logFunc      func(logger Logger)
		shouldOutput bool
		description  string
	}{
		{
			name:  "Debug logged at Debug level",
			level: slog.LevelDebug,
			logFunc: func(logger Logger) {
				logger.Debug("debug msg")
			},
			shouldOutput: true,
		},
		{
			name:  "Debug not logged at Info level",
			level: slog.LevelInfo,
			logFunc: func(logger Logger) {
				logger.Debug("debug msg")
			},
			shouldOutput: false,
		},
		{
			name:  "Info logged at Info level",
			level: slog.LevelInfo,
			logFunc: func(logger Logger) {
				logger.Info("info msg")
			},
			shouldOutput: true,
		},
		{
			name:  "Info logged at Debug level",
			level: slog.LevelDebug,
			logFunc: func(logger Logger) {
				logger.Info("info msg")
			},
			shouldOutput: true,
		},
		{
			name:  "Warn logged at Warn level",
			level: slog.LevelWarn,
			logFunc: func(logger Logger) {
				logger.Warn("warn msg")
			},
			shouldOutput: true,
		},
		{
			name:  "Error logged at Error level",
			level: slog.LevelError,
			logFunc: func(logger Logger) {
				logger.Error("error msg")
			},
			shouldOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewSlogLogger(&buf, tt.level)

			tt.logFunc(logger)

			output := buf.String()
			hasOutput := len(output) > 0

			if hasOutput != tt.shouldOutput {
				t.Errorf("Expected output=%v, got output=%v (len=%d)", tt.shouldOutput, hasOutput, len(output))
			}
		})
	}
}

// TestJSONLogger tests JSON format logging
func TestJSONLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewJSONLogger(&buf, slog.LevelInfo)

	logger.Info("test message", "key1", "value1", "key2", 42)

	output := buf.String()

	// Check that output contains JSON structure
	if !strings.Contains(output, `"msg"`) {
		t.Error("Output should contain JSON 'msg' field")
	}

	if !strings.Contains(output, "test message") {
		t.Error("Output should contain the log message")
	}

	// Should be valid JSON-ish (has braces)
	if !strings.Contains(output, "{") || !strings.Contains(output, "}") {
		t.Error("Output should be JSON formatted")
	}
}

// TestDefaultLogger tests the default logger factory
func TestDefaultLogger(t *testing.T) {
	logger := DefaultLogger()

	// Should not be nil
	if logger == nil {
		t.Fatal("DefaultLogger() should not return nil")
	}

	// Should implement Logger interface
	var _ = Logger(logger)

	// Should be a SlogLogger
	if _, ok := logger.(*SlogLogger); !ok {
		t.Error("DefaultLogger() should return *SlogLogger")
	}
}

// TestDebugLogger tests the debug logger factory
func TestDebugLogger(t *testing.T) {
	logger := DebugLogger()

	// Should not be nil
	if logger == nil {
		t.Fatal("DebugLogger() should not return nil")
	}

	// Should implement Logger interface
	var _ = Logger(logger)

	// Should be a SlogLogger
	if _, ok := logger.(*SlogLogger); !ok {
		t.Error("DebugLogger() should return *SlogLogger")
	}
}

// TestSanitizeHeaders tests header sanitization
func TestSanitizeHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string][]string
		expected map[string]string
	}{
		{
			name: "redact authorization header",
			input: map[string][]string{
				"Authorization": {"Bearer secret-token"},
				"Content-Type":  {"application/json"},
			},
			expected: map[string]string{
				"Authorization": "[REDACTED]",
				"Content-Type":  "application/json",
			},
		},
		{
			name: "preserve other headers",
			input: map[string][]string{
				"Accept":       {"application/json"},
				"Content-Type": {"application/json"},
				"User-Agent":   {"go-scalr/1.0"},
			},
			expected: map[string]string{
				"Accept":       "application/json",
				"Content-Type": "application/json",
				"User-Agent":   "go-scalr/1.0",
			},
		},
		{
			name:     "empty headers",
			input:    map[string][]string{},
			expected: map[string]string{},
		},
		{
			name: "multiple values - take first",
			input: map[string][]string{
				"Accept": {"application/json", "text/plain"},
			},
			expected: map[string]string{
				"Accept": "application/json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeHeaders(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d headers, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("Header %q = %q, want %q", key, result[key], expectedValue)
				}
			}
		})
	}
}

// TestSanitizeURL tests URL sanitization
func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple URL",
			input:    "https://example.scalr.io/api/v3/workspaces",
			expected: "https://example.scalr.io/api/v3/workspaces",
		},
		{
			name:     "URL with query params",
			input:    "https://example.scalr.io/api/v3/workspaces?page[size]=10",
			expected: "https://example.scalr.io/api/v3/workspaces?page[size]=10",
		},
		{
			name:     "empty URL",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeURL(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestLoggerInterface tests that all logger types implement the Logger interface
func TestLoggerInterface(t *testing.T) {
	tests := []struct {
		name   string
		logger Logger
	}{
		{"NoOpLogger", NewNoOpLogger()},
		{"SlogLogger", NewSlogLogger(&bytes.Buffer{}, slog.LevelInfo)},
		{"DefaultLogger", DefaultLogger()},
		{"DebugLogger", DebugLogger()},
		{"JSONLogger", NewJSONLogger(&bytes.Buffer{}, slog.LevelInfo)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			tt.logger.Debug("debug")
			tt.logger.Info("info")
			tt.logger.Warn("warn")
			tt.logger.Error("error")

			// With key-value pairs
			tt.logger.Debug("debug", "key", "value")
			tt.logger.Info("info", "key", "value")
			tt.logger.Warn("warn", "key", "value")
			tt.logger.Error("error", "key", "value")
		})
	}
}

// TestSlogLoggerWithKeyValues tests logging with key-value pairs
func TestSlogLoggerWithKeyValues(t *testing.T) {
	var buf bytes.Buffer
	logger := NewSlogLogger(&buf, slog.LevelInfo)

	logger.Info("message", "user", "john", "count", 42, "active", true)

	output := buf.String()

	expectedParts := []string{"message", "user", "john", "count", "42", "active", "true"}
	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Output %q should contain %q", output, part)
		}
	}
}
