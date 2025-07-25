package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Compile-time proof of interface implementation.
var _ AccessTokens = (*accessTokens)(nil)

// AccessTokens describes all the access token related methods that the
// Scalr IACP API supports.
type AccessTokens interface {
	Read(ctx context.Context, accessTokenID string) (*AccessToken, error)
	Update(ctx context.Context, accessTokenID string, options AccessTokenUpdateOptions) (*AccessToken, error)
	Delete(ctx context.Context, accessTokenID string) error
}

// accessTokens implements AccessTokens.
type accessTokens struct {
	client *Client
}

// AccessTokenList represents a list of access tokens.
type AccessTokenList struct {
	*Pagination
	Items []*AccessToken
}

// AccessToken represents a Scalr access token.
type AccessToken struct {
	ID          string    `jsonapi:"primary,access-tokens"`
	CreatedAt   time.Time `jsonapi:"attr,created-at,iso8601"`
	Description string    `jsonapi:"attr,description"`
	Token       string    `jsonapi:"attr,token"`
	Name        string    `jsonapi:"attr,name"`
	ExpiresIn   int       `jsonapi:"attr,expires-in"`
}

// AccessTokenListOptions represents the options for listing access tokens.
type AccessTokenListOptions struct {
	ListOptions
}

// AccessTokenCreateOptions represents the options for creating a new AccessToken.
type AccessTokenCreateOptions struct {
	// For internal use only!
	ID string `jsonapi:"primary,access-tokens"`

	Description *string `jsonapi:"attr,description,omitempty"`
	Name        *string `jsonapi:"attr,name,omitempty"`
	ExpiresIn   *int    `jsonapi:"attr,expires-in,omitempty"`
}

// AccessTokenUpdateOptions represents the options for updating an AccessToken.
type AccessTokenUpdateOptions struct {
	// For internal use only!
	ID string `jsonapi:"primary,access-tokens"`

	Description *string `jsonapi:"attr,description,omitempty"`
	Name        *string `jsonapi:"attr,name,omitempty"`
}

// Read access token by its ID
func (s *accessTokens) Read(ctx context.Context, accessTokenID string) (*AccessToken, error) {
	if !validStringID(&accessTokenID) {
		return nil, errors.New("invalid value for access token ID")
	}

	u := fmt.Sprintf("access-tokens/%s", url.QueryEscape(accessTokenID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	at := &AccessToken{}
	err = s.client.do(ctx, req, at)
	if err != nil {
		return nil, err
	}

	return at, nil
}

// Update is used to update an AccessToken.
func (s *accessTokens) Update(ctx context.Context, accessTokenID string, options AccessTokenUpdateOptions) (*AccessToken, error) {

	// Make sure we don't send a user provided ID.
	options.ID = ""

	if !validStringID(&accessTokenID) {
		return nil, fmt.Errorf("invalid value for access token ID: '%s'", accessTokenID)
	}

	req, err := s.client.newRequest("PATCH", fmt.Sprintf("access-tokens/%s", url.QueryEscape(accessTokenID)), &options)
	if err != nil {
		return nil, err
	}

	accessToken := &AccessToken{}
	err = s.client.do(ctx, req, accessToken)
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

// Delete an access token by its ID.
func (s *accessTokens) Delete(ctx context.Context, accessTokenID string) error {
	if !validStringID(&accessTokenID) {
		return fmt.Errorf("invalid value for access token ID: '%s'", accessTokenID)
	}

	t := fmt.Sprintf("access-tokens/%s", url.QueryEscape(accessTokenID))
	req, err := s.client.newRequest("DELETE", t, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
