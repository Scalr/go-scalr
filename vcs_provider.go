package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ VcsProviders = (*vcsProviders)(nil)

// VcsProviders describes all the VcsProviders related methods that the Scalr
// IACP API supports.
//
// IACP API docs: https://www.scalr.com/docs/en/latest/api/index.html
type VcsProviders interface {
	// List the vcs providers.
	List(ctx context.Context, options VcsProvidersListOptions) (*VcsProvidersList, error)
	Create(ctx context.Context, options VcsProviderCreateOptions) (*VcsProvider, error)
	Read(ctx context.Context, vcsProvider string) (*VcsProvider, error)
	Update(ctx context.Context, vcsProvider string, options VcsProviderUpdateOptions) (*VcsProvider, error)
	Delete(ctx context.Context, vcsProvider string) error
}

// vcsProviders implements VcsProviders.
type vcsProviders struct {
	client *Client
}

// VcsType represents a type VCS provider.
type VcsType string

const (
	Github              VcsType = "github"
	GithubEnterprise    VcsType = "github_enterprise"
	Gitlab              VcsType = "gitlab"
	GitlabEnterprise    VcsType = "gitlab_enterprise"
	Bitbucket           VcsType = "bitbucket"
	BitbucketEnterprise VcsType = "bitbucket_enterprise"
	AzureDevOpsServices VcsType = "azure_dev_ops_services"
)

// AuthType represents the authorization type used in VCS provider.
type AuthType string

const (
	Oauth2        AuthType = "oauth2"
	PersonalToken AuthType = "personal_token"
)

// VcsProvidersList represents a list of VCS providers.
type VcsProvidersList struct {
	*Pagination
	Items []*VcsProvider
}

// OAuth contains the properties required for 'oauth2' authorization type.
type OAuth struct {
	ClientId     string `json:"client-id"`
	ClientSecret string `json:"client-secret"`
}

// VcsProvider represents a Scalr IACP VcsProvider.
type VcsProvider struct {
	ID                 string   `jsonapi:"primary,vcs-providers"`
	Name               string   `jsonapi:"attr,name"`
	Url                string   `jsonapi:"attr,url"`
	VcsType            VcsType  `jsonapi:"attr,vcs-type"`
	AuthType           AuthType `jsonapi:"attr,auth-type"`
	OAuth              *OAuth   `jsonapi:"attr,oauth"`
	Token              *string  `jsonapi:"attr,token"`
	Username           *string  `jsonapi:"attr,username"`
	IsShared           bool     `jsonapi:"attr,is-shared"`
	DraftPrRunsEnabled bool     `jsonapi:"attr,draft-pr-runs-enabled"`

	// Relations
	Environments []*Environment `jsonapi:"relation,environments"`
	Account      *Account       `jsonapi:"relation,account"`
	AgentPool    *AgentPool     `jsonapi:"relation,agent-pool"`
}

// VcsProvidersListOptions represents the options for listing vcs providers.
type VcsProvidersListOptions struct {
	ListOptions

	// Filter by identifier.
	ID *string `url:"filter[vcs-provider],omitempty"`

	// Query string.
	Query *string `url:"query,omitempty"`

	// The comma-separated list of attributes.
	Sort *string `url:"sort,omitempty"`

	// Filter by vcs-type
	VcsType *VcsType `url:"filter[vcs-type],omitempty"`

	// Scope filters.
	Environment *string `url:"filter[environment],omitempty"`
	Account     *string `url:"filter[account],omitempty"`
	AgentPool   *string `url:"filter[agent-pool],omitempty"`
}

// List the vcs providers.
func (s *vcsProviders) List(ctx context.Context, options VcsProvidersListOptions) (*VcsProvidersList, error) {
	req, err := s.client.newRequest("GET", "vcs-providers", &options)
	if err != nil {
		return nil, err
	}

	providersList := &VcsProvidersList{}
	err = s.client.do(ctx, req, providersList)
	if err != nil {
		return nil, err
	}

	return providersList, nil
}

// VcsProviderCreateOptions represents the options for creating a new vcs provider.
type VcsProviderCreateOptions struct {
	ID                 string   `jsonapi:"primary,vcs-providers"`
	Name               *string  `jsonapi:"attr,name"`
	VcsType            VcsType  `jsonapi:"attr,vcs-type"`
	AuthType           AuthType `jsonapi:"attr,auth-type"`
	OAuth              *OAuth   `jsonapi:"attr,oauth"`
	Token              string   `jsonapi:"attr,token"`
	Url                *string  `jsonapi:"attr,url"`
	Username           *string  `jsonapi:"attr,username"`
	IsShared           *bool    `jsonapi:"attr,is-shared,omitempty"`
	DraftPrRunsEnabled *bool    `jsonapi:"attr,draft-pr-runs-enabled,omitempty"`

	// Relations
	Environments []*Environment `jsonapi:"relation,environments,omitempty"`
	Account      *Account       `jsonapi:"relation,account,omitempty"`
	AgentPool    *AgentPool     `jsonapi:"relation,agent-pool,omitempty"`
}

// Create is used to create a new vcs provider.
func (s *vcsProviders) Create(ctx context.Context, options VcsProviderCreateOptions) (*VcsProvider, error) {
	// Make sure we don't send a user provided ID.
	options.ID = ""

	req, err := s.client.newRequest("POST", "vcs-providers", &options)
	if err != nil {
		return nil, err
	}

	w := &VcsProvider{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

// Read a vcs provider by its ID.
func (s *vcsProviders) Read(ctx context.Context, vcsProviderID string) (*VcsProvider, error) {
	if !validStringID(&vcsProviderID) {
		return nil, errors.New("invalid value for vcs provider ID")
	}

	u := fmt.Sprintf("vcs-providers/%s", url.QueryEscape(vcsProviderID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	w := &VcsProvider{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

// VcsProviderUpdateOptions represents the options for updating a vcs provider.
type VcsProviderUpdateOptions struct {
	// For internal use only!
	ID                 string  `jsonapi:"primary,vcs-providers"`
	Name               *string `jsonapi:"attr,name,omitempty"`
	Token              *string `jsonapi:"attr,token,omitempty"`
	Url                *string `jsonapi:"attr,url,omitempty"`
	Username           *string `jsonapi:"attr,username,omitempty"`
	IsShared           *bool   `jsonapi:"attr,is-shared,omitempty"`
	DraftPrRunsEnabled *bool   `jsonapi:"attr,draft-pr-runs-enabled,omitempty"`

	// Relations
	Environments []*Environment `jsonapi:"relation,environments"`
	AgentPool    *AgentPool     `jsonapi:"relation,agent-pool"`
}

// Update settings of an existing vcs provider.
func (s *vcsProviders) Update(ctx context.Context, vcsProviderId string, options VcsProviderUpdateOptions) (*VcsProvider, error) {
	if !validStringID(&vcsProviderId) {
		return nil, errors.New("invalid value for vcs provider ID")
	}
	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("vcs-providers/%s", url.QueryEscape(vcsProviderId))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	vcs := &VcsProvider{}
	err = s.client.do(ctx, req, vcs)
	if err != nil {
		return nil, err
	}

	return vcs, nil
}

// Delete a vcs provider by its ID.
func (s *vcsProviders) Delete(ctx context.Context, vcsProviderId string) error {
	if !validStringID(&vcsProviderId) {
		return errors.New("invalid value for vcs provider ID")
	}

	u := fmt.Sprintf("vcs-providers/%s", url.QueryEscape(vcsProviderId))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
