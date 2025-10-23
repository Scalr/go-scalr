package generator

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
)

//go:embed templates/operations.tpl
var operationsTemplate string

// generateOperations generates resource client files
func (g *Generator) generateOperations(doc *openapi3.T, outputDir string) error {
	// Group operations by x-resource
	resourceOps := make(map[string][]Operation)
	// Collect operations without x-resource separately
	var standaloneOps []Operation

	for path, pathItem := range doc.Paths.Map() {
		for method, op := range pathItem.Operations() {
			if op == nil {
				continue
			}

			resource := getResourceName(op)
			if resource == "" {
				// No x-resource: add to standalone operations
				operation := g.parseOperation(path, method, op, doc, "")
				standaloneOps = append(standaloneOps, operation)
			} else {
				operation := g.parseOperation(path, method, op, doc, resource)
				resourceOps[resource] = append(resourceOps[resource], operation)
			}
		}
	}

	// Each resource operation resides in its own package
	tmpl, err := template.New("operations").Funcs(template.FuncMap{
		"trimPrefix": strings.TrimPrefix,
		"contains":   strings.Contains,
	}).Parse(operationsTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	for resource, ops := range resourceOps {
		resourceDir := filepath.Join(outputDir, strcase.ToSnake(resource))
		if err := os.MkdirAll(resourceDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", resourceDir, err)
		}

		fileName := strcase.ToSnake(resource) + ".gen.go"
		filePath := filepath.Join(resourceDir, fileName)

		// Sort operations by name
		sort.Slice(ops, func(i, j int) bool {
			return ops[i].Name < ops[j].Name
		})

		data := ResourceClientData{
			PackageName:    strcase.ToSnake(resource),
			ResourceName:   resource,
			ApiPackageName: g.pkgName,
			Operations:     ops,
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", resource, err)
		}

		if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", fileName, err)
		}
	}

	// Generate standalone operations (without x-resource)
	if len(standaloneOps) > 0 {
		standaloneDir := filepath.Join(outputDir, "misc")
		if err := os.MkdirAll(standaloneDir, 0755); err != nil {
			return fmt.Errorf("failed to create misc directory: %w", err)
		}

		filePath := filepath.Join(standaloneDir, "misc.gen.go")

		// Sort operations by name
		sort.Slice(standaloneOps, func(i, j int) bool {
			return standaloneOps[i].Name < standaloneOps[j].Name
		})

		data := ResourceClientData{
			PackageName:    "misc",
			ResourceName:   "Misc",
			ApiPackageName: g.pkgName,
			Operations:     standaloneOps,
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("failed to execute template for standalone operations: %w", err)
		}

		if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write misc.gen.go: %w", err)
		}
	}

	return nil
}

// ResourceClientData holds template data for a resource client
type ResourceClientData struct {
	PackageName    string
	ResourceName   string
	ApiPackageName string
	Operations     []Operation
}

// Operation represents an API operation
type Operation struct {
	Name                 string
	Method               string
	Path                 string
	Description          string
	PathParameters       []Parameter
	QueryParams          []QueryParam
	Returns              string // Return type
	RequestType          string // Request body type (schemas.WorkspaceRequest, schemas.TagRelationshipFieldsetsListingDocument, etc.)
	IsRelationshipOp     bool   // Is this a relationship operation (needs special handling)
	IsList               bool   // Is this a listing operation
	HasBody              bool   // Has request body
	ReturnsData          bool   // Returns data (vs void)
	ReturnsText          bool   // Returns plain text (not JSON)
	ReturnsRelationships bool   // Whether the return type has relationships field
	UsesPlainJSON        bool   // True if request body is plain JSON (not JSON:API)
}

// Parameter represents an operation path parameter
type Parameter struct {
	Name   string
	GoName string
	Type   string
}

// QueryParam represents a query parameter
type QueryParam struct {
	Name         string
	GoName       string
	Type         string
	Description  string
	IsFilter     bool
	IsSort       bool
	IsInclude    bool
	IsPagination bool
}

