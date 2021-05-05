package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ Teams = (*teams)(nil)

// Teams describes all the team related methods that the
// Scalr IACP API supports.
type Teams interface {
	List(ctx context.Context) (*TeamList, error)
	Read(ctx context.Context, teamID string) (*Team, error)
	Create(ctx context.Context, options TeamCreateOptions) (*Team, error)
	Update(ctx context.Context, teamID string, options TeamUpdateOptions) (*Team, error)
	Delete(ctx context.Context, teamID string) error
}

// teams implements Teams.
type teams struct {
	client *Client
}

// TeamList represents a list of teams.
type TeamList struct {
	*Pagination
	Items []*Team
}

// Team represents a Scalr team.
type Team struct {
	ID          string `jsonapi:"primary,teams"`
	Name        string `jsonapi:"attr,name"`
	Description string `jsonapi:"attr,description"`

	Account          *Account          `jsonapi:"relation,account"`
	IdentityProvider *IdentityProvider `jsonapi:"relation,identity-provider"`
	Users            []*User           `jsonapi:"relation,users"`
}

// TeamCreateOptions represents the options for creating a new Team.
type TeamCreateOptions struct {
	ID          string  `jsonapi:"primary,teams"`
	Name        *string `jsonapi:"attr,name"`
	Description *string `jsonapi:"attr,description,omitempty"`

	// Relations
	IdentityProvider *IdentityProvider `jsonapi:"relation,identity-provider"`
	Account          *Account          `jsonapi:"relation,account,omitempty"`
	Users            []*User           `jsonapi:"relation,users,omitempty"`
}

func (o TeamCreateOptions) valid() error {
	if o.IdentityProvider == nil {
		return errors.New("identity provider is required")
	}
	if !validStringID(&o.IdentityProvider.ID) {
		return errors.New("invalid value for identity provider ID")
	}
	if o.Account == nil {
		return errors.New("account is required")
	}
	if o.Account != nil && !validStringID(&o.Account.ID) {
		return errors.New("invalid value for account ID")
	}

	if o.Name == nil {
		return errors.New("name is required")
	}

	for i, usr := range o.Users {
		if usr != nil && !validStringID(&usr.ID) {
			return errors.New(fmt.Sprintf("invalid value for user ID: %v (idx: %v)", usr.ID, i))
		}
	}
	return nil
}

// Create is used to create a new Team.
func (s *teams) Create(ctx context.Context, options TeamCreateOptions) (*Team, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}
	// Make sure we don't send an team provided ID.
	options.ID = ""
	req, err := s.client.newRequest("POST", "teams", &options)
	if err != nil {
		return nil, err
	}

	team := &Team{}
	err = s.client.do(ctx, req, team)
	if err != nil {
		return nil, err
	}

	return team, nil
}

// List all the teams.
func (s *teams) List(ctx context.Context) (*TeamList, error) {

	options := struct {
		Include string `url:"include"`
	}{
		Include: "users",
	}
	req, err := s.client.newRequest("GET", "teams", &options)
	if err != nil {
		return nil, err
	}

	tl := &TeamList{}
	err = s.client.do(ctx, req, tl)
	if err != nil {
		return nil, err
	}

	return tl, nil
}

// Read an team by its ID.
func (s *teams) Read(ctx context.Context, teamID string) (*Team, error) {
	if !validStringID(&teamID) {
		return nil, errors.New("invalid value for team ID")
	}
	options := struct {
		Include string `url:"include"`
	}{
		Include: "users",
	}

	q := fmt.Sprintf("teams/%s", url.QueryEscape(teamID))
	req, err := s.client.newRequest("GET", q, &options)
	if err != nil {
		return nil, err
	}

	team := &Team{}
	err = s.client.do(ctx, req, team)
	if err != nil {
		return nil, err
	}

	return team, nil
}

// TeamUpdateOptions represents the options for updating a team.
type TeamUpdateOptions struct {
	ID          string  `jsonapi:"primary,teams"`
	Name        *string `jsonapi:"attr,name,omitempty"`
	Description *string `jsonapi:"attr,description,omitempty"`

	// Relations
	Users []*User `jsonapi:"relation,users,omitempty"`
}

// Update settings of an existing team.
func (s *teams) Update(ctx context.Context, teamID string, options TeamUpdateOptions) (*Team, error) {
	// Make sure we don't send a team provided ID.
	options.ID = ""

	u := fmt.Sprintf("teams/%s", url.QueryEscape(teamID))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	team := &Team{}
	err = s.client.do(ctx, req, team)
	if err != nil {
		return nil, err
	}

	return team, nil
}

// Delete an team by its ID.
func (s *teams) Delete(ctx context.Context, teamID string) error {
	if !validStringID(&teamID) {
		return errors.New("invalid value for team ID")
	}

	u := fmt.Sprintf("teams/%s", url.QueryEscape(teamID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
