package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Compile-time proof of interface implementation.
var _ Users = (*users)(nil)

// Users describes all the user related methods that the
// Scalr IACP API supports.
type Users interface {
	List(ctx context.Context) (*UserList, error)
	Read(ctx context.Context, userID string) (*User, error)
	Create(ctx context.Context, options UserCreateOptions) (*User, error)
	Update(ctx context.Context, userID string, options UserUpdateOptions) (*User, error)
	Delete(ctx context.Context, userID string) error
}

// users implements Users.
type users struct {
	client *Client
}

// UserStatus represents an user status.
type UserStatus = string

// List of available user statuses.
const (
	UserStatusActive   UserStatus = "Active"
	UserStatusInactive UserStatus = "Inactive"
	UserStatusPending  UserStatus = "Pending"
)

// UserList represents a list of users.
type UserList struct {
	*Pagination
	Items []*User
}

// User represents a Scalr user.
type User struct {
	ID               string            `jsonapi:"primary,users"`
	Email            string            `jsonapi:"attr,email"`
	Username         string            `jsonapi:"attr,username"`
	FullName         string            `jsonapi:"attr,full-name"`
	Status           UserStatus        `jsonapi:"attr,status"`
	CreatedAt        time.Time         `jsonapi:"attr,created-at,iso8601"`
	IdentityProvider *IdentityProvider `jsonapi:"relation,identity-provider"`
}

// UserCreateOptions represents the options for creating a new User.
type UserCreateOptions struct {
	ID       string     `jsonapi:"primary,users"`
	Username *string    `jsonapi:"attr,username,omitempty"`
	Email    *string    `jsonapi:"attr,email"`
	FullName *string    `jsonapi:"attr,full-name,omitempty"`
	Password *string    `jsonapi:"attr,password,omitempty"`
	Status   UserStatus `jsonapi:"attr,status"`

	// Relations
	IdentityProvider *IdentityProvider `jsonapi:"relation,identity-provider"`
}

func (o UserCreateOptions) valid() error {
	if o.IdentityProvider == nil {
		return errors.New("identity provider is required")
	}
	if !validStringID(&o.IdentityProvider.ID) {
		return errors.New("invalid value for identity provider ID")
	}
	if o.Email == nil {
		return errors.New("email is required")
	}
	if o.Status == "" {
		return errors.New("status is required")
	}
	if !validEmail(o.Email) {
		return errors.New("invalid value for email")
	}
	return nil
}

// Create is used to create a new User.
func (s *users) Create(ctx context.Context, options UserCreateOptions) (*User, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}
	// Make sure we don't send an user provided ID.
	options.ID = ""
	req, err := s.client.newRequest("POST", "users", &options)
	if err != nil {
		return nil, err
	}

	user := &User{}
	err = s.client.do(ctx, req, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// List all the users.
func (s *users) List(ctx context.Context) (*UserList, error) {
	req, err := s.client.newRequest("GET", "users", nil)
	if err != nil {
		return nil, err
	}

	usrl := &UserList{}
	err = s.client.do(ctx, req, usrl)
	if err != nil {
		return nil, err
	}

	return usrl, nil
}

// Read an user by its ID.
func (s *users) Read(ctx context.Context, userID string) (*User, error) {
	if !validStringID(&userID) {
		return nil, errors.New("invalid value for user ID")
	}

	u := fmt.Sprintf("users/%s", url.QueryEscape(userID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	usr := &User{}
	err = s.client.do(ctx, req, usr)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

// UserUpdateOptions represents the options for updating an user.
type UserUpdateOptions struct {
	ID       string     `jsonapi:"primary,users"`
	FullName *string    `jsonapi:"attr,full-name,omitempty"`
	Status   UserStatus `jsonapi:"attr,status,omitempty"`
}

// Update settings of an existing user.
func (s *users) Update(ctx context.Context, userID string, options UserUpdateOptions) (*User, error) {
	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("users/%s", url.QueryEscape(userID))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	usr := &User{}
	err = s.client.do(ctx, req, usr)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

// Delete an user by its ID.
func (s *users) Delete(ctx context.Context, userID string) error {
	if !validStringID(&userID) {
		return errors.New("invalid value for user ID")
	}

	u := fmt.Sprintf("users/%s", url.QueryEscape(userID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