// parseOperation parses an OpenAPI operation
func (g *Generator) parseOperation(path, method string, op *openapi3.Operation, doc *openapi3.T, resourceName string) Operation {
	operation := Operation{
		Name:        strcase.ToCamel(op.OperationID),
		Method:      strings.ToUpper(method),
		Path:        path,
		Description: cleanDescription(op.Description),
	}

	// Parse path parameters
	for _, paramRef := range op.Parameters {
		if paramRef.Value == nil {
			continue
		}
		param := paramRef.Value

		switch param.In {
		case "path":
			paramName := sanitizeGoName(param.Name)
			operation.PathParameters = append(operation.PathParameters, Parameter{
				GoName: paramName,
				Name:   param.Name,
				Type:   "string",
			})
		case "query":
			operation.QueryParams = append(operation.QueryParams, g.parseQueryParam(param))
		}
	}

	// Check if has request body and extract request type
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		operation.HasBody = true
		reqType, isRelationship, usesPlainJSON := g.getRequestBodyType(op.RequestBody.Value, doc, resourceName, path, method)
		operation.RequestType = reqType
		operation.IsRelationshipOp = isRelationship
		operation.UsesPlainJSON = usesPlainJSON
	}

	// Determine return type from response
	if responses := op.Responses; responses != nil {
		resp := responses.Status(200)
		if resp == nil {
			resp = responses.Status(201)
		}
		if resp != nil && resp.Value != nil {
			operation.ReturnsData = true
			returnType, isText := g.getResponseType(resp.Value, doc, path, method)
			operation.Returns = returnType
			operation.ReturnsText = isText
			operation.IsList = strings.Contains(operation.Returns, "[]")
			operation.ReturnsRelationships = g.schemaHasRelationships(resp.Value, doc)
		} else if responses.Status(204) != nil {
			operation.ReturnsData = false
		}
	}

	return operation
}

// parseQueryParam parses a query parameter
func (g *Generator) parseQueryParam(param *openapi3.Parameter) QueryParam {
	qp := QueryParam{
		Name:        param.Name,
		GoName:      strcase.ToCamel(sanitizeGoName(param.Name)),
		Description: cleanDescription(param.Description),
	}

	// Detect special parameter types
	switch {
	case strings.HasPrefix(param.Name, "filter["):
		qp.IsFilter = true
		qp.Type = "string"
	case param.Name == "sort":
		qp.IsSort = true
		qp.Type = "[]string"
	case param.Name == "include":
		qp.IsInclude = true
		qp.Type = "[]string"
	case param.Name == "page[number]":
		qp.IsPagination = true
		qp.GoName = "PageNumber"
		qp.Type = "int"
	case param.Name == "page[size]":
		qp.IsPagination = true
		qp.GoName = "PageSize"
		qp.Type = "int"
	case strings.HasPrefix(param.Name, "page["):
		qp.IsPagination = true
		qp.Type = "int"
	default:
		// Infer type from schema
		if param.Schema != nil && param.Schema.Value != nil {
			qp.Type = g.schemaToGoType(param.Schema.Value)
		} else {
			qp.Type = "string"
		}
	}

	return qp
}

