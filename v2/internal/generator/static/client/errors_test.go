package client

import (
	"errors"
	"testing"
)

// TestSentinelErrors tests that sentinel errors are defined and usable with errors.Is
func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
		want     string
	}{
		{"ErrUnauthorized", ErrUnauthorized, "unauthorized"},
		{"ErrForbidden", ErrForbidden, "forbidden"},
		{"ErrNotFound", ErrNotFound, "not found"},
		{"ErrConflict", ErrConflict, "conflict"},
		{"ErrUnprocessableEntity", ErrUnprocessableEntity, "unprocessable entity"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.sentinel == nil {
				t.Error("Sentinel error should not be nil")
			}
			if tt.sentinel.Error() != tt.want {
				t.Errorf("Error message = %q, want %q", tt.sentinel.Error(), tt.want)
			}
		})
	}
}

// TestUnauthorizedError tests the UnauthorizedError type
func TestUnauthorizedError(t *testing.T) {
	tests := []struct {
		name        string
		err         *UnauthorizedError
		expectedMsg string
	}{
		{
			name:        "basic message",
			err:         &UnauthorizedError{Message: "invalid token"},
			expectedMsg: "unauthorized: invalid token",
		},
		{
			name:        "with status code",
			err:         &UnauthorizedError{Message: "invalid token", StatusCode: 401},
			expectedMsg: "unauthorized: invalid token",
		},
		{
			name:        "with wrapped error",
			err:         &UnauthorizedError{Err: errors.New("token expired")},
			expectedMsg: "unauthorized: token expired",
		},
		{
			name:        "no message or error",
			err:         &UnauthorizedError{},
			expectedMsg: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expectedMsg {
				t.Errorf("Error() = %q, want %q", got, tt.expectedMsg)
			}

			// Test errors.Is
			if !errors.Is(tt.err, ErrUnauthorized) {
				t.Error("errors.Is(err, ErrUnauthorized) should be true")
			}

			// Test Unwrap
			if tt.err.Err != nil {
				if unwrapped := tt.err.Unwrap(); unwrapped != tt.err.Err {
					t.Errorf("Unwrap() = %v, want %v", unwrapped, tt.err.Err)
				}
			}
		})
	}
}

// TestForbiddenError tests the ForbiddenError type
func TestForbiddenError(t *testing.T) {
	err := &ForbiddenError{
		Message:    "insufficient permissions",
		StatusCode: 403,
	}

	if got := err.Error(); got != "forbidden: insufficient permissions" {
		t.Errorf("Error() = %q, want %q", got, "forbidden: insufficient permissions")
	}

	if !errors.Is(err, ErrForbidden) {
		t.Error("errors.Is(err, ErrForbidden) should be true")
	}

	// Test with wrapped error
	underlyingErr := errors.New("access denied")
	err = &ForbiddenError{Err: underlyingErr}

	if got := err.Error(); got != "forbidden: access denied" {
		t.Errorf("Error() = %q, want %q", got, "forbidden: access denied")
	}

	if err.Unwrap() != underlyingErr {
		t.Error("Unwrap() should return underlying error")
	}
}

// TestNotFoundError tests the NotFoundError type
func TestNotFoundError(t *testing.T) {
	err := &NotFoundError{
		Message:    "workspace not found",
		StatusCode: 404,
	}

	if got := err.Error(); got != "not found: workspace not found" {
		t.Errorf("Error() = %q, want %q", got, "not found: workspace not found")
	}

	if !errors.Is(err, ErrNotFound) {
		t.Error("errors.Is(err, ErrNotFound) should be true")
	}
}

// TestConflictError tests the ConflictError type
func TestConflictError(t *testing.T) {
	err := &ConflictError{
		Message:    "resource already exists",
		StatusCode: 409,
	}

	if got := err.Error(); got != "conflict: resource already exists" {
		t.Errorf("Error() = %q, want %q", got, "conflict: resource already exists")
	}

	if !errors.Is(err, ErrConflict) {
		t.Error("errors.Is(err, ErrConflict) should be true")
	}
}

// TestUnprocessableEntityError tests the UnprocessableEntityError type
func TestUnprocessableEntityError(t *testing.T) {
	err := &UnprocessableEntityError{
		Message:    "validation failed",
		StatusCode: 422,
	}

	if got := err.Error(); got != "unprocessable entity: validation failed" {
		t.Errorf("Error() = %q, want %q", got, "unprocessable entity: validation failed")
	}

	if !errors.Is(err, ErrUnprocessableEntity) {
		t.Error("errors.Is(err, ErrUnprocessableEntity) should be true")
	}
}

// TestHTTPError tests the generic HTTPError type
func TestHTTPError(t *testing.T) {
	tests := []struct {
		name        string
		err         *HTTPError
		expectedMsg string
	}{
		{
			name:        "with message",
			err:         &HTTPError{StatusCode: 500, Message: "server error"},
			expectedMsg: "HTTP 500: server error",
		},
		{
			name:        "with wrapped error",
			err:         &HTTPError{StatusCode: 503, Err: errors.New("service unavailable")},
			expectedMsg: "HTTP 503: service unavailable",
		},
		{
			name:        "status code only",
			err:         &HTTPError{StatusCode: 429},
			expectedMsg: "HTTP 429",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expectedMsg {
				t.Errorf("Error() = %q, want %q", got, tt.expectedMsg)
			}

			if tt.err.Err != nil {
				if unwrapped := tt.err.Unwrap(); unwrapped != tt.err.Err {
					t.Error("Unwrap() should return underlying error")
				}
			}
		})
	}
}

// TestErrorWrapping tests that errors properly implement error wrapping
func TestErrorWrapping(t *testing.T) {
	// Create a chain: JSONAPIError -> UnauthorizedError
	apiErr := &JSONAPIError{
		Title:  "Authentication Failed",
		Detail: "Token expired",
	}

	authErr := &UnauthorizedError{
		Message: "auth failed",
		Err:     apiErr,
	}

	// Test errors.As with wrapped error
	var jsonAPIErr *JSONAPIError
	if !errors.As(authErr, &jsonAPIErr) {
		t.Error("errors.As should find wrapped JSONAPIError")
	}

	if jsonAPIErr.Title != "Authentication Failed" {
		t.Errorf("Wrapped error title = %q, want %q", jsonAPIErr.Title, "Authentication Failed")
	}

	// Test errors.Is with sentinel
	if !errors.Is(authErr, ErrUnauthorized) {
		t.Error("errors.Is should match sentinel error")
	}
}

// TestErrorComparison tests comparing different error types
func TestErrorComparison(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		sentinel error
		want     bool
	}{
		{
			name:     "UnauthorizedError matches ErrUnauthorized",
			err:      &UnauthorizedError{Message: "test"},
			sentinel: ErrUnauthorized,
			want:     true,
		},
		{
			name:     "UnauthorizedError doesn't match ErrForbidden",
			err:      &UnauthorizedError{Message: "test"},
			sentinel: ErrForbidden,
			want:     false,
		},
		{
			name:     "NotFoundError matches ErrNotFound",
			err:      &NotFoundError{Message: "test"},
			sentinel: ErrNotFound,
			want:     true,
		},
		{
			name:     "HTTPError doesn't match any sentinel",
			err:      &HTTPError{StatusCode: 500},
			sentinel: ErrUnauthorized,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, tt.sentinel); got != tt.want {
				t.Errorf("errors.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}
