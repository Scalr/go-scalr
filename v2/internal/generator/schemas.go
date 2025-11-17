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

//go:embed templates/schema.tpl
var schemaTemplate string

//go:embed templates/simple_schema.tpl
var simpleSchemaTemplate string

//go:embed templates/document_schema.tpl
var documentSchemaTemplate string

// generateSchemas generates all schema types
func (g *Generator) generateSchemas(doc *openapi3.T, outputDir string) error {
	tmpl, err := template.New("schema").Parse(schemaTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Collect all top-level schema names to avoid collisions with nested structs
	topLevelSchemas := make(map[string]bool)
	for name := range doc.Components.Schemas {
		topLevelSchemas[name] = true
	}

	// Collect canonical resource names from x-resource in operations
	// These are the "main" schemas for each resource
	canonicalResources := make(map[string]bool)
	for _, pathItem := range doc.Paths.Map() {
		for _, op := range pathItem.Operations() {
			if op == nil {
				continue
			}
			if resource := getResourceName(op); resource != "" {
				canonicalResources[resource] = true
			}
		}
	}

	// Build mapping of JSON:API type -> Schema name
	// This is used for relationships to find the correct Go type
	// When multiple schemas share the same JSON:API type, prefer the one referenced in x-resource
	g.typeToSchemaMap = make(map[string]string)
	for name, schemaRef := range doc.Components.Schemas {
		if schemaRef.Value == nil || !isResourceSchema(schemaRef.Value) {
			continue
		}

		// Extract the JSON:API type from the schema's type field
		typeName := extractTypeName(schemaRef.Value)
		if typeName != "" {
			// If type is already mapped, prefer the canonical resource (from x-resource)
			if existing, ok := g.typeToSchemaMap[typeName]; ok {
				// Replace if current schema is canonical and existing is not
				if canonicalResources[name] && !canonicalResources[existing] {
					g.typeToSchemaMap[typeName] = name
				}
				// Keep existing if it's canonical or both/neither are canonical
			} else {
				// First time seeing this type
				g.typeToSchemaMap[typeName] = name
			}
		}
	}

	// Collect schema names used in request bodies
	requestBodySchemas := make(map[string]bool)
	for _, pathItem := range doc.Paths.Map() {
		for _, op := range pathItem.Operations() {
			if op == nil || op.RequestBody == nil || op.RequestBody.Value == nil {
				continue
			}

			// Check JSON:API content
			content := op.RequestBody.Value.Content.Get("application/vnd.api+json")
			if content == nil {
				content = op.RequestBody.Value.Content.Get("application/json")
			}

			if content != nil && content.Schema != nil && content.Schema.Ref != "" {
				parts := strings.Split(content.Schema.Ref, "/")
				schemaName := parts[len(parts)-1]
				requestBodySchemas[schemaName] = true
			}
		}
	}

	// Generate each resource schema into the schemas directory
	for name, schemaRef := range doc.Components.Schemas {
		if schemaRef.Value == nil {
			continue
		}

		// Skip non-JSON:API resource schemas (they don't have 'type' field)
		if !isResourceSchema(schemaRef.Value) {
			continue
		}

		fileName := strcase.ToSnake(name) + ".gen.go"
		filePath := filepath.Join(outputDir, fileName)

		data := g.buildSchemaData(name, schemaRef.Value, topLevelSchemas)

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", name, err)
		}

		if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", fileName, err)
		}
	}

	// Collect all schemas referenced by Document schemas
	referencedSchemas := make(map[string]bool)
	for name := range requestBodySchemas {
		schemaRef := doc.Components.Schemas[name]
		if schemaRef == nil || schemaRef.Value == nil {
			continue
		}

		// Collect schemas referenced in the data field
		if dataProp := schemaRef.Value.Properties["data"]; dataProp != nil {
			if dataProp.Ref != "" {
				parts := strings.Split(dataProp.Ref, "/")
				referencedSchemas[parts[len(parts)-1]] = true
			} else if dataProp.Value != nil && dataProp.Value.Items != nil && dataProp.Value.Items.Ref != "" {
				parts := strings.Split(dataProp.Value.Items.Ref, "/")
				referencedSchemas[parts[len(parts)-1]] = true
			}
		}
	}

	// Generate Document schemas used in request bodies
	for name := range requestBodySchemas {
		schemaRef := doc.Components.Schemas[name]
		if schemaRef == nil || schemaRef.Value == nil {
			continue
		}

		// Skip if already generated (it's a resource schema)
		if isResourceSchema(schemaRef.Value) {
			continue
		}

		// Generate Document schema
		if err := g.generateDocumentSchema(name, schemaRef.Value, outputDir); err != nil {
			return fmt.Errorf("failed to generate document schema %s: %w", name, err)
		}
	}

	// Generate referenced schemas (like TagRelationship)
	for name := range referencedSchemas {
		schemaRef := doc.Components.Schemas[name]
		if schemaRef == nil || schemaRef.Value == nil {
			continue
		}

		// Skip if already generated
		if isResourceSchema(schemaRef.Value) {
			continue
		}

		// Generate simple schema
		if err := g.generateSimpleSchema(name, schemaRef.Value, outputDir); err != nil {
			return fmt.Errorf("failed to generate simple schema %s: %w", name, err)
		}
	}

	return nil
}

