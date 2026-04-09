package client

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewHTTPClient tests creating a new HTTP client
func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient("https://example.com", "test-token")

	if client == nil {
		t.Fatal("NewHTTPClient() returned nil")
	}

	if client.baseURL != "https://example.com" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, "https://example.com")
	}

	if client.token != "test-token" {
		t.Errorf("token = %q, want %q", client.token, "test-token")
	}

	// Check defaults
	if client.retryMax != 5 {
		t.Errorf("retryMax = %d, want 5", client.retryMax)
	}

	if client.timeout != 30*time.Second {
		t.Errorf("timeout = %v, want %v", client.timeout, 30*time.Second)
	}

	if client.retryServerErrors {
		t.Error("retryServerErrors should default to false")
	}
}

// TestHTTPClientOptions tests HTTP client options
func TestHTTPClientOptions(t *testing.T) {
	customHTTP := &http.Client{Timeout: 10 * time.Second}
	headers := map[string]string{"X-Custom": "value"}

	client := NewHTTPClient(
		"https://example.com",
		"token",
		WithRetryMax(10),
		WithRetryServerErrors(true),
		WithTimeout(60*time.Second),
		WithHTTPClient(customHTTP),
		WithHeaders(headers),
	)

	if client.retryMax != 10 {
		t.Errorf("retryMax = %d, want 10", client.retryMax)
	}

	if !client.retryServerErrors {
		t.Error("retryServerErrors should be true")
	}

	if client.timeout != 60*time.Second {
		t.Errorf("timeout = %v, want %v", client.timeout, 60*time.Second)
	}

	if client.httpClient != customHTTP {
		t.Error("httpClient should be the custom client")
	}

	if client.defaultHeaders["X-Custom"] != "value" {
		t.Error("defaultHeaders should contain custom header")
	}
}

// TestWithHeader tests creating a client with additional headers
func TestWithHeader(t *testing.T) {
	original := NewHTTPClient("https://example.com", "token")
	original.defaultHeaders["Existing"] = "value"

	// Create new client with additional header
	withHeader := original.WithHeader("Prefer", "profile=internal")

	// Original should be unchanged
	if _, ok := original.defaultHeaders["Prefer"]; ok {
		t.Error("Original client should not have Prefer header")
	}

	// New client should have both headers
	if withHeader.defaultHeaders["Existing"] != "value" {
		t.Error("New client should inherit existing headers")
	}

	if withHeader.defaultHeaders["Prefer"] != "profile=internal" {
		t.Error("New client should have new header")
	}

	// Verify other fields are copied
	if withHeader.baseURL != original.baseURL {
		t.Error("baseURL should be copied")
	}

	if withHeader.token != original.token {
		t.Error("token should be copied")
	}
}

// TestHTTPClientGet tests GET requests
func TestHTTPClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Method = %q, want GET", r.Method)
		}

		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("Authorization header not set correctly")
		}

		if r.Header.Get("Content-Type") != "application/vnd.api+json" {
			t.Error("Content-Type header not set correctly")
		}

		// Send response
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {"id": "123"}}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, "test-token", WithRetryMax(0))
	resp, err := client.Get(context.Background(), "/test", nil)

	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestHTTPClientPost tests POST requests
func TestHTTPClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %q, want POST", r.Method)
		}

		// Read and verify body
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		_ = json.Unmarshal(body, &data)

		if data["name"] != "test" {
			t.Error("Request body not sent correctly")
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data": {"id": "456"}}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, "test-token", WithRetryMax(0))
	resp, err := client.Post(context.Background(), "/test", map[string]string{"name": "test"}, nil)

	if err != nil {
		t.Fatalf("Post() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}

// TestHTTPClientPatch tests PATCH requests
func TestHTTPClientPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("Method = %q, want PATCH", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, "test-token", WithRetryMax(0))
	resp, err := client.Patch(context.Background(), "/test", map[string]string{"name": "updated"}, nil)

	if err != nil {
		t.Fatalf("Patch() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestHTTPClientDelete tests DELETE requests
func TestHTTPClientDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, "test-token", WithRetryMax(0))
	resp, err := client.Delete(context.Background(), "/test", nil, nil)

	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

// TestHTTPClientDeleteWithBody tests DELETE with a request body (JSON:API relationship delete)
func TestHTTPClientDeleteWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %s, want DELETE", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization header = %q, want %q", r.Header.Get("Authorization"), "Bearer test-token")
		}

		// Verify body is present
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read body: %v", err)
		}
		if len(body) == 0 {
			t.Error("Expected body in DELETE request, got empty body")
		}

		// Verify it's valid JSON
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			t.Errorf("Body is not valid JSON: %v", err)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, "test-token", WithRetryMax(0))
	body := map[string]interface{}{
		"data": []map[string]interface{}{
			{"id": "ws-123", "type": "workspaces"},
		},
	}
	resp, err := client.Delete(context.Background(), "/test", body, nil)

	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

