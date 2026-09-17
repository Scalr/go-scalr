package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ ModuleTestConfigurations = (*moduleTestConfigurations)(nil)

// ModuleTestConfigurations describes all the module test configuration related
// methods that the Scalr API supports. A test configuration controls whether
// tofu tests are run for a module, what triggers a test run, and how a test
// failure is handled.
//
// This resource is not part of the public Scalr API yet, so requests are sent
// with the "Prefer: profile=internal" header.
type ModuleTestConfigurations interface {
	// Read a module test configuration by its ID.
	Read(ctx context.Context, testConfigurationID string) (*ModuleTestConfiguration, error)
	// Update creates or updates the test configuration of a module.
	Update(ctx context.Context, moduleID string, options ModuleTestConfigurationUpdateOptions) (*ModuleTestConfiguration, error)
}

// moduleTestConfigurations implements ModuleTestConfigurations.
type moduleTestConfigurations struct {
	client *Client
}

// ModuleTestFailureBehavior represents the behavior to apply when a module test fails.
type ModuleTestFailureBehavior string

// List all available module test failure behaviors.
const (
	// ModuleTestFailureBehaviorFailure fails the module setup/publish on a test failure.
	ModuleTestFailureBehaviorFailure ModuleTestFailureBehavior = "failure"
	// ModuleTestFailureBehaviorNotify only notifies about a test failure, without blocking the module.
	ModuleTestFailureBehaviorNotify ModuleTestFailureBehavior = "notify"
)

// ModuleTestConfiguration represents a Scalr module test configuration.
type ModuleTestConfiguration struct {
	ID                         string                    `jsonapi:"primary,test-configurations"`
	Enabled                    bool                      `jsonapi:"attr,enabled"`
	FailureBehavior            ModuleTestFailureBehavior `jsonapi:"attr,failure-behavior"`
	TriggerOnPrActivityEnabled bool                      `jsonapi:"attr,trigger-on-pr-activity-enabled"`
	TriggerOnNewVersionEnabled bool                      `jsonapi:"attr,trigger-on-new-version-enabled"`

	// Relations
	Module *Module `jsonapi:"relation,module,omitempty"`
}

// ModuleTestConfigurationUpdateOptions represents the options for creating or
// updating a module test configuration.
type ModuleTestConfigurationUpdateOptions struct {
	// For internal use only!
	ID string `jsonapi:"primary,test-configurations"`

	Enabled                    *bool                      `jsonapi:"attr,enabled,omitempty"`
	FailureBehavior            *ModuleTestFailureBehavior `jsonapi:"attr,failure-behavior,omitempty"`
	TriggerOnPrActivityEnabled *bool                      `jsonapi:"attr,trigger-on-pr-activity-enabled,omitempty"`
	TriggerOnNewVersionEnabled *bool                      `jsonapi:"attr,trigger-on-new-version-enabled,omitempty"`
}

// Read a module test configuration by its ID.
func (s *moduleTestConfigurations) Read(ctx context.Context, testConfigurationID string) (*ModuleTestConfiguration, error) {
	if !validStringID(&testConfigurationID) {
		return nil, errors.New("invalid value for test configuration ID")
	}

	u := fmt.Sprintf("test-configurations/%s", url.QueryEscape(testConfigurationID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Prefer", "profile=internal")

	tc := &ModuleTestConfiguration{}
	err = s.client.do(ctx, req, tc)
	if err != nil {
		return nil, err
	}

	return tc, nil
}

// Update creates or updates the test configuration of a module.
func (s *moduleTestConfigurations) Update(
	ctx context.Context, moduleID string, options ModuleTestConfigurationUpdateOptions,
) (*ModuleTestConfiguration, error) {
	if !validStringID(&moduleID) {
		return nil, errors.New("invalid value for module ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("modules/%s/test-configuration", url.QueryEscape(moduleID))
	req, err := s.client.newRequest("PUT", u, &options)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Prefer", "profile=internal")

	tc := &ModuleTestConfiguration{}
	err = s.client.do(ctx, req, tc)
	if err != nil {
		return nil, err
	}

	return tc, nil
}
