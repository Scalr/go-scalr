package generator_test

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scalr/go-scalr/v2/internal/generator"
)

// update rewrites golden files when set. Use: go test -run TestE2EGenerate -update
var update = flag.Bool("update", false, "update golden files")

// specGeneratedFiles lists the files whose content depends on the OpenAPI spec.
// Static support files (client/, value/) are deterministic copies — their
// presence is verified separately, not byte-compared.
var specGeneratedFiles = []string{
	"client.gen.go",
	filepath.Join("schemas", "widget.gen.go"),
	filepath.Join("ops", "widget", "widget.gen.go"),
}

// staticGeneratedFiles are a subset of static files that must exist after generation.
var staticGeneratedFiles = []string{
	filepath.Join("client", "http.gen.go"),
	filepath.Join("client", "errors.gen.go"),
	filepath.Join("client", "iterator.gen.go"),
	filepath.Join("value", "value.gen.go"),
}

func TestE2EGenerate(t *testing.T) {
	specPath := filepath.Join("testdata", "spec.yml")
	goldenDir := filepath.Join("testdata", "golden")

	tmpDir := t.TempDir()
	gen := generator.New(tmpDir, "testpkg")
	if err := gen.Generate(specPath); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Verify static files are present.
	for _, rel := range staticGeneratedFiles {
		if _, err := os.Stat(filepath.Join(tmpDir, rel)); err != nil {
			t.Errorf("static file missing: %s", rel)
		}
	}

	// Compare (or update) spec-generated files.
	for _, rel := range specGeneratedFiles {
		got, err := os.ReadFile(filepath.Join(tmpDir, rel))
		if err != nil {
			t.Errorf("generated file missing: %s: %v", rel, err)
			continue
		}

		goldenPath := filepath.Join(goldenDir, filepath.FromSlash(rel))

		if *update {
			if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
				t.Fatalf("mkdir %s: %v", filepath.Dir(goldenPath), err)
			}
			if err := os.WriteFile(goldenPath, got, 0644); err != nil {
				t.Fatalf("write golden %s: %v", goldenPath, err)
			}
			t.Logf("updated golden: %s", goldenPath)
			continue
		}

		want, err := os.ReadFile(goldenPath)
		if err != nil {
			t.Errorf("golden file missing: %s\n  run: go test -run TestE2EGenerate -update", goldenPath)
			continue
		}

		if !bytes.Equal(got, want) {
			t.Errorf("generated %s differs from golden:\n%s\n  run: go test -run TestE2EGenerate -update to refresh",
				rel, firstDiff(want, got))
		}
	}
}

// firstDiff returns a short human-readable description of the first difference
// between two byte slices compared line by line.
func firstDiff(a, b []byte) string {
	aLines := strings.Split(string(a), "\n")
	bLines := strings.Split(string(b), "\n")

	n := len(aLines)
	if len(bLines) < n {
		n = len(bLines)
	}

	for i := 0; i < n; i++ {
		if aLines[i] != bLines[i] {
			return fmt.Sprintf("  first diff at line %d:\n    golden: %q\n    got:    %q",
				i+1, aLines[i], bLines[i])
		}
	}

	return fmt.Sprintf("  line count: golden %d, got %d", len(aLines), len(bLines))
}
