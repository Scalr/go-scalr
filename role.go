package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Compile-time proof of interface implementation.
var _ Roles = (*roles)(nil)

// Roles describes all the role related methods that the
// Scalr IACP API supports.
type Roles interface {
	List(ctx context.Context, options RoleListOptions) (*RoleList, error)
	Read(ctx context.Context, roleID string) (*Role, error)
	Create(ctx context.Context, options RoleCreateOptions) (*Role, error)
	Update(ctx context.Context, roleID string, options RoleUpdateOptions) (*Role, error)
	Delete(ctx context.Context, roleID string) error
}

// roles implements Roles.
type roles struct {
	client *Client
}

// Permission relationship
type Permission struct {
	ID string `jsonapi:"primary,permissions,omitempty"`
}

// RoleList represents a list of roles.
type RoleList struct {
	*Pagination
	Items []*Role
}

// Role represents a Scalr role.
type Role struct {
	ID          string `jsonapi:"primary,roles"`
	Name        string `jsonapi:"attr,name"`
	Description string `jsonapi:"attr,description"`
	IsSystem    bool   `jsonapi:"attr,is-system"`

	// Relations
	Account     *Account      `jsonapi:"relation,account"`
	Permissions []*Permission `jsonapi:"relation,permissions,omitempty"`
}

// RoleCreateOptions represents the options for creating a new Role.
type RoleCreateOptions struct {
	ID          string  `jsonapi:"primary,roles"`
	Name        *string `jsonapi:"attr,name"`
	Description *string `jsonapi:"attr,description,omitempty"`

	// Relations
	Permissions []*Permission `jsonapi:"relation,permissions,omitempty"`
}

func (o RoleCreateOptions) valid() error {
	if o.Name == nil {
		return errors.New("name is required")
	}
	if strings.TrimSpace(*o.Name) == "" {
		return errors.New("invalid value for name")
	}
	return nil
}

// RoleListOptions represents the options for listing roles.
type RoleListOptions struct {
	ListOptions

	Account *string `url:"filter[account],omitempty"`
	Name    string  `url:"filter[name],omitempty"`
	Role    string  `url:"filter[role],omitempty"`
	Query   string  `url:"query,omitempty"`
	Include string  `url:"include,omitempty"`
}

// List all the roles.
func (s *roles) List(ctx context.Context, options RoleListOptions) (*RoleList, error) {
	req, err := s.client.newRequest("GET", "roles", &options)
	if err != nil {
		return nil, err
	}

	rolel := &RoleList{}
	err = s.client.do(ctx, req, rolel)
	if err != nil {
		return nil, err
	}

	return rolel, nil
}

// Create is used to create a new Role.
func (s *roles) Create(ctx context.Context, options RoleCreateOptions) (*Role, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}
	// Make sure we don't send a user provided ID.
	options.ID = ""
	req, err := s.client.newRequest("POST", "roles", &options)
	if err != nil {
		return nil, err
	}

	role := &Role{}
	err = s.client.do(ctx, req, role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// Read an role by its ID.
func (s *roles) Read(ctx context.Context, roleID string) (*Role, error) {
	if !validStringID(&roleID) {
		return nil, errors.New("invalid value for role ID")
	}

	u := fmt.Sprintf("roles/%s", url.QueryEscape(roleID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	role := &Role{}
	err = s.client.do(ctx, req, role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// RoleUpdateOptions represents the options for updating an role.
type RoleUpdateOptions struct {
	// For internal use only!
	ID          string  `jsonapi:"primary,roles"`
	Name        *string `jsonapi:"attr,name,omitempty"`
	Description *string `jsonapi:"attr,description,omitempty"`

	// Relations
	Permissions []*Permission `jsonapi:"relation,permissions,omitempty"`
}

// Update settings of an existing role.
func (s *roles) Update(ctx context.Context, roleID string, options RoleUpdateOptions) (*Role, error) {
	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("roles/%s", url.QueryEscape(roleID))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	role := &Role{}
	err = s.client.do(ctx, req, role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// Delete an role by its ID.
func (s *roles) Delete(ctx context.Context, roleID string) error {
	if !validStringID(&roleID) {
		return errors.New("invalid value for role ID")
	}

	u := fmt.Sprintf("roles/%s", url.QueryEscape(roleID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
