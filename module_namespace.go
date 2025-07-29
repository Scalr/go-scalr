package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ ModuleNamespaces = (*moduleNamespaces)(nil)

// ModuleNamespaces describes all the module namespace related methods that the Scalr API supports.
type ModuleNamespaces interface {
	List(ctx context.Context, options ModuleNamespacesListOptions) (*ModuleNamespacesList, error)
	Create(ctx context.Context, options ModuleNamespaceCreateOptions) (*ModuleNamespace, error)
	Read(ctx context.Context, namespaceID string) (*ModuleNamespace, error)
	Delete(ctx context.Context, namespaceID string) error
	Update(ctx context.Context, namespaceID string, options ModuleNamespaceUpdateOptions) (*ModuleNamespace, error)
}

// moduleNamespaces implements ModuleNamespaces.
type moduleNamespaces struct {
	client *Client
}

// ModuleNamespacesList represents a list of module namespaces.
type ModuleNamespacesList struct {
	*Pagination
	Items []*ModuleNamespace
}

// ModuleNamespace represents a Scalr module namespace.
type ModuleNamespace struct {
	ID       string `jsonapi:"primary,module-namespaces"`
	Name     string `jsonapi:"attr,name"`
	IsShared bool   `jsonapi:"attr,is-shared"`

	Environments []*Environment `jsonapi:"relation,environments"`
	Modules      []*Module      `jsonapi:"relation,modules"`
	Owners       []*Team        `jsonapi:"relation,owners"`
}

// ModuleNamespacesListOptions represents the options for listing module namespaces.
type ModuleNamespacesListOptions struct {
	ListOptions

	Sort   string                 `url:"sort,omitempty"`
	Filter *ModuleNamespaceFilter `url:"filter,omitempty"`
}

// ModuleNamespaceFilter represents the options for filtering module namespaces.
type ModuleNamespaceFilter struct {
	Name        string `url:"name,omitempty"`
	Environment string `url:"environment,omitempty"`
}

func (s *moduleNamespaces) List(ctx context.Context, options ModuleNamespacesListOptions) (*ModuleNamespacesList, error) {
	req, err := s.client.newRequest("GET", "module-namespaces", &options)
	if err != nil {
		return nil, err
	}

	mnl := &ModuleNamespacesList{}
	err = s.client.do(ctx, req, mnl)
	if err != nil {
		return nil, err
	}

	return mnl, nil
}

// ModuleNamespaceCreateOptions represents the options for creating a new module namespace.
type ModuleNamespaceCreateOptions struct {
	ID           string         `jsonapi:"primary,module-namespaces"`
	Name         *string        `jsonapi:"attr,name"`
	IsShared     *bool          `jsonapi:"attr,is-shared,omitempty"`
	Environments []*Environment `jsonapi:"relation,environments,omitempty"`
	Owners       []*Team        `jsonapi:"relation,owners"`
}

func (s *moduleNamespaces) Create(ctx context.Context, options ModuleNamespaceCreateOptions) (*ModuleNamespace, error) {
	options.ID = ""

	req, err := s.client.newRequest("POST", "module-namespaces", &options)
	if err != nil {
		return nil, err
	}

	mn := &ModuleNamespace{}
	err = s.client.do(ctx, req, mn)
	if err != nil {
		return nil, err
	}

	return mn, nil
}

func (s *moduleNamespaces) Read(ctx context.Context, namespaceID string) (*ModuleNamespace, error) {
	if !validStringID(&namespaceID) {
		return nil, errors.New("invalid value for module namespace ID")
	}
	url_path := fmt.Sprintf("module-namespaces/%s", url.QueryEscape(namespaceID))
	req, err := s.client.newRequest("GET", url_path, nil)
	if err != nil {
		return nil, err
	}

	namespace := &ModuleNamespace{}
	err = s.client.do(ctx, req, namespace)
	if err != nil {
		return nil, err
	}

	return namespace, nil
}

// ModuleNamespaceUpdateOptions represents the options for updating a module namespace.
type ModuleNamespaceUpdateOptions struct {
	ID           string         `jsonapi:"primary,module-namespaces"`
	IsShared     *bool          `jsonapi:"attr,is-shared,omitempty"`
	Environments []*Environment `jsonapi:"relation,environments"`
	Owners       []*Team        `jsonapi:"relation,owners"`
}

func (s *moduleNamespaces) Update(ctx context.Context, namespaceID string, options ModuleNamespaceUpdateOptions) (*ModuleNamespace, error) {
	if !validStringID(&namespaceID) {
		return nil, errors.New("invalid value for module namespace ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	url_path := fmt.Sprintf("module-namespaces/%s", url.QueryEscape(namespaceID))
	req, err := s.client.newRequest("PATCH", url_path, &options)
	if err != nil {
		return nil, err
	}

	namespace := &ModuleNamespace{}
	err = s.client.do(ctx, req, namespace)
	if err != nil {
		return nil, err
	}

	return namespace, nil
}

func (s *moduleNamespaces) Delete(ctx context.Context, namespaceID string) error {
	if !validStringID(&namespaceID) {
		return errors.New("invalid value for module namespace ID")
	}

	url_path := fmt.Sprintf("module-namespaces/%s", url.QueryEscape(namespaceID))
	req, err := s.client.newRequest("DELETE", url_path, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
