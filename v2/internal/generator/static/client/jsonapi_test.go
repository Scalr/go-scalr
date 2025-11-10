package client

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestJSONAPIError tests the JSONAPIError type
func TestJSONAPIError(t *testing.T) {
	tests := []struct {
		name        string
		err         *JSONAPIError
		expectedMsg string
	}{
		{
			name: "title only",
			err: &JSONAPIError{
				Title: "Bad Request",
			},
			expectedMsg: "Bad Request",
		},
		{
			name: "title and detail",
			err: &JSONAPIError{
				Title:  "Bad Request",
				Detail: "Invalid field value",
			},
			expectedMsg: "Bad Request: Invalid field value",
		},
		{
			name: "with source pointer",
			err: &JSONAPIError{
				Title:  "Validation Error",
				Detail: "Name is required",
				Source: &JSONAPIErrorSource{
					Pointer: "/data/attributes/name",
				},
			},
			expectedMsg: "Validation Error: Name is required (/data/attributes/name)",
		},
		{
			name: "with all fields",
			err: &JSONAPIError{
				Title:  "Conflict",
				Status: "409",
				Detail: "Resource already exists",
				Code:   "CONFLICT",
				Source: &JSONAPIErrorSource{
					Pointer: "/data/attributes/name",
				},
			},
			expectedMsg: "Conflict: Resource already exists (/data/attributes/name)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expectedMsg {
				t.Errorf("Error() = %q, want %q", got, tt.expectedMsg)
			}
		})
	}
}

// TestJSONAPIErrorSource tests the source field
func TestJSONAPIErrorSource(t *testing.T) {
	source := &JSONAPIErrorSource{
		Pointer:   "/data/attributes/name",
		Parameter: "name",
	}

	err := &JSONAPIError{
		Title:  "Validation Error",
		Source: source,
	}

	if err.Source.Pointer != "/data/attributes/name" {
		t.Errorf("Source.Pointer = %q, want %q", err.Source.Pointer, "/data/attributes/name")
	}

	if err.Source.Parameter != "name" {
		t.Errorf("Source.Parameter = %q, want %q", err.Source.Parameter, "name")
	}
}