// isResourceSchema checks if a schema is a JSON:API resource
func isResourceSchema(schema *openapi3.Schema) bool {
	if schema.Properties == nil {
		return false
	}
	// JSON:API resources have 'type', 'id', 'attributes', 'relationships'
	_, hasType := schema.Properties["type"]
	_, hasAttrs := schema.Properties["attributes"]
	return hasType && hasAttrs
}

// extractTypeName extracts the JSON:API type name from the schema's type field enum
func extractTypeName(schema *openapi3.Schema) string {
	if schema.Properties == nil {
		return ""
	}

	typeField := schema.Properties["type"]
	if typeField == nil || typeField.Value == nil {
		return ""
	}

	// The type field should have an enum with a single value
	if len(typeField.Value.Enum) > 0 {
		if typeName, ok := typeField.Value.Enum[0].(string); ok {
			return typeName
		}
	}

	return ""
}

// SchemaData holds template data for a schema
type SchemaData struct {
	ApiPackageName       string
	Name                 string
	TypeName             string // JSON:API type name (e.g., "workspaces")
	Description          string
	Attributes           []Attribute
	Relationships        []Relationship
	NestedStructs        []NestedStruct
	RequestNestedStructs []NestedStruct
	EnumTypes            []EnumType
}

// EnumType represents an enum type definition
type EnumType struct {
	Name        string
	Description string
	BaseType    string // "string" or "int"
	Values      []EnumValue
}

// EnumValue represents a single enum constant
type EnumValue struct {
	Name        string
	Value       string
	Description string
}

// Attribute represents a schema attribute
type Attribute struct {
	Name         string
	JSONName     string
	ResponseType string // Type for response structs (plain types)
	RequestType  string // Type for request structs (Value wrapped)
	Description  string
	ReadOnly     bool
}

// NestedStruct represents a nested object structure within attributes
type NestedStruct struct {
	Name        string
	Description string
	Fields      []NestedField
}

// NestedField represents a field in a nested struct
type NestedField struct {
	Name        string
	JSONName    string
	Type        string
	Description string
	ReadOnly    bool
}

// Relationship represents a schema relationship
type Relationship struct {
	Name        string
	JSONName    string
	Type        string // The target schema type
	Description string
	ToMany      bool
	ReadOnly    bool // If true, only include in response, not in request
}

