package client

import (
	"errors"
	"fmt"
)

// Sentinel errors for common HTTP status codes to use in errors.Is() checks
var (
	// ErrUnauthorized indicates a 401 Unauthorized response
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates a 403 Forbidden response
	ErrForbidden = errors.New("forbidden")

	// ErrNotFound indicates a 404 Not Found response
	ErrNotFound = errors.New("not found")

	// ErrConflict indicates a 409 Conflict response
	ErrConflict = errors.New("conflict")

	// ErrUnprocessableEntity indicates a 422 Unprocessable Entity response
	ErrUnprocessableEntity = errors.New("unprocessable entity")

	// ErrTooManyRequests indicates a 429 Too Many Requests response
	ErrTooManyRequests = errors.New("too many requests")
)

// UnauthorizedError represents a 401 Unauthorized response
type UnauthorizedError struct {
	Message    string
	StatusCode int
	// Underlying JSONAPIError if available
	Err error
}

func (e *UnauthorizedError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("unauthorized: %s", e.Err.Error())
	}
	if e.Message != "" {
		return fmt.Sprintf("unauthorized: %s", e.Message)
	}
	return "unauthorized"
}

func (e *UnauthorizedError) Unwrap() error        { return e.Err }
func (e *UnauthorizedError) Is(target error) bool { return target == ErrUnauthorized }

// ForbiddenError represents a 403 Forbidden response
type ForbiddenError struct {
	Message    string
	StatusCode int
	Err        error
}

func (e *ForbiddenError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("forbidden: %s", e.Err.Error())
	}
	if e.Message != "" {
		return fmt.Sprintf("forbidden: %s", e.Message)
	}
	return "forbidden"
}

func (e *ForbiddenError) Unwrap() error        { return e.Err }
func (e *ForbiddenError) Is(target error) bool { return target == ErrForbidden }

// NotFoundError represents a 404 Not Found response
type NotFoundError struct {
	Message    string
	StatusCode int
	Err        error
}

func (e *NotFoundError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("not found: %s", e.Err.Error())
	}
	if e.Message != "" {
		return fmt.Sprintf("not found: %s", e.Message)
	}
	return "not found"
}

func (e *NotFoundError) Unwrap() error        { return e.Err }
func (e *NotFoundError) Is(target error) bool { return target == ErrNotFound }

// ConflictError represents a 409 Conflict response
type ConflictError struct {
	Message    string
	StatusCode int
	Err        error
}

func (e *ConflictError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("conflict: %s", e.Err.Error())
	}
	if e.Message != "" {
		return fmt.Sprintf("conflict: %s", e.Message)
	}
	return "conflict"
}

func (e *ConflictError) Unwrap() error        { return e.Err }
func (e *ConflictError) Is(target error) bool { return target == ErrConflict }

// UnprocessableEntityError represents a 422 Unprocessable Entity response
type UnprocessableEntityError struct {
	Message    string
	StatusCode int
	Err        error
}

func (e *UnprocessableEntityError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("unprocessable entity: %s", e.Err.Error())
	}
	if e.Message != "" {
		return fmt.Sprintf("unprocessable entity: %s", e.Message)
	}
	return "unprocessable entity"
}

func (e *UnprocessableEntityError) Unwrap() error        { return e.Err }
func (e *UnprocessableEntityError) Is(target error) bool { return target == ErrUnprocessableEntity }

// TooManyRequestsError represents a 429 Too Many Requests response
type TooManyRequestsError struct {
	Message    string
	StatusCode int
	Err        error
}

func (e *TooManyRequestsError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("too many requests: %s", e.Err.Error())
	}
	if e.Message != "" {
		return fmt.Sprintf("too many requests: %s", e.Message)
	}
	return "too many requests"
}

func (e *TooManyRequestsError) Unwrap() error        { return e.Err }
func (e *TooManyRequestsError) Is(target error) bool { return target == ErrTooManyRequests }

// HTTPError represents a generic HTTP error response for status codes
// that don't have specific error types
type HTTPError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Err.Error())
	}
	if e.Message != "" {
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

func (e *HTTPError) Unwrap() error { return e.Err }
