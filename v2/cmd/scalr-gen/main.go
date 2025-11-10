package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/scalr/go-scalr/v2/internal/generator"
)

func main() {
	var (
		specPath = flag.String("spec", "", "Path to OpenAPI spec file (required)")
		pkgName  = flag.String("package", "scalr", "API client package name. Default: scalr")
	)
	flag.Parse()

	if *specPath == "" {
		log.Fatal("--spec flag is required")
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Sanitize and validate package name
	safePkgName, err := sanitizePackageName(*pkgName)
	if err != nil {
		log.Fatalf("Invalid package name %q: %v", *pkgName, err)
	}

	if safePkgName != *pkgName {
		log.Printf("- Package name %q sanitized to %q", *pkgName, safePkgName)
	}

	clientRoot := filepath.Join(wd, safePkgName)

	log.Printf("- Generating API client...")

	gen := generator.New(clientRoot, safePkgName)
	if err := gen.Generate(*specPath); err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	log.Println("- Done.")
}

// sanitizePackageName ensures the package name is valid as Go package and as directory name
func sanitizePackageName(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("package name cannot be empty")
	}

	name = strings.ToLower(name)

	// Replace invalid characters with underscores
	// Valid: letters, digits, underscores
	var sb strings.Builder
	for i, r := range name {
		if unicode.IsLetter(r) || (i > 0 && unicode.IsDigit(r)) || r == '_' {
			sb.WriteRune(r)
		} else if unicode.IsDigit(r) || unicode.IsSpace(r) || r == '-' {
			// Convert spaces and hyphens to underscores, skip leading digits
			if i > 0 {
				sb.WriteRune('_')
			}
		}
		// Skip other invalid characters
	}

	sanitized := sb.String()

	if sanitized == "" {
		return "", fmt.Errorf("package name %q contains only invalid characters", name)
	}

	// Prefix with "pkg_" if not starts with a letter
	if !unicode.IsLetter(rune(sanitized[0])) {
		sanitized = "pkg_" + sanitized
	}

	// Suffix with "_pkg" if the sanitized name is a Go reserved word
	if generator.IsGoReservedWord(sanitized) {
		sanitized = sanitized + "_pkg"
	}

	if !isValidPackageName(sanitized) {
		return "", fmt.Errorf("cannot create valid package name from %q", name)
	}

	return sanitized, nil
}

// isValidPackageName checks if a string is a valid Go package name
func isValidPackageName(name string) bool {
	if name == "" {
		return false
	}

	// Must start with a letter
	if !unicode.IsLetter(rune(name[0])) {
		return false
	}

	// Must contain only letters, digits, and underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, name)
	return matched
}
