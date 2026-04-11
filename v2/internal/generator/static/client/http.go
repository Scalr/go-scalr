package client

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// HTTPClient handles HTTP requests
type HTTPClient struct {
	baseURL           string
	token             string
	httpClient        *http.Client
	retryMax          int
	timeout           time.Duration
	defaultHeaders    map[string]string
	retryServerErrors bool
	logger            Logger
	userAgent         string
	sleepFunc         func(time.Duration) // For testing - allows mocking sleep
}

type HTTPClientOption func(*HTTPClient)

// WithRetryMax sets the maximum number of retries. Default: 5
func WithRetryMax(max int) HTTPClientOption {
	return func(c *HTTPClient) {
		c.retryMax = max
	}
}

// WithRetryServerErrors enables retrying on 5xx server errors (in addition to 429)
func WithRetryServerErrors(retry bool) HTTPClientOption {
	return func(c *HTTPClient) {
		c.retryServerErrors = retry
	}
}

// WithLogger sets a custom logger for the HTTP client
func WithLogger(logger Logger) HTTPClientOption {
	return func(c *HTTPClient) {
		c.logger = logger
	}
}

// WithTimeout sets the request timeout. Default: 30 seconds
func WithTimeout(timeout time.Duration) HTTPClientOption {
	return func(c *HTTPClient) {
		c.timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) HTTPClientOption {
	return func(c *HTTPClient) {
		c.httpClient = client
	}
}

// WithHeaders sets custom headers to be included in all requests
func WithHeaders(headers map[string]string) HTTPClientOption {
	return func(c *HTTPClient) {
		if c.defaultHeaders == nil {
			c.defaultHeaders = make(map[string]string)
		}
		for k, v := range headers {
			c.defaultHeaders[k] = v
		}
	}
}

// WithUserAgent sets a custom User-Agent header.
// This replaces the default User-Agent.
// For appending application info to the default User-Agent, use WithAppInfo instead.
func WithUserAgent(userAgent string) HTTPClientOption {
	return func(c *HTTPClient) {
		c.userAgent = userAgent
	}
}

// WithAppInfo appends application information to the User-Agent header.
// This is the preferred way for applications to identify themselves
// and preserve the go-scalr client information.
//
// Example:
//
//	WithAppInfo("terraform-provider-scalr", "v3.9.0")
//	Result: "terraform-provider-scalr/v3.9.0 go-scalr/v2.0.0-rc1 (Go 1.24; darwin/arm64)"
func WithAppInfo(appName, appVersion string) HTTPClientOption {
	return func(c *HTTPClient) {
		c.userAgent = UserAgentWithApp(appName, appVersion)
	}
}

// withSleepFunc sets a custom sleep function (for testing only - not exported)
func withSleepFunc(fn func(time.Duration)) HTTPClientOption {
	return func(c *HTTPClient) {
		c.sleepFunc = fn
	}
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(baseURL, token string, opts ...HTTPClientOption) *HTTPClient {
	// Clone the default transport and raise MaxIdleConnsPerHost to match MaxIdleConns.
	// Because this client talks to a single host, the default cap of 2 idle connections per host
	// would cause connection thrashing under concurrent use.
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = transport.MaxIdleConns

	client := &HTTPClient{
		baseURL:        baseURL,
		token:          token,
		retryMax:       5,
		timeout:        30 * time.Second,
		httpClient:     &http.Client{Transport: transport},
		defaultHeaders: make(map[string]string),
		logger:         NewNoOpLogger(), // Default: no logging
		userAgent:      UserAgent(),     // Default User-Agent
		sleepFunc:      time.Sleep,      // Default: real sleep
	}

	for _, opt := range opts {
		opt(client)
	}

	// Do NOT set httpClient.Timeout — we apply the timeout as a per-request
	// context deadline inside do() so that all timeout errors are uniformly
	// context.DeadlineExceeded rather than the opaque url.Error that
	// http.Client.Timeout produces.

	return client
}

// Close releases idle connections held by the underlying transport.
// Call this when the client is no longer needed to free resources promptly.
func (c *HTTPClient) Close() {
	c.httpClient.CloseIdleConnections()
}

// WithHeader creates a new HTTPClient with an additional default header
func (c *HTTPClient) WithHeader(key, value string) *HTTPClient {
	newClient := &HTTPClient{
		baseURL:           c.baseURL,
		token:             c.token,
		retryMax:          c.retryMax,
		timeout:           c.timeout,
		httpClient:        c.httpClient,
		defaultHeaders:    make(map[string]string),
		retryServerErrors: c.retryServerErrors,
		logger:            c.logger,
		userAgent:         c.userAgent,
		sleepFunc:         c.sleepFunc,
	}

	// Copy existing default headers
	for k, v := range c.defaultHeaders {
		newClient.defaultHeaders[k] = v
	}

	// Add new header
	newClient.defaultHeaders[key] = value

	return newClient
}

// Get performs a GET request
func (c *HTTPClient) Get(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, "GET", path, nil, headers)
}

// Post performs a POST request
func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, "POST", path, body, headers)
}