// TestJSONAPIDocumentParsing tests parsing JSON:API documents
func TestJSONAPIDocumentParsing(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		check   func(t *testing.T, doc *JSONAPIDocument)
	}{
		{
			name: "success response with data",
			json: `{"data": {"type": "workspaces", "id": "ws-123"}}`,
			check: func(t *testing.T, doc *JSONAPIDocument) {
				if doc.Data == nil {
					t.Error("Data should not be nil")
				}
				if len(doc.Errors) > 0 {
					t.Error("Errors should be empty")
				}
			},
		},
		{
			name: "error response",
			json: `{"errors": [{"title": "Not Found", "status": "404"}]}`,
			check: func(t *testing.T, doc *JSONAPIDocument) {
				if len(doc.Errors) != 1 {
					t.Errorf("Expected 1 error, got %d", len(doc.Errors))
				}
				if doc.Errors[0].Title != "Not Found" {
					t.Errorf("Error title = %q, want %q", doc.Errors[0].Title, "Not Found")
				}
				if doc.Errors[0].Status != "404" {
					t.Errorf("Error status = %q, want %q", doc.Errors[0].Status, "404")
				}
			},
		},
		{
			name: "response with included",
			json: `{"data": {}, "included": [{"type": "users", "id": "user-123"}]}`,
			check: func(t *testing.T, doc *JSONAPIDocument) {
				if len(doc.Included) != 1 {
					t.Errorf("Expected 1 included resource, got %d", len(doc.Included))
				}
			},
		},
		{
			name: "response with meta",
			json: `{"data": {}, "meta": {"pagination": {"total-count": 10}}}`,
			check: func(t *testing.T, doc *JSONAPIDocument) {
				if doc.Meta == nil {
					t.Error("Meta should not be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc JSONAPIDocument
			err := json.Unmarshal([]byte(tt.json), &doc)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.check != nil {
				tt.check(t, &doc)
			}
		})
	}
}

// TestListingMetaParsing tests parsing pagination metadata
func TestListingMetaParsing(t *testing.T) {
	jsonData := `{
		"pagination": {
			"current-page": 2,
			"page-size": 20,
			"prev-page": 1,
			"next-page": 3,
			"total-pages": 5,
			"total-count": 100
		}
	}`

	var meta ListingMeta
	err := json.Unmarshal([]byte(jsonData), &meta)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if meta.Pagination.CurrentPage != 2 {
		t.Errorf("CurrentPage = %d, want 2", meta.Pagination.CurrentPage)
	}

	if meta.Pagination.PageSize != 20 {
		t.Errorf("PageSize = %d, want 20", meta.Pagination.PageSize)
	}

	if meta.Pagination.TotalPages != 5 {
		t.Errorf("TotalPages = %d, want 5", meta.Pagination.TotalPages)
	}

	if meta.Pagination.TotalCount != 100 {
		t.Errorf("TotalCount = %d, want 100", meta.Pagination.TotalCount)
	}

	if meta.Pagination.PrevPage == nil || *meta.Pagination.PrevPage != 1 {
		t.Error("PrevPage should be 1")
	}

	if meta.Pagination.NextPage == nil || *meta.Pagination.NextPage != 3 {
		t.Error("NextPage should be 3")
	}
}

// TestListingMetaLastPage tests pagination on the last page
func TestListingMetaLastPage(t *testing.T) {
	jsonData := `{
		"pagination": {
			"current-page": 5,
			"page-size": 20,
			"prev-page": 4,
			"next-page": null,
			"total-pages": 5,
			"total-count": 100
		}
	}`

	var meta ListingMeta
	err := json.Unmarshal([]byte(jsonData), &meta)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if meta.Pagination.NextPage != nil {
		t.Error("NextPage should be nil on last page")
	}

	if meta.Pagination.PrevPage == nil {
		t.Error("PrevPage should not be nil")
	}
}

// TestResourceIdentifier tests the ResourceIdentifier type
func TestResourceIdentifier(t *testing.T) {
	identifier := ResourceIdentifier{
		ID:   "ws-123",
		Type: "workspaces",
	}

	if identifier.ID != "ws-123" {
		t.Errorf("ID = %q, want %q", identifier.ID, "ws-123")
	}

	if identifier.Type != "workspaces" {
		t.Errorf("Type = %q, want %q", identifier.Type, "workspaces")
	}
}

// TestResourceIdentifierMarshaling tests JSON marshaling of resource identifiers
func TestResourceIdentifierMarshaling(t *testing.T) {
	identifier := ResourceIdentifier{
		ID:   "tag-456",
		Type: "tags",
	}

	data, err := json.Marshal(identifier)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	expected := `{"id":"tag-456","type":"tags"}`
	if string(data) != expected {
		t.Errorf("Marshal() = %q, want %q", string(data), expected)
	}

	// Test unmarshaling
	var decoded ResourceIdentifier
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.ID != identifier.ID {
		t.Errorf("Unmarshaled ID = %q, want %q", decoded.ID, identifier.ID)
	}

	if decoded.Type != identifier.Type {
		t.Errorf("Unmarshaled Type = %q, want %q", decoded.Type, identifier.Type)
	}
}

// TestResponseWrapper tests the Response wrapper
func TestResponseWrapper(t *testing.T) {
	// Create a mock HTTP response
	httpResp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
	}

	pagination := &Pagination{
		CurrentPage: 1,
		TotalPages:  5,
		PageSize:    20,
		TotalCount:  100,
	}

	resp := &Response{
		Response:   httpResp,
		Pagination: pagination,
	}

	// Test that Response wraps http.Response
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}

	if resp.Status != "200 OK" {
		t.Errorf("Status = %q, want %q", resp.Status, "200 OK")
	}

	// Test pagination data
	if resp.Pagination.CurrentPage != 1 {
		t.Errorf("Pagination.CurrentPage = %d, want 1", resp.Pagination.CurrentPage)
	}

	if resp.Pagination.TotalPages != 5 {
		t.Errorf("Pagination.TotalPages = %d, want 5", resp.Pagination.TotalPages)
	}
}

// TestPaginationMarshaling tests JSON marshaling of Pagination
func TestPaginationMarshaling(t *testing.T) {
	nextPage := 2
	pagination := &Pagination{
		CurrentPage: 1,
		PageSize:    20,
		NextPage:    &nextPage,
		TotalPages:  5,
		TotalCount:  100,
	}

	data, err := json.Marshal(pagination)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal to verify structure
	var decoded Pagination
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.CurrentPage != pagination.CurrentPage {
		t.Errorf("CurrentPage = %d, want %d", decoded.CurrentPage, pagination.CurrentPage)
	}

	if decoded.NextPage == nil || *decoded.NextPage != *pagination.NextPage {
		t.Error("NextPage mismatch")
	}
}

// Mock type implementing ResourceLike for testing
type MockResource struct {
	ID   string
	Type string
}

func (m MockResource) GetID() string {
	return m.ID
}

func (m MockResource) GetResourceType() string {
	return m.Type
}

// TestResourceLikeInterface tests that ResourceLike interface works
func TestResourceLikeInterface(t *testing.T) {
	resource := MockResource{
		ID:   "mock-123",
		Type: "mock-resources",
	}

	// Test that MockResource implements ResourceLike
	var _ ResourceLike = resource

	if resource.GetID() != "mock-123" {
		t.Errorf("GetID() = %q, want %q", resource.GetID(), "mock-123")
	}

	if resource.GetResourceType() != "mock-resources" {
		t.Errorf("GetResourceType() = %q, want %q", resource.GetResourceType(), "mock-resources")
	}
}
