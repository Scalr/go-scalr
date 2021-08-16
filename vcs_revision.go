package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ VcsRevisions = (*vcsRevisions)(nil)

// VCS revision implements VcsRevisions.
type vcsRevisions struct {
	client *Client
}

// VcsRevisions describes all the vcs revisions related methods that the Scalr API supports.
type VcsRevisions interface {
	// Read reads a VCS revision by its ID.
	Read(ctx context.Context, vcsRevisionID string) (*VcsRevision, error)
}

// VcsRevision represents the VCS metadata
type VcsRevision struct {
	ID             string `jsonapi:"primary,vcs-revisions"`
	Branch         string `jsonapi:"attr,branch"`
	CommitSha      string `jsonapi:"attr,commit-sha"`
	CommitMessage  string `jsonapi:"attr,commit-message"`
	SenderUsername string `jsonapi:"attr,sender-username"`
}

// Read a VCS revision by its ID.
func (s *vcsRevisions) Read(ctx context.Context, vcsRevisionID string) (*VcsRevision, error) {
	if !validStringID(&vcsRevisionID) {
		return nil, errors.New("invalid value for vcs revision ID")
	}

	u := fmt.Sprintf("vcs-revisions/%s", url.QueryEscape(vcsRevisionID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	r := &VcsRevision{}
	err = s.client.do(ctx, req, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}
