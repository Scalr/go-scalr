package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ ModuleVersions = (*moduleVersions)(nil)

// ModuleVersions describes all the run related methods that the Scalr API supports.
type ModuleVersions interface {
	// List all the module versions within a module.
	List(ctx context.Context, options ModuleVersionListOptions) (*ModuleVersionList, error)
	// Read a module version by its ID.
	Read(ctx context.Context, moduleVersionID string) (*ModuleVersion, error)
}

// moduleVersions implements ModuleVersions.
type moduleVersions struct {
	client *Client
}

// ModuleVersionList represents a list of module versions.
type ModuleVersionList struct {
	*Pagination
	Items []*ModuleVersion
}

// ModuleVersion represents a Scalr module version.
type ModuleVersion struct {
	ID           string              `jsonapi:"primary,module-versions"`
	IsRootModule bool                `jsonapi:"attr"`
	Status       ModuleVersionStatus `jsonapi:"attr,status"`
}

type ModuleVersionStatus string

const (
	ModuleVersionNotUploaded   ModuleVersionStatus = "not_uploaded"
	ModuleVersionPending       ModuleVersionStatus = "pending"
	ModuleVersionOk            ModuleVersionStatus = "ok"
	ModuleVersionErrored       ModuleVersionStatus = "reg_ingress_failed"
	ModuleVersionPendingDelete ModuleVersionStatus = "pending_delete"
)

type ModuleVersionListOptions struct {
	ListOptions
	Module  string  `url:"filter[module]"`
	Status  *string `url:"filter[status],omitempty"`
	Version *string `url:"filter[version],omitempty"`
	Include string  `url:"include,omitempty"`
}

func (o ModuleVersionListOptions) validate() error {
	if o.Module == "" {
		return errors.New("filter[module] is required")
	}

	return nil
}

// Read a module version by its ID.
func (s *moduleVersions) Read(ctx context.Context, moduleVersionID string) (*ModuleVersion, error) {
	if !validStringID(&moduleVersionID) {
		return nil, errors.New("invalid value for module version ID")
	}

	u := fmt.Sprintf("module-versions/%s", url.QueryEscape(moduleVersionID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	m := &ModuleVersion{}
	err = s.client.do(ctx, req, m)
	if err != nil {
		return nil, err
	}

	return m, err
}

// List the list of module versions
func (s *moduleVersions) List(ctx context.Context, options ModuleVersionListOptions) (*ModuleVersionList, error) {
	if err := options.validate(); err != nil {
		return nil, err
	}

	req, err := s.client.newRequest("GET", "module-versions", &options)
	if err != nil {
		return nil, err
	}

	mv := &ModuleVersionList{}
	err = s.client.do(ctx, req, mv)
	if err != nil {
		return nil, err
	}

	return mv, nil
}
