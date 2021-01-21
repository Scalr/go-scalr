package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ Variables = (*variables)(nil)

// Variables describes all the variable related methods that the Scalr API supports.
type Variables interface {
	// List all the variables associated with the given workspace.
	List(ctx context.Context, options VariableListOptions) (*VariableList, error)

	// Create is used to create a new variable.
	Create(ctx context.Context, options VariableCreateOptions, force bool) (*Variable, error)

	// Read a variable by its ID.
	Read(ctx context.Context, variableID string) (*Variable, error)

	// Update values of an existing variable.
	Update(ctx context.Context, variableID string, options VariableUpdateOptions, force bool) (*Variable, error)

	// Delete a variable by its ID.
	Delete(ctx context.Context, variableID string) error
}

// variables implements Variables.
type variables struct {
	client *Client
}

// CategoryType represents a category type.
type CategoryType string

//List all available categories.
const (
	CategoryEnv       CategoryType = "env"
	CategoryTerraform CategoryType = "terraform"
)

// VariableList represents a list of variables.
type VariableList struct {
	*Pagination
	Items []*Variable
}

// Variable represents a Scalr variable.
type Variable struct {
	ID        string       `jsonapi:"primary,vars"`
	Key       string       `jsonapi:"attr,key"`
	Value     string       `jsonapi:"attr,value"`
	Category  CategoryType `jsonapi:"attr,category"`
	HCL       bool         `jsonapi:"attr,hcl"`
	Sensitive bool         `jsonapi:"attr,sensitive"`
	Final     bool         `jsonapi:"attr,final"`

	// Relations
	Workspace   *Workspace   `jsonapi:"relation,workspace"`
	Environment *Environment `jsonapi:"relation,environment"`
	Account     *Account     `jsonapi:"relation,account"`
}

// VariableListOptions represents the options for listing variables.
type VariableListOptions struct {
	ListOptions
	Account     *string `url:"filter[account],omitempty"`
	Environment *string `url:"filter[environment],omitempty"`
	Workspace   *string `url:"filter[workspace],omitempty"`
}

func (o VariableListOptions) valid() error {
	return nil
}

// List all the variables associated with the given workspace.
func (s *variables) List(ctx context.Context, options VariableListOptions) (*VariableList, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}

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
func (s *variables) Create(ctx context.Context, options VariableCreateOptions, force bool) (*Variable, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("vars?force=%t", force)
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
	req, err := s.client.newRequest("GET", u, nil)
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

	// Whether to evaluate the value of the variable as a string of HCL code.
	HCL *bool `jsonapi:"attr,hcl,omitempty"`

	// Whether the value is sensitive.
	Sensitive *bool `jsonapi:"attr,sensitive,omitempty"`

	// Whether the value is final.
	Final *bool `jsonapi:"attr,final,omitempty"`
}

// Update values of an existing variable.
func (s *variables) Update(ctx context.Context, variableID string, options VariableUpdateOptions, force bool) (*Variable, error) {
	if !validStringID(&variableID) {
		return nil, errors.New("invalid value for variable ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = variableID

	u := fmt.Sprintf("vars/%s?force=%t", url.QueryEscape(variableID), force)
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