// TestHTTPClient401Error tests 401 Unauthorized responses
func TestHTTPClient401Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors": [{"title": "Unauthorized", "detail": "Invalid token"}]}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, "test-token", WithRetryMax(0))
	_, err := client.Get(context.Background(), "/test", nil)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	var authErr *UnauthorizedError
	if !errors.As(err, &authErr) {
		t.Fatalf("Error type = %T, want *UnauthorizedError", err)
	}

	if !errors.Is(err, ErrUnauthorized) {
		t.Error("Error should match ErrUnauthorized sentinel")
	}
}

// TestHTTPClient404Error tests 404 Not Found responses
func TestHTTPClient404Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors": [{"title": "Not Found"}]}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, "test-token", WithRetryMax(0))
	_, err := client.Get(context.Background(), "/test", nil)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("Error type = %T, want *NotFoundError", err)
	}

	if !errors.Is(err, ErrNotFound) {
		t.Error("Error should match ErrNotFound sentinel")
	}
}

// TestHTTPClientRetryOn429 tests retry logic for 429 responses
func TestHTTPClientRetryOn429(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	// Use mock sleep to make test fast
	client := NewHTTPClient(server.URL, "test-token", WithRetryMax(5), withSleepFunc(func(time.Duration) {}))
	resp, err := client.Get(context.Background(), "/test", nil)

	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestParseRetryAfter tests the Retry-After header parser.
func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected time.Duration
	}{
		{name: "empty header", header: "", expected: 0},
		{name: "integer seconds", header: "30", expected: 30 * time.Second},
		{name: "one second", header: "1", expected: time.Second},
		{name: "zero seconds", header: "0", expected: 0},
		{name: "negative integer", header: "-5", expected: 0},
		{name: "invalid string", header: "soon", expected: 0},
		{name: "past HTTP-date", header: "Thu, 01 Jan 1970 00:00:00 GMT", expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRetryAfter(tt.header)
			if got != tt.expected {
				t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.header, got, tt.expected)
			}
		})
	}

	t.Run("future HTTP-date", func(t *testing.T) {
		future := time.Now().Add(10 * time.Second).UTC().Format(http.TimeFormat)
		got := parseRetryAfter(future)
		if got <= 0 || got > 11*time.Second {
			t.Errorf("parseRetryAfter(future date) = %v, expected ~10s", got)
		}
	})
}

// TestHTTPClientRetryAfterHeader verifies that the Retry-After header value is used
// as the sleep duration instead of exponential backoff when present on a 429 response.
func TestHTTPClientRetryAfterHeader(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.Header().Set("Retry-After", "42")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	var sleptFor time.Duration
	client := NewHTTPClient(server.URL, "test-token",
		WithRetryMax(3),
		withSleepFunc(func(d time.Duration) { sleptFor = d }),
	)
	resp, err := client.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if sleptFor != 42*time.Second {
		t.Errorf("sleep duration = %v, want 42s (from Retry-After header)", sleptFor)
	}
}

