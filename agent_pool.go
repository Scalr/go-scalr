package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Compile-time proof of interface implementation.
var _ AgentPools = (*agentPools)(nil)

// AgentPools describes all the Agent Pool related methods that the
// Scalr IACP API supports.
type AgentPools interface {
	List(ctx context.Context, options AgentPoolListOptions) (*AgentPoolList, error)
	Read(ctx context.Context, agentPoolID string) (*AgentPool, error)
	Create(ctx context.Context, options AgentPoolCreateOptions) (*AgentPool, error)
	Update(ctx context.Context, agentPoolID string, options AgentPoolUpdateOptions) (*AgentPool, error)
	Delete(ctx context.Context, agentPoolID string) error
}

// agentPools implements AgentPools.
type agentPools struct {
	client *Client
}

// AgentPoolList represents a list of agent pools.
type AgentPoolList struct {
	*Pagination
	Items []*AgentPool
}

// AgentPool represents a Scalr agent pool.
type AgentPool struct {
	ID         string `jsonapi:"primary,agent-pools"`
	Name       string `jsonapi:"attr,name"`
	VcsEnabled bool   `jsonapi:"attr,vcs-enabled"`
	IsShared   bool   `jsonapi:"attr,is-shared"`
	// Relations

	// The agent pool's scope
	Account     *Account     `jsonapi:"relation,account"`
	Environment *Environment `jsonapi:"relation,environment"`

	// Workspaces this pool is connected to
	Workspaces []*Workspace `jsonapi:"relation,workspaces"`
	// Connected agents
	Agents []*Agent `jsonapi:"relation,agents"`
	// Environments this pool is linked to
	Environments []*Environment `jsonapi:"relation,environments"`
}

// AgentPoolCreateOptions represents the options for creating a new AgentPool.
type AgentPoolCreateOptions struct {
	ID         string  `jsonapi:"primary,agent-pools"`
	Name       *string `jsonapi:"attr,name"`
	VcsEnabled *bool   `jsonapi:"attr,vcs-enabled,omitempty"`
	IsShared   *bool   `jsonapi:"attr,is-shared,omitempty"`

	// The agent pool's scope
	Environment *Environment `jsonapi:"relation,environment,omitempty"`

	// Workspaces this pool is connected to
	Workspaces   []*Workspace   `jsonapi:"relation,workspaces,omitempty"`
	Environments []*Environment `jsonapi:"relation,environments,omitempty"`
}

func (o AgentPoolCreateOptions) valid() error {
	if o.Environment != nil && !validStringID(&o.Environment.ID) {
		return fmt.Errorf("invalid value for environment ID: '%s'", o.Environment.ID)
	}
	if len(o.Workspaces) != 0 {
		for i, ws := range o.Workspaces {
			if !validStringID(&ws.ID) {
				return fmt.Errorf("%d: invalid value for workspace ID: '%s'", i, ws.ID)
			}
		}
	}
	if o.Name == nil {
		return errors.New("name is required")
	}
	if strings.TrimSpace(*o.Name) == "" {
		return fmt.Errorf("invalid value for agent pool name: '%s'", *o.Name)
	}
	return nil
}

// AgentPoolListOptions represents the options for listing agent pools.
type AgentPoolListOptions struct {
	ListOptions

	Account     *string `url:"filter[account],omitempty"`
	Environment *string `url:"filter[environment],omitempty"`
	Name        string  `url:"filter[name],omitempty"`
	AgentPool   string  `url:"filter[agent-pool],omitempty"`
	VcsEnabled  *bool   `url:"filter[vcs-enabled],omitempty"`
	Include     string  `url:"include,omitempty"`
}

// List all the agent pools.
func (s *agentPools) List(ctx context.Context, options AgentPoolListOptions) (*AgentPoolList, error) {
	req, err := s.client.newRequest("GET", "agent-pools", &options)
	if err != nil {
		return nil, err
	}

	apl := &AgentPoolList{}
	err = s.client.do(ctx, req, apl)
	if err != nil {
		return nil, err
	}

	return apl, nil
}

// Create is used to create a new AgentPool.
func (s *agentPools) Create(ctx context.Context, options AgentPoolCreateOptions) (*AgentPool, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}
	// Make sure we don't send a user provided ID.
	options.ID = ""
	req, err := s.client.newRequest("POST", "agent-pools", &options)
	if err != nil {
		return nil, err
	}

	agentPool := &AgentPool{}
	err = s.client.do(ctx, req, agentPool)
	if err != nil {
		return nil, err
	}

	return agentPool, nil
}

// Read an agent pool by its ID.
func (s *agentPools) Read(ctx context.Context, agentPoolID string) (*AgentPool, error) {
	if !validStringID(&agentPoolID) {
		return nil, fmt.Errorf("invalid value for agent pool ID: '%s'", agentPoolID)
	}

	u := fmt.Sprintf("agent-pools/%s", url.QueryEscape(agentPoolID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	agentPool := &AgentPool{}
	err = s.client.do(ctx, req, agentPool)
	if err != nil {
		return nil, err
	}

	return agentPool, nil
}

// AgentPoolUpdateOptions represents the options for updating an agent pool.
type AgentPoolUpdateOptions struct {
	ID       string  `jsonapi:"primary,agent-pools"`
	Name     *string `jsonapi:"attr,name,omitempty"`
	IsShared *bool   `jsonapi:"attr,is-shared,omitempty"`

	// Workspaces this pool is connected to
	Workspaces   []*Workspace   `jsonapi:"relation,workspaces"`
	Environments []*Environment `jsonapi:"relation,environments"`
}

// Update settings of an existing agent pool.
func (s *agentPools) Update(ctx context.Context, agentPoolID string, options AgentPoolUpdateOptions) (*AgentPool, error) {
	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("agent-pools/%s", url.QueryEscape(agentPoolID))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	agentPool := &AgentPool{}
	err = s.client.do(ctx, req, agentPool)
	if err != nil {
		return nil, err
	}

	return agentPool, nil
}

// Delete an agent pool by its ID.
func (s *agentPools) Delete(ctx context.Context, agentPoolID string) error {
	if !validStringID(&agentPoolID) {
		return errors.New("invalid value for agent pool ID")
	}

	u := fmt.Sprintf("agent-pools/%s", url.QueryEscape(agentPoolID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