// getResponseType extracts the return type from a response
// Returns (type, isPlainText)
func (g *Generator) getResponseType(resp *openapi3.Response, doc *openapi3.T, path, method string) (string, bool) {
	// Check if this is a relationship endpoint (path contains /relationships/)
	isRelationshipEndpoint := strings.Contains(path, "/relationships/")

	// Check for JSON:API content
	if content := resp.Content.Get("application/vnd.api+json"); content != nil {
		if content.Schema != nil && content.Schema.Ref != "" {
			// Extract schema name from $ref
			parts := strings.Split(content.Schema.Ref, "/")
			documentSchemaName := parts[len(parts)-1]

			// Handle relationship listing endpoints (GET /resources/{id}/relationships/{name})
			if isRelationshipEndpoint && method == "GET" {
				// Extract the inner resource type from the relationship document
				itemType := g.extractRelationshipItemType(doc, documentSchemaName)
				if itemType != "" {
					// Return pointer slice for consistency with other list operations
					return "[]*schemas." + itemType, false
				}
				// Fallback: try schema name-based detection for edge cases
				if strings.Contains(documentSchemaName, "Relationship") && strings.HasSuffix(documentSchemaName, "ListingDocument") {
					schemaName := strings.TrimSuffix(documentSchemaName, "FieldsetsListingDocument")
					if schemaRef := doc.Components.Schemas[schemaName]; schemaRef != nil {
						return "[]*schemas." + schemaName, false
					}
				}
			}

			// Remove "Document" or "ListingDocument" suffix for regular resources
			schemaName := documentSchemaName
			isList := false
			if strings.HasSuffix(schemaName, "ListingDocument") {
				schemaName = strings.TrimSuffix(schemaName, "ListingDocument")
				isList = true
			} else if strings.HasSuffix(schemaName, "Document") {
				schemaName = strings.TrimSuffix(schemaName, "Document")
			}

			// Check if schema actually exists in the spec
			if _, exists := doc.Components.Schemas[schemaName]; !exists {
				// Schema doesn't exist, use interface{}
				if isList {
					return "[]interface{}", false
				}
				return "interface{}", false
			}

			if isList {
				return "[]*schemas." + schemaName, false
			}
			return "*schemas." + schemaName, false
		}
	}

	// Check for plain text / binary content (application/octet-stream, text/plain, etc.)
	// If no JSON:API content is found, assume it's plain text
	if len(resp.Content) > 0 {
		// Has content but not JSON:API - likely plain text or binary
		return "string", true
	}

	// No content schema defined at all in OpenAPI spec
	// This typically means it returns plain text or binary data (like logs, files, etc.)
	// Common for endpoints that stream logs or download files
	return "string", true
}

// getRequestBodyType extracts the request body type from a request body
// Returns (requestType, isRelationshipOp, usesPlainJSON)
func (g *Generator) getRequestBodyType(reqBody *openapi3.RequestBody, doc *openapi3.T, resourceName, path, method string) (string, bool, bool) {
	// Check if this is a relationship endpoint (path contains /relationships/)
	isRelationshipEndpoint := strings.Contains(path, "/relationships/")

	// Check for JSON:API content first
	content := reqBody.Content.Get("application/vnd.api+json")
	usesPlainJSON := false
	if content == nil {
		// Fall back to plain JSON
		content = reqBody.Content.Get("application/json")
		usesPlainJSON = true // This is not a JSON:API request
	}

	if content != nil && content.Schema != nil && content.Schema.Ref != "" {
		// Extract schema name from $ref
		parts := strings.Split(content.Schema.Ref, "/")
		schemaName := parts[len(parts)-1]

		// Check if it's a simple resource Document (e.g., WorkspaceDocument)
		// These should be unwrapped to *Request types
		if schemaName == resourceName+"Document" {
			return "*schemas." + resourceName + "Request", false, usesPlainJSON
		}

		// Handle relationship modification endpoints (POST/PATCH/DELETE /resources/{id}/relationships/{name})
		if isRelationshipEndpoint && (method == "POST" || method == "PATCH" || method == "DELETE") {
			// Extract the inner type from the Document schema
			itemType := g.extractRelationshipItemType(doc, schemaName)
			if itemType != "" {
				// Return the item type as the user-facing type
				return "[]schemas." + itemType, true, usesPlainJSON
			}
			// Fallback: try schema name-based detection for edge cases
			if strings.Contains(schemaName, "Relationship") || strings.Contains(schemaName, "Fieldsets") {
				// Fallback to original document type if we can't extract
				return "*schemas." + schemaName, false, usesPlainJSON
			}
		}

		// For other Document types, also unwrap *Document -> *Request
		if strings.HasSuffix(schemaName, "Document") {
			baseName := strings.TrimSuffix(schemaName, "Document")
			return "*schemas." + baseName + "Request", false, usesPlainJSON
		}

		// Otherwise keep the schema name as-is (for non-Document schemas)
		return "*schemas." + schemaName, false, usesPlainJSON
	}

	// No $ref found, fall back to default pattern: *schemas.{Resource}Request
	// This is the standard for JSON:API resource operations
	return "*schemas." + resourceName + "Request", false, usesPlainJSON
}

