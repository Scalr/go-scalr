package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ Endpoints = (*endpoints)(nil)

// Endpoints describes all the endpoints related methods that the Scalr
// IACP API supports.
//
// IACP API docs: https://www.scalr.com/docs/en/latest/api/index.html
type Endpoints interface {
	// List the endpoints.
	List(ctx context.Context, options EndpointListOptions) (*EndpointList, error)
	Create(ctx context.Context, options EndpointCreateOptions) (*Endpoint, error)
	Read(ctx context.Context, endpoint string) (*Endpoint, error)
	Update(ctx context.Context, endpoint string, options EndpointUpdateOptions) (*Endpoint, error)
	Delete(ctx context.Context, endpoint string) error
}

// endpoints implements Endpoints.
type endpoints struct {
	client *Client
}

// EndpointList represents a list of endpoints.
type EndpointList struct {
	*Pagination
	Items []*Endpoint
}

// Endpoint represents a Scalr IACP endpoint.
type Endpoint struct {
	ID          string `jsonapi:"primary,endpoints"`
	MaxAttempts int    `jsonapi:"attr,max-attempts"`
	Name        string `jsonapi:"attr,name"`
	SecretKey   string `jsonapi:"attr,secret-key"`
	Timeout     int    `jsonapi:"attr,timeout"`
	Url         string `jsonapi:"attr,url"`

	// Relations
	Environment *Environment `jsonapi:"relation,environment"`
	Account     *Account     `jsonapi:"relation,account"`
}

// EndpointListOptions represents the options for listing endpoints.
type EndpointListOptions struct {
	ListOptions

	// Query string.
	Query *string `url:"query,omitempty"`

	Name *string `url:"query,omitempty"`

	// The comma-separated list of attributes.
	Sort *string `url:"sort,omitempty"`

	// Scope filters.
	Environment *string `url:"filter[environment],omitempty"`
	Account     *string `url:"filter[account],omitempty"`
}

// List the endpoints.
func (s *endpoints) List(ctx context.Context, options EndpointListOptions) (*EndpointList, error) {
	req, err := s.client.newRequest("GET", "endpoints", &options)
	if err != nil {
		return nil, err
	}

	el := &EndpointList{}
	err = s.client.do(ctx, req, el)
	if err != nil {
		return nil, err
	}

	return el, nil
}

// EndpointCreateOptions represents the options for creating a new endpoint.
type EndpointCreateOptions struct {
	// For internal use only!
	ID          string  `jsonapi:"primary,endpoints"`
	MaxAttempts *int    `jsonapi:"attr,max-attempts,omitempty"`
	Name        *string `jsonapi:"attr,name"`
	Url         *string `jsonapi:"attr,url"`
	SecretKey   *string `jsonapi:"attr,secret-key,omitempty"`
	Timeout     *int    `jsonapi:"attr,timeout,omitempty"`

	// Relations
	Environment *Environment `jsonapi:"relation,environment,omitempty"`
	Account     *Account     `jsonapi:"relation,account"`
}

func (o EndpointCreateOptions) valid() error {
	if !validString(o.Name) {
		return errors.New("name is required")
	}
	if !validString(o.Url) {
		return errors.New("url is required")
	}
	return nil
}

// Create is used to create a new endpoint.
func (s *endpoints) Create(ctx context.Context, options EndpointCreateOptions) (*Endpoint, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	req, err := s.client.newRequest("POST", "endpoints", &options)
	if err != nil {
		return nil, err
	}

	w := &Endpoint{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

// Read a endpoint by its ID.
func (s *endpoints) Read(ctx context.Context, endpointID string) (*Endpoint, error) {
	if !validStringID(&endpointID) {
		return nil, errors.New("invalid value for endpoint ID")
	}

	u := fmt.Sprintf("endpoints/%s", url.QueryEscape(endpointID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	e := &Endpoint{}
	err = s.client.do(ctx, req, e)
	if err != nil {
		return nil, err
	}

	return e, nil
}

// EndpointUpdateOptions represents the options for updating an endpoint.
type EndpointUpdateOptions struct {
	ID          string  `jsonapi:"primary,endpoints"`
	Name        *string `jsonapi:"attr,name,omitempty"`
	MaxAttempts *int    `jsonapi:"attr,max-attempts,omitempty"`
	Url         *string `jsonapi:"attr,url,omitempty"`
	SecretKey   *string `jsonapi:"attr,secret-key,omitempty"`
	Timeout     *int    `jsonapi:"attr,timeout,omitempty"`
}

// Update settings of an existing endpoint.
func (s *endpoints) Update(ctx context.Context, endpointID string, options EndpointUpdateOptions) (*Endpoint, error) {
	if !validStringID(&endpointID) {
		return nil, errors.New("invalid value for endpoint ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("endpoints/%s", url.QueryEscape(endpointID))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	e := &Endpoint{}
	err = s.client.do(ctx, req, e)
	if err != nil {
		return nil, err
	}

	return e, nil
}

// Delete an endpoint by its ID.
func (s *endpoints) Delete(ctx context.Context, endpointID string) error {
	if !validStringID(&endpointID) {
		return errors.New("invalid value for endpoint ID")
	}

	u := fmt.Sprintf("endpoints/%s", url.QueryEscape(endpointID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
