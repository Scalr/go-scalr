package scalr

import (
	"context"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ SSHKeysLinks = (*sshKeysLinks)(nil)

type SSHKeysLinks interface {
	Create(ctx context.Context, workspaceID string, sshKeyID string) (*Workspace, error)
	Delete(ctx context.Context, workspaceID string) error
}

// sshKeysLinks implements SSHKeysLinks.
type sshKeysLinks struct {
	client *Client
}

// SSHKeysLink represents a single SSH key workspace link.
type SSHKeysLink struct {
	SSHKeyID string `jsonapi:"attr,ssh-key"`
}

type SSHKeysLinkCreateOptions struct {
	SSHKeyID string `json:"ssh-key"`
}

// Create creates a SSH key workspace link.
func (s *sshKeysLinks) Create(ctx context.Context, workspaceID string, sshKeyID string) (*Workspace, error) {
	urlPath := fmt.Sprintf("workspaces/%s/ssh-key-links", url.QueryEscape(workspaceID))
	linkOptions := SSHKeysLinkCreateOptions{
		SSHKeyID: sshKeyID,
	}
	req, err := s.client.newRequest("POST", urlPath, linkOptions)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/vnd.api+json")

	workspace := &Workspace{}
	if err := s.client.do(ctx, req, workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

// Delete deletes a SSH key workspace link.
func (s *sshKeysLinks) Delete(ctx context.Context, workspaceID string) error {
	urlPath := fmt.Sprintf("workspaces/%s/ssh-key-links/", url.QueryEscape(workspaceID))
	req, err := s.client.newRequest("DELETE", urlPath, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
