package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Compile-time proof of interface implementation.
var _ Environments = (*environments)(nil)

// Environments describes all the environment related methods that the
// Scalr IACP API supports.
type Environments interface {
	List(ctx context.Context, options EnvironmentListOptions) (*EnvironmentList, error)
	Read(ctx context.Context, environmentID string) (*Environment, error)
	Create(ctx context.Context, options EnvironmentCreateOptions) (*Environment, error)
	Update(ctx context.Context, environmentID string, options EnvironmentUpdateOptions) (*Environment, error)
	Delete(ctx context.Context, environmentID string) error
}

// environments implements Environments.
type environments struct {
	client *Client
}

// EnvironmentStatus represents an environment status.
type EnvironmentStatus string

// List of available environment statuses.
const (
	EnvironmentStatusActive   EnvironmentStatus = "Active"
	EnvironmentStatusInactive EnvironmentStatus = "Inactive"
)

// CloudCredential relationship
type CloudCredential struct {
	ID string `jsonapi:"primary,cloud-credentials"`
}

// EnvironmentList represents a list of environments.
type EnvironmentList struct {
	*Pagination
	Items []*Environment
}

// Environment represents a Scalr environment.
type Environment struct {
	ID                    string            `jsonapi:"primary,environments"`
	Name                  string            `jsonapi:"attr,name"`
	CostEstimationEnabled bool              `jsonapi:"attr,cost-estimation-enabled"`
	CreatedAt             time.Time         `jsonapi:"attr,created-at,iso8601"`
	Status                EnvironmentStatus `jsonapi:"attr,status"`

	// Relations
	Account                       *Account                 `jsonapi:"relation,account"`
	CloudCredentials              []*CloudCredential       `jsonapi:"relation,cloud-credentials"`
	PolicyGroups                  []*PolicyGroup           `jsonapi:"relation,policy-groups"`
	DefaultProviderConfigurations []*ProviderConfiguration `jsonapi:"relation,default-provider-configurations"`
	ProviderConfigurations        []*ProviderConfiguration `jsonapi:"relation,provider-configurations"`
	CreatedBy                     *User                    `jsonapi:"relation,created-by"`
}

// Organization is Environment included in Workspace - always prefer Environment
type Organization struct {
	ID                    string            `jsonapi:"primary,organizations"`
	Name                  string            `jsonapi:"attr,name"`
	CostEstimationEnabled bool              `jsonapi:"attr,cost-estimation-enabled"`
	CreatedAt             time.Time         `jsonapi:"attr,created-at,iso8601"`
	CreatedBy             string            `jsonapi:"attr,created-by"`
	Status                EnvironmentStatus `jsonapi:"attr,status"`

	// Relations
	Account *Account `jsonapi:"relation,account"`
}

// EnvironmentCreateOptions represents the options for creating a new Environment.
type EnvironmentCreateOptions struct {
	ID                    string  `jsonapi:"primary,environments"`
	Name                  *string `jsonapi:"attr,name"`
	CostEstimationEnabled *bool   `jsonapi:"attr,cost-estimation-enabled,omitempty"`

	// Relations
	Account                       *Account                 `jsonapi:"relation,account"`
	CloudCredentials              []*CloudCredential       `jsonapi:"relation,cloud-credentials,omitempty"`
	PolicyGroups                  []*PolicyGroup           `jsonapi:"relation,policy-groups,omitempty"`
	DefaultProviderConfigurations []*ProviderConfiguration `jsonapi:"relation,default-provider-configurations,omitempty"`
}

func (o EnvironmentCreateOptions) valid() error {
	if o.Account == nil {
		return errors.New("account is required")
	}
	if !validStringID(&o.Account.ID) {
		return errors.New("invalid value for account ID")
	}
	if o.Name == nil {
		return errors.New("name is required")
	}
	return nil
}

type EnvironmentListOptions struct {
	ListOptions

	Include *string `url:"include,omitempty"`
	Name    *string `url:"query,omitempty"`
	Id      *string `url:"query,omitempty"`
	Account *string `url:"filter[account],omitempty"`
}

// List all the environmens.
func (s *environments) List(ctx context.Context, options EnvironmentListOptions) (*EnvironmentList, error) {
	req, err := s.client.newRequest("GET", "environments", &options)
	if err != nil {
		return nil, err
	}

	envl := &EnvironmentList{}
	err = s.client.do(ctx, req, envl)
	if err != nil {
		return nil, err
	}

	return envl, nil
}

// Create is used to create a new Environment.
func (s *environments) Create(ctx context.Context, options EnvironmentCreateOptions) (*Environment, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}
	// Make sure we don't send a user provided ID.
	options.ID = ""
	req, err := s.client.newRequest("POST", "environments", &options)
	if err != nil {
		return nil, err
	}

	environment := &Environment{}
	err = s.client.do(ctx, req, environment)
	if err != nil {
		return nil, err
	}

	return environment, nil
}

// Read an environment by its ID.
func (s *environments) Read(ctx context.Context, environmentID string) (*Environment, error) {
	if !validStringID(&environmentID) {
		return nil, errors.New("invalid value for environment ID")
	}

	options := struct {
		Include string `url:"include"`
	}{
		Include: "created-by",
	}
	u := fmt.Sprintf("environments/%s", url.QueryEscape(environmentID))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	env := &Environment{}
	err = s.client.do(ctx, req, env)
	if err != nil {
		return nil, err
	}

	return env, nil
}

// EnvironmentUpdateOptions represents the options for updating an environment.
type EnvironmentUpdateOptions struct {
	// For internal use only!
	ID                    string  `jsonapi:"primary,environments"`
	Name                  *string `jsonapi:"attr,name,omitempty"`
	CostEstimationEnabled *bool   `jsonapi:"attr,cost-estimation-enabled,omitempty"`

	// Relations
	CloudCredentials              []*CloudCredential       `jsonapi:"relation,cloud-credentials"`
	PolicyGroups                  []*PolicyGroup           `jsonapi:"relation,policy-groups"`
	DefaultProviderConfigurations []*ProviderConfiguration `jsonapi:"relation,default-provider-configurations,omitempty"`
}

// Update settings of an existing environment.
func (s *environments) Update(ctx context.Context, environmentID string, options EnvironmentUpdateOptions) (*Environment, error) {
	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("environments/%s", url.QueryEscape(environmentID))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	env := &Environment{}
	err = s.client.do(ctx, req, env)
	if err != nil {
		return nil, err
	}

	return env, nil
}

// Delete an environment by its ID.
func (s *environments) Delete(ctx context.Context, environmentID string) error {
	if !validStringID(&environmentID) {
		return errors.New("invalid value for environment ID")
	}

	u := fmt.Sprintf("environments/%s", url.QueryEscape(environmentID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
