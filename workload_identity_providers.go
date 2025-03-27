package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ WorkloadIdentityProviders = (*workloadIdentityProviders)(nil)

// WorkloadIdentityProviders describes all the workload identity provider related methods that the Scalr API supports.
type WorkloadIdentityProviders interface {
	List(ctx context.Context, options WorkloadIdentityProvidersListOptions) (*WorkloadIdentityProvidersList, error)
	Create(ctx context.Context, options WorkloadIdentityProviderCreateOptions) (*WorkloadIdentityProvider, error)
	Read(ctx context.Context, providerID string) (*WorkloadIdentityProvider, error)
	Update(ctx context.Context, providerID string, options WorkloadIdentityProviderUpdateOptions) (*WorkloadIdentityProvider, error)
	Delete(ctx context.Context, providerID string) error
}

// workloadIdentityProviders implements WorkloadIdentityProviders.
type workloadIdentityProviders struct {
	client *Client
}

// WorkloadIdentityProvidersList represents a list of workload identity providers.
type WorkloadIdentityProvidersList struct {
	*Pagination
	Items []*WorkloadIdentityProvider
}

// WorkloadIdentityProvider represents a Scalr workload identity provider.
type WorkloadIdentityProvider struct {
	ID                           string                        `jsonapi:"primary,workload-identity-providers"`
	Name                         string                        `jsonapi:"attr,name"`
	URL                          string                        `jsonapi:"attr,url"`
	AllowedAudiences             []string                      `jsonapi:"attr,allowed-audiences"`
	CreatedAt                    string                        `jsonapi:"attr,created-at"`
	CreatedByEmail               *string                       `jsonapi:"attr,created-by-email"`
	Status                       string                        `jsonapi:"attr,status"`
	AssumeServiceAccountPolicies []*AssumeServiceAccountPolicy `jsonapi:"relation,assume-service-account-policies"`
}

// WorkloadIdentityProvidersListOptions represents the options for listing workload identity providers.
type WorkloadIdentityProvidersListOptions struct {
	ListOptions
	Sort   string                          `url:"sort,omitempty"`
	Query  *string                         `url:"query,omitempty"`
	Filter *WorkloadIdentityProviderFilter `url:"filter,omitempty"`
}

type WorkloadIdentityProviderFilter struct {
	WorkloadIdentityProvider string `url:"workload-identity-provider,omitempty"`
	Name                     string `url:"name,omitempty"`
	Url                      string `url:"url,omitempty"`
}

// List all workload identity providers within a Scalr account.
func (s *workloadIdentityProviders) List(ctx context.Context, options WorkloadIdentityProvidersListOptions) (*WorkloadIdentityProvidersList, error) {
	req, err := s.client.newRequest("GET", "workload-identity-providers", &options)
	if err != nil {
		return nil, err
	}

	wipList := &WorkloadIdentityProvidersList{}
	err = s.client.do(ctx, req, wipList)
	if err != nil {
		return nil, err
	}

	return wipList, nil
}

// WorkloadIdentityProviderCreateOptions represents the options for creating a new workload identity provider.
type WorkloadIdentityProviderCreateOptions struct {
	ID               string   `jsonapi:"primary,workload-identity-providers"`
	Name             *string  `jsonapi:"attr,name"`
	URL              *string  `jsonapi:"attr,url"`
	AllowedAudiences []string `jsonapi:"attr,allowed-audiences"`
}

// Create a new workload identity provider.
func (s *workloadIdentityProviders) Create(ctx context.Context, options WorkloadIdentityProviderCreateOptions) (*WorkloadIdentityProvider, error) {
	options.ID = ""

	req, err := s.client.newRequest("POST", "workload-identity-providers", &options)
	if err != nil {
		return nil, err
	}

	provider := &WorkloadIdentityProvider{}
	err = s.client.do(ctx, req, provider)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// Read a workload identity provider by provider ID.
func (s *workloadIdentityProviders) Read(ctx context.Context, providerID string) (*WorkloadIdentityProvider, error) {
	if !validStringID(&providerID) {
		return nil, errors.New("invalid value for workload identity provider ID")
	}

	urlPath := fmt.Sprintf("workload-identity-providers/%s", url.QueryEscape(providerID))
	req, err := s.client.newRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}

	provider := &WorkloadIdentityProvider{}
	err = s.client.do(ctx, req, provider)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// WorkloadIdentityProviderUpdateOptions represents the options for updating a workload identity provider.
type WorkloadIdentityProviderUpdateOptions struct {
	ID               string   `jsonapi:"primary,workload-identity-providers"`
	Name             *string  `jsonapi:"attr,name"`
	AllowedAudiences []string `jsonapi:"attr,allowed-audiences"`
}

// Update an existing workload identity provider.
func (s *workloadIdentityProviders) Update(ctx context.Context, providerID string, options WorkloadIdentityProviderUpdateOptions) (*WorkloadIdentityProvider, error) {
	if !validStringID(&providerID) {
		return nil, errors.New("invalid value for workload identity provider ID")
	}

	options.ID = ""

	urlPath := fmt.Sprintf("workload-identity-providers/%s", url.QueryEscape(providerID))
	req, err := s.client.newRequest("PATCH", urlPath, &options)
	if err != nil {
		return nil, err
	}

	provider := &WorkloadIdentityProvider{}
	err = s.client.do(ctx, req, provider)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// Delete a workload identity provider by its ID.
func (s *workloadIdentityProviders) Delete(ctx context.Context, providerID string) error {
	if !validStringID(&providerID) {
		return errors.New("invalid value for workload identity provider ID")
	}

	urlPath := fmt.Sprintf("workload-identity-providers/%s", url.QueryEscape(providerID))
	req, err := s.client.newRequest("DELETE", urlPath, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
