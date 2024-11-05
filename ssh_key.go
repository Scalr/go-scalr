package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ SSHKeys = (*sshKeys)(nil)

// SSHKeys describes all the SSH keys related methods that the Scalr API supports.
type SSHKeys interface {
	List(ctx context.Context, options SSHKeysListOptions) (*SSHKeysList, error)
	Create(ctx context.Context, options SSHKeyCreateOptions) (*SSHKey, error)
	Read(ctx context.Context, sshKeyID string) (*SSHKey, error)
	Update(ctx context.Context, sshKeyID string, options SSHKeyUpdateOptions) (*SSHKey, error)
	Delete(ctx context.Context, sshKeyID string) error
}

// sshKeys implements SSHKeys.
type sshKeys struct {
	client *Client
}

// SSHKeysList represents a list of SSH keys.
type SSHKeysList struct {
	*Pagination
	Items []*SSHKey
}

// SSHKey represents a Scalr SSH key.
type SSHKey struct {
	ID           string         `jsonapi:"primary,account-ssh-keys"`
	Name         string         `jsonapi:"attr,name"`
	PrivateKey   string         `jsonapi:"attr,private_key,omitempty"`
	IsShared     bool           `jsonapi:"attr,is-shared"`
	Account      *Account       `jsonapi:"relation,account"`
	Environments []*Environment `jsonapi:"relation,environments"`
}

// SSHKeysListOptions represents the options for listing SSH keys.
type SSHKeysListOptions struct {
	ListOptions
	Sort    string        `url:"sort,omitempty"`
	Include string        `url:"include,omitempty"`
	Filter  *SSHKeyFilter `url:"filter,omitempty"`
}

// SSHKeyFilter represents the options for filtering SSH keys.
type SSHKeyFilter struct {
	Name      string `url:"name,omitempty"`
	AccountID string `url:"account,omitempty"`
}

// SSHKeyCreateOptions represents the options for creating a new SSH key.
type SSHKeyCreateOptions struct {
	ID           string         `jsonapi:"primary,account-ssh-keys"`
	Name         *string        `jsonapi:"attr,name"`
	PrivateKey   *string        `jsonapi:"attr,private_key"`
	IsShared     *bool          `jsonapi:"attr,is_shared,omitempty"`
	Account      *Account       `jsonapi:"relation,account"`
	Environments []*Environment `jsonapi:"relation,environments,omitempty"`
}

// Create is used to create a new SSH key.
func (s *sshKeys) Create(ctx context.Context, options SSHKeyCreateOptions) (*SSHKey, error) {
	options.ID = ""

	req, err := s.client.newRequest("POST", "ssh-keys", &options)
	if err != nil {
		return nil, err
	}

	sshKey := &SSHKey{}
	err = s.client.do(ctx, req, sshKey)
	if err != nil {
		return nil, err
	}

	return sshKey, nil
}

// Read an SSH key by its ID.
func (s *sshKeys) Read(ctx context.Context, sshKeyID string) (*SSHKey, error) {
	if !validStringID(&sshKeyID) {
		return nil, errors.New("invalid value for SSH key ID")
	}
	urlPath := fmt.Sprintf("ssh-keys/%s", url.QueryEscape(sshKeyID))
	req, err := s.client.newRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}

	sshKey := &SSHKey{}
	err = s.client.do(ctx, req, sshKey)
	if err != nil {
		return nil, err
	}

	return sshKey, nil
}

// SSHKeyUpdateOptions represents the options for updating an existing SSH key.
type SSHKeyUpdateOptions struct {
	ID           string         `jsonapi:"primary,account-ssh-keys"`
	Name         *string        `jsonapi:"attr,name,omitempty"`
	PrivateKey   *string        `jsonapi:"attr,private_key,omitempty"`
	IsShared     *bool          `jsonapi:"attr,is-shared,omitempty"`
	Environments []*Environment `jsonapi:"relation,environments"`
}

// Update an existing SSH key.
func (s *sshKeys) Update(ctx context.Context, sshKeyID string, options SSHKeyUpdateOptions) (*SSHKey, error) {
	if !validStringID(&sshKeyID) {
		return nil, errors.New("invalid value for SSH key ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	urlPath := fmt.Sprintf("ssh-keys/%s", url.QueryEscape(sshKeyID))
	req, err := s.client.newRequest("PATCH", urlPath, &options)
	if err != nil {
		return nil, err
	}

	sshKey := &SSHKey{}
	err = s.client.do(ctx, req, sshKey)
	if err != nil {
		return nil, err
	}

	return sshKey, nil
}

// Delete deletes an SSH key by its ID.
func (s *sshKeys) Delete(ctx context.Context, sshKeyID string) error {
	if !validStringID(&sshKeyID) {
		return errors.New("invalid value for SSH key ID")
	}

	urlPath := fmt.Sprintf("ssh-keys/%s", url.QueryEscape(sshKeyID))
	req, err := s.client.newRequest("DELETE", urlPath, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// List all SSH keys within a scalr account.
func (s *sshKeys) List(ctx context.Context, options SSHKeysListOptions) (*SSHKeysList, error) {
	req, err := s.client.newRequest("GET", "ssh-keys", &options)
	if err != nil {
		return nil, err
	}

	sshKeysList := &SSHKeysList{}
	err = s.client.do(ctx, req, sshKeysList)
	if err != nil {
		return nil, err
	}

	return sshKeysList, nil
}
