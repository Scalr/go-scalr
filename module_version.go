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
	// ReadBySemanticVersion read module version by module and semantic version
	ReadBySemanticVersion(ctx context.Context, moduleId string, version string) (*ModuleVersion, error)
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
	IsRootModule bool                `jsonapi:"attr,is-root-module"`
	Status       ModuleVersionStatus `jsonapi:"attr,status"`
	Version      string              `jsonapi:"attr,version"`
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

func (s *moduleVersions) ReadBySemanticVersion(ctx context.Context, moduleID string, version string) (*ModuleVersion, error) {
	if !validStringID(&moduleID) {
		return nil, errors.New("invalid value for module id")
	}

	v := &version
	if !validString(v) {
		return nil, errors.New("invalid value for version")
	}

	req, err := s.client.newRequest("GET", "module-versions", &ModuleVersionListOptions{Module: moduleID, Version: v})
	if err != nil {
		return nil, err
	}

	mvl := &ModuleVersionList{}
	err = s.client.do(ctx, req, mvl)
	if err != nil {
		return nil, err
	}
	if len(mvl.Items) != 1 {
		return nil, ErrResourceNotFound
	}

	return mvl.Items[0], nil
}
