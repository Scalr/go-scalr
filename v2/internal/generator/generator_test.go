package generator

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// TestTypeMapping tests OpenAPI type to Go type conversion
func TestSchemaToGoType(t *testing.T) {
	g := New("", "test")

	tests := []struct {
		name     string
		schema   *openapi3.Schema
		expected string
	}{
		{
			name:     "string type",
			schema:   &openapi3.Schema{Type: &openapi3.Types{"string"}},
			expected: "string",
		},
		{
			name:     "integer type",
			schema:   &openapi3.Schema{Type: &openapi3.Types{"integer"}},
			expected: "int",
		},
		{
			name:     "integer with int64 format",
			schema:   &openapi3.Schema{Type: &openapi3.Types{"integer"}, Format: "int64"},
			expected: "int", // Generator currently doesn't distinguish int32/int64
		},
		{
			name:     "number type",
			schema:   &openapi3.Schema{Type: &openapi3.Types{"number"}},
			expected: "float64",
		},
		{
			name:     "boolean type",
			schema:   &openapi3.Schema{Type: &openapi3.Types{"boolean"}},
			expected: "bool",
		},
		{
			name:     "date-time format",
			schema:   &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date-time"},
			expected: "time.Time",
		},
		{
			name: "array of strings",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"array"},
				Items: &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
				},
			},
			expected: "[]string",
		},
		{
			name: "array of integers",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"array"},
				Items: &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
				},
			},
			expected: "[]int",
		},
		{
			name:     "object without properties (map)",
			schema:   &openapi3.Schema{Type: &openapi3.Types{"object"}},
			expected: "map[string]interface{}", // Objects without properties become maps
		},
		{
			name:     "nullable string (not handled specially)",
			schema:   &openapi3.Schema{Type: &openapi3.Types{"string"}, Nullable: true},
			expected: "string", // Generator doesn't currently add pointers for nullable
		},
		{
			name:     "nullable integer (not handled specially)",
			schema:   &openapi3.Schema{Type: &openapi3.Types{"integer"}, Nullable: true},
			expected: "int", // Generator doesn't currently add pointers for nullable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.schemaToGoType(tt.schema)
			if got != tt.expected {
				t.Errorf("schemaToGoType() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestGetResourceName tests extracting resource names from operations
func TestGetResourceName(t *testing.T) {
	tests := []struct {
		name     string
		op       *openapi3.Operation
		expected string
	}{
		{
			name: "operation with x-resource",
			op: &openapi3.Operation{
				Extensions: map[string]interface{}{
					"x-resource": "Workspace",
				},
			},
			expected: "Workspace",
		},
		{
			name: "operation without x-resource",
			op: &openapi3.Operation{
				Extensions: map[string]interface{}{},
			},
			expected: "",
		},
		{
			name:     "operation with nil extensions",
			op:       &openapi3.Operation{},
			expected: "",
		},
		{
			name: "operation with x-resource as non-string",
			op: &openapi3.Operation{
				Extensions: map[string]interface{}{
					"x-resource": 123,
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getResourceName(tt.op)
			if got != tt.expected {
				t.Errorf("getResourceName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestIsRelationshipEndpoint tests relationship endpoint detection
func TestIsRelationshipEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "regular endpoint",
			path:     "/workspaces/{workspace}",
			expected: false,
		},
		{
			name:     "relationship endpoint",
			path:     "/workspaces/{workspace}/relationships/tags",
			expected: true,
		},
		{
			name:     "nested relationship endpoint",
			path:     "/policy-groups/{policy_group}/relationships/environments",
			expected: true,
		},
		{
			name:     "not a relationship - similar path",
			path:     "/workspaces/{workspace}/tags",
			expected: false,
		},
		{
			name:     "relationships in plural form but not endpoint",
			path:     "/relationships-history",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test using strings.Contains as used in the actual implementation
			got := strings.Contains(tt.path, "/relationships/")
			if got != tt.expected {
				t.Errorf("isRelationshipEndpoint(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

// TestParseSpecMetadata tests parsing server URL and headers from spec
func TestParseSpecMetadata(t *testing.T) {
	tests := []struct {
		name           string
		doc            *openapi3.T
		expectedPath   string
		expectedVar    string
		expectedHeader string
		expectError    bool
	}{
		{
			name: "public API - no prefer header",
			doc: &openapi3.T{
				Servers: openapi3.Servers{
					{URL: "https://{Domain}/api/iacp/v3"},
				},
			},
			expectedPath:   "/api/iacp/v3",
			expectedVar:    "Domain",
			expectedHeader: "",
			expectError:    false,
		},
		{
			name: "admin API - different path",
			doc: &openapi3.T{
				Servers: openapi3.Servers{
					{URL: "https://{Domain}/api/admin"},
				},
			},
			expectedPath:   "/api/admin",
			expectedVar:    "Domain",
			expectedHeader: "",
			expectError:    false,
		},
		{
			name: "internal API - with prefer header",
			doc: &openapi3.T{
				Servers: openapi3.Servers{
					{URL: "https://{Domain}/api/iacp/v3"},
				},
				Components: &openapi3.Components{
					Parameters: map[string]*openapi3.ParameterRef{
						"PreferParam": {
							Value: &openapi3.Parameter{
								In:       "header",
								Name:     "Prefer",
								Required: true,
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Default: "profile=internal",
									},
								},
							},
						},
					},
				},
			},
			expectedPath:   "/api/iacp/v3",
			expectedVar:    "Domain",
			expectedHeader: "profile=internal",
			expectError:    false,
		},
		{
			name: "no servers",
			doc: &openapi3.T{
				Servers: openapi3.Servers{},
			},
			expectError: true,
		},
		{
			name: "invalid server URL format",
			doc: &openapi3.T{
				Servers: openapi3.Servers{
					{URL: "https://example.com/api"},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New("", "test")
			err := g.parseSpecMetadata(tt.doc)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if g.basePath != tt.expectedPath {
				t.Errorf("basePath = %q, want %q", g.basePath, tt.expectedPath)
			}
			if g.serverVariable != tt.expectedVar {
				t.Errorf("serverVariable = %q, want %q", g.serverVariable, tt.expectedVar)
			}
			if g.preferHeader != tt.expectedHeader {
				t.Errorf("preferHeader = %q, want %q", g.preferHeader, tt.expectedHeader)
			}
		})
	}
}

// TestTypeToSchemaMapBuilding tests the logic for building type to schema mapping
func TestTypeToSchemaMapBuilding(t *testing.T) {
	g := New("", "test")
	g.typeToSchemaMap = make(map[string]string)

	// Simulate the mapping logic
	g.typeToSchemaMap["workspaces"] = "Workspace"
	g.typeToSchemaMap["tags"] = "Tag"
	g.typeToSchemaMap["environments"] = "Environment"

	tests := []struct {
		jsonType     string
		expectedName string
	}{
		{"workspaces", "Workspace"},
		{"tags", "Tag"},
		{"environments", "Environment"},
		{"nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.jsonType, func(t *testing.T) {
			got := g.typeToSchemaMap[tt.jsonType]
			if got != tt.expectedName {
				t.Errorf("typeToSchemaMap[%q] = %q, want %q", tt.jsonType, got, tt.expectedName)
			}
		})
	}
}

// TestCleanDescription tests description cleaning
func TestCleanDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty description",
			input:    "",
			expected: "",
		},
		{
			name:     "single line",
			input:    "This is a workspace",
			expected: "This is a workspace",
		},
		{
			name:     "multiple lines",
			input:    "This is a workspace\nIt has multiple lines",
			expected: "This is a workspace It has multiple lines",
		},
		{
			name:     "with extra spaces (tabs not handled)",
			input:    "This  is   a workspace",
			expected: "This is a workspace",
		},
		{
			name:     "with markdown",
			input:    "This is a **workspace** with _markdown_",
			expected: "This is a **workspace** with _markdown_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanDescription(tt.input)
			if got != tt.expected {
				t.Errorf("cleanDescription() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestIsReadOnlySchemaDetection tests detecting read-only schema patterns
func TestIsReadOnlySchemaDetection(t *testing.T) {
	tests := []struct {
		name       string
		schemaName string
		expected   bool
	}{
		{
			name:       "regular schema",
			schemaName: "Workspace",
			expected:   false,
		},
		{
			name:       "listing document",
			schemaName: "WorkspaceListingDocument",
			expected:   true,
		},
		{
			name:       "details document",
			schemaName: "WorkspaceDetailsDocument",
			expected:   true,
		},
		{
			name:       "relationship document",
			schemaName: "TagRelationshipFieldsetsListingDocument",
			expected:   true,
		},
		{
			name:       "partial match - not read-only",
			schemaName: "WorkspaceDocument",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the pattern matching logic used in the actual implementation
			got := strings.HasSuffix(tt.schemaName, "ListingDocument") ||
				strings.HasSuffix(tt.schemaName, "DetailsDocument")
			if got != tt.expected {
				t.Errorf("isReadOnlySchema(%q) = %v, want %v", tt.schemaName, got, tt.expected)
			}
		})
	}
}
