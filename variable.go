package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/google/go-querystring/query"
)

// Compile-time proof of interface implementation.
var _ Variables = (*variables)(nil)

// Variables describes all the variable related methods that the Scalr API supports.
type Variables interface {
	// List variables by filter options.
	List(ctx context.Context, options VariableListOptions) (*VariableList, error)

	// Create is used to create a new variable.
	Create(ctx context.Context, options VariableCreateOptions) (*Variable, error)

	// Read a variable by its ID.
	Read(ctx context.Context, variableID string) (*Variable, error)

	// Update values of an existing variable.
	Update(ctx context.Context, variableID string, options VariableUpdateOptions) (*Variable, error)

	// Delete a variable by its ID.
	Delete(ctx context.Context, variableID string) error
}

// variables implements Variables.
type variables struct {
	client *Client
}

// CategoryType represents a category type.
type CategoryType string

// List all available categories.
const (
	CategoryEnv       CategoryType = "env"
	CategoryTerraform CategoryType = "terraform"
	CategoryShell     CategoryType = "shell"
)

// VariableList represents a list of variables.
type VariableList struct {
	*Pagination
	Items []*Variable
}

// Variable represents a Scalr variable.
type Variable struct {
	ID             string       `jsonapi:"primary,vars"`
	Key            string       `jsonapi:"attr,key"`
	Value          string       `jsonapi:"attr,value"`
	Category       CategoryType `jsonapi:"attr,category"`
	Description    string       `jsonapi:"attr,description"`
	HCL            bool         `jsonapi:"attr,hcl"`
	Sensitive      bool         `jsonapi:"attr,sensitive"`
	Final          bool         `jsonapi:"attr,final"`
	UpdatedByEmail string       `jsonapi:"attr,updated-by-email"`
	UpdatedAt      *time.Time   `jsonapi:"attr,updated-at,iso8601"`

	// Relations
	Workspace   *Workspace   `jsonapi:"relation,workspace"`
	Environment *Environment `jsonapi:"relation,environment"`
	Account     *Account     `jsonapi:"relation,account"`
	UpdatedBy   *User        `jsonapi:"relation,updated-by"`
}

// VariableListOptions represents the options for listing variables.
type VariableListOptions struct {
	ListOptions

	// The comma-separated list of attributes.
	Sort *string `url:"sort,omitempty"`

	// The comma-separated list of relationship paths.
	Include *string `url:"include,omitempty"`

	// Filters
	Filter *VariableFilter `url:"filter,omitempty"`
}

type VariableFilter struct {
	// Filter by ID
	Var *string `url:"var,omitempty"`

	// Filter by key
	Key *string `url:"key,omitempty"`

	// Filter by category
	Category *string `url:"category,omitempty"`

	// Scope filters.
	Workspace   *string `url:"workspace,omitempty"`
	Environment *string `url:"environment,omitempty"`
	Account     *string `url:"account,omitempty"`
}

// List the variables.
func (s *variables) List(ctx context.Context, options VariableListOptions) (*VariableList, error) {
	req, err := s.client.newRequest("GET", "vars", &options)
	if err != nil {
		return nil, err
	}

	vl := &VariableList{}
	err = s.client.do(ctx, req, vl)
	if err != nil {
		return nil, err
	}

	return vl, nil
}

type VariableWriteQueryOptions struct {
	Force *bool `url:"force,omitempty"`

	// The comma-separated list of relationship paths.
	Include *string `url:"include,omitempty"`
}

