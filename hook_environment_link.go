package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ HookEnvironmentLinks = (*hookEnvironmentLinks)(nil)

// HookEnvironmentLinks interface for hook environment link related operations
type HookEnvironmentLinks interface {
	List(ctx context.Context, options HookEnvironmentLinkListOptions) (*HookEnvironmentLinkList, error)
	Create(ctx context.Context, options HookEnvironmentLinkCreateOptions) (*HookEnvironmentLink, error)
	Read(ctx context.Context, id string) (*HookEnvironmentLink, error)
	Update(ctx context.Context, id string, options HookEnvironmentLinkUpdateOptions) (*HookEnvironmentLink, error)
	Delete(ctx context.Context, id string) error
}

// hookEnvironmentLinks implements HookEnvironmentLinks interface
type hookEnvironmentLinks struct {
	client *Client
}

// HookEnvironmentLinkList represents a list of hook environment links
type HookEnvironmentLinkList struct {
	*Pagination
	Items []*HookEnvironmentLink
}

// HookEnvironmentLink represents a Scalr hook environment link
type HookEnvironmentLink struct {
	ID     string   `jsonapi:"primary,hook-environment-links"`
	Events []string `jsonapi:"attr,events,omitempty"`

	// Relations
	Environment *Environment `jsonapi:"relation,environment"`
	Hook        *Hook        `jsonapi:"relation,hook,omitempty"`
}

// HookEnvironmentLinkListOptions represents the options for listing hook environment links
type HookEnvironmentLinkListOptions struct {
	ListOptions

	Environment *string `url:"filter[environment],omitempty"`
	Events      *string `url:"filter[events],omitempty"`
	Query       *string `url:"query,omitempty"`
	Sort        *string `url:"sort,omitempty"`
	Include     *string `url:"include,omitempty"`
}

// HookEnvironmentLinkCreateOptions represents the options for creating a hook environment link
type HookEnvironmentLinkCreateOptions struct {
	ID     string    `jsonapi:"primary,hook-environment-links"`
	Events *[]string `jsonapi:"attr,events,omitempty"`

	// Relations
	Environment *Environment `jsonapi:"relation,environment"`
	Hook        *Hook        `jsonapi:"relation,hook"`
}

// HookEnvironmentLinkUpdateOptions represents the options for updating a hook environment link
type HookEnvironmentLinkUpdateOptions struct {
	ID     string    `jsonapi:"primary,hook-environment-links"`
	Events *[]string `jsonapi:"attr,events,omitempty"`

	// Relations
	Environment *Environment `jsonapi:"relation,environment,omitempty"`
	Hook        *Hook        `jsonapi:"relation,hook,omitempty"`
}

// List lists all hook environment links based on the provided options
func (s *hookEnvironmentLinks) List(ctx context.Context, options HookEnvironmentLinkListOptions) (*HookEnvironmentLinkList, error) {
	if options.Environment == nil {
		return nil, errors.New("environment is required")
	}

	req, err := s.client.newRequest("GET", "hook-environment-links", &options)
	if err != nil {
		return nil, err
	}

	linkList := &HookEnvironmentLinkList{}
	err = s.client.do(ctx, req, linkList)
	if err != nil {
		return nil, err
	}

	return linkList, nil
}

// Create creates a new hook environment link
func (s *hookEnvironmentLinks) Create(ctx context.Context, options HookEnvironmentLinkCreateOptions) (*HookEnvironmentLink, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}

	// Make sure we don't send a user provided ID
	options.ID = ""

	req, err := s.client.newRequest("POST", "hook-environment-links", &options)
	if err != nil {
		return nil, err
	}

	link := &HookEnvironmentLink{}
	err = s.client.do(ctx, req, link)
	if err != nil {
		return nil, err
	}

	return link, nil
}

// Read reads a hook environment link by its ID
func (s *hookEnvironmentLinks) Read(ctx context.Context, id string) (*HookEnvironmentLink, error) {
	if !validStringID(&id) {
		return nil, errors.New("invalid value for Hook Environment Link ID")
	}

	u := fmt.Sprintf("hook-environment-links/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	link := &HookEnvironmentLink{}
	err = s.client.do(ctx, req, link)
	if err != nil {
		return nil, err
	}

	return link, nil
}

// Update updates a hook environment link by its ID
func (s *hookEnvironmentLinks) Update(ctx context.Context, id string, options HookEnvironmentLinkUpdateOptions) (*HookEnvironmentLink, error) {
	if !validStringID(&id) {
		return nil, errors.New("invalid value for Hook Environment Link ID")
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

	link := &HookEnvironmentLink{}
	err = s.client.do(ctx, req, link)
	if err != nil {
		return nil, err
	}

	return link, nil
}

// Delete deletes a hook environment link by its ID
func (s *hookEnvironmentLinks) Delete(ctx context.Context, id string) error {
	if !validStringID(&id) {
		return errors.New("invalid value for Hook Environment Link ID")
	}

	u := fmt.Sprintf("hook-environment-links/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// valid validates the hook environment link create options
func (o HookEnvironmentLinkCreateOptions) valid() error {
	if o.Environment == nil {
		return errors.New("environment is required")
	}
	if o.Hook == nil {
		return errors.New("hook is required")
	}
	return nil
}

// valid validates the hook environment link update options
func (o HookEnvironmentLinkUpdateOptions) valid() error {
	return nil
}