// buildSchemaData builds template data from OpenAPI schema
func (g *Generator) buildSchemaData(name string, schema *openapi3.Schema, topLevelSchemas map[string]bool) SchemaData {
	data := SchemaData{
		ApiPackageName: g.pkgName,
		Name:           name,
		TypeName:       extractTypeName(schema),
		Description:    cleanDescription(schema.Description),
	}

	// Process attributes
	if attrSchema := schema.Properties["attributes"]; attrSchema != nil && attrSchema.Value != nil {
		for attrName, attrRef := range attrSchema.Value.Properties {
			if attrRef.Value == nil {
				continue
			}

			// Check if this is a nested object that should be a struct
			var responseType, requestType string
			if attrRef.Value.Type.Is("object") && attrRef.Value.Properties != nil && len(attrRef.Value.Properties) > 0 {
				// Create nested structs for this object (both response and request versions)
				baseStructName := name + strcase.ToCamel(attrName)

				// Avoid collision with top-level schemas
				if topLevelSchemas[baseStructName] {
					baseStructName = name + strcase.ToCamel(attrName) + "Nested"
				}

				// Response version (plain types)
				responseNested := g.buildNestedStruct(baseStructName, attrRef.Value, false)
				data.NestedStructs = append(data.NestedStructs, responseNested)

				// Add pointer if nullable for response type
				if attrRef.Value.Nullable {
					responseType = "*" + baseStructName
				} else {
					responseType = baseStructName
				}

				// Request version (Value types)
				requestStructName := baseStructName + "Request"
				requestNested := g.buildNestedStruct(requestStructName, attrRef.Value, true)
				data.RequestNestedStructs = append(data.RequestNestedStructs, requestNested)

				// Request type: value.Value handles null, so inner type doesn't need pointer
				// value.Value already provides tri-state: unset/null/set
				requestType = "*value.Value[" + requestStructName + "]"
			} else {
				// Check if this is an enum field
				var enumTypeName string
				if len(attrRef.Value.Enum) > 0 {
					enumTypeName = name + strcase.ToCamel(attrName)
					enumType := g.buildEnumType(enumTypeName, attrRef.Value)
					data.EnumTypes = append(data.EnumTypes, enumType)
				}

				// For non-nested types, use schemaToGoType which handles nullable
				if enumTypeName != "" {
					// Use enum type constant as field type
					if attrRef.Value.Nullable {
						responseType = "*" + enumTypeName
					} else {
						responseType = enumTypeName
					}
					requestType = "*value.Value[" + enumTypeName + "]"
				} else {
					// Use standard type
					responseType = g.schemaToGoType(attrRef.Value)

					// For request, get base type without nullable pointer
					// We need to call schemaToGoType with nullable=false to get base type
					baseAttrSchema := *attrRef.Value
					baseAttrSchema.Nullable = false
					baseType := g.schemaToGoType(&baseAttrSchema)

					requestType = "*value.Value[" + baseType + "]"
				}
			}

			attr := Attribute{
				Name:         strcase.ToCamel(attrName),
				JSONName:     attrName,
				ResponseType: responseType,
				RequestType:  requestType,
				Description:  cleanDescription(attrRef.Value.Description),
				ReadOnly:     attrRef.Value.ReadOnly,
			}

			data.Attributes = append(data.Attributes, attr)
		}

		// Sort attributes by name for consistent output
		sort.Slice(data.Attributes, func(i, j int) bool {
			return data.Attributes[i].Name < data.Attributes[j].Name
		})

		// Sort nested structs by name for consistent output
		sort.Slice(data.NestedStructs, func(i, j int) bool {
			return data.NestedStructs[i].Name < data.NestedStructs[j].Name
		})
		sort.Slice(data.RequestNestedStructs, func(i, j int) bool {
			return data.RequestNestedStructs[i].Name < data.RequestNestedStructs[j].Name
		})
	}

	// Process relationships
	if relSchema := schema.Properties["relationships"]; relSchema != nil && relSchema.Value != nil {
		for relName, relRef := range relSchema.Value.Properties {
			if relRef.Value == nil {
				continue
			}

			rel := g.parseRelationship(relName, relRef.Value)

			// Skip relationships with no type (schema not found)
			if rel.Type == "" {
				continue
			}

			// Mark read-only relationships (they'll be included in responses but not requests)
			if relRef.Value.ReadOnly {
				rel.ReadOnly = true
			}
			data.Relationships = append(data.Relationships, rel)
		}

		// Sort relationships by name for consistent output
		sort.Slice(data.Relationships, func(i, j int) bool {
			return data.Relationships[i].Name < data.Relationships[j].Name
		})
	}

	return data
}

