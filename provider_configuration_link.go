package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ ProviderConfigurationLinks = (*providerConfigurationLinks)(nil)

// ProviderConfigurationLinks describes all the provider configurartion link related methods that the Scalr API supports.
type ProviderConfigurationLinks interface {
	List(ctx context.Context, workspaceID string, options ProviderConfigurationLinksListOptions) (*ProviderConfigurationLinksList, error)
	Create(ctx context.Context, workspaceID string, options ProviderConfigurationLinkCreateOptions) (*ProviderConfigurationLink, error)
	Read(ctx context.Context, linkID string) (*ProviderConfigurationLink, error)
	Delete(ctx context.Context, linkID string) error
	Update(ctx context.Context, linkID string, options ProviderConfigurationLinkUpdateOptions) (*ProviderConfigurationLink, error)
}

// providerConfigurationLinks implements ProviderConfigurationLinks.
type providerConfigurationLinks struct {
	client *Client
}

// ProviderConfigurationLinksList represents a list of provider configuration links.
type ProviderConfigurationLinksList struct {
	*Pagination
	Items []*ProviderConfigurationLink
}

// ProviderConfigurationLink represents a Scalr provider configuration link.
type ProviderConfigurationLink struct {
	ID      string `jsonapi:"primary,provider-configuration-links"`
	Default bool   `jsonapi:"attr,default"`
	Alias   string `jsonapi:"attr,alias"`

	ProviderConfiguration *ProviderConfiguration `jsonapi:"relation,provider-configuration,omitempty"`
	Environment           *Environment           `jsonapi:"relation,environment,omitempty"`
	Workspace             *Workspace             `jsonapi:"relation,workspace,omitempty"`
}

// ProviderConfigurationLinksListOptions represents the options for listing provider configuration links.
type ProviderConfigurationLinksListOptions struct {
	ListOptions

	Include string `url:"include,omitempty"`
}

// List all provider configuration applied to the workspace.
func (s *providerConfigurationLinks) List(ctx context.Context, workspaceID string, options ProviderConfigurationLinksListOptions) (*ProviderConfigurationLinksList, error) {
	if !validStringID(&workspaceID) {
		return nil, errors.New("invalid value for provider configuration ID")
	}

	url_path := fmt.Sprintf("workspaces/%s/provider-configuration-links", url.QueryEscape(workspaceID))

	req, err := s.client.newRequest("GET", url_path, &options)
	if err != nil {
		return nil, err
	}

	linksList := &ProviderConfigurationLinksList{}
	err = s.client.do(ctx, req, linksList)
	if err != nil {
		return nil, err
	}

	return linksList, nil
}

// ProviderConfigurationLinkCreateOptions represents the options for creating a new provider configuration workspace link.
type ProviderConfigurationLinkCreateOptions struct {
	ID    string  `jsonapi:"primary,provider-configuration-links"`
	Alias *string `jsonapi:"attr,alias"`

	ProviderConfiguration *ProviderConfiguration `jsonapi:"relation,provider-configuration"`
}

// Create is used to create a new provider configuration workspace link.
func (s *providerConfigurationLinks) Create(ctx context.Context, workspaceID string, options ProviderConfigurationLinkCreateOptions) (*ProviderConfigurationLink, error) {
	options.ID = ""

	url_path := fmt.Sprintf("workspaces/%s/provider-configuration-links", url.QueryEscape(workspaceID))
	req, err := s.client.newRequest("POST", url_path, &options)
	if err != nil {
		return nil, err
	}

	link := &ProviderConfigurationLink{}
	err = s.client.do(ctx, req, link)

	if err != nil {
		return nil, err
	}

	return link, nil
}

// Read a provider configuration link by link ID.
func (s *providerConfigurationLinks) Read(ctx context.Context, linkID string) (*ProviderConfigurationLink, error) {
	if !validStringID(&linkID) {
		return nil, errors.New("invalid value for provider configuration link ID")
	}

	url_path := fmt.Sprintf("provider-configuration-links/%s", url.QueryEscape(linkID))

	options := struct {
		Include string `url:"include"`
	}{
		Include: "provider-configuration",
	}

	req, err := s.client.newRequest("GET", url_path, options)
	if err != nil {
		return nil, err
	}

	link := &ProviderConfigurationLink{}
	err = s.client.do(ctx, req, link)
	if err != nil {
		return nil, err
	}

	return link, nil
}

// ProviderConfigurationLinkUpdateOptions represents the options for updating a provider configuration link.
type ProviderConfigurationLinkUpdateOptions struct {
	ID    string  `jsonapi:"primary,provider-configuration-links"`
	Alias *string `jsonapi:"attr,alias"`
}

// Update an existing provider configuration link.
func (s *providerConfigurationLinks) Update(ctx context.Context, linkID string, options ProviderConfigurationLinkUpdateOptions) (*ProviderConfigurationLink, error) {
	if !validStringID(&linkID) {
		return nil, errors.New("invalid value for provider configuration link ID")
	}

	url_path := fmt.Sprintf("provider-configuration-links/%s", url.QueryEscape(linkID))

	req, err := s.client.newRequest("PATCH", url_path, &options)
	if err != nil {
		return nil, err
	}

	link := &ProviderConfigurationLink{}
	err = s.client.do(ctx, req, link)
	if err != nil {
		return nil, err
	}

	return link, nil
}

// Delete deletes a provider configuration link by its ID.
func (s *providerConfigurationLinks) Delete(ctx context.Context, linkID string) error {
	if !validStringID(&linkID) {
		return errors.New("invalid value for provider link ID")
	}

	url_path := fmt.Sprintf("provider-configuration-links/%s", url.QueryEscape(linkID))
	req, err := s.client.newRequest("DELETE", url_path, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
