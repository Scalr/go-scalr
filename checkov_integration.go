package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ CheckovIntegrations = (*checkovIntegrations)(nil)

type CheckovIntegrations interface {
	List(ctx context.Context, options CheckovIntegrationListOptions) (*CheckovIntegrationList, error)
	Create(ctx context.Context, options CheckovIntegrationCreateOptions) (*CheckovIntegration, error)
	Read(ctx context.Context, id string) (*CheckovIntegration, error)
	Update(ctx context.Context, id string, options CheckovIntegrationUpdateOptions) (*CheckovIntegration, error)
	Delete(ctx context.Context, id string) error
}

// checkovIntegrations implements CheckovIntegrations.
type checkovIntegrations struct {
	client *Client
}

// CheckovIntegration represents a Scalr IACP Checkov integration.
type CheckovIntegration struct {
	ID                    string                     `jsonapi:"primary,checkov-integrations"`
	Name                  string                     `jsonapi:"attr,name"`
	Version               string                     `jsonapi:"attr,version"`
	CliArgs               string                     `jsonapi:"attr,cli-args"`
	IsShared              bool                       `jsonapi:"attr,is-shared"`
	VCSRepo               *CheckovIntegrationVCSRepo `jsonapi:"attr,vcs-repo"`
	ExternalChecksEnabled bool                       `jsonapi:"attr,external-checks-enabled"`

	// Relations
	Environments []*Environment `jsonapi:"relation,environments"`
	VcsProvider  *VcsProvider   `jsonapi:"relation,vcs-provider"`
}

// CheckovIntegrationVCSRepo contains the configuration of a VCS integration.
type CheckovIntegrationVCSRepo struct {
	Identifier string `json:"identifier"`
	Branch     string `json:"branch"`
	Path       string `json:"path"`
}

type CheckovIntegrationList struct {
	*Pagination
	Items []*CheckovIntegration
}

type CheckovIntegrationListOptions struct {
	ListOptions
}

// CheckovIntegrationVCSRepoOptions represents the configuration options of a VCS integration.
type CheckovIntegrationVCSRepoOptions struct {
	Identifier *string `json:"identifier"`
	Branch     *string `json:"branch,omitempty"`
	Path       *string `json:"path,omitempty"`
}

type CheckovIntegrationCreateOptions struct {
	ID                    string                            `jsonapi:"primary,checkov-integrations"`
	Name                  *string                           `jsonapi:"attr,name"`
	Version               *string                           `jsonapi:"attr,version,omitempty"`
	CliArgs               *string                           `jsonapi:"attr,cli-args,omitempty"`
	IsShared              *bool                             `jsonapi:"attr,is-shared,omitempty"`
	VCSRepo               *CheckovIntegrationVCSRepoOptions `jsonapi:"attr,vcs-repo,omitempty"`
	ExternalChecksEnabled *bool                             `jsonapi:"attr,external-checks-enabled,omitempty"`

	// Relations
	Environments []*Environment `jsonapi:"relation,environments"`
	VcsProvider  *VcsProvider   `jsonapi:"relation,vcs-provider,omitempty"`
}

type CheckovIntegrationUpdateOptions struct {
	ID                    string                            `jsonapi:"primary,checkov-integrations"`
	Name                  *string                           `jsonapi:"attr,name,omitempty"`
	Version               *string                           `jsonapi:"attr,version,omitempty"`
	CliArgs               *string                           `jsonapi:"attr,cli-args,omitempty"`
	IsShared              *bool                             `jsonapi:"attr,is-shared,omitempty"`
	VCSRepo               *CheckovIntegrationVCSRepoOptions `jsonapi:"attr,vcs-repo"`
	ExternalChecksEnabled *bool                             `jsonapi:"attr,external-checks-enabled,omitempty"`

	// Relations
	Environments []*Environment `jsonapi:"relation,environments"`
	VcsProvider  *VcsProvider   `jsonapi:"relation,vcs-provider"`
}

func (s *checkovIntegrations) List(
	ctx context.Context, options CheckovIntegrationListOptions,
) (*CheckovIntegrationList, error) {
	req, err := s.client.newRequest("GET", "integrations/checkov", &options)
	if err != nil {
		return nil, err
	}

	cil := &CheckovIntegrationList{}
	err = s.client.do(ctx, req, cil)
	if err != nil {
		return nil, err
	}

	return cil, nil
}

func (s *checkovIntegrations) Create(
	ctx context.Context, options CheckovIntegrationCreateOptions,
) (*CheckovIntegration, error) {
	// Make sure we don't send a user provided ID.
	options.ID = ""

	req, err := s.client.newRequest("POST", "integrations/checkov", &options)
	if err != nil {
		return nil, err
	}

	ci := &CheckovIntegration{}
	err = s.client.do(ctx, req, ci)
	if err != nil {
		return nil, err
	}

	return ci, nil
}

func (s *checkovIntegrations) Read(ctx context.Context, id string) (*CheckovIntegration, error) {
	if !validStringID(&id) {
		return nil, errors.New("invalid value for Checkov integration ID")
	}

	u := fmt.Sprintf("integrations/checkov/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	ci := &CheckovIntegration{}
	err = s.client.do(ctx, req, ci)
	if err != nil {
		return nil, err
	}

	return ci, nil
}

func (s *checkovIntegrations) Update(
	ctx context.Context, id string, options CheckovIntegrationUpdateOptions,
) (*CheckovIntegration, error) {
	if !validStringID(&id) {
		return nil, errors.New("invalid value for Checkov integration ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("integrations/checkov/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	ci := &CheckovIntegration{}
	err = s.client.do(ctx, req, ci)
	if err != nil {
		return nil, err
	}

	return ci, nil
}

func (s *checkovIntegrations) Delete(ctx context.Context, id string) error {
	if !validStringID(&id) {
		return errors.New("invalid value for Checkov integration ID")
	}

	u := fmt.Sprintf("integrations/checkov/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