// buildEnumType creates an enum type definition from a schema with enum values
func (g *Generator) buildEnumType(typeName string, schema *openapi3.Schema) EnumType {
	enumType := EnumType{
		Name:        typeName,
		Description: cleanDescription(schema.Description),
		BaseType:    "string",
	}

	if schema.Type.Is("integer") {
		enumType.BaseType = "int"
	}

	// Generate enum values
	for _, enumVal := range schema.Enum {
		var valStr string
		var constName string

		switch v := enumVal.(type) {
		case string:
			valStr = v
			constName = typeName + strcase.ToCamel(v)
		case int, int64, float64:
			valStr = fmt.Sprintf("%v", v)
			constName = typeName + strcase.ToCamel(valStr)
		default:
			valStr = fmt.Sprintf("%v", v)
			constName = typeName + strcase.ToCamel(valStr)
		}

		enumType.Values = append(enumType.Values, EnumValue{
			Name:  constName,
			Value: valStr,
		})
	}

	return enumType
}

// buildNestedStruct creates a nested struct definition for an object attribute
// If useValue is true, all fields will be wrapped in value.Value (for request structs)
// If useValue is false, fields will use plain types (for response structs)
func (g *Generator) buildNestedStruct(name string, schema *openapi3.Schema, useValue bool) NestedStruct {
	nested := NestedStruct{
		Name:        name,
		Description: cleanDescription(schema.Description),
	}

	// Process all properties of the nested object
	for fieldName, fieldRef := range schema.Properties {
		if fieldRef.Value == nil {
			continue
		}

		var fieldType string

		if useValue {
			// For request structs: value.Value handles null, so get base type without nullable pointer
			baseFieldSchema := *fieldRef.Value
			baseFieldSchema.Nullable = false
			baseType := g.schemaToGoType(&baseFieldSchema)
			fieldType = "*value.Value[" + baseType + "]"
		} else {
			// For response structs: include nullable pointer if needed
			fieldType = g.schemaToGoType(fieldRef.Value)
		}

		field := NestedField{
			Name:        strcase.ToCamel(fieldName),
			JSONName:    fieldName,
			Type:        fieldType,
			Description: cleanDescription(fieldRef.Value.Description),
			ReadOnly:    fieldRef.Value.ReadOnly,
		}

		nested.Fields = append(nested.Fields, field)
	}

	// Sort fields by name for consistent output
	sort.Slice(nested.Fields, func(i, j int) bool {
		return nested.Fields[i].Name < nested.Fields[j].Name
	})

	return nested
}

// parseRelationship extracts relationship information
func (g *Generator) parseRelationship(name string, schema *openapi3.Schema) Relationship {
	rel := Relationship{
		Name:        strcase.ToCamel(name),
		JSONName:    name,
		Description: cleanDescription(schema.Description),
	}

	// Look for data property
	dataSchema := schema.Properties["data"]
	if dataSchema == nil || dataSchema.Value == nil {
		return rel
	}

	// Check if to-many (array)
	if dataSchema.Value.Type.Is("array") {
		rel.ToMany = true
		if dataSchema.Value.Items != nil && dataSchema.Value.Items.Value != nil {
			if typeEnum := dataSchema.Value.Items.Value.Properties["type"]; typeEnum != nil && typeEnum.Value != nil {
				if len(typeEnum.Value.Enum) > 0 {
					resourceType := fmt.Sprintf("%v", typeEnum.Value.Enum[0])
					// Look up the schema name from the JSON:API type
					if schemaName, ok := g.typeToSchemaMap[resourceType]; ok {
						rel.Type = schemaName
					} else {
						// Type not found in schema map - this relationship type doesn't have a schema
						// This can happen for internal/admin-only types
						// Empty type means we'll skip generating this relationship
						rel.Type = ""
					}
				}
			}
		}
	} else {
		// To-one relationship
		if typeEnum := dataSchema.Value.Properties["type"]; typeEnum != nil && typeEnum.Value != nil {
			if len(typeEnum.Value.Enum) > 0 {
				resourceType := fmt.Sprintf("%v", typeEnum.Value.Enum[0])
				// Look up the schema name from the JSON:API type
				if schemaName, ok := g.typeToSchemaMap[resourceType]; ok {
					rel.Type = schemaName
				} else {
					// Type not found in schema map - this relationship type doesn't have a schema
					// This can happen for internal/admin-only types
					// Empty type means we'll skip generating this relationship
					rel.Type = ""
				}
			}
		}
	}

	return rel
}

