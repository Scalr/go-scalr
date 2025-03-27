package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ AssumeServiceAccountPolicies = (*assumeServiceAccountPolicies)(nil)

type AssumeServiceAccountPolicies interface {
	List(ctx context.Context, options AssumeServiceAccountPoliciesListOptions) (*AssumeServiceAccountPoliciesList, error)
	Create(ctx context.Context, serviceAccountID string, options AssumeServiceAccountPolicyCreateOptions) (*AssumeServiceAccountPolicy, error)
	Read(ctx context.Context, serviceAccountID, policyID string) (*AssumeServiceAccountPolicy, error)
	Update(ctx context.Context, serviceAccountID, policyID string, options AssumeServiceAccountPolicyUpdateOptions) (*AssumeServiceAccountPolicy, error)
	Delete(ctx context.Context, serviceAccountID, policyID string) error
}

type assumeServiceAccountPolicies struct {
	client *Client
}

type AssumeServiceAccountPoliciesList struct {
	*Pagination
	Items []*AssumeServiceAccountPolicy
}

type AssumeServiceAccountPolicy struct {
	ID                     string                    `jsonapi:"primary,assume-service-account-policies"`
	Name                   string                    `jsonapi:"attr,name"`
	Provider               *WorkloadIdentityProvider `jsonapi:"relation,provider"`
	ServiceAccount         *ServiceAccount           `jsonapi:"relation,service-account"`
	MaximumSessionDuration int                       `jsonapi:"attr,maximum-session-duration"`
	ClaimConditions        []ClaimCondition          `jsonapi:"attr,claim-conditions"`
	CreatedAt              string                    `jsonapi:"attr,created-at"`
	CreatedByEmail         *string                   `jsonapi:"attr,created-by-email"`
}

type ClaimCondition struct {
	Claim    string  `json:"claim"`
	Value    string  `json:"value"`
	Operator *string `json:"operator,omitempty"`
}

type AssumeServiceAccountPolicyCreateOptions struct {
	ID                     string                    `jsonapi:"primary,assume-service-account-policies"`
	Name                   *string                   `jsonapi:"attr,name"`
	Provider               *WorkloadIdentityProvider `jsonapi:"relation,provider"`
	MaximumSessionDuration *int                      `jsonapi:"attr,maximum-session-duration,omitempty"`
	ClaimConditions        []ClaimCondition          `jsonapi:"attr,claim-conditions"`
}

type AssumeServiceAccountPolicyUpdateOptions struct {
	ID string `jsonapi:"primary,assume-service-account-policies"`

	Name                   *string           `jsonapi:"attr,name,omitempty"`
	MaximumSessionDuration *int              `jsonapi:"attr,maximum-session-duration,omitempty"`
	ClaimConditions        *[]ClaimCondition `jsonapi:"attr,claim-conditions,omitempty"`
}

type AssumeServiceAccountPoliciesListOptions struct {
	ListOptions
	Query  *string                           `url:"query,omitempty"`
	Filter *AssumeServiceAccountPolicyFilter `url:"filter,omitempty"`
}

type AssumeServiceAccountPolicyFilter struct {
	AssumeServiceAccountPolicy string `url:"assume-service-account-policy,omitempty"`
	ServiceAccount             string `url:"service-account,omitempty"`
	WorkloadIdentityProvider   string `url:"workload-identity-provider,omitempty"`
	Name                       string `url:"name,omitempty"`
}

func (s *assumeServiceAccountPolicies) List(ctx context.Context, options AssumeServiceAccountPoliciesListOptions) (*AssumeServiceAccountPoliciesList, error) {
	req, err := s.client.newRequest("GET", "assume-service-account-policies", &options)
	if err != nil {
		return nil, err
	}

	policies := &AssumeServiceAccountPoliciesList{}
	err = s.client.do(ctx, req, policies)
	if err != nil {
		return nil, err
	}

	return policies, nil
}

func (s *assumeServiceAccountPolicies) Create(ctx context.Context, serviceAccountID string, options AssumeServiceAccountPolicyCreateOptions) (*AssumeServiceAccountPolicy, error) {
	options.ID = ""

	urlPath := fmt.Sprintf("service-accounts/%s/assume-policies", url.QueryEscape(serviceAccountID))
	req, err := s.client.newRequest("POST", urlPath, &options)
	if err != nil {
		return nil, err
	}

	policy := &AssumeServiceAccountPolicy{}
	err = s.client.do(ctx, req, policy)
	if err != nil {
		return nil, err
	}

	return policy, nil
}

func (s *assumeServiceAccountPolicies) Read(ctx context.Context, serviceAccountID, policyID string) (*AssumeServiceAccountPolicy, error) {
	options := struct {
		Include string `url:"include"`
	}{
		Include: "service-account,provider",
	}

	urlPath := fmt.Sprintf("service-accounts/%s/assume-policies/%s", url.QueryEscape(serviceAccountID), url.QueryEscape(policyID))
	req, err := s.client.newRequest("GET", urlPath, options)
	if err != nil {
		return nil, err
	}

	policy := &AssumeServiceAccountPolicy{}
	err = s.client.do(ctx, req, policy)
	if err != nil {
		return nil, err
	}

	return policy, nil
}

func (s *assumeServiceAccountPolicies) Update(ctx context.Context, serviceAccountID, policyID string, options AssumeServiceAccountPolicyUpdateOptions) (*AssumeServiceAccountPolicy, error) {
	options.ID = ""

	urlPath := fmt.Sprintf("service-accounts/%s/assume-policies/%s", url.QueryEscape(serviceAccountID), url.QueryEscape(policyID))
	req, err := s.client.newRequest("PATCH", urlPath, &options)
	if err != nil {
		return nil, err
	}

	policy := &AssumeServiceAccountPolicy{}
	err = s.client.do(ctx, req, policy)
	if err != nil {
		return nil, err
	}

	return policy, nil
}

func (s *assumeServiceAccountPolicies) Delete(ctx context.Context, serviceAccountID, policyID string) error {
	if !validStringID(&serviceAccountID) {
		return errors.New("invalid value for workload identity provider ID")
	}

	if !validStringID(&policyID) {
		return errors.New("invalid value for workload identity provider ID")
	}

	urlPath := fmt.Sprintf("service-accounts/%s/assume-policies/%s", url.QueryEscape(serviceAccountID), url.QueryEscape(policyID))
	req, err := s.client.newRequest("DELETE", urlPath, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
