package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Compile-time proof of interface implementation.
var _ ServiceAccounts = (*serviceAccounts)(nil)

// ServiceAccounts describes all the service account related methods that the Scalr API supports.
type ServiceAccounts interface {
	// List all the service accounts.
	List(ctx context.Context, options ServiceAccountListOptions) (*ServiceAccountList, error)
	// Create is used to create a new service account.
	Create(ctx context.Context, options ServiceAccountCreateOptions) (*ServiceAccount, error)
	// Read reads a service account by its ID.
	Read(ctx context.Context, serviceAccountID string) (*ServiceAccount, error)
	// Update existing service account by its ID.
	Update(ctx context.Context, serviceAccountID string, options ServiceAccountUpdateOptions) (*ServiceAccount, error)
	// Delete service account by its ID.
	Delete(ctx context.Context, serviceAccountID string) error
}

// serviceAccounts implements ServiceAccounts.
type serviceAccounts struct {
	client *Client
}

// ServiceAccountList represents a list of service accounts.
type ServiceAccountList struct {
	*Pagination
	Items []*ServiceAccount
}

// ServiceAccountStatus represents the status of service account.
type ServiceAccountStatus string

// List of available service account statuses.
const (
	ServiceAccountStatusActive   ServiceAccountStatus = "Active"
	ServiceAccountStatusInactive ServiceAccountStatus = "Inactive"
)

type ServiceAccount struct {
	ID          string               `jsonapi:"primary,service-accounts"`
	Name        string               `jsonapi:"attr,name"`
	Email       string               `jsonapi:"attr,email"`
	Description string               `jsonapi:"attr,description"`
	Status      ServiceAccountStatus `jsonapi:"attr,status"`
	CreatedAt   time.Time            `jsonapi:"attr,created-at,iso8601"`

	// Relations
	Account   *Account `jsonapi:"relation,account,omitempty"`
	CreatedBy *User    `jsonapi:"relation,created-by,omitempty"`
	Owners    []*Team  `jsonapi:"relation,owners"`
}

// ServiceAccountListOptions represents the options for listing service accounts.
type ServiceAccountListOptions struct {
	ListOptions

	Account        *string `url:"filter[account],omitempty"`
	Email          *string `url:"filter[email],omitempty"`
	ServiceAccount *string `url:"filter[service-account],omitempty"`
	Query          *string `url:"query,omitempty"`
	Include        *string `url:"include,omitempty"`
}

// ServiceAccountCreateOptions represents the options for creating a new service account.
type ServiceAccountCreateOptions struct {
	// For internal use only!
	ID string `jsonapi:"primary,service-accounts"`

	// The name of the service account, it must be unique within the account.
	Name        *string               `jsonapi:"attr,name"`
	Description *string               `jsonapi:"attr,description,omitempty"`
	Status      *ServiceAccountStatus `jsonapi:"attr,status,omitempty"`
	Account     *Account              `jsonapi:"relation,account"`
	Owners      []*Team               `jsonapi:"relation,owners"`
}

func (o ServiceAccountCreateOptions) valid() error {
	if o.Account == nil {
		return errors.New("account is required")
	}
	if !validStringID(&o.Account.ID) {
		return errors.New("invalid value for account ID")
	}
	if o.Name == nil {
		return errors.New("name is required")
	}
	return nil
}

// ServiceAccountUpdateOptions represents the options for updating a service account.
type ServiceAccountUpdateOptions struct {
	// For internal use only!
	ID string `jsonapi:"primary,service-accounts"`

	Description *string               `jsonapi:"attr,description,omitempty"`
	Status      *ServiceAccountStatus `jsonapi:"attr,status,omitempty"`
	Owners      []*Team               `jsonapi:"relation,owners"`
}

// Read a service account by its ID.
func (s *serviceAccounts) Read(ctx context.Context, serviceAccountID string) (*ServiceAccount, error) {
	if !validStringID(&serviceAccountID) {
		return nil, errors.New("invalid value for service account ID")
	}

	options := struct {
		Include string `url:"include"`
	}{
		Include: "created-by",
	}
	u := fmt.Sprintf("service-accounts/%s", url.QueryEscape(serviceAccountID))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	sa := &ServiceAccount{}
	err = s.client.do(ctx, req, sa)
	if err != nil {
		return nil, err
	}

	return sa, nil
}

// List all the service accounts.
func (s *serviceAccounts) List(ctx context.Context, options ServiceAccountListOptions) (*ServiceAccountList, error) {
	req, err := s.client.newRequest("GET", "service-accounts", &options)
	if err != nil {
		return nil, err
	}

	sal := &ServiceAccountList{}
	err = s.client.do(ctx, req, sal)
	if err != nil {
		return nil, err
	}

	return sal, nil
}

// Create is used to create a new service account.
func (s *serviceAccounts) Create(ctx context.Context, options ServiceAccountCreateOptions) (*ServiceAccount, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}
	// Make sure we don't send a user provided ID.
	options.ID = ""

	req, err := s.client.newRequest("POST", "service-accounts", &options)
	if err != nil {
		return nil, err
	}

	sa := &ServiceAccount{}
	err = s.client.do(ctx, req, sa)
	if err != nil {
		return nil, err
	}

	return sa, nil
}

// Update is used to update a service account.
func (s *serviceAccounts) Update(ctx context.Context, serviceAccountID string, options ServiceAccountUpdateOptions) (*ServiceAccount, error) {
	if !validStringID(&serviceAccountID) {
		return nil, errors.New("invalid value for service account ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("service-accounts/%s", url.QueryEscape(serviceAccountID))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	sa := &ServiceAccount{}
	err = s.client.do(ctx, req, sa)
	if err != nil {
		return nil, err
	}

	return sa, nil
}

// Delete service account by its ID.
func (s *serviceAccounts) Delete(ctx context.Context, serviceAccountID string) error {
	if !validStringID(&serviceAccountID) {
		return errors.New("invalid value for service account ID")
	}

	u := fmt.Sprintf("service-accounts/%s", url.QueryEscape(serviceAccountID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