// schemaToGoType converts OpenAPI schema type to Go type
func (g *Generator) schemaToGoType(schema *openapi3.Schema) string {
	if schema.Title != "" {
		// Title is set when resolving $refs - use it as the type name
		// Check if nullable to add pointer
		if schema.Nullable {
			return "*" + schema.Title
		}
		return schema.Title
	}

	var baseType string

	if schema.Type.Is("array") {
		if schema.Items != nil {
			// Check if items has a $ref
			if schema.Items.Ref != "" {
				// Extract schema name from $ref
				parts := strings.Split(schema.Items.Ref, "/")
				schemaName := parts[len(parts)-1]
				baseType = "[]" + schemaName
			} else if schema.Items.Value != nil {
				// Otherwise recursively process the items schema
				itemType := g.schemaToGoType(schema.Items.Value)
				baseType = "[]" + itemType
			} else {
				baseType = "[]interface{}"
			}
		} else {
			baseType = "[]interface{}"
		}

		// For slices, nullable means the slice itself can be nil
		if schema.Nullable {
			return "*" + baseType
		}
		return baseType
	}

	// Handle enums
	if len(schema.Enum) > 0 {
		baseType = "string" // Enums are strings for now
		if schema.Nullable {
			return "*" + baseType
		}
		return baseType
	}

	// Handle basic types
	switch {
	case schema.Type.Is("string"):
		if schema.Format == "date-time" {
			baseType = "time.Time"
		} else {
			baseType = "string"
		}
	case schema.Type.Is("integer"):
		baseType = "int"
	case schema.Type.Is("number"):
		baseType = "float64"
	case schema.Type.Is("boolean"):
		baseType = "bool"
	case schema.Type.Is("object"):
		// For nested objects, use interface{} or map
		if schema.AdditionalProperties.Has != nil && *schema.AdditionalProperties.Has {
			baseType = "map[string]interface{}"
		} else {
			baseType = "map[string]interface{}"
		}
	default:
		baseType = "interface{}"
	}

	// Add pointer if nullable
	if schema.Nullable {
		return "*" + baseType
	}

	return baseType
}

