package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
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
}

type HTTPClientOption func(*HTTPClient)

// WithRetryMax sets the maximum number of retries
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

// WithTimeout sets the request timeout
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

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(baseURL, token string, opts ...HTTPClientOption) *HTTPClient {
	client := &HTTPClient{
		baseURL:        baseURL,
		token:          token,
		retryMax:       5,
		timeout:        30 * time.Second,
		httpClient:     &http.Client{},
		defaultHeaders: make(map[string]string),
		logger:         NewNoOpLogger(), // Default: no logging
		userAgent:      UserAgent(),     // Default User-Agent
	}

	for _, opt := range opts {
		opt(client)
	}

	client.httpClient.Timeout = client.timeout

	return client
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
func (c *HTTPClient) Delete(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, "DELETE", path, nil, headers)
}

func (c *HTTPClient) do(ctx context.Context, method, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	url := c.baseURL + path

	// Log request start
	c.logger.Debug("Starting HTTP request",
		"method", method,
		"path", path,
		"url", sanitizeURL(url),
	)

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			c.logger.Error("Failed to marshal request body",
				"error", err,
				"method", method,
				"path", path,
			)
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)

		c.logger.Debug("Request body prepared",
			"method", method,
			"path", path,
			"bodySize", len(bodyBytes),
		)
	}

	var lastErr error
	var lastStatusCode int

	for attempt := 0; attempt <= c.retryMax; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter: base * 2^attempt + jitter
			// e.g., attempt 1: 1-2s, attempt 2: 2-4s, attempt 3: 4-8s
			base := time.Second
			maxBackoff := base * time.Duration(math.Pow(2, float64(attempt)))
			jitter := time.Duration(rand.Int63n(int64(maxBackoff)))
			backoff := maxBackoff + jitter

			// Cap backoff at 32 seconds
			if backoff > 32*time.Second {
				backoff = 32*time.Second + time.Duration(rand.Int63n(int64(time.Second)))
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

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				c.logger.Error("Request cancelled during retry backoff",
					"method", method,
					"path", path,
					"error", ctx.Err(),
				)
				return nil, ctx.Err()
			}

			// Reset body reader for retry
			if body != nil {
				bodyBytes, _ := json.Marshal(body)
				bodyReader = bytes.NewReader(bodyBytes)
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
			"headers", sanitizeHeaders(req.Header),
		)

		resp, err := c.httpClient.Do(req)
		if err != nil {
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
			_ = resp.Body.Close() // Ignore error on retry

			if resp.StatusCode == 429 {
				c.logger.Warn("Rate limit encountered",
					"method", method,
					"path", path,
					"status", 429,
					"willRetry", true,
				)
			} else if resp.StatusCode >= 500 {
				c.logger.Warn("Server error encountered",
					"method", method,
					"path", path,
					"status", resp.StatusCode,
					"willRetry", true,
				)
			}

			lastErr = fmt.Errorf("request failed with status %d", resp.StatusCode)
			continue
		}

		if resp.StatusCode >= 400 {
			defer func() { _ = resp.Body.Close() }()

			bodyBytes, err := io.ReadAll(resp.Body)
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
				underlyingErr = doc.Errors[0]
				message = doc.Errors[0].Error()
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

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", c.retryMax, lastErr)
	}

	return nil, fmt.Errorf("request failed after %d retries", c.retryMax)
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
