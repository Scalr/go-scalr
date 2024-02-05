package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Compile-time proof of interface implementation.
var _ SlackIntegrations = (*slackIntegrations)(nil)

// SlackIntegrations describes all the SlackIntegration related methods that the Scalr
// IACP API supports.
//
// IACP API docs: https://www.scalr.com/docs/en/latest/api/index.html
type SlackIntegrations interface {
	List(ctx context.Context, options SlackIntegrationListOptions) (*SlackIntegrationList, error)
	Create(ctx context.Context, options SlackIntegrationCreateOptions) (*SlackIntegration, error)
	Read(ctx context.Context, slackIntegration string) (*SlackIntegration, error)
	Update(ctx context.Context, slackIntegration string, options SlackIntegrationUpdateOptions) (*SlackIntegration, error)
	Delete(ctx context.Context, slackIntegration string) error
	GetConnection(ctx context.Context, accID string) (*SlackConnection, error)
}

// slackIntegrations implements SlackIntegrations.
type slackIntegrations struct {
	client *Client
}

const (
	SlackIntegrationEventRunApprovalRequired string = "run_approval_required"
	SlackIntegrationEventRunSuccess          string = "run_success"
	SlackIntegrationEventRunErrored          string = "run_errored"
)

// SlackIntegration represents a Scalr IACP slack integration.
type SlackIntegration struct {
	ID        string            `jsonapi:"primary,slack-integrations"`
	Name      string            `jsonapi:"attr,name"`
	Status    IntegrationStatus `jsonapi:"attr,status"`
	ChannelId string            `jsonapi:"attr,channel-id"`
	Events    []string          `jsonapi:"attr,events"`
	RunMode   string            `jsonapi:"attr,run-mode"`

	// Relations
	Account      *Account       `jsonapi:"relation,account"`
	Environments []*Environment `jsonapi:"relation,environments"`
	Workspaces   []*Workspace   `jsonapi:"relation,workspaces"`
}

type SlackIntegrationList struct {
	*Pagination
	Items []*SlackIntegration
}

type SlackIntegrationListOptions struct {
	ListOptions

	Filter *SlackIntegrationFilter `url:"filter,omitempty"`
}

// SlackIntegrationFilter represents the options for filtering Slack integrations.
type SlackIntegrationFilter struct {
	Account *string `url:"account,omitempty"`
}

type SlackIntegrationCreateOptions struct {
	ID        string   `jsonapi:"primary,slack-integrations"`
	Name      *string  `jsonapi:"attr,name"`
	ChannelId *string  `jsonapi:"attr,channel-id"`
	Events    []string `jsonapi:"attr,events"`
	RunMode   *string  `jsonapi:"attr,run-mode"`

	Account      *Account         `jsonapi:"relation,account"`
	Connection   *SlackConnection `jsonapi:"relation,connection"`
	Environments []*Environment   `jsonapi:"relation,environments"`
	Workspaces   []*Workspace     `jsonapi:"relation,workspaces,omitempty"`
}

type SlackIntegrationUpdateOptions struct {
	ID        string             `jsonapi:"primary,slack-integrations"`
	Name      *string            `jsonapi:"attr,name,omitempty"`
	ChannelId *string            `jsonapi:"attr,channel-id,omitempty"`
	Status    *IntegrationStatus `jsonapi:"attr,status,omitempty"`
	Events    []string           `jsonapi:"attr,events,omitempty"`
	RunMode   *string            `jsonapi:"attr,run-mode,omitempty"`

	Environments []*Environment `jsonapi:"relation,environments,omitempty"`
	Workspaces   []*Workspace   `jsonapi:"relation,workspaces"`
}

type SlackConnection struct {
	ID                 string `jsonapi:"primary,slack-connections"`
	SlackWorkspaceName string `jsonapi:"attr,slack-workspace-name"`

	// Relations
	Account *Account `jsonapi:"relation,account"`
}

func (s *slackIntegrations) List(
	ctx context.Context, options SlackIntegrationListOptions,
) (*SlackIntegrationList, error) {
	req, err := s.client.newRequest("GET", "integrations/slack", &options)
	if err != nil {
		return nil, err
	}

	wl := &SlackIntegrationList{}
	err = s.client.do(ctx, req, wl)
	if err != nil {
		return nil, err
	}

	return wl, nil
}

func (s *slackIntegrations) Create(
	ctx context.Context, options SlackIntegrationCreateOptions,
) (*SlackIntegration, error) {
	// Make sure we don't send a user provided ID.
	options.ID = ""

	req, err := s.client.newRequest("POST", "integrations/slack", &options)
	if err != nil {
		return nil, err
	}

	w := &SlackIntegration{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (s *slackIntegrations) Read(ctx context.Context, si string) (*SlackIntegration, error) {
	if !validStringID(&si) {
		return nil, errors.New("invalid value for Slack integration ID")
	}

	u := fmt.Sprintf("integrations/slack/%s", url.QueryEscape(si))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	w := &SlackIntegration{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (s *slackIntegrations) Update(
	ctx context.Context, si string, options SlackIntegrationUpdateOptions,
) (*SlackIntegration, error) {
	if !validStringID(&si) {
		return nil, errors.New("invalid value for slack integration ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("integrations/slack/%s", url.QueryEscape(si))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	w := &SlackIntegration{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (s *slackIntegrations) Delete(ctx context.Context, si string) error {
	if !validStringID(&si) {
		return errors.New("invalid value for slack integration ID")
	}

	u := fmt.Sprintf("integrations/slack/%s", url.QueryEscape(si))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

func (s *slackIntegrations) GetConnection(ctx context.Context, accID string) (*SlackConnection, error) {
	if !validStringID(&accID) {
		return nil, errors.New("invalid value for account ID")
	}

	u := fmt.Sprintf("integrations/slack/%s/connection", url.QueryEscape(accID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	c := &SlackConnection{}
	err = s.client.do(ctx, req, c)
	if err != nil {
		if strings.Contains(err.Error(), "data is not a jsonapi representation") {
			// workaround for jsonapi serializer that can't handle nil 'data' structure we use for missing connection
			return c, nil
		}
		return nil, err
	}

	return c, nil
}
