package client

import (
	"runtime"
	"strings"
	"testing"
)

// TestUserAgent tests the default User-Agent generation
func TestUserAgent(t *testing.T) {
	ua := UserAgent()

	// Should contain "go-scalr"
	if !strings.Contains(ua, "go-scalr") {
		t.Errorf("User-Agent should contain 'go-scalr', got: %s", ua)
	}

	// Should contain Go version
	if !strings.Contains(ua, "Go") {
		t.Errorf("User-Agent should contain Go version, got: %s", ua)
	}

	// Should contain OS
	if !strings.Contains(ua, runtime.GOOS) {
		t.Errorf("User-Agent should contain OS (%s), got: %s", runtime.GOOS, ua)
	}

	// Should contain architecture
	if !strings.Contains(ua, runtime.GOARCH) {
		t.Errorf("User-Agent should contain architecture (%s), got: %s", runtime.GOARCH, ua)
	}

	// Should follow format: go-scalr/VERSION (Go GOVERSION; OS/ARCH)
	// Example: go-scalr/dev (Go 1.23; darwin/arm64)
	if !strings.HasPrefix(ua, "go-scalr/") {
		t.Errorf("User-Agent should start with 'go-scalr/', got: %s", ua)
	}

	t.Logf("User-Agent: %s", ua)
}

// TestUserAgentWithApp tests combining application and client User-Agent
func TestUserAgentWithApp(t *testing.T) {
	tests := []struct {
		name           string
		appName        string
		appVersion     string
		expectedPrefix string
		shouldContain  []string
	}{
		{
			name:           "with app name and version",
			appName:        "terraform-provider-scalr",
			appVersion:     "3.9.0",
			expectedPrefix: "terraform-provider-scalr/3.9.0 go-scalr/",
			shouldContain:  []string{"terraform-provider-scalr/3.9.0", "go-scalr/"},
		},
		{
			name:           "with app name only",
			appName:        "terraform-provider-scalr",
			appVersion:     "",
			expectedPrefix: "terraform-provider-scalr go-scalr/",
			shouldContain:  []string{"terraform-provider-scalr", "go-scalr/"},
		},
		{
			name:           "empty app name returns default",
			appName:        "",
			appVersion:     "1.0.0",
			expectedPrefix: "go-scalr/",
			shouldContain:  []string{"go-scalr/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := UserAgentWithApp(tt.appName, tt.appVersion)

			// Check prefix
			if !strings.HasPrefix(ua, tt.expectedPrefix) {
				t.Errorf("User-Agent should start with %q, got: %s", tt.expectedPrefix, ua)
			}

			// Check all expected substrings
			for _, substr := range tt.shouldContain {
				if !strings.Contains(ua, substr) {
					t.Errorf("User-Agent should contain %q, got: %s", substr, ua)
				}
			}

			t.Logf("User-Agent: %s", ua)
		})
	}
}

// TestUserAgentFormat tests the exact format of User-Agent
func TestUserAgentFormat(t *testing.T) {
	// Save original version
	originalVersion := Version
	Version = "1.0.0"
	defer func() { Version = originalVersion }()

	ua := UserAgent()

	// Should match format: go-scalr/1.0.0 (Go 1.XX; os/arch)
	parts := strings.Split(ua, " ")
	if len(parts) < 2 {
		t.Fatalf("User-Agent should have at least 2 parts, got: %s", ua)
	}

	// First part: go-scalr/1.0.0
	if parts[0] != "go-scalr/1.0.0" {
		t.Errorf("First part should be 'go-scalr/1.0.0', got: %s", parts[0])
	}

	// Second part: (Go 1.XX; os/arch)
	if !strings.HasPrefix(parts[1], "(Go") {
		t.Errorf("Second part should start with '(Go', got: %s", parts[1])
	}

	if !strings.HasSuffix(ua, ")") {
		t.Errorf("User-Agent should end with ')', got: %s", ua)
	}
}

// TestUserAgentWithAppFormat tests the exact format when combining with app
func TestUserAgentWithAppFormat(t *testing.T) {
	// Save original version
	originalVersion := Version
	Version = "1.0.0"
	defer func() { Version = originalVersion }()

	ua := UserAgentWithApp("terraform-provider-scalr", "3.9.0")

	// Should match format: terraform-provider-scalr/3.9.0 go-scalr/1.0.0 (Go 1.XX; os/arch)
	if !strings.HasPrefix(ua, "terraform-provider-scalr/3.9.0 go-scalr/1.0.0") {
		t.Errorf("User-Agent should start with 'terraform-provider-scalr/3.9.0 go-scalr/1.0.0', got: %s", ua)
	}

	// Should have both application and client info
	parts := strings.Split(ua, " ")
	if len(parts) < 3 {
		t.Errorf("User-Agent should have at least 3 parts, got: %s", ua)
	}

	// First part: application
	if parts[0] != "terraform-provider-scalr/3.9.0" {
		t.Errorf("First part should be 'terraform-provider-scalr/3.9.0', got: %s", parts[0])
	}

	// Second part: client
	if parts[1] != "go-scalr/1.0.0" {
		t.Errorf("Second part should be 'go-scalr/1.0.0', got: %s", parts[1])
	}

	// Third part onwards: (Go version; os/arch)
	if !strings.HasPrefix(parts[2], "(Go") {
		t.Errorf("Third part should start with '(Go', got: %s", parts[2])
	}
}

// TestVersionVariable tests that Version can be set
func TestVersionVariable(t *testing.T) {
	// Save original
	originalVersion := Version
	defer func() { Version = originalVersion }()

	// Set custom version
	Version = "2.0.0-beta1"

	ua := UserAgent()
	if !strings.Contains(ua, "go-scalr/2.0.0-beta1") {
		t.Errorf("User-Agent should contain custom version, got: %s", ua)
	}
}
