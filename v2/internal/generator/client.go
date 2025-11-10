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

//go:embed templates/client.tpl
var clientTemplate string

// generateClient generates the main client file
func (g *Generator) generateClient(doc *openapi3.T, outputDir string) error {
	// Collect all resource names and check operations without x-resource (standalone)
	resources := make(map[string]bool)
	hasStandaloneOps := false

	for _, pathItem := range doc.Paths.Map() {
		for _, op := range pathItem.Operations() {
			if op == nil {
				continue
			}
			resource := getResourceName(op)
			if resource != "" {
				resources[resource] = true
			} else {
				hasStandaloneOps = true
			}
		}
	}

	// Convert to sorted slice
	var resourceList []string
	for resource := range resources {
		resourceList = append(resourceList, resource)
	}
	sort.Strings(resourceList)

	// Misc client holds all operations without x-resource
	if hasStandaloneOps {
		resourceList = append(resourceList, "Misc")
	}

	data := MainClientData{
		PackageName:    g.pkgName,
		Resources:      resourceList,
		BasePath:       g.basePath,
		ServerVariable: g.serverVariable,
		PreferHeader:   g.preferHeader,
	}

	tmpl, err := template.New("client").Funcs(template.FuncMap{
		"toSnake": strcase.ToSnake,
		"toLower": strings.ToLower,
	}).Parse(clientTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	filePath := filepath.Join(outputDir, "client.gen.go")
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write client.gen.go: %w", err)
	}

	return nil
}

// MainClientData holds template data for the main client
type MainClientData struct {
	PackageName    string
	Resources      []string
	BasePath       string
	ServerVariable string
	PreferHeader   string
}
