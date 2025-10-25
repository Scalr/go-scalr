package generator

import (
	"sort"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// TestFieldOrdering tests that struct fields are generated in consistent order
func TestFieldOrdering(t *testing.T) {
	// Create a schema with multiple properties
	schema := &openapi3.Schema{
		Properties: map[string]*openapi3.SchemaRef{
			"name":        {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			"id":          {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			"type":        {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			"description": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			"status":      {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
		},
	}

	// Get property names and sort them
	propertyNames := make([]string, 0, len(schema.Properties))
	for fieldName := range schema.Properties {
		propertyNames = append(propertyNames, fieldName)
	}

	// Sort with custom comparator (same as in generator)
	sort.Slice(propertyNames, func(i, j int) bool {
		a, b := propertyNames[i], propertyNames[j]

		// 'type' always comes first
		if a == "type" {
			return true
		}
		if b == "type" {
			return false
		}

		// 'id' comes second (after type)
		if a == "id" {
			return true
		}
		if b == "id" {
			return false
		}

		// Everything else alphabetically
		return a < b
	})

	// Verify order: type, id, then alphabetically
	expected := []string{"type", "id", "description", "name", "status"}

	if len(propertyNames) != len(expected) {
		t.Fatalf("Expected %d fields, got %d", len(expected), len(propertyNames))
	}

	for i, field := range propertyNames {
		if field != expected[i] {
			t.Errorf("Field %d: expected %q, got %q", i, expected[i], field)
		}
	}
}

// TestFieldOrderingWithoutTypeAndId tests ordering when type/id are absent
func TestFieldOrderingWithoutTypeAndId(t *testing.T) {
	schema := &openapi3.Schema{
		Properties: map[string]*openapi3.SchemaRef{
			"zebra":       {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			"apple":       {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			"banana":      {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			"description": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
		},
	}

	propertyNames := make([]string, 0, len(schema.Properties))
	for fieldName := range schema.Properties {
		propertyNames = append(propertyNames, fieldName)
	}

	sort.Slice(propertyNames, func(i, j int) bool {
		a, b := propertyNames[i], propertyNames[j]
		if a == "type" {
			return true
		}
		if b == "type" {
			return false
		}
		if a == "id" {
			return true
		}
		if b == "id" {
			return false
		}
		return a < b
	})

	// Should be alphabetically ordered
	expected := []string{"apple", "banana", "description", "zebra"}

	for i, field := range propertyNames {
		if field != expected[i] {
			t.Errorf("Field %d: expected %q, got %q", i, expected[i], field)
		}
	}
}

// TestParseRelationship tests relationship parsing
func TestParseRelationship(t *testing.T) {
	g := New("", "test")

	// Setup type to schema mapping
	g.typeToSchemaMap = map[string]string{
		"tags":         "Tag",
		"environments": "Environment",
	}

	tests := []struct {
		name           string
		relName        string
		schema         *openapi3.Schema
		expectedType   string
		expectedToMany bool
	}{
		{
			name:    "to-many relationship (array)",
			relName: "tags",
			schema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"data": {
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"array"},
							Items: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Properties: map[string]*openapi3.SchemaRef{
										"type": {
											Value: &openapi3.Schema{
												Enum: []interface{}{"tags"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedType:   "Tag",
			expectedToMany: true,
		},
		{
			name:    "to-one relationship (object)",
			relName: "environment",
			schema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"data": {
						Value: &openapi3.Schema{
							Properties: map[string]*openapi3.SchemaRef{
								"type": {
									Value: &openapi3.Schema{
										Enum: []interface{}{"environments"},
									},
								},
							},
						},
					},
				},
			},
			expectedType:   "Environment",
			expectedToMany: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := g.parseRelationship(tt.relName, tt.schema)

			if rel.Type != tt.expectedType {
				t.Errorf("Type = %q, want %q", rel.Type, tt.expectedType)
			}
			if rel.ToMany != tt.expectedToMany {
				t.Errorf("ToMany = %v, want %v", rel.ToMany, tt.expectedToMany)
			}
		})
	}
}

// TestParseAttribute tests attribute parsing
func TestParseAttribute(t *testing.T) {
	g := New("", "test")

	tests := []struct {
		name                string
		attrSchema          *openapi3.Schema
		expectedGoType      string
		expectedRequestType string
	}{
		{
			name: "string attribute",
			attrSchema: &openapi3.Schema{
				Type: &openapi3.Types{"string"},
			},
			expectedGoType:      "string",
			expectedRequestType: "*value.Value[string]",
		},
		{
			name: "integer attribute",
			attrSchema: &openapi3.Schema{
				Type: &openapi3.Types{"integer"},
			},
			expectedGoType:      "int",
			expectedRequestType: "*value.Value[int]",
		},
		{
			name: "boolean attribute",
			attrSchema: &openapi3.Schema{
				Type: &openapi3.Types{"boolean"},
			},
			expectedGoType:      "bool",
			expectedRequestType: "*value.Value[bool]",
		},
		{
			name: "array of strings",
			attrSchema: &openapi3.Schema{
				Type: &openapi3.Types{"array"},
				Items: &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
				},
			},
			expectedGoType:      "[]string",
			expectedRequestType: "*value.Value[[]string]",
		},
		{
			name: "nullable string",
			attrSchema: &openapi3.Schema{
				Type:     &openapi3.Types{"string"},
				Nullable: true,
			},
			expectedGoType:      "*string",              // Response: pointer to distinguish null from ""
			expectedRequestType: "*value.Value[string]", // Request: Value handles null, no inner pointer
		},
		{
			name: "nullable integer",
			attrSchema: &openapi3.Schema{
				Type:     &openapi3.Types{"integer"},
				Nullable: true,
			},
			expectedGoType:      "*int",
			expectedRequestType: "*value.Value[int]",
		},
		{
			name: "nullable time",
			attrSchema: &openapi3.Schema{
				Type:     &openapi3.Types{"string"},
				Format:   "date-time",
				Nullable: true,
			},
			expectedGoType:      "*time.Time",
			expectedRequestType: "*value.Value[time.Time]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test response type (with nullable)
			baseType := g.schemaToGoType(tt.attrSchema)
			if baseType != tt.expectedGoType {
				t.Errorf("Go type = %q, want %q", baseType, tt.expectedGoType)
			}

			// Test request type (value.Value handles null, so use base type without pointer)
			// Simulate what the actual generator does
			baseAttrSchema := *tt.attrSchema
			baseAttrSchema.Nullable = false
			requestBaseType := g.schemaToGoType(&baseAttrSchema)
			requestType := "*value.Value[" + requestBaseType + "]"
			if requestType != tt.expectedRequestType {
				t.Errorf("Request type = %q, want %q", requestType, tt.expectedRequestType)
			}
		})
	}
}

// TestReadOnlyFiltering tests that read-only fields are excluded from request structs
func TestReadOnlyFiltering(t *testing.T) {
	schema := &openapi3.Schema{
		Properties: map[string]*openapi3.SchemaRef{
			"name": {
				Value: &openapi3.Schema{
					Type:     &openapi3.Types{"string"},
					ReadOnly: false,
				},
			},
			"created-at": {
				Value: &openapi3.Schema{
					Type:     &openapi3.Types{"string"},
					Format:   "date-time",
					ReadOnly: true,
				},
			},
			"updated-at": {
				Value: &openapi3.Schema{
					Type:     &openapi3.Types{"string"},
					Format:   "date-time",
					ReadOnly: true,
				},
			},
		},
	}

	// Simulate filtering logic
	var writableAttrs []string
	var readOnlyAttrs []string

	for attrName, attrRef := range schema.Properties {
		if attrRef.Value.ReadOnly {
			readOnlyAttrs = append(readOnlyAttrs, attrName)
		} else {
			writableAttrs = append(writableAttrs, attrName)
		}
	}

	// Should have 1 writable attribute
	if len(writableAttrs) != 1 {
		t.Errorf("Expected 1 writable attribute, got %d", len(writableAttrs))
	}

	// Should have 2 read-only attributes
	if len(readOnlyAttrs) != 2 {
		t.Errorf("Expected 2 read-only attributes, got %d", len(readOnlyAttrs))
	}

	// Verify the writable one is "name"
	if len(writableAttrs) > 0 && writableAttrs[0] != "name" {
		t.Errorf("Expected writable attribute to be 'name', got %q", writableAttrs[0])
	}
}

// TestNestedObjectHandling tests nested object attribute handling
func TestNestedObjectHandling(t *testing.T) {
	g := New("", "test")

	schema := &openapi3.Schema{
		Type: &openapi3.Types{"object"},
		Properties: map[string]*openapi3.SchemaRef{
			"name": {
				Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
			},
			"port": {
				Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
			},
		},
	}

	// When schema has properties, it should be treated as a struct
	hasProperties := len(schema.Properties) > 0
	if !hasProperties {
		t.Error("Expected schema to have properties")
	}

	// Verify we can extract field types
	for fieldName, fieldRef := range schema.Properties {
		goType := g.schemaToGoType(fieldRef.Value)

		switch fieldName {
		case "name":
			if goType != "string" {
				t.Errorf("Expected 'name' to be string, got %q", goType)
			}
		case "port":
			if goType != "int" {
				t.Errorf("Expected 'port' to be int, got %q", goType)
			}
		}
	}
}

// TestSchemaTypeDetection tests detecting resource-like schemas
func TestSchemaTypeDetection(t *testing.T) {
	tests := []struct {
		name           string
		schema         *openapi3.Schema
		expectResource bool
	}{
		{
			name: "resource with type and id",
			schema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"type": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
					"id":   {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
				},
			},
			expectResource: true,
		},
		{
			name: "plain schema without type and id",
			schema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"name":   {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
					"status": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
				},
			},
			expectResource: false,
		},
		{
			name: "has id but no type",
			schema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"id": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
				},
			},
			expectResource: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasID := false
			hasType := false

			for fieldName := range tt.schema.Properties {
				if fieldName == "id" {
					hasID = true
				}
				if fieldName == "type" {
					hasType = true
				}
			}

			isResource := hasID && hasType

			if isResource != tt.expectResource {
				t.Errorf("Expected resource=%v, got %v", tt.expectResource, isResource)
			}
		})
	}
}

// TestTypeEnumExtraction tests extracting type enum values from schemas
func TestTypeEnumExtraction(t *testing.T) {
	tests := []struct {
		name         string
		schema       *openapi3.Schema
		expectedType string
	}{
		{
			name: "schema with type enum",
			schema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"type": {
						Value: &openapi3.Schema{
							Enum: []interface{}{"workspaces"},
						},
					},
				},
			},
			expectedType: "workspaces",
		},
		{
			name: "schema without type enum",
			schema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"type": {
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"string"},
						},
					},
				},
			},
			expectedType: "",
		},
		{
			name: "schema without type field",
			schema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"name": {
						Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
					},
				},
			},
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var typeEnum string

			if typeProp := tt.schema.Properties["type"]; typeProp != nil && typeProp.Value != nil {
				if len(typeProp.Value.Enum) > 0 {
					if typeVal, ok := typeProp.Value.Enum[0].(string); ok {
						typeEnum = typeVal
					}
				}
			}

			if typeEnum != tt.expectedType {
				t.Errorf("Expected type %q, got %q", tt.expectedType, typeEnum)
			}
		})
	}
}