// VariableCreateOptions represents the options for creating a new variable.
type VariableCreateOptions struct {
	// For internal use only!
	ID string `jsonapi:"primary,vars"`

	// The name of the variable.
	Key *string `jsonapi:"attr,key"`

	// The value of the variable.
	Value *string `jsonapi:"attr,value,omitempty"`

	// Whether this is a Terraform or environment variable.
	Category *CategoryType `jsonapi:"attr,category"`

	// Variable description.
	Description *string `jsonapi:"attr,description"`

	// Whether to evaluate the value of the variable as a string of HCL code.
	HCL *bool `jsonapi:"attr,hcl,omitempty"`

	// Whether the value is sensitive.
	Sensitive *bool `jsonapi:"attr,sensitive,omitempty"`

	// Whether the value is final.
	Final *bool `jsonapi:"attr,final,omitempty"`

	// The workspace that owns the variable.
	Workspace *Workspace `jsonapi:"relation,workspace,omitempty"`

	// The environment that owns the variable.
	Environment *Environment `jsonapi:"relation,environment,omitempty"`

	// The account  that owns the variable.
	Account *Account `jsonapi:"relation,account,omitempty"`

	QueryOptions *VariableWriteQueryOptions
}

func (o VariableCreateOptions) valid() error {
	if !validString(o.Key) {
		return errors.New("key is required")
	}
	if o.Category == nil {
		return errors.New("category is required")
	}
	return nil
}

// Create is used to create a new variable.
func (s *variables) Create(ctx context.Context, options VariableCreateOptions) (*Variable, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := "vars"
	if options.QueryOptions != nil {
		q, err := query.Values(options.QueryOptions)
		if err != nil {
			return nil, err
		}
		u = fmt.Sprintf("vars?%s", q.Encode())
	}
	req, err := s.client.newRequest("POST", u, &options)

	if err != nil {
		return nil, err
	}

	v := &Variable{}
	err = s.client.do(ctx, req, v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// Read a variable by its ID.
func (s *variables) Read(ctx context.Context, variableID string) (*Variable, error) {
	if !validStringID(&variableID) {
		return nil, errors.New("invalid value for variable ID")
	}

	u := fmt.Sprintf("vars/%s", url.QueryEscape(variableID))

	options := struct {
		Include string `url:"include"`
	}{
		Include: "updated-by",
	}

	req, err := s.client.newRequest("GET", u, &options)
	if err != nil {
		return nil, err
	}

	v := &Variable{}
	err = s.client.do(ctx, req, v)
	if err != nil {
		return nil, err
	}

	return v, err
}

// VariableUpdateOptions represents the options for updating a variable.
type VariableUpdateOptions struct {
	// For internal use only!
	ID string `jsonapi:"primary,vars"`

	// The name of the variable.
	Key *string `jsonapi:"attr,key,omitempty"`

	// The value of the variable.
	Value *string `jsonapi:"attr,value,omitempty"`

	// The description of the variable.
	Description *string `jsonapi:"attr,description,omitempty"`

	// Whether to evaluate the value of the variable as a string of HCL code.
	HCL *bool `jsonapi:"attr,hcl,omitempty"`

	// Whether the value is sensitive.
	Sensitive *bool `jsonapi:"attr,sensitive,omitempty"`

	// Whether the value is final.
	Final        *bool `jsonapi:"attr,final,omitempty"`
	QueryOptions *VariableWriteQueryOptions
}

// Update values of an existing variable.
func (s *variables) Update(ctx context.Context, variableID string, options VariableUpdateOptions) (*Variable, error) {
	if !validStringID(&variableID) {
		return nil, errors.New("invalid value for variable ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = variableID

	u := fmt.Sprintf("vars/%s", url.QueryEscape(variableID))
	if options.QueryOptions != nil {
		q, err := query.Values(options.QueryOptions)
		if err != nil {
			return nil, err
		}
		u = fmt.Sprintf("%s?%s", u, q.Encode())
	}

	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	v := &Variable{}
	err = s.client.do(ctx, req, v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// Delete a variable by its ID.
func (s *variables) Delete(ctx context.Context, variableID string) error {
	if !validStringID(&variableID) {
		return errors.New("invalid value for variable ID")
	}

	u := fmt.Sprintf("vars/%s", url.QueryEscape(variableID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
