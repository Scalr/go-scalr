package client

import (
	"fmt"
	"runtime"
)

// Version is the current version of the client.
var Version = "dev"

// UserAgent returns the default User-Agent string for the client.
//
// Example: go-scalr/v2.0.0-rc1 (Go 1.24; darwin/arm64)
func UserAgent() string {
	return fmt.Sprintf("go-scalr/%s (Go %s; %s/%s)",
		Version,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}

// UserAgentWithApp returns a User-Agent string that includes both the
// application and the client library.
//
// Example: terraform-provider-scalr/v3.9.0 go-scalr/v2.0.0-rc1 (Go 1.24; darwin/arm64)
func UserAgentWithApp(appName, appVersion string) string {
	if appName == "" {
		return UserAgent()
	}

	clientUA := UserAgent()

	if appVersion != "" {
		return fmt.Sprintf("%s/%s %s", appName, appVersion, clientUA)
	}

	return fmt.Sprintf("%s %s", appName, clientUA)
}