// TestHTTPClientRetryOn5xx tests retry logic for 5xx responses (when enabled)
func TestHTTPClientRetryOn5xx(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	// Use mock sleep to make test fast
	client := NewHTTPClient(
		server.URL,
		"test-token",
		WithRetryMax(5),
		WithRetryServerErrors(true),
		withSleepFunc(func(time.Duration) {}),
	)
	resp, err := client.Get(context.Background(), "/test", nil)

	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

// TestHTTPClientNoRetryOn5xxByDefault tests that 5xx doesn't retry by default
func TestHTTPClientNoRetryOn5xxByDefault(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors": [{"title": "Server Error"}]}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, "test-token", WithRetryMax(5))
	_, err := client.Get(context.Background(), "/test", nil)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if attempts != 1 {
		t.Errorf("Should only attempt once (no retry on 5xx by default), got %d attempts", attempts)
	}
}

// TestHTTPClientMaxRetriesExhausted tests behavior when max retries are exhausted
func TestHTTPClientMaxRetriesExhausted(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	// Use mock sleep to make test fast
	client := NewHTTPClient(server.URL, "test-token", WithRetryMax(2), withSleepFunc(func(time.Duration) {}))
	_, err := client.Get(context.Background(), "/test", nil)

	if err == nil {
		t.Fatal("Expected error after exhausting retries")
	}

	// Initial attempt + 2 retries = 3 total attempts
	if attempts != 3 {
		t.Errorf("Expected 3 attempts (1 initial + 2 retries), got %d", attempts)
	}

	if !strings.Contains(err.Error(), "after 2 retries") {
		t.Errorf("Error should mention retries, got: %v", err)
	}
}

// TestHTTPClientContextCancellation tests context cancellation
func TestHTTPClientContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := NewHTTPClient(server.URL, "test-token", WithRetryMax(0))
	_, err := client.Get(ctx, "/test", nil)

	if err == nil {
		t.Fatal("Expected error from cancelled context")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
}

// TestHTTPClientCustomHeaders tests custom headers
func TestHTTPClientCustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "value" {
			t.Error("Custom header not set")
		}

		if r.Header.Get("X-Request") != "specific" {
			t.Error("Request-specific header not set")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(
		server.URL,
		"test-token",
		WithRetryMax(0),
		WithHeaders(map[string]string{"X-Custom": "value"}),
	)

	_, err := client.Get(context.Background(), "/test", map[string]string{"X-Request": "specific"})

	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
}

// TestHTTPClientLogger tests logging integration
func TestHTTPClientLogger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	// Create a mock logger to capture logs
	mockLogger := &MockLogger{logs: make([]LogEntry, 0)}

	client := NewHTTPClient(
		server.URL,
		"test-token",
		WithRetryMax(0),
		WithLogger(mockLogger),
	)

	_, err := client.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	// Verify logs were generated
	if len(mockLogger.logs) == 0 {
		t.Error("Expected log entries, got none")
	}

	// Should have Debug, Info logs at minimum
	hasDebug := false
	hasInfo := false
	for _, log := range mockLogger.logs {
		if log.Level == "Debug" {
			hasDebug = true
		}
		if log.Level == "Info" {
			hasInfo = true
		}
	}

	if !hasDebug {
		t.Error("Expected Debug log entries")
	}

	if !hasInfo {
		t.Error("Expected Info log entries")
	}
}

// TestShouldRetry tests the shouldRetry logic
func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name              string
		statusCode        int
		retryServerErrors bool
		want              bool
	}{
		{"429 always retries", 429, false, true},
		{"429 with server errors enabled", 429, true, true},
		{"500 without server errors enabled", 500, false, false},
		{"500 with server errors enabled", 500, true, true},
		{"503 with server errors enabled", 503, true, true},
		{"400 never retries", 400, false, false},
		{"404 never retries", 404, false, false},
		{"200 never retries", 200, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &HTTPClient{retryServerErrors: tt.retryServerErrors}
			if got := client.shouldRetry(tt.statusCode); got != tt.want {
				t.Errorf("shouldRetry(%d) = %v, want %v", tt.statusCode, got, tt.want)
			}
		})
	}
}

// TestUserAgentDefault tests that default User-Agent is set
func TestUserAgentDefault(t *testing.T) {
	var capturedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, "test-token", WithRetryMax(0))
	_, err := client.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if capturedUA == "" {
		t.Error("User-Agent header was not set")
	}

	if !strings.Contains(capturedUA, "go-scalr") {
		t.Errorf("User-Agent should contain 'go-scalr', got: %s", capturedUA)
	}

	t.Logf("Default User-Agent: %s", capturedUA)
}

// TestUserAgentCustom tests WithUserAgent option
func TestUserAgentCustom(t *testing.T) {
	var capturedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	customUA := "MyCustomClient/1.0.0"
	client := NewHTTPClient(
		server.URL,
		"test-token",
		WithRetryMax(0),
		WithUserAgent(customUA),
	)

	_, err := client.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if capturedUA != customUA {
		t.Errorf("User-Agent = %q, want %q", capturedUA, customUA)
	}

	t.Logf("Custom User-Agent: %s", capturedUA)
}

// TestUserAgentWithAppInfo tests WithAppInfo option
func TestUserAgentWithAppInfo(t *testing.T) {
	var capturedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	client := NewHTTPClient(
		server.URL,
		"test-token",
		WithRetryMax(0),
		WithAppInfo("terraform-provider-scalr", "3.9.0"),
	)

	_, err := client.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	// Should contain both app and client
	if !strings.Contains(capturedUA, "terraform-provider-scalr/3.9.0") {
		t.Errorf("User-Agent should contain 'terraform-provider-scalr/3.9.0', got: %s", capturedUA)
	}

	if !strings.Contains(capturedUA, "go-scalr") {
		t.Errorf("User-Agent should contain 'go-scalr', got: %s", capturedUA)
	}

	// App should come before client
	appIndex := strings.Index(capturedUA, "terraform-provider-scalr")
	clientIndex := strings.Index(capturedUA, "go-scalr")
	if appIndex > clientIndex {
		t.Errorf("App should come before client in User-Agent, got: %s", capturedUA)
	}

	t.Logf("App User-Agent: %s", capturedUA)
}