// generateDocumentSchema generates a simple Document schema (e.g., TagRelationshipFieldsetsListingDocument)
// These are wrapper types used in request/response bodies, or plain schemas (like Reason)
func (g *Generator) generateDocumentSchema(name string, schema *openapi3.Schema, outputDir string) error {
	fileName := strcase.ToSnake(name) + ".gen.go"
	filePath := filepath.Join(outputDir, fileName)

	// Build the struct
	var fields []string
	isJSONAPIDocument := false

	// Check if has data field (indicates JSON:API document)
	if dataProp := schema.Properties["data"]; dataProp != nil && dataProp.Value != nil {
		dataType := g.schemaToGoType(dataProp.Value)
		fields = append(fields, fmt.Sprintf("\tData %s `json:\"data,omitempty\"`", dataType))
		isJSONAPIDocument = true
	}

	// Add other common JSON:API document fields
	if _, ok := schema.Properties["included"]; ok {
		fields = append(fields, "\tIncluded []map[string]interface{} `json:\"included,omitempty\"`")
		isJSONAPIDocument = true
	}

	if _, ok := schema.Properties["links"]; ok {
		fields = append(fields, "\tLinks map[string]string `json:\"links,omitempty\"`")
		isJSONAPIDocument = true
	}

	if _, ok := schema.Properties["meta"]; ok {
		fields = append(fields, "\tMeta map[string]interface{} `json:\"meta,omitempty\"`")
		isJSONAPIDocument = true
	}

	// If not a JSON:API document, generate all properties (plain schema)
	if !isJSONAPIDocument {
		// Get all property names and sort them for consistent output
		propertyNames := make([]string, 0, len(schema.Properties))
		for fieldName := range schema.Properties {
			propertyNames = append(propertyNames, fieldName)
		}
		sort.Strings(propertyNames)

		for _, fieldName := range propertyNames {
			propRef := schema.Properties[fieldName]
			if propRef == nil || propRef.Value == nil {
				continue
			}

			goFieldName := strcase.ToCamel(fieldName)
			fieldType := g.schemaToGoType(propRef.Value)
			jsonTag := fmt.Sprintf("`json:\"%s,omitempty\"`", fieldName)
			field := fmt.Sprintf("\t%s %s %s", goFieldName, fieldType, jsonTag)
			fields = append(fields, field)
		}
	}

	// Generate the file using template
	tmpl, err := template.New("document").Parse(documentSchemaTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse document template: %w", err)
	}

	data := struct {
		Name        string
		Description string
		Fields      []string
	}{
		Name:        name,
		Description: cleanDescription(schema.Description),
		Fields:      fields,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute document template: %w", err)
	}

	return os.WriteFile(filePath, buf.Bytes(), 0644)
}

// generateSimpleSchema generates a simple schema (e.g., TagRelationship, resource identifiers)
// These are basic structs with a few fields, not full JSON:API resources
func (g *Generator) generateSimpleSchema(name string, schema *openapi3.Schema, outputDir string) error {
	fileName := strcase.ToSnake(name) + ".gen.go"
	filePath := filepath.Join(outputDir, fileName)

	// Build the struct fields
	var fields []string
	hasID := false
	hasType := false
	var typeEnum string

	// Get all property names and sort them for consistent output
	// Exception: for JSON:API resources, put 'type' before 'id' (conventional order)
	propertyNames := make([]string, 0, len(schema.Properties))
	for fieldName := range schema.Properties {
		propertyNames = append(propertyNames, fieldName)
	}

	// Sort with custom comparator: type first, id second, rest alphabetically
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

	for _, fieldName := range propertyNames {
		propRef := schema.Properties[fieldName]
		if propRef == nil || propRef.Value == nil {
			continue
		}

		goFieldName := strcase.ToCamel(sanitizeGoName(fieldName))
		fieldType := g.schemaToGoType(propRef.Value)
		jsonTag := fmt.Sprintf("`json:\"%s,omitempty\"`", fieldName)

		field := fmt.Sprintf("\t%s %s %s", goFieldName, fieldType, jsonTag)
		fields = append(fields, field)

		// Track if this is a resource-like schema (has id and type)
		if fieldName == "id" {
			hasID = true
		}
		if fieldName == "type" {
			hasType = true
			// Extract type enum value if available
			if len(propRef.Value.Enum) > 0 {
				if typeVal, ok := propRef.Value.Enum[0].(string); ok {
					typeEnum = typeVal
				}
			}
		}
	}

	// Generate the file using template
	tmpl, err := template.New("simple").Parse(simpleSchemaTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse simple schema template: %w", err)
	}

	data := struct {
		Name                   string
		Description            string
		Fields                 []string
		HasResourceLikeMethods bool
		TypeEnum               string
	}{
		Name:                   name,
		Description:            cleanDescription(schema.Description),
		Fields:                 fields,
		HasResourceLikeMethods: hasID && hasType,
		TypeEnum:               typeEnum,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute simple schema template: %w", err)
	}

	return os.WriteFile(filePath, buf.Bytes(), 0644)
}
