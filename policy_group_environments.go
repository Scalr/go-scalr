package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/svanharmelen/jsonapi"
)

// Compile-time proof of interface implementation.
var _ PolicyGroupEnvironments = (*policyGroupEnvironment)(nil)

// PolicyGroupEnvironments describes all the policy group environments related methods that the
// Scalr API supports.
type PolicyGroupEnvironments interface {
	Create(ctx context.Context, options PolicyGroupEnvironmentsCreateOptions) error
	Update(ctx context.Context, options PolicyGroupEnvironmentsUpdateOptions) error
	Delete(ctx context.Context, options PolicyGroupEnvironmentDeleteOptions) error
}

// policyGroupEnvironments implements PolicyGroupEnvironments.
type policyGroupEnvironment struct {
	client *Client
}

// PolicyGroupEnvironment represents a single policy group environment relation.
type PolicyGroupEnvironment struct {
	ID string `jsonapi:"primary,environments"`
}

// PolicyGroupEnvironmentsCreateOptions represents options for creating new policy group environment linkage
type PolicyGroupEnvironmentsCreateOptions struct {
	PolicyGroupID           string
	PolicyGroupEnvironments []*PolicyGroupEnvironment
}

// PolicyGroupEnvironmentsUpdateOptions represents options for updating policy group environment linkage
type PolicyGroupEnvironmentsUpdateOptions struct {
	PolicyGroupID           string
	PolicyGroupEnvironments []*PolicyGroupEnvironment
}

type PolicyGroupEnvironmentDeleteOptions struct {
	PolicyGroupID string
	EnvironmentID string
}

func (o PolicyGroupEnvironmentsCreateOptions) valid() error {
	if !validStringID(&o.PolicyGroupID) {
		return errors.New("invalid value for policy group ID")
	}
	if o.PolicyGroupEnvironments == nil || len(o.PolicyGroupEnvironments) < 1 {
		return errors.New("list of environments is required")
	}
	return nil
}

func (o PolicyGroupEnvironmentsUpdateOptions) valid() error {
	if !validStringID(&o.PolicyGroupID) {
		return errors.New("invalid value for policy group ID")
	}
	return nil
}

func (o PolicyGroupEnvironmentDeleteOptions) valid() error {
	if !validStringID(&o.PolicyGroupID) {
		return errors.New("invalid value for policy group ID")
	}

	if !validStringID(&o.EnvironmentID) {
		return errors.New("invalid value for environment ID")
	}

	return nil
}

// Create a new policy group environments linkage.
func (s *policyGroupEnvironment) Create(ctx context.Context, options PolicyGroupEnvironmentsCreateOptions) error {
	if err := options.valid(); err != nil {
		return err
	}
	u := fmt.Sprintf("policy-groups/%s/relationships/environments", url.QueryEscape(options.PolicyGroupID))
	payload, err := jsonapi.Marshal(options.PolicyGroupEnvironments)
	if err != nil {
		return err
	}
	req, err := s.client.newRequest("POST", u, payload)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// Update policy group environments linkage
func (s *policyGroupEnvironment) Update(ctx context.Context, options PolicyGroupEnvironmentsUpdateOptions) error {
	if err := options.valid(); err != nil {
		return err
	}
	u := fmt.Sprintf("policy-groups/%s/relationships/environments", url.QueryEscape(options.PolicyGroupID))
	payload, err := jsonapi.Marshal(options.PolicyGroupEnvironments)
	if err != nil {
		return err
	}
	req, err := s.client.newRequest("PATCH", u, payload)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// Delete policy group environments linkage by its ID.
func (s *policyGroupEnvironment) Delete(ctx context.Context, options PolicyGroupEnvironmentDeleteOptions) error {
	if err := options.valid(); err != nil {
		return err
	}

	u := fmt.Sprintf(
		"policy-groups/%s/relationships/environments/%s",
		url.QueryEscape(options.PolicyGroupID),
		url.QueryEscape(options.EnvironmentID),
	)
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
