package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Hooks describes all the hook related methods that the Scalr API supports
type Hooks interface {
	List(ctx context.Context, options HookListOptions) (*HookList, error)
	Create(ctx context.Context, options HookCreateOptions) (*Hook, error)
	Read(ctx context.Context, id string) (*Hook, error)
	Update(ctx context.Context, id string, options HookUpdateOptions) (*Hook, error)
	Delete(ctx context.Context, id string) error
}

// hooks implements Hooks
type hooks struct {
	client *Client
}

// Hook represents a Scalr hook
type Hook struct {
	ID             string       `jsonapi:"primary,hooks"`
	Name           string       `jsonapi:"attr,name"`
	Description    string       `jsonapi:"attr,description,omitempty"`
	Interpreter    string       `jsonapi:"attr,interpreter,omitempty"`
	ScriptfilePath string       `jsonapi:"attr,scriptfile-path,omitempty"`
	VcsRepo        *HookVcsRepo `jsonapi:"attr,vcs-repo,omitempty"`

	// Relations
	VcsProvider *VcsProvider `jsonapi:"relation,vcs-provider,omitempty"`
	Account     *Account     `jsonapi:"relation,account"`
}

// HookVcsRepo represents a repository in a VCS provider
type HookVcsRepo struct {
	Identifier string `json:"identifier,omitempty"`
	Branch     string `json:"branch,omitempty"`
}

// HookList represents a list of hooks
type HookList struct {
	*Pagination
	Items []*Hook
}

// HookListOptions represents the options for listing hooks
type HookListOptions struct {
	ListOptions

	Account string `url:"filter[account],omitempty"`
	Name    string `url:"filter[name],omitempty"`
	Events  string `url:"filter[events],omitempty"`
	Query   string `url:"query,omitempty"`
	Sort    string `url:"sort,omitempty"`
	Include string `url:"include,omitempty"`
}

// HookCreateOptions represents the options for creating a hook
type HookCreateOptions struct {
	ID             string       `jsonapi:"primary,hooks"`
	Name           *string      `jsonapi:"attr,name"`
	Description    *string      `jsonapi:"attr,description,omitempty"`
	Interpreter    *string      `jsonapi:"attr,interpreter,omitempty"`
	ScriptfilePath *string      `jsonapi:"attr,scriptfile-path,omitempty"`
	VcsRepo        *HookVcsRepo `jsonapi:"attr,vcs-repo,omitempty"`

	// Relations
	Account     *Account     `jsonapi:"relation,account"`
	VcsProvider *VcsProvider `jsonapi:"relation,vcs-provider,omitempty"`
}

// HookUpdateOptions represents the options for updating a hook
type HookUpdateOptions struct {
	ID             string       `jsonapi:"primary,hooks"`
	Name           *string      `jsonapi:"attr,name,omitempty"`
	Description    *string      `jsonapi:"attr,description,omitempty"`
	Interpreter    *string      `jsonapi:"attr,interpreter,omitempty"`
	ScriptfilePath *string      `jsonapi:"attr,scriptfile-path,omitempty"`
	VcsRepo        *HookVcsRepo `jsonapi:"attr,vcs-repo,omitempty"`

	// Relations
	VcsProvider *VcsProvider `jsonapi:"relation,vcs-provider,omitempty"`
}

// List lists all hooks based on the provided options
func (s *hooks) List(ctx context.Context, options HookListOptions) (*HookList, error) {
	req, err := s.client.newRequest("GET", "hooks", &options)
	if err != nil {
		return nil, err
	}

	hookList := &HookList{}
	err = s.client.do(ctx, req, hookList)
	if err != nil {
		return nil, err
	}

	return hookList, nil
}

// Create creates a new hook
func (s *hooks) Create(ctx context.Context, options HookCreateOptions) (*Hook, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}

	req, err := s.client.newRequest("POST", "hooks", &options)
	if err != nil {
		return nil, err
	}

	hook := &Hook{}
	err = s.client.do(ctx, req, hook)
	if err != nil {
		return nil, err
	}

	return hook, nil
}

// Read reads a hook by its ID
func (s *hooks) Read(ctx context.Context, id string) (*Hook, error) {
	if !validStringID(&id) {
		return nil, errors.New("invalid value for Hook ID")
	}

	u := fmt.Sprintf("hooks/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	hook := &Hook{}
	err = s.client.do(ctx, req, hook)
	if err != nil {
		return nil, err
	}

	return hook, nil
}

// Update updates a hook by its ID
func (s *hooks) Update(ctx context.Context, id string, options HookUpdateOptions) (*Hook, error) {
	if !validStringID(&id) {
		return nil, errors.New("invalid value for Hook ID")
	}

	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("hooks/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	hook := &Hook{}
	err = s.client.do(ctx, req, hook)
	if err != nil {
		return nil, err
	}

	return hook, nil
}

// Delete deletes a hook by its ID
func (s *hooks) Delete(ctx context.Context, id string) error {
	if !validStringID(&id) {
		return errors.New("invalid value for Hook ID")
	}

	u := fmt.Sprintf("hooks/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

func (o HookCreateOptions) valid() error {
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
	if o.VcsRepo == nil {
		return errors.New("vcs repo is required")
	}

	if o.Name == nil {
		return errors.New("name is required")
	}

	if o.Interpreter == nil {
		return errors.New("interpreter is required")
	}

	if o.ScriptfilePath == nil {
		return errors.New("scriptfile path is required")
	}

	return nil
}

// valid validates the hook update options
func (o HookUpdateOptions) valid() error {
	return nil
}
