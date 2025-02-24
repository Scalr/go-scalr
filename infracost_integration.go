package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ InfracostIntegrations = (*infracostIntegrations)(nil)

type InfracostIntegrations interface {
	List(ctx context.Context, options InfracostIntegrationListOptions) (*InfracostIntegrationList, error)
	Create(ctx context.Context, options InfracostIntegrationCreateOptions) (*InfracostIntegration, error)
	Read(ctx context.Context, id string) (*InfracostIntegration, error)
	Update(ctx context.Context, id string, options InfracostIntegrationUpdateOptions) (*InfracostIntegration, error)
	Delete(ctx context.Context, id string) error
}

// infracostIntegrations implements InfracostIntegrations.
type infracostIntegrations struct {
	client *Client
}

type InfracostIntegrationList struct {
	*Pagination
	Items []*InfracostIntegration
}

// InfracostIntegration represents a Scalr IACP Infracost integration.
type InfracostIntegration struct {
	ID       string            `jsonapi:"primary,infracost-integration"`
	Name     string            `jsonapi:"attr,name"`
	Status   IntegrationStatus `jsonapi:"attr,status"`
	ApiKey   string            `jsonapi:"attr,api-key"`
	IsShared bool              `jsonapi:"attr,is-shared,omitempty"`

	// Relations
	Environments []*Environment `jsonapi:"relation,environments"`
}

type InfracostIntegrationListOptions struct {
	ListOptions

	InfracostIntegration *string `url:"filter[infracost-integration],omitempty"`
	Name                 *string `url:"filter[name],omitempty"`
}

type InfracostIntegrationCreateOptions struct {
	ID       string  `jsonapi:"primary,infracost-integration"`
	Name     *string `jsonapi:"attr,name"`
	ApiKey   *string `jsonapi:"attr,api-key"`
	IsShared *bool   `jsonapi:"attr,is-shared,omitempty"`

	// Relations
	Environments []*Environment `jsonapi:"relation,environments"`
}

type InfracostIntegrationUpdateOptions struct {
	ID       string  `jsonapi:"primary,infracost-integration"`
	Name     *string `jsonapi:"attr,name,omitempty"`
	ApiKey   *string `jsonapi:"attr,api-key,omitempty"`
	IsShared *bool   `jsonapi:"attr,is-shared,omitempty"`

	// Relations
	Environments []*Environment `jsonapi:"relation,environments"`
}

func (s *infracostIntegrations) List(
	ctx context.Context, options InfracostIntegrationListOptions,
) (*InfracostIntegrationList, error) {
	req, err := s.client.newRequest("GET", "integrations/infracost", &options)
	if err != nil {
		return nil, err
	}

	iil := &InfracostIntegrationList{}
	err = s.client.do(ctx, req, iil)
	if err != nil {
		return nil, err
	}

	return iil, nil
}

func (s *infracostIntegrations) Create(
	ctx context.Context, options InfracostIntegrationCreateOptions,
) (*InfracostIntegration, error) {
	// Make sure we don't send a user provided ID.
	options.ID = ""

	req, err := s.client.newRequest("POST", "integrations/infracost", &options)
	if err != nil {
		return nil, err
	}

	ii := &InfracostIntegration{}
	err = s.client.do(ctx, req, ii)
	if err != nil {
		return nil, err
	}

	return ii, nil
}

func (s *infracostIntegrations) Read(ctx context.Context, id string) (*InfracostIntegration, error) {
	if !validStringID(&id) {
		return nil, errors.New("invalid value for Infracost integration ID")
	}

	u := fmt.Sprintf("integrations/infracost/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	ii := &InfracostIntegration{}
	err = s.client.do(ctx, req, ii)
	if err != nil {
		return nil, err
	}

	return ii, nil
}

func (s *infracostIntegrations) Update(
	ctx context.Context, id string, options InfracostIntegrationUpdateOptions,
) (*InfracostIntegration, error) {
	if !validStringID(&id) {
		return nil, errors.New("invalid value for Infracost integration ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("integrations/infracost/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	ii := &InfracostIntegration{}
	err = s.client.do(ctx, req, ii)
	if err != nil {
		return nil, err
	}

	return ii, nil
}

func (s *infracostIntegrations) Delete(ctx context.Context, id string) error {
	if !validStringID(&id) {
		return errors.New("invalid value for Infracost integration ID")
	}

	u := fmt.Sprintf("integrations/infracost/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
