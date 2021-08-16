package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ Modules = (*modules)(nil)

// Modules describes all the module related methods that the Scalr API supports.
type Modules interface {
	// List all the modules .
	List(ctx context.Context, options ModuleListOptions) (*ModuleList, error)

	// Create the module
	Create(ctx context.Context, options ModuleCreateOptions) (*Module, error)

	// Read a module by its ID.
	Read(ctx context.Context, moduleID string) (*Module, error)

	// Delete a module by its ID.
	Delete(ctx context.Context, moduleID string) error
}

// runs implements Runs.
type modules struct {
	client *Client
}

type Module struct {
	ID       string         `jsonapi:"primary,modules"`
	Name     string         `jsonapi:"attr,name"`
	Provider string         `jsonapi:"attr,provider"`
	VCSRepo  *ModuleVCSRepo `jsonapi:"attr,vcs-repo"`
	Status   ModuleStatus   `jsonapi:"attr,status"`
	Source   string         `jsonapi:"attr,source"`
	// Relation
	VcsProvider *VcsProviderOptions `jsonapi:"relation,vcs-provider"`
	Account     *Account            `jsonapi:"relation,account,omitempty"`
	Environment *Environment        `jsonapi:"relation,environment,omitempty"`
}

// ModuleStatus represents a module state.
type ModuleStatus string

//List all available run statuses.
const (
	ModuleNoVersionTags ModuleStatus = "no_version_tag"
	ModulePending       ModuleStatus = "pending"
	ModuleSetupComplete ModuleStatus = "setup_complete"
	ModuleErrored       ModuleStatus = "errored"
)

// ModuleVCSRepo contains the configuration of a VCS integration.
type ModuleVCSRepo struct {
	Identifier string  `json:"identifier"`
	Path       *string `json:"path"`
	TagPrefix  *string `json:"tag-prefix,omitempty"`
}

// ModuleList represents a list of module.
type ModuleList struct {
	*Pagination
	Items []*Module
}

// ModuleListOptions represents the options for listing modules.
type ModuleListOptions struct {
	ListOptions
	Name        *string       `url:"filter[name],omitempty"`
	Status      *ModuleStatus `url:"filter[status],omitempty"`
	Provider    *string       `url:"filter[provider],omitempty"`
	Account     *string       `url:"filter[account],omitempty"`
	Environment *string       `url:"filter[environment],omitempty"`
	//Include     string  `url:"include,omitempty"`
}

// List all the modules
func (s *modules) List(ctx context.Context, options ModuleListOptions) (*ModuleList, error) {
	req, err := s.client.newRequest("GET", "modules", &options)
	if err != nil {
		return nil, err
	}

	ml := &ModuleList{}
	err = s.client.do(ctx, req, ml)
	if err != nil {
		return nil, err
	}

	return ml, nil
}

type ModuleCreateOptions struct {
	//// For internal use only!
	ID string `jsonapi:"primary,modules"`

	// Settings for the module VCS repository.
	VCSRepo *ModuleVCSRepo `jsonapi:"attr,vcs-repo"`

	// Specifies the VcsProvider for module vcs-repo.
	VcsProvider *VcsProviderOptions `jsonapi:"relation,vcs-provider"`

	// Specifies the Account for module
	Account *Account `jsonapi:"relation,account,omitempty"`

	// Specifies the Environment for module
	Environment *Environment `jsonapi:"relation,environment,omitempty"`
}

func (o ModuleCreateOptions) valid() error {
	if o.VCSRepo == nil {
		return errors.New("vcs repo is required")
	}

	if o.VcsProvider == nil {
		return errors.New("vcs provider is required")
	}

	return nil
}

// Create is used to create a new module.
func (s *modules) Create(ctx context.Context, options ModuleCreateOptions) (*Module, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}
	//// Make sure we don't send a user provided ID.
	options.ID = ""

	req, err := s.client.newRequest("POST", "modules", &options)
	if err != nil {
		return nil, err
	}

	m := &Module{}
	err = s.client.do(ctx, req, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *modules) Read(ctx context.Context, moduleID string) (*Module, error) {
	if !validStringID(&moduleID) {
		return nil, errors.New("invalid value for module ID")
	}

	u := fmt.Sprintf("modules/%s", url.QueryEscape(moduleID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	m := &Module{}
	err = s.client.do(ctx, req, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Delete deletes a module by its ID.
func (s *modules) Delete(ctx context.Context, moduleID string) error {
	if !validStringID(&moduleID) {
		return errors.New("invalid value for module ID")
	}

	u := fmt.Sprintf("modules/%s", url.QueryEscape(moduleID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