// extractRelationshipItemType extracts the inner type from a relationship document
// e.g., TagRelationshipFieldsetsListingDocument -> "Tag"
func (g *Generator) extractRelationshipItemType(doc *openapi3.T, documentSchemaName string) string {
	schemaRef := doc.Components.Schemas[documentSchemaName]
	if schemaRef == nil || schemaRef.Value == nil {
		return ""
	}

	// Get the data property
	dataProp := schemaRef.Value.Properties["data"]
	if dataProp == nil || dataProp.Value == nil {
		return ""
	}

	// Check if it's an array
	if !dataProp.Value.Type.Is("array") {
		return ""
	}

	// Get the items type
	if dataProp.Value.Items == nil {
		return ""
	}

	// Extract schema name from $ref
	if dataProp.Value.Items.Ref != "" {
		parts := strings.Split(dataProp.Value.Items.Ref, "/")
		relationshipSchemaName := parts[len(parts)-1]

		// Get the relationship schema to extract its JSON:API type
		relationshipSchema := doc.Components.Schemas[relationshipSchemaName]
		if relationshipSchema != nil && relationshipSchema.Value != nil {
			// Look for the 'type' field in the relationship schema
			if typeProp := relationshipSchema.Value.Properties["type"]; typeProp != nil && typeProp.Value != nil {
				// Extract the type enum value
				if len(typeProp.Value.Enum) > 0 {
					if typeVal, ok := typeProp.Value.Enum[0].(string); ok && typeVal != "" {
						// Look up the resource schema for this JSON:API type
						if resourceSchema, exists := g.typeToSchemaMap[typeVal]; exists {
							return resourceSchema
						}
					}
				}
			}
		}

		// Fallback: try to strip "Relationship" suffix
		if strings.HasSuffix(relationshipSchemaName, "Relationship") {
			baseName := strings.TrimSuffix(relationshipSchemaName, "Relationship")

			// Check if the base resource schema exists
			if _, exists := doc.Components.Schemas[baseName]; exists {
				return baseName
			}
		}

		// Last resort: return the relationship schema name as-is
		return relationshipSchemaName
	}

	return ""
}

// schemaHasRelationships checks if a response schema has relationships
func (g *Generator) schemaHasRelationships(resp *openapi3.Response, doc *openapi3.T) bool {
	// Check for JSON:API content
	if content := resp.Content.Get("application/vnd.api+json"); content != nil {
		if content.Schema != nil && content.Schema.Ref != "" {
			// Extract schema name from $ref
			parts := strings.Split(content.Schema.Ref, "/")
			schemaName := parts[len(parts)-1]

			// Remove "Document" or "ListingDocument" suffix
			if strings.HasSuffix(schemaName, "ListingDocument") {
				schemaName = strings.TrimSuffix(schemaName, "ListingDocument")
			} else if strings.HasSuffix(schemaName, "Document") {
				schemaName = strings.TrimSuffix(schemaName, "Document")
			}

			// Check if schema exists and has relationships
			if schemaRef, exists := doc.Components.Schemas[schemaName]; exists {
				if schemaRef.Value != nil {
					// Check if schema has a relationships property
					if schemaRef.Value.Properties != nil {
						if relProp := schemaRef.Value.Properties["relationships"]; relProp != nil && relProp.Value != nil {
							// Has relationships property - check if it's not empty
							return len(relProp.Value.Properties) > 0
						}
					}
				}
			}
		}
	}

	return false
}