// TestUserAgentCanBeOverridden tests that custom headers can override User-Agent
func TestUserAgentCanBeOverridden(t *testing.T) {
	var capturedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, "test-token", WithRetryMax(0))

	// Pass User-Agent in custom headers (should override)
	customHeaders := map[string]string{
		"User-Agent": "OverriddenAgent/1.0.0",
	}

	_, err := client.Get(context.Background(), "/test", customHeaders)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if capturedUA != "OverriddenAgent/1.0.0" {
		t.Errorf("User-Agent should be overridden, got: %s", capturedUA)
	}

	t.Logf("Overridden User-Agent: %s", capturedUA)
}

// TestUserAgentPreservedInWithHeader tests that UserAgent is copied in WithHeader
func TestUserAgentPreservedInWithHeader(t *testing.T) {
	var capturedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	originalClient := NewHTTPClient(
		server.URL,
		"test-token",
		WithRetryMax(0),
		WithAppInfo("myapp", "1.0.0"),
	)

	// Create new client with additional header
	newClient := originalClient.WithHeader("X-Custom", "value")

	_, err := newClient.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	// User-Agent should be preserved
	if !strings.Contains(capturedUA, "myapp/1.0.0") {
		t.Errorf("User-Agent should be preserved in WithHeader, got: %s", capturedUA)
	}

	if !strings.Contains(capturedUA, "go-scalr") {
		t.Errorf("User-Agent should contain go-scalr, got: %s", capturedUA)
	}

	t.Logf("Preserved User-Agent: %s", capturedUA)
}

// MockLogger for testing
type MockLogger struct {
	logs []LogEntry
}

type LogEntry struct {
	Level   string
	Message string
	KeyVals []interface{}
}

func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.logs = append(m.logs, LogEntry{"Debug", msg, keysAndValues})
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.logs = append(m.logs, LogEntry{"Info", msg, keysAndValues})
}

func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.logs = append(m.logs, LogEntry{"Warn", msg, keysAndValues})
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.logs = append(m.logs, LogEntry{"Error", msg, keysAndValues})
}

// TestIsRetryableNetworkError tests the network error classifier.
func TestIsRetryableNetworkError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "context canceled",
			err:       context.Canceled,
			retryable: false,
		},
		{
			name:      "context deadline exceeded",
			err:       context.DeadlineExceeded,
			retryable: false,
		},
		{
			name:      "DNS not found",
			err:       &net.DNSError{IsNotFound: true},
			retryable: false,
		},
		{
			name:      "DNS temporary",
			err:       &net.DNSError{IsTemporary: true},
			retryable: true,
		},
		{
			name:      "x509 certificate invalid",
			err:       x509.CertificateInvalidError{},
			retryable: false,
		},
		{
			name:      "x509 unknown authority",
			err:       x509.UnknownAuthorityError{},
			retryable: false,
		},
		{
			name:      "generic error is retryable",
			err:       errors.New("connection reset by peer"),
			retryable: true,
		},
		{
			name:      "wrapped context canceled",
			err:       &net.OpError{Err: context.Canceled},
			retryable: false,
		},
		{
			name:      "wrapped DNS not found",
			err:       &net.OpError{Err: &net.DNSError{IsNotFound: true}},
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryableNetworkError(tt.err)
			if got != tt.retryable {
				t.Errorf("isRetryableNetworkError(%v) = %v, want %v", tt.err, got, tt.retryable)
			}
		})
	}
}

// TestHTTPClientClose verifies that Close() does not panic and drains idle connections.
func TestHTTPClientClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	c := NewHTTPClient(server.URL, "token", WithRetryMax(0))
	resp, err := c.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	_ = resp.Body.Close()

	// Close must not panic; idle connection should be released.
	c.Close()
}

// TestHTTPClientMultiError verifies that all JSON:API errors are joined into the
// returned error, not just the first one.
func TestHTTPClientMultiError(t *testing.T) {
	body := `{"errors":[{"title":"first","detail":"d1"},{"title":"second","detail":"d2"}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	c := NewHTTPClient(server.URL, "token", WithRetryMax(0))
	_, err := c.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, "first") || !strings.Contains(msg, "second") {
		t.Errorf("expected both errors in message, got: %s", msg)
	}

	// Underlying errors should be accessible as *JSONAPIError via errors.As
	var apiErr *JSONAPIError
	if !errors.As(err, &apiErr) {
		t.Error("expected to find *JSONAPIError via errors.As")
	}
}

// TestHTTPClientNonRetryableDNSError verifies that a permanent DNS error does not
// trigger retries — the client should fail immediately.
func TestHTTPClientNonRetryableDNSError(t *testing.T) {
	// Use an invalid domain that will produce a permanent DNS failure.
	client := NewHTTPClient("http://this.hostname.does.not.exist.invalid", "token",
		WithRetryMax(5),
		WithTimeout(5*time.Second),
		withSleepFunc(func(time.Duration) { t.Error("sleep called — retry happened unexpectedly") }),
	)
	_, err := client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
