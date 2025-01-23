package scalr

import (
	"context"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ RemoteStateConsumers = (*remoteStateConsumers)(nil)

type RemoteStateConsumers interface {
	List(ctx context.Context, wsID string, options ListOptions) (*RemoteStateConsumersList, error)
	Add(ctx context.Context, wsID string, wsRels []*WorkspaceRelation) error
	Replace(ctx context.Context, wsID string, wsRels []*WorkspaceRelation) error
	Delete(ctx context.Context, wsID string, wsRels []*WorkspaceRelation) error
}

type RemoteStateConsumersList struct {
	*Pagination
	Items []*WorkspaceRelation
}

type remoteStateConsumers struct {
	client *Client
}

func (s *remoteStateConsumers) List(ctx context.Context, wsID string, options ListOptions) (*RemoteStateConsumersList, error) {
	u := fmt.Sprintf("workspaces/%s/relationships/remote-state-consumers", url.QueryEscape(wsID))
	req, err := s.client.newRequest("GET", u, &options)
	if err != nil {
		return nil, err
	}

	cl := &RemoteStateConsumersList{}
	err = s.client.do(ctx, req, cl)
	if err != nil {
		return nil, err
	}

	return cl, nil
}

func (s *remoteStateConsumers) Add(ctx context.Context, wsID string, wsRels []*WorkspaceRelation) error {
	u := fmt.Sprintf("workspaces/%s/relationships/remote-state-consumers", url.QueryEscape(wsID))
	req, err := s.client.newRequest("POST", u, wsRels)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

func (s *remoteStateConsumers) Replace(ctx context.Context, wsID string, wsRels []*WorkspaceRelation) error {
	u := fmt.Sprintf("workspaces/%s/relationships/remote-state-consumers", url.QueryEscape(wsID))
	req, err := s.client.newRequest("PATCH", u, wsRels)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

func (s *remoteStateConsumers) Delete(ctx context.Context, wsID string, wsRels []*WorkspaceRelation) error {
	u := fmt.Sprintf("workspaces/%s/relationships/remote-state-consumers", url.QueryEscape(wsID))
	req, err := s.client.newRequest("DELETE", u, wsRels)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
