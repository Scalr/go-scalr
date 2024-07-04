package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ EventBridgeIntegrations = (*eventBridgeIntegrations)(nil)

type EventBridgeIntegrations interface {
	List(ctx context.Context, options EventBridgeIntegrationListOptions) (*EventBridgeIntegrationList, error)
	Create(ctx context.Context, options EventBridgeIntegrationCreateOptions) (*EventBridgeIntegration, error)
	Read(ctx context.Context, wi string) (*EventBridgeIntegration, error)
	Update(ctx context.Context, wi string, options EventBridgeIntegrationUpdateOptions) (*EventBridgeIntegration, error)
	Delete(ctx context.Context, wi string) error
}

// eventBridgeIntegrations implements EventBridgeIntegrations.
type eventBridgeIntegrations struct {
	client *Client
}

type EventBridgeIntegrationList struct {
	*Pagination
	Items []*EventBridgeIntegration
}

// EventBridgeIntegration represents a Scalr IACP eventBridge integration.
type EventBridgeIntegration struct {
	ID             string            `jsonapi:"primary,aws-event-bridge-integrations"`
	Name           string            `jsonapi:"attr,name"`
	Status         IntegrationStatus `jsonapi:"attr,status"`
	EventSource    string            `jsonapi:"attr,event-source"`
	EventSourceARN string            `jsonapi:"attr,event-source-arn"`
	AWSAccountId   string            `jsonapi:"attr,aws-account-id"`
	Region         string            `jsonapi:"attr,region"`

	// Relations
	Account *Account `jsonapi:"relation,account"`
}

type EventBridgeIntegrationListOptions struct {
	ListOptions
}

type EventBridgeIntegrationCreateOptions struct {
	ID           string  `jsonapi:"primary,aws-event-bridge-integrations"`
	Name         *string `jsonapi:"attr,name"`
	AWSAccountId *string `jsonapi:"attr,aws-account-id"`
	Region       *string `jsonapi:"attr,region"`
}

type EventBridgeIntegrationUpdateOptions struct {
	ID     string            `jsonapi:"primary,aws-event-bridge-integrations"`
	Status IntegrationStatus `jsonapi:"attr,status"`
}

func (s *eventBridgeIntegrations) List(
	ctx context.Context, options EventBridgeIntegrationListOptions,
) (*EventBridgeIntegrationList, error) {
	req, err := s.client.newRequest("GET", "integrations/aws-event-bridge", &options)
	if err != nil {
		return nil, err
	}

	wl := &EventBridgeIntegrationList{}
	err = s.client.do(ctx, req, wl)
	if err != nil {
		return nil, err
	}

	return wl, nil
}

func (s *eventBridgeIntegrations) Create(
	ctx context.Context, options EventBridgeIntegrationCreateOptions,
) (*EventBridgeIntegration, error) {
	// Make sure we don't send a user provided ID.
	options.ID = ""

	req, err := s.client.newRequest("POST", "integrations/aws-event-bridge", &options)
	if err != nil {
		return nil, err
	}

	w := &EventBridgeIntegration{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (s *eventBridgeIntegrations) Read(ctx context.Context, ei string) (*EventBridgeIntegration, error) {
	if !validStringID(&ei) {
		return nil, errors.New("invalid value for EventBridge integration ID")
	}

	u := fmt.Sprintf("integrations/aws-event-bridge/%s", url.QueryEscape(ei))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	w := &EventBridgeIntegration{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (s *eventBridgeIntegrations) Update(
	ctx context.Context, si string, options EventBridgeIntegrationUpdateOptions,
) (*EventBridgeIntegration, error) {
	if !validStringID(&si) {
		return nil, errors.New("invalid value for slack integration ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("integrations/aws-event-bridge/%s", url.QueryEscape(si))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	w := &EventBridgeIntegration{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (s *eventBridgeIntegrations) Delete(ctx context.Context, si string) error {
	if !validStringID(&si) {
		return errors.New("invalid value for slack integration ID")
	}

	u := fmt.Sprintf("integrations/aws-event-bridge/%s", url.QueryEscape(si))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
