package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ PolicyGroups = (*policyGroups)(nil)

// PolicyGroups describes all the policy group related methods that the
// Scalr API supports.
type PolicyGroups interface {
	List(ctx context.Context, options PolicyGroupListOptions) (*PolicyGroupList, error)
	Read(ctx context.Context, policyGroupID string) (*PolicyGroup, error)
	Create(ctx context.Context, options PolicyGroupCreateOptions) (*PolicyGroup, error)
	Update(ctx context.Context, policyGroupID string, options PolicyGroupUpdateOptions) (*PolicyGroup, error)
	Delete(ctx context.Context, policyGroupID string) error
}

// policyGroups implements PolicyGroups.
type policyGroups struct {
	client *Client
}

// PolicyGroupStatus represents a policy group status.
type PolicyGroupStatus string

// List of available policy group statuses.
const (
	PolicyGroupStatusFetching PolicyGroupStatus = "fetching"
	PolicyGroupStatusActive   PolicyGroupStatus = "active"
	PolicyGroupStatusErrored  PolicyGroupStatus = "errored"
)

// PolicyEnforcementLevel represents enforcement level of an OPA policy.
type PolicyEnforcementLevel string

// List of available policy enforcement levels.
const (
	PolicyEnforcementLevelHard     = "hard-mandatory"
	PolicyEnforcementLevelSoft     = "soft-mandatory"
	PolicyEnforcementLevelAdvisory = "advisory"
)

// Policy represents a single OPA policy.
type Policy struct {
	ID               string                 `jsonapi:"primary,policies"`
	Name             string                 `jsonapi:"attr,name"`
	Enabled          bool                   `jsonapi:"attr,enabled"`
	EnforcementLevel PolicyEnforcementLevel `jsonapi:"attr,enforced-level"`

	// Relations
	PolicyGroup *PolicyGroup `jsonapi:"relation,policy-groups"`
}

// PolicyGroupVCSRepo contains the configuration of a VCS integration.
type PolicyGroupVCSRepo struct {
	Identifier string `json:"identifier"`
	Branch     string `json:"branch"`
	Path       string `json:"path"`
}

// PolicyGroupVCSRepoOptions contains the configuration options of a VCS integration.
type PolicyGroupVCSRepoOptions struct {
	Identifier *string `json:"identifier"`
	Branch     *string `json:"branch,omitempty"`
	Path       *string `json:"path,omitempty"`
}

// PolicyGroup represents a Scalr policy group.
type PolicyGroup struct {
	ID                    string              `jsonapi:"primary,policy-groups"`
	Name                  string              `jsonapi:"attr,name"`
	Status                PolicyGroupStatus   `jsonapi:"attr,status"`
	ErrorMessage          string              `jsonapi:"attr,error-message"`
	OpaVersion            string              `jsonapi:"attr,opa-version"`
	VCSRepo               *PolicyGroupVCSRepo `jsonapi:"attr,vcs-repo"`
	IsEnforced            bool                `jsonapi:"attr,is-enforced"`
	CommonFunctionsFolder string              `jsonapi:"attr,common-functions-folder"`

	// Relations
	Account      *Account       `jsonapi:"relation,account"`
	VcsProvider  *VcsProvider   `jsonapi:"relation,vcs-provider"`
	VcsRevision  *VcsRevision   `jsonapi:"relation,vcs-revision"`
	Policies     []*Policy      `jsonapi:"relation,policies"`
	Environments []*Environment `jsonapi:"relation,environments"`
}

// PolicyGroupList represents a list of policy groups.
type PolicyGroupList struct {
	*Pagination
	Items []*PolicyGroup
}

// PolicyGroupListOptions represents the options for listing policy groups.
type PolicyGroupListOptions struct {
	ListOptions

	Account     string `url:"filter[account],omitempty"`
	Environment string `url:"filter[environment],omitempty"`
	Name        string `url:"filter[name],omitempty"`
	PolicyGroup string `url:"filter[policy-group],omitempty"`
	Query       string `url:"query,omitempty"`
	Sort        string `url:"sort,omitempty"`
	Include     string `url:"include,omitempty"`
}

