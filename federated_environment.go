package scalr

import (
	"context"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ FederatedEnvironments = (*federatedEnvironments)(nil)

type FederatedEnvironments interface {
	List(ctx context.Context, envID string, options ListOptions) (*FederatedEnvironmentsList, error)
	Add(ctx context.Context, envID string, envs []*EnvironmentRelation) error
	Replace(ctx context.Context, envID string, envs []*EnvironmentRelation) error
	Delete(ctx context.Context, envID string, envs []*EnvironmentRelation) error
}

type FederatedEnvironmentsList struct {
	*Pagination
	Items []*EnvironmentRelation
}

type federatedEnvironments struct {
	client *Client
}

func (s *federatedEnvironments) List(ctx context.Context, envID string, options ListOptions) (*FederatedEnvironmentsList, error) {
	u := fmt.Sprintf("environments/%s/relationships/federated-environments", url.QueryEscape(envID))
	req, err := s.client.newRequest("GET", u, &options)
	if err != nil {
		return nil, err
	}

	el := &FederatedEnvironmentsList{}
	err = s.client.do(ctx, req, el)
	if err != nil {
		return nil, err
	}

	return el, nil
}

func (s *federatedEnvironments) Add(ctx context.Context, envID string, envs []*EnvironmentRelation) error {
	u := fmt.Sprintf("environments/%s/relationships/federated-environments", url.QueryEscape(envID))
	req, err := s.client.newRequest("POST", u, envs)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

func (s *federatedEnvironments) Replace(ctx context.Context, envID string, envs []*EnvironmentRelation) error {
	u := fmt.Sprintf("environments/%s/relationships/federated-environments", url.QueryEscape(envID))
	req, err := s.client.newRequest("PATCH", u, envs)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

func (s *federatedEnvironments) Delete(ctx context.Context, envID string, envs []*EnvironmentRelation) error {
	u := fmt.Sprintf("environments/%s/relationships/federated-environments", url.QueryEscape(envID))
	req, err := s.client.newRequest("DELETE", u, envs)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
