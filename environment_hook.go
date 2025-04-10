package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ EnvironmentHooks = (*environmentHooks)(nil)

// EnvironmentHooks interface for environment hook related operations
type EnvironmentHooks interface {
	List(ctx context.Context, options EnvironmentHookListOptions) (*EnvironmentHookList, error)
	Create(ctx context.Context, options EnvironmentHookCreateOptions) (*EnvironmentHook, error)
	Read(ctx context.Context, id string) (*EnvironmentHook, error)
	Update(ctx context.Context, id string, options EnvironmentHookUpdateOptions) (*EnvironmentHook, error)
	Delete(ctx context.Context, id string) error
}

// environmentHooks implements EnvironmentHooks interface
type environmentHooks struct {
	client *Client
}

// EnvironmentHookList represents a list of environment hooks
type EnvironmentHookList struct {
	*Pagination
	Items []*EnvironmentHook
}

// EnvironmentHook represents a Scalr environment hook
type EnvironmentHook struct {
	ID     string   `jsonapi:"primary,hook-environment-links"`
	Events []string `jsonapi:"attr,events"`

	// Relations
	Environment *Environment `jsonapi:"relation,environment"`
	Hook        *Hook        `jsonapi:"relation,hook"`
}

// EnvironmentHookListOptions represents the options for listing environment hooks
type EnvironmentHookListOptions struct {
	ListOptions

	Environment *string `url:"filter[environment],omitempty"`
	Events      *string `url:"filter[events],omitempty"`
	Query       *string `url:"query,omitempty"`
	Sort        *string `url:"sort,omitempty"`
	Include     *string `url:"include,omitempty"`
}

// EnvironmentHookCreateOptions represents the options for creating an environment hook
type EnvironmentHookCreateOptions struct {
	ID     string   `jsonapi:"primary,hook-environment-links"`
	Events []string `jsonapi:"attr,events"`

	// Relations
	Environment *Environment `jsonapi:"relation,environment"`
	Hook        *Hook        `jsonapi:"relation,hook"`
}

// EnvironmentHookUpdateOptions represents the options for updating an environment hook
type EnvironmentHookUpdateOptions struct {
	ID     string    `jsonapi:"primary,hook-environment-links"`
	Events *[]string `jsonapi:"attr,events,omitempty"`
}

// List lists all environment hooks based on the provided options
func (s *environmentHooks) List(ctx context.Context, options EnvironmentHookListOptions) (*EnvironmentHookList, error) {
	if options.Environment == nil {
		return nil, errors.New("environment is required")
	}

	req, err := s.client.newRequest("GET", "hook-environment-links", &options)
	if err != nil {
		return nil, err
	}

	hookList := &EnvironmentHookList{}
	err = s.client.do(ctx, req, hookList)
	if err != nil {
		return nil, err
	}

	return hookList, nil
}

// Create creates a new environment hook
func (s *environmentHooks) Create(ctx context.Context, options EnvironmentHookCreateOptions) (*EnvironmentHook, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}

	// Make sure we don't send a user provided ID
	options.ID = ""

	req, err := s.client.newRequest("POST", "hook-environment-links", &options)
	if err != nil {
		return nil, err
	}

	hook := &EnvironmentHook{}
	err = s.client.do(ctx, req, hook)
	if err != nil {
		return nil, err
	}

	return hook, nil
}

// Read reads an environment hook by its ID
func (s *environmentHooks) Read(ctx context.Context, id string) (*EnvironmentHook, error) {
	if !validStringID(&id) {
		return nil, errors.New("invalid value for Environment Hook ID")
	}

	u := fmt.Sprintf("hook-environment-links/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	hook := &EnvironmentHook{}
	err = s.client.do(ctx, req, hook)
	if err != nil {
		return nil, err
	}

	return hook, nil
}

// Update updates an environment hook by its ID
func (s *environmentHooks) Update(ctx context.Context, id string, options EnvironmentHookUpdateOptions) (*EnvironmentHook, error) {
	if !validStringID(&id) {
		return nil, errors.New("invalid value for Environment Hook ID")
	}

	if err := options.valid(); err != nil {
		return nil, err
	}

	// Make sure we don't send a user provided ID
	options.ID = ""

	u := fmt.Sprintf("hook-environment-links/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	hook := &EnvironmentHook{}
	err = s.client.do(ctx, req, hook)
	if err != nil {
		return nil, err
	}

	return hook, nil
}

// Delete deletes an environment hook by its ID
func (s *environmentHooks) Delete(ctx context.Context, id string) error {
	if !validStringID(&id) {
		return errors.New("invalid value for Environment Hook ID")
	}

	u := fmt.Sprintf("hook-environment-links/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// valid validates the environment hook create options
func (o EnvironmentHookCreateOptions) valid() error {
	if o.Environment == nil {
		return errors.New("environment is required")
	}
	if o.Hook == nil {
		return errors.New("hook is required")
	}
	return nil
}

// valid validates the environment hook update options
func (o EnvironmentHookUpdateOptions) valid() error {
	return nil
}
