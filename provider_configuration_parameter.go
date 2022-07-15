package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ ProviderConfigurationParameters = (*providerConfigurationParameters)(nil)

// ProviderConfigurationParameters describes all the provider configurartion parameter related methods that the Scalr API supports.
type ProviderConfigurationParameters interface {
	List(ctx context.Context, configurationID string, options ProviderConfigurationParametersListOptions) (*ProviderConfigurationParametersList, error)
	Create(ctx context.Context, configurationID string, options ProviderConfigurationParameterCreateOptions) (*ProviderConfigurationParameter, error)
	Read(ctx context.Context, parameterID string) (*ProviderConfigurationParameter, error)
	Delete(ctx context.Context, parameterID string) error
	Update(ctx context.Context, parameterID string, options ProviderConfigurationParameterUpdateOptions) (*ProviderConfigurationParameter, error)
}

// providerConfigurationParameters implements ProviderConfigurationParameters.
type providerConfigurationParameters struct {
	client *Client
}

// ProviderConfigurationParametersList represents a list of provider configuration parameters.
type ProviderConfigurationParametersList struct {
	*Pagination
	Items []*ProviderConfigurationParameter
}

// ProviderConfigurationParameter represents a Scalr provider configuration parameter.
type ProviderConfigurationParameter struct {
	ID          string `jsonapi:"primary,provider-configuration-parameters"`
	Key         string `jsonapi:"attr,key"`
	Sensitive   bool   `jsonapi:"attr,sensitive"`
	Value       string `jsonapi:"attr,value"`
	Description string `jsonapi:"attr,description"`
}

// ProviderConfigurationParametersListOptions represents the options for listing provider configuration parameters.
type ProviderConfigurationParametersListOptions struct {
	ListOptions

	Sort string `url:"sort,omitempty"`
}

// List all the parameters of the provider configuration.
func (s *providerConfigurationParameters) List(ctx context.Context, configurationID string, options ProviderConfigurationParametersListOptions) (*ProviderConfigurationParametersList, error) {
	if !validStringID(&configurationID) {
		return nil, errors.New("invalid value for provider configuration ID")
	}

	url_path := fmt.Sprintf("provider-configurations/%s/parameters", url.QueryEscape(configurationID))

	req, err := s.client.newRequest("GET", url_path, &options)
	if err != nil {
		return nil, err
	}

	parametersList := &ProviderConfigurationParametersList{}
	err = s.client.do(ctx, req, parametersList)
	if err != nil {
		return nil, err
	}

	return parametersList, nil
}

// ProviderConfigurationParameterCreateOptions represents the options for creating a new provider configuration parameter.
type ProviderConfigurationParameterCreateOptions struct {
	ID          string  `jsonapi:"primary,provider-configuration-parameters"`
	Key         *string `jsonapi:"attr,key"`
	Sensitive   *bool   `jsonapi:"attr,sensitive,omitempty"`
	Value       *string `jsonapi:"attr,value"`
	Description *string `jsonapi:"attr,description,omitempty"`
}

// Create is used to create a new provider configuration parameter.
func (s *providerConfigurationParameters) Create(ctx context.Context, configurationID string, options ProviderConfigurationParameterCreateOptions) (*ProviderConfigurationParameter, error) {
	options.ID = ""

	url_path := fmt.Sprintf("provider-configurations/%s/parameters", url.QueryEscape(configurationID))
	req, err := s.client.newRequest("POST", url_path, &options)
	if err != nil {
		return nil, err
	}

	parameter := &ProviderConfigurationParameter{}
	err = s.client.do(ctx, req, parameter)

	if err != nil {
		return nil, err
	}

	return parameter, nil
}

// Read a provider configuration parameter by parameter ID.
func (s *providerConfigurationParameters) Read(ctx context.Context, parameterID string) (*ProviderConfigurationParameter, error) {
	if !validStringID(&parameterID) {
		return nil, errors.New("invalid value for provider configuration parameter ID")
	}

	url_path := fmt.Sprintf("provider-configuration-parameters/%s", url.QueryEscape(parameterID))

	req, err := s.client.newRequest("GET", url_path, nil)
	if err != nil {
		return nil, err
	}

	parameter := &ProviderConfigurationParameter{}
	err = s.client.do(ctx, req, parameter)
	if err != nil {
		return nil, err
	}

	return parameter, nil
}

// ProviderConfigurationParameterUpdateOptions represents the options for updating a provider configuration.
type ProviderConfigurationParameterUpdateOptions struct {
	ID          string  `jsonapi:"primary,provider-configuration-parameters"`
	Key         *string `jsonapi:"attr,key,omitempty"`
	Sensitive   *bool   `jsonapi:"attr,sensitive,omitempty"`
	Value       *string `jsonapi:"attr,value,omitempty"`
	Description *string `jsonapi:"attr,description,omitempty"`
}

// Update an existing provider configuration parameter.
func (s *providerConfigurationParameters) Update(ctx context.Context, parameterID string, options ProviderConfigurationParameterUpdateOptions) (*ProviderConfigurationParameter, error) {
	options.ID = ""

	if !validStringID(&parameterID) {
		return nil, errors.New("invalid value for provider configuration parameter ID")
	}

	url_path := fmt.Sprintf("provider-configuration-parameters/%s", url.QueryEscape(parameterID))

	req, err := s.client.newRequest("PATCH", url_path, &options)
	if err != nil {
		return nil, err
	}

	parameter := &ProviderConfigurationParameter{}
	err = s.client.do(ctx, req, parameter)
	if err != nil {
		return nil, err
	}

	return parameter, nil
}

// Delete deletes a provider configuration parameter by its ID.
func (s *providerConfigurationParameters) Delete(ctx context.Context, parameterID string) error {
	if !validStringID(&parameterID) {
		return errors.New("invalid value for provider parameter ID")
	}

	url_path := fmt.Sprintf("provider-configuration-parameters/%s", url.QueryEscape(parameterID))
	req, err := s.client.newRequest("DELETE", url_path, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
