package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Compile-time proof of interface implementation.
var _ WebhookIntegrations = (*webhookIntegrations)(nil)

type WebhookIntegrations interface {
	List(ctx context.Context, options WebhookIntegrationListOptions) (*WebhookIntegrationList, error)
	Create(ctx context.Context, options WebhookIntegrationCreateOptions) (*WebhookIntegration, error)
	Read(ctx context.Context, wi string) (*WebhookIntegration, error)
	Update(ctx context.Context, wi string, options WebhookIntegrationUpdateOptions) (*WebhookIntegration, error)
	Delete(ctx context.Context, wi string) error
}

// webhookIntegrations implements WebhookIntegrations.
type webhookIntegrations struct {
	client *Client
}

type WebhookIntegrationList struct {
	*Pagination
	Items []*WebhookIntegration
}

// WebhookIntegration represents a Scalr IACP webhook integration.
type WebhookIntegration struct {
	ID              string           `jsonapi:"primary,webhook-integrations"`
	Name            string           `jsonapi:"attr,name"`
	Enabled         bool             `jsonapi:"attr,enabled"`
	IsShared        bool             `jsonapi:"attr,is-shared"`
	LastTriggeredAt *time.Time       `jsonapi:"attr,last-triggered-at,iso8601"`
	Url             string           `jsonapi:"attr,url"`
	SecretKey       string           `jsonapi:"attr,secret-key"`
	Timeout         int              `jsonapi:"attr,timeout"`
	MaxAttempts     int              `jsonapi:"attr,max-attempts"`
	HttpMethod      string           `jsonapi:"attr,http-method"`
	Headers         []*WebhookHeader `jsonapi:"attr,headers"`

	// Relations
	Environments []*Environment     `jsonapi:"relation,environments"`
	Account      *Account           `jsonapi:"relation,account"`
	Events       []*EventDefinition `jsonapi:"relation,events"`
}

type WebhookHeader struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Sensitive bool   `json:"sensitive"`
}

type EventDefinition struct {
	ID string `jsonapi:"primary,event-definitions"`
}

type WebhookIntegrationListOptions struct {
	ListOptions

	Query       *string `url:"query,omitempty"`
	Sort        *string `url:"sort,omitempty"`
	Enabled     *bool   `url:"filter[enabled],omitempty"`
	Event       *string `url:"filter[event],omitempty"`
	Environment *string `url:"filter[environment],omitempty"`
	Account     *string `url:"filter[account],omitempty"`
}

type WebhookIntegrationCreateOptions struct {
	ID       string  `jsonapi:"primary,webhook-integrations"`
	Name     *string `jsonapi:"attr,name"`
	Enabled  *bool   `jsonapi:"attr,enabled,omitempty"`
	IsShared *bool   `jsonapi:"attr,is-shared,omitempty"`

	Url         *string          `jsonapi:"attr,url"`
	SecretKey   *string          `jsonapi:"attr,secret-key,omitempty"`
	Timeout     *int             `jsonapi:"attr,timeout,omitempty"`
	MaxAttempts *int             `jsonapi:"attr,max-attempts,omitempty"`
	Headers     []*WebhookHeader `jsonapi:"attr,headers,omitempty"`

	Environments []*Environment     `jsonapi:"relation,environments,omitempty"`
	Account      *Account           `jsonapi:"relation,account"`
	Events       []*EventDefinition `jsonapi:"relation,events,omitempty"`
}

type WebhookIntegrationUpdateOptions struct {
	ID       string  `jsonapi:"primary,webhook-integrations"`
	Name     *string `jsonapi:"attr,name,omitempty"`
	Enabled  *bool   `jsonapi:"attr,enabled,omitempty"`
	IsShared *bool   `jsonapi:"attr,is-shared,omitempty"`

	Url         *string          `jsonapi:"attr,url,omitempty"`
	SecretKey   *string          `jsonapi:"attr,secret-key,omitempty"`
	Timeout     *int             `jsonapi:"attr,timeout,omitempty"`
	MaxAttempts *int             `jsonapi:"attr,max-attempts,omitempty"`
	Headers     []*WebhookHeader `jsonapi:"attr,headers,omitempty"`

	Environments []*Environment     `jsonapi:"relation,environments"`
	Events       []*EventDefinition `jsonapi:"relation,events"`
}

func (s *webhookIntegrations) List(
	ctx context.Context, options WebhookIntegrationListOptions,
) (*WebhookIntegrationList, error) {
	req, err := s.client.newRequest("GET", "integrations/webhooks", &options)
	if err != nil {
		return nil, err
	}

	wl := &WebhookIntegrationList{}
	err = s.client.do(ctx, req, wl)
	if err != nil {
		return nil, err
	}

	return wl, nil
}

func (s *webhookIntegrations) Create(
	ctx context.Context, options WebhookIntegrationCreateOptions,
) (*WebhookIntegration, error) {
	// Make sure we don't send a user provided ID.
	options.ID = ""

	req, err := s.client.newRequest("POST", "integrations/webhooks", &options)
	if err != nil {
		return nil, err
	}

	w := &WebhookIntegration{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (s *webhookIntegrations) Read(ctx context.Context, wi string) (*WebhookIntegration, error) {
	if !validStringID(&wi) {
		return nil, errors.New("invalid value for webhook ID")
	}

	u := fmt.Sprintf("integrations/webhooks/%s", url.QueryEscape(wi))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	w := &WebhookIntegration{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (s *webhookIntegrations) Update(
	ctx context.Context, wi string, options WebhookIntegrationUpdateOptions,
) (*WebhookIntegration, error) {
	if !validStringID(&wi) {
		return nil, errors.New("invalid value for webhook ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("integrations/webhooks/%s", url.QueryEscape(wi))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	w := &WebhookIntegration{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (s *webhookIntegrations) Delete(ctx context.Context, wi string) error {
	if !validStringID(&wi) {
		return errors.New("invalid value for webhook ID")
	}

	u := fmt.Sprintf("integrations/webhooks/%s", url.QueryEscape(wi))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