// PolicyGroupCreateOptions represents the options for creating a new PolicyGroup.
type PolicyGroupCreateOptions struct {
	ID                    string                     `jsonapi:"primary,policy-groups"`
	Name                  *string                    `jsonapi:"attr,name"`
	OpaVersion            *string                    `jsonapi:"attr,opa-version,omitempty"`
	VCSRepo               *PolicyGroupVCSRepoOptions `jsonapi:"attr,vcs-repo"`
	IsEnforced            *bool                      `jsonapi:"attr,is-enforced,omitempty"`
	CommonFunctionsFolder *string                    `jsonapi:"attr,common-functions-folder,omitempty"`

	// Relations
	Account      *Account       `jsonapi:"relation,account"`
	VcsProvider  *VcsProvider   `jsonapi:"relation,vcs-provider"`
	Environments []*Environment `jsonapi:"relation,environments"`
}

func (o PolicyGroupCreateOptions) valid() error {
	if !validString(o.Name) {
		return errors.New("name is required")
	}
	if o.Account == nil {
		return errors.New("account is required")
	}
	if !validStringID(&o.Account.ID) {
		return errors.New("invalid value for account ID")
	}
	if o.VcsProvider == nil {
		return errors.New("vcs provider is required")
	}
	if !validStringID(&o.VcsProvider.ID) {
		return errors.New("invalid value for vcs provider ID")
	}
	if o.VCSRepo == nil {
		return errors.New("vcs repo is required")
	}
	return nil
}

// PolicyGroupUpdateOptions represents the options for updating a PolicyGroup.
type PolicyGroupUpdateOptions struct {
	ID                    string                     `jsonapi:"primary,policy-groups"`
	Name                  *string                    `jsonapi:"attr,name,omitempty"`
	OpaVersion            *string                    `jsonapi:"attr,opa-version,omitempty"`
	VCSRepo               *PolicyGroupVCSRepoOptions `jsonapi:"attr,vcs-repo,omitempty"`
	IsEnforced            *bool                      `jsonapi:"attr,is-enforced,omitempty"`
	CommonFunctionsFolder *string                    `jsonapi:"attr,common-functions-folder"`

	// Relations
	VcsProvider *VcsProvider `jsonapi:"relation,vcs-provider,omitempty"`
}

// List all the policy groups.
func (s *policyGroups) List(ctx context.Context, options PolicyGroupListOptions) (*PolicyGroupList, error) {
	req, err := s.client.newRequest("GET", "policy-groups", &options)
	if err != nil {
		return nil, err
	}

	pgl := &PolicyGroupList{}
	err = s.client.do(ctx, req, pgl)
	if err != nil {
		return nil, err
	}

	return pgl, nil
}

// Create a new policy group.
func (s *policyGroups) Create(ctx context.Context, options PolicyGroupCreateOptions) (*PolicyGroup, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}
	// Make sure we don't send a user provided ID.
	options.ID = ""
	req, err := s.client.newRequest("POST", "policy-groups", &options)
	if err != nil {
		return nil, err
	}

	pg := &PolicyGroup{}
	err = s.client.do(ctx, req, pg)
	if err != nil {
		return nil, err
	}

	return pg, nil
}

// Read policy group by its ID.
func (s *policyGroups) Read(ctx context.Context, policyGroupID string) (*PolicyGroup, error) {
	if !validStringID(&policyGroupID) {
		return nil, errors.New("invalid value for policy group ID")
	}

	options := struct {
		Include string `url:"include"`
	}{
		Include: "policies",
	}
	u := fmt.Sprintf("policy-groups/%s", url.QueryEscape(policyGroupID))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	pg := &PolicyGroup{}
	err = s.client.do(ctx, req, pg)
	if err != nil {
		return nil, err
	}

	return pg, nil
}

// Update settings of existing policy group.
func (s *policyGroups) Update(ctx context.Context, policyGroupID string, options PolicyGroupUpdateOptions) (*PolicyGroup, error) {
	if !validStringID(&policyGroupID) {
		return nil, errors.New("invalid value for policy group ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("policy-groups/%s", url.QueryEscape(policyGroupID))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	pg := &PolicyGroup{}
	err = s.client.do(ctx, req, pg)
	if err != nil {
		return nil, err
	}

	return pg, nil
}

// Delete policy group by its ID.
func (s *policyGroups) Delete(ctx context.Context, policyGroupID string) error {
	if !validStringID(&policyGroupID) {
		return errors.New("invalid value for policy group ID")
	}

	u := fmt.Sprintf("policy-groups/%s", url.QueryEscape(policyGroupID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
