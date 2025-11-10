package generator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
	"golang.org/x/tools/imports"
)

// Generator generates Go API client from OpenAPI specs
type Generator struct {
	outputDir       string
	pkgName         string
	typeToSchemaMap map[string]string // JSON:API type to Schema name mapping ("workspaces" -> "Workspace")
	basePath        string            // /api/iacp/v3
	serverVariable  string
	preferHeader    string // Value for Prefer header if required
}

// New creates a new generator
func New(out, pkg string) *Generator {
	return &Generator{
		outputDir:       out,
		pkgName:         pkg,
		typeToSchemaMap: make(map[string]string),
	}
}

// Generate generates the API client
func (g *Generator) Generate(specPath string) error {
	log.Printf("Reading OpenAPI spec from %s", specPath)

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to load spec: %w", err)
	}

	log.Printf("Loaded %d schemas, %d paths", len(doc.Components.Schemas), len(doc.Paths.Map()))

	if err := g.parseSpecMetadata(doc); err != nil {
		return fmt.Errorf("failed to parse spec metadata: %w", err)
	}

	log.Printf("Detected API base path: %s", g.basePath)
	if g.preferHeader != "" {
		log.Printf("Detected \"Prefer\" header: %q", g.preferHeader)
	}

	log.Printf("Preparing output directory: %s", g.outputDir)
	if err := os.RemoveAll(g.outputDir); err != nil {
		return fmt.Errorf("failed to clean output directory: %w", err)
	}

	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	schemasDir := filepath.Join(g.outputDir, "schemas")
	if err := os.MkdirAll(schemasDir, 0755); err != nil {
		return fmt.Errorf("failed to create schemas directory: %w", err)
	}

	log.Println("Generating schemas...")
	if err := g.generateSchemas(doc, schemasDir); err != nil {
		return fmt.Errorf("failed to generate schemas: %w", err)
	}

	opsDir := filepath.Join(g.outputDir, "ops")
	log.Println("Generating operations...")
	if err := g.generateOperations(doc, opsDir); err != nil {
		return fmt.Errorf("failed to generate operations: %w", err)
	}

	log.Println("Generating main client...")
	if err := g.generateClient(doc, g.outputDir); err != nil {
		return fmt.Errorf("failed to generate client: %w", err)
	}

	log.Println("Generating common files...")
	if err := g.generateStatic(g.outputDir); err != nil {
		return fmt.Errorf("failed to generate common files: %w", err)
	}

	log.Println("Formatting generated code...")
	if err := g.formatCode(g.outputDir); err != nil {
		return fmt.Errorf("failed to format code: %w", err)
	}

	log.Println("Generation completed.")
	return nil
}

// formatCode runs goimports on all generated Go files
func (g *Generator) formatCode(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".gen.go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		formatted, err := imports.Process(path, content, nil)
		if err != nil {
			log.Printf("Warning: failed to format %s: %v", path, err)
			return nil // Don't fail on format errors
		}

		return os.WriteFile(path, formatted, info.Mode())
	})
}

// Helper to get resource name from operation
func getResourceName(op *openapi3.Operation) string {
	if op.Extensions["x-resource"] != nil {
		if resource, ok := op.Extensions["x-resource"].(string); ok {
			return resource
		}
	}
	return ""
}

// IsGoReservedWord checks if a word is a Go reserved keyword
func IsGoReservedWord(word string) bool {
	reserved := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}
	return reserved[word]
}

// parseSpecMetadata extracts API configuration from the spec
func (g *Generator) parseSpecMetadata(doc *openapi3.T) error {
	if len(doc.Servers) == 0 {
		return fmt.Errorf("no servers defined in spec")
	}

	server := doc.Servers[0]
	serverURL := server.URL

	// "https://{Domain}/api/iacp/v3" -> "/api/iacp/v3"
	varStart := strings.Index(serverURL, "{")
	varEnd := strings.Index(serverURL, "}")
	if varStart == -1 || varEnd == -1 {
		return fmt.Errorf("invalid server URL format: %s", serverURL)
	}

	g.serverVariable = serverURL[varStart+1 : varEnd]

	pathStart := varEnd + 1
	if pathStart < len(serverURL) {
		g.basePath = serverURL[pathStart:]
	} else {
		g.basePath = ""
	}

	// Detect if Prefer header is required
	if doc.Components != nil && doc.Components.Parameters != nil {
		if preferParam := doc.Components.Parameters["PreferParam"]; preferParam != nil {
			param := preferParam.Value
			if param != nil {
				if param.In == "header" && param.Name == "Prefer" && param.Required {
					if param.Schema != nil && param.Schema.Value != nil {
						if defaultValue, ok := param.Schema.Value.Default.(string); ok {
							g.preferHeader = defaultValue
						}
					}
				}
			}
		}
	}

	return nil
}

// cleanDescription cleans up a description for use in comments
func cleanDescription(desc string) string {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return ""
	}
	// Replace newlines with spaces
	desc = strings.ReplaceAll(desc, "\n", " ")
	// Remove multiple spaces
	for strings.Contains(desc, "  ") {
		desc = strings.ReplaceAll(desc, "  ", " ")
	}
	return desc
}

// sanitizeGoName converts names to valid Go identifiers
// Returns result in lowerCamelCase, call strcase.ToCamel if needed
func sanitizeGoName(name string) string {
	// Handle page[number] -> PageNumber
	if strings.Contains(name, "[") {
		name = strings.ReplaceAll(name, "[", " ")
		name = strings.ReplaceAll(name, "]", "")
	}
	// Handle filter[some-thing] -> FilterSomeThing
	name = strcase.ToLowerCamel(name)
	// Escape Go reserved words
	if IsGoReservedWord(name) {
		name = name + "_"
	}
	return name
}
