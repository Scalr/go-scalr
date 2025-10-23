package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// JSONAPIError represents a JSON:API error object
type JSONAPIError struct {
	Title  string              `json:"title"`
	Status string              `json:"status"`
	Detail string              `json:"detail"`
	Source *JSONAPIErrorSource `json:"source"`
	Code   string              `json:"code"`
}

// JSONAPIErrorSource represents the source of a JSON:API error
type JSONAPIErrorSource struct {
	Pointer   string `json:"pointer"`
	Parameter string `json:"parameter"`
}

// Error implements the error interface
func (e *JSONAPIError) Error() string {
	msg := e.Title
	if e.Detail != "" {
		msg = fmt.Sprintf("%s: %s", e.Title, e.Detail)
	}
	if e.Source != nil && e.Source.Pointer != "" {
		msg = fmt.Sprintf("%s (%s)", msg, e.Source.Pointer)
	}
	return msg
}

// JSONAPIDocument represents a JSON:API document
type JSONAPIDocument struct {
	Data     json.RawMessage   `json:"data"`
	Included []json.RawMessage `json:"included"`
	Meta     json.RawMessage   `json:"meta"`
	Errors   []*JSONAPIError   `json:"errors"`
}

// ListingMeta represents pagination metadata
type ListingMeta struct {
	Pagination struct {
		CurrentPage int  `json:"current-page"`
		PageSize    int  `json:"page-size"`
		PrevPage    *int `json:"prev-page"`
		NextPage    *int `json:"next-page"`
		TotalPages  int  `json:"total-pages"`
		TotalCount  int  `json:"total-count"`
	} `json:"pagination"`
}

// ResourceIdentifier represents a JSON:API resource identifier
type ResourceIdentifier struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// ResourceLike is an interface for types that can be used as resource identifiers
// All schema types implement this interface
type ResourceLike interface {
	GetID() string
	GetResourceType() string
}

// Response wraps the standard http.Response with additional metadata
type Response struct {
	*http.Response
	Pagination *Pagination
}

// Pagination holds pagination metadata from JSON:API responses
type Pagination struct {
	CurrentPage int  `json:"current-page"`
	PageSize    int  `json:"page-size"`
	PrevPage    *int `json:"prev-page"`
	NextPage    *int `json:"next-page"`
	TotalPages  int  `json:"total-pages"`
	TotalCount  int  `json:"total-count"`
}