// Patch performs a PATCH request
func (c *HTTPClient) Patch(ctx context.Context, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, "PATCH", path, body, headers)
}

// Put performs a PUT request
func (c *HTTPClient) Put(ctx context.Context, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, "PUT", path, body, headers)
}

// Delete performs a DELETE request
func (c *HTTPClient) Delete(ctx context.Context, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, "DELETE", path, body, headers)
}

func (c *HTTPClient) do(ctx context.Context, method, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	// Apply per-request timeout via context so all deadline errors are
	// context.DeadlineExceeded regardless of whether the limit came from
	// WithTimeout or an external context.
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	url := c.baseURL + path

	// Log request start
	c.logger.Debug("Starting HTTP request",
		"method", method,
		"path", path,
		"url", sanitizeURL(url),
	)

	// Marshal the body once; retries reuse the same bytes.
	var marshaledBody []byte
	var bodyReader io.Reader
	if body != nil {
		var err error
		marshaledBody, err = json.Marshal(body)
		if err != nil {
			c.logger.Error("Failed to marshal request body",
				"error", err,
				"method", method,
				"path", path,
			)
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(marshaledBody)

		c.logger.Debug("Request body prepared",
			"method", method,
			"path", path,
			"bodySize", len(marshaledBody),
		)
	}

	var lastErr error
	var lastStatusCode int
	var retryAfter time.Duration     // backoff for next retry
	var lastRetryAfter time.Duration // most recent Retry-After from server (for final error)

	for attempt := 0; attempt <= c.retryMax; attempt++ {
		if attempt > 0 {
			var backoff time.Duration
			if retryAfter > 0 {
				// Honour the server's requested delay
				backoff = retryAfter
				retryAfter = 0
			} else {
				// Exponential backoff with jitter: base * 2^attempt + jitter
				// e.g., attempt 1: 1-2s, attempt 2: 2-4s, attempt 3: 4-8s
				base := time.Second
				maxBackoff := base * time.Duration(math.Pow(2, float64(attempt)))
				jitter := time.Duration(rand.Int63n(int64(maxBackoff)))
				backoff = maxBackoff + jitter

				// Cap backoff at 32 seconds
				if backoff > 32*time.Second {
					backoff = 32*time.Second + time.Duration(rand.Int63n(int64(time.Second)))
				}
			}

			// Log retry attempt
			c.logger.Warn("Retrying HTTP request",
				"attempt", attempt,
				"maxRetries", c.retryMax,
				"method", method,
				"path", path,
				"lastStatus", lastStatusCode,
				"backoff", backoff.String(),
			)

			// Check if context is cancelled before sleeping
			if ctx.Err() != nil {
				c.logger.Error("Request cancelled before retry backoff",
					"method", method,
					"path", path,
					"error", ctx.Err(),
				)
				return nil, ctx.Err()
			}

			// Sleep (can be mocked in tests)
			c.sleepFunc(backoff)

			// Reset body reader for retry (reuse already-marshaled bytes)
			if marshaledBody != nil {
				bodyReader = bytes.NewReader(marshaledBody)
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			c.logger.Error("Failed to create HTTP request",
				"error", err,
				"method", method,
				"path", path,
			)
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set default headers
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("User-Agent", c.userAgent)
		// Defaults to JSON:API content type. Some operations may override this if needed.
		req.Header.Set("Content-Type", "application/vnd.api+json")
		req.Header.Set("Accept", "application/vnd.api+json")

		// Set client default headers
		for key, value := range c.defaultHeaders {
			req.Header.Set(key, value)
		}

		// Set custom headers (can override default headers)
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		c.logger.Debug("Sending HTTP request",
			"method", method,
			"path", path,
		)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Context cancellation/deadline is never retryable — return immediately.
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if !isRetryableNetworkError(err) {
				c.logger.Error("Non-retryable network error",
					"error", err,
					"method", method,
					"path", path,
				)
				return nil, fmt.Errorf("request failed: %w", err)
			}
			lastErr = err
			c.logger.Error("HTTP request failed",
				"error", err,
				"method", method,
				"path", path,
				"attempt", attempt,
			)
			continue
		}

		c.logger.Debug("Received HTTP response",
			"method", method,
			"path", path,
			"status", resp.StatusCode,
			"statusText", resp.Status,
		)

		if c.shouldRetry(resp.StatusCode) {
			lastStatusCode = resp.StatusCode

			if resp.StatusCode == 429 {
				retryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
				if retryAfter > 0 {
					lastRetryAfter = retryAfter
				}
				c.logger.Warn("Rate limit encountered",
					"method", method,
					"path", path,
					"status", 429,
					"willRetry", true,
					"retryAfter", retryAfter.String(),
				)
			} else if resp.StatusCode >= 500 {
				c.logger.Warn("Server error encountered",
					"method", method,
					"path", path,
					"status", resp.StatusCode,
					"willRetry", true,
				)
			}

			_ = resp.Body.Close()
			lastErr = fmt.Errorf("request failed with status %d", resp.StatusCode)
			continue
		}

		if resp.StatusCode >= 400 {
			defer func() { _ = resp.Body.Close() }()

			bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MiB cap
			if err != nil {
				c.logger.Error("Failed to read error response body",
					"error", err,
					"method", method,
					"path", path,
					"status", resp.StatusCode,
				)
				return nil, &HTTPError{
					StatusCode: resp.StatusCode,
					Message:    "failed to read error response",
					Err:        err,
				}
			}

			// Try to parse JSON:API error
			var doc JSONAPIDocument
			var underlyingErr error
			message := string(bodyBytes)
			if err := json.Unmarshal(bodyBytes, &doc); err == nil && len(doc.Errors) > 0 {
				msgs := make([]string, len(doc.Errors))
				errs := make([]error, len(doc.Errors))
				for i, e := range doc.Errors {
					msgs[i] = e.Error()
					errs[i] = e
				}
				message = strings.Join(msgs, "; ")
				underlyingErr = errors.Join(errs...)
			}

			c.logger.Error("HTTP request returned error",
				"method", method,
				"path", path,
				"status", resp.StatusCode,
				"message", message,
			)

			// Return specific error types for common status codes
			switch resp.StatusCode {
			case 401:
				return nil, &UnauthorizedError{
					Message:    message,
					StatusCode: resp.StatusCode,
					Err:        underlyingErr,
				}
			case 403:
				return nil, &ForbiddenError{
					Message:    message,
					StatusCode: resp.StatusCode,
					Err:        underlyingErr,
				}
			case 404:
				return nil, &NotFoundError{
					Message:    message,
					StatusCode: resp.StatusCode,
					Err:        underlyingErr,
				}
			case 409:
				return nil, &ConflictError{
					Message:    message,
					StatusCode: resp.StatusCode,
					Err:        underlyingErr,
				}
			case 422:
				return nil, &UnprocessableEntityError{
					Message:    message,
					StatusCode: resp.StatusCode,
					Err:        underlyingErr,
				}
			case 429:
				return nil, &TooManyRequestsError{
					Message:    message,
					StatusCode: resp.StatusCode,
					Err:        underlyingErr,
				}
			default:
				// For other status codes, return generic HTTPError
				return nil, &HTTPError{
					StatusCode: resp.StatusCode,
					Message:    message,
					Err:        underlyingErr,
				}
			}
		}

		c.logger.Info("HTTP request completed successfully",
			"method", method,
			"path", path,
			"status", resp.StatusCode,
			"attempts", attempt+1,
		)

		return resp, nil
	}

	// All retries exhausted
	c.logger.Error("HTTP request failed after all retries",
		"method", method,
		"path", path,
		"maxRetries", c.retryMax,
		"lastError", lastErr,
		"lastStatus", lastStatusCode,
	)

	// Return a typed TooManyRequestsError so callers can inspect RetryAfter.
	if lastStatusCode == 429 {
		return nil, &TooManyRequestsError{
			StatusCode: 429,
			RetryAfter: lastRetryAfter,
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", c.retryMax, lastErr)
	}

	return nil, fmt.Errorf("request failed after %d retries", c.retryMax)
}

// isRetryableNetworkError returns true if the error is a transient network condition
// worth retrying. Permanent failures (DNS not found, TLS certificate errors) return
// false so the caller gets an immediate result instead of waiting through backoff.
func isRetryableNetworkError(err error) bool {
	// Context cancellation/deadline exceeded is never retryable.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	// DNS errors: only retry if the resolver itself says it's temporary.
	// A not-found or authoritative NXDOMAIN will never succeed on retry.
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.Temporary()
	}
	// TLS certificate errors are permanent — retrying won't fix an invalid cert.
	var certInvalid x509.CertificateInvalidError
	if errors.As(err, &certInvalid) {
		return false
	}
	var unknownAuth x509.UnknownAuthorityError
	if errors.As(err, &unknownAuth) {
		return false
	}
	// All other network errors (connection reset, brief timeout, etc.) are
	// treated as transient and worth retrying.
	return true
}

// parseRetryAfter parses the Retry-After header value and returns the duration to wait.
// Supports both integer (seconds) and HTTP-date formats per RFC 7231.
// Returns 0 if the header is absent, malformed, or in the past.
func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}
	// Try integer format first (number of seconds)
	if secs, err := strconv.Atoi(header); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	// Try HTTP-date format
	if t, err := http.ParseTime(header); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}

func (c *HTTPClient) shouldRetry(statusCode int) bool {
	// Always retry on 429 (rate limit)
	if statusCode == 429 {
		return true
	}

	// Optionally retry on 5xx server errors
	if statusCode >= 500 && c.retryServerErrors {
		return true
	}

	return false
}
