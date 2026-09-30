package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ ModuleTestProviderConfigurationLinks = (*moduleTestProviderConfigurationLinks)(nil)

type ModuleTestProviderConfigurationLinks interface {
	// List all the provider configuration links attached to a test configuration.
	List(ctx context.Context, testConfigurationID string, options ListOptions) (*ModuleTestProviderConfigurationLinkList, error)
	// Create attaches a provider configuration to a test configuration.
	Create(
		ctx context.Context, testConfigurationID string, options ModuleTestProviderConfigurationLinkCreateOptions,
	) (*ModuleTestProviderConfigurationLink, error)
	// Read a module test provider configuration link by its ID.
	Read(ctx context.Context, linkID string) (*ModuleTestProviderConfigurationLink, error)
	// Update an existing module test provider configuration link.
	Update(
		ctx context.Context, linkID string, options ModuleTestProviderConfigurationLinkUpdateOptions,
	) (*ModuleTestProviderConfigurationLink, error)
	// Delete a module test provider configuration link by its ID.
	Delete(ctx context.Context, linkID string) error
}

// moduleTestProviderConfigurationLinks implements ModuleTestProviderConfigurationLinks.
type moduleTestProviderConfigurationLinks struct {
	client *Client
}

// ModuleTestProviderConfigurationLink represents a link between a module test
// configuration and the provider configuration (credentials) used to run it.
type ModuleTestProviderConfigurationLink struct {
	ID string `jsonapi:"primary,module-test-provider-configuration-links"`

	// Relations
	ProviderConfiguration *ProviderConfiguration   `jsonapi:"relation,provider-configuration"`
	TestConfiguration     *ModuleTestConfiguration `jsonapi:"relation,test-configuration,omitempty"`
}

// ModuleTestProviderConfigurationLinkList represents a list of module test
// provider configuration links.
type ModuleTestProviderConfigurationLinkList struct {
	*Pagination
	Items []*ModuleTestProviderConfigurationLink
}

// List all the provider configuration links attached to a test configuration.
func (s *moduleTestProviderConfigurationLinks) List(
	ctx context.Context, testConfigurationID string, options ListOptions,
) (*ModuleTestProviderConfigurationLinkList, error) {
	if !validStringID(&testConfigurationID) {
		return nil, errors.New("invalid value for test configuration ID")
	}

	u := fmt.Sprintf("test-configurations/%s/provider-configuration-links", url.QueryEscape(testConfigurationID))
	req, err := s.client.newRequest("GET", u, &options)
	if err != nil {
		return nil, err
	}

	l := &ModuleTestProviderConfigurationLinkList{}
	err = s.client.do(ctx, req, l)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// ModuleTestProviderConfigurationLinkCreateOptions represents the options for
// attaching a provider configuration to a test configuration.
type ModuleTestProviderConfigurationLinkCreateOptions struct {
	// For internal use only!
	ID string `jsonapi:"primary,module-test-provider-configuration-links"`

	ProviderConfiguration *ProviderConfiguration `jsonapi:"relation,provider-configuration"`
}

// Create attaches a provider configuration to a test configuration.
func (s *moduleTestProviderConfigurationLinks) Create(
	ctx context.Context, testConfigurationID string, options ModuleTestProviderConfigurationLinkCreateOptions,
) (*ModuleTestProviderConfigurationLink, error) {
	if !validStringID(&testConfigurationID) {
		return nil, errors.New("invalid value for test configuration ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("test-configurations/%s/provider-configuration-links", url.QueryEscape(testConfigurationID))
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return nil, err
	}

	l := &ModuleTestProviderConfigurationLink{}
	err = s.client.do(ctx, req, l)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// Read a module test provider configuration link by its ID.
func (s *moduleTestProviderConfigurationLinks) Read(ctx context.Context, linkID string) (*ModuleTestProviderConfigurationLink, error) {
	if !validStringID(&linkID) {
		return nil, errors.New("invalid value for module test provider configuration link ID")
	}

	u := fmt.Sprintf("module-test-provider-configuration-links/%s", url.QueryEscape(linkID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	l := &ModuleTestProviderConfigurationLink{}
	err = s.client.do(ctx, req, l)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// ModuleTestProviderConfigurationLinkUpdateOptions represents the options for
// updating a module test provider configuration link.
type ModuleTestProviderConfigurationLinkUpdateOptions struct {
	// For internal use only!
	ID string `jsonapi:"primary,module-test-provider-configuration-links"`

	ProviderConfiguration *ProviderConfiguration `jsonapi:"relation,provider-configuration,omitempty"`
}

// Update an existing module test provider configuration link.
func (s *moduleTestProviderConfigurationLinks) Update(
	ctx context.Context, linkID string, options ModuleTestProviderConfigurationLinkUpdateOptions,
) (*ModuleTestProviderConfigurationLink, error) {
	if !validStringID(&linkID) {
		return nil, errors.New("invalid value for module test provider configuration link ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("module-test-provider-configuration-links/%s", url.QueryEscape(linkID))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	l := &ModuleTestProviderConfigurationLink{}
	err = s.client.do(ctx, req, l)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// Delete a module test provider configuration link by its ID.
func (s *moduleTestProviderConfigurationLinks) Delete(ctx context.Context, linkID string) error {
	if !validStringID(&linkID) {
		return errors.New("invalid value for module test provider configuration link ID")
	}

	u := fmt.Sprintf("module-test-provider-configuration-links/%s", url.QueryEscape(linkID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
