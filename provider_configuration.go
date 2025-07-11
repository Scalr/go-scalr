package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ ProviderConfigurations = (*providerConfigurations)(nil)

// ProviderConfigurations describes all the provider configuration related methods that the Scalr API supports.
type ProviderConfigurations interface {
	List(ctx context.Context, options ProviderConfigurationsListOptions) (*ProviderConfigurationsList, error)
	Create(ctx context.Context, options ProviderConfigurationCreateOptions) (*ProviderConfiguration, error)
	Read(ctx context.Context, configurationID string) (*ProviderConfiguration, error)
	Delete(ctx context.Context, configurationID string) error
	Update(ctx context.Context, configurationID string, options ProviderConfigurationUpdateOptions) (*ProviderConfiguration, error)
}

type AwsDefaultTagsStrategy string

const (
	AwsDefaultTagsStrategySkip   AwsDefaultTagsStrategy = "skip"
	AwsDefaultTagsStrategyUpdate AwsDefaultTagsStrategy = "update"
)

// providerConfigurations implements ProviderConfigurations.
type providerConfigurations struct {
	client *Client
}

// ProviderConfigurationsList represents a list of provider configurations.
type ProviderConfigurationsList struct {
	*Pagination
	Items []*ProviderConfiguration
}

// ProviderConfiguration represents a Scalr provider configuration.
type ProviderConfiguration struct {
	ID                         string                 `jsonapi:"primary,provider-configurations"`
	Name                       string                 `jsonapi:"attr,name"`
	ProviderName               string                 `jsonapi:"attr,provider-name"`
	ExportShellVariables       bool                   `jsonapi:"attr,export-shell-variables"`
	IsShared                   bool                   `jsonapi:"attr,is-shared"`
	IsCustom                   bool                   `jsonapi:"attr,is-custom"`
	AwsAccessKey               string                 `jsonapi:"attr,aws-access-key"`
	AwsSecretKey               string                 `jsonapi:"attr,aws-secret-key"`
	AwsAccountType             string                 `jsonapi:"attr,aws-account-type"`
	AwsCredentialsType         string                 `jsonapi:"attr,aws-credentials-type"`
	AwsTrustedEntityType       string                 `jsonapi:"attr,aws-trusted-entity-type"`
	AwsRoleArn                 string                 `jsonapi:"attr,aws-role-arn"`
	AwsExternalId              string                 `jsonapi:"attr,aws-external-id"`
	AwsAudience                string                 `jsonapi:"attr,aws-audience"`
	AwsDefaultTags             *map[string]string     `jsonapi:"attr,aws-default-tags"`
	AwsDefaultTagsStrategy     AwsDefaultTagsStrategy `jsonapi:"attr,aws-default-tags-strategy"`
	AzurermClientId            string                 `jsonapi:"attr,azurerm-client-id"`
	AzurermClientSecret        string                 `jsonapi:"attr,azurerm-client-secret"`
	AzurermSubscriptionId      string                 `jsonapi:"attr,azurerm-subscription-id"`
	AzurermTenantId            string                 `jsonapi:"attr,azurerm-tenant-id"`
	AzurermAuthType            string                 `jsonapi:"attr,azurerm-auth-type"`
	AzurermAudience            string                 `jsonapi:"attr,azurerm-audience"`
	GoogleAuthType             string                 `jsonapi:"attr,google-auth-type"`
	GoogleServiceAccountEmail  string                 `jsonapi:"attr,google-service-account-email"`
	GoogleWorkloadProviderName string                 `jsonapi:"attr,google-workload-provider-name"`
	GoogleProject              string                 `jsonapi:"attr,google-project"`
	GoogleCredentials          string                 `jsonapi:"attr,google-credentials"`
	GoogleUseDefaultProject    bool                   `jsonapi:"attr,google-use-default-project"`
	ScalrHostname              string                 `jsonapi:"attr,scalr-hostname"`
	ScalrToken                 string                 `jsonapi:"attr,scalr-token"`

	Account      *Account                          `jsonapi:"relation,account"`
	Parameters   []*ProviderConfigurationParameter `jsonapi:"relation,parameters"`
	Environments []*Environment                    `jsonapi:"relation,environments"`
	Owners       []*Team                           `jsonapi:"relation,owners"`
}

// ProviderConfigurationsListOptions represents the options for listing provider configurations.
type ProviderConfigurationsListOptions struct {
	ListOptions

	Sort    string                       `url:"sort,omitempty"`
	Include string                       `url:"include,omitempty"`
	Filter  *ProviderConfigurationFilter `url:"filter,omitempty"`
}

// ProviderConfigurationFilter represents the options for filtering provider configurations.
type ProviderConfigurationFilter struct {
	ProviderConfiguration string `url:"provider-configuration,omitempty"`
	ProviderName          string `url:"provider-name,omitempty"`
	Name                  string `url:"name,omitempty"`
	AccountID             string `url:"account,omitempty"`
}

// List all the provider configurations within a scalr account.
func (s *providerConfigurations) List(ctx context.Context, options ProviderConfigurationsListOptions) (*ProviderConfigurationsList, error) {
	req, err := s.client.newRequest("GET", "provider-configurations", &options)
	if err != nil {
		return nil, err
	}

	pcfgl := &ProviderConfigurationsList{}
	err = s.client.do(ctx, req, pcfgl)
	if err != nil {
		return nil, err
	}

	return pcfgl, nil
}

// ProviderConfigurationCreateOptions represents the options for creating a new provider configuration.
type ProviderConfigurationCreateOptions struct {
	ID                         string                  `jsonapi:"primary,provider-configurations"`
	Name                       *string                 `jsonapi:"attr,name"`
	ProviderName               *string                 `jsonapi:"attr,provider-name"`
	ExportShellVariables       *bool                   `jsonapi:"attr,export-shell-variables,omitempty"`
	IsShared                   *bool                   `jsonapi:"attr,is-shared,omitempty"`
	IsCustom                   *bool                   `jsonapi:"attr,is-custom,omitempty"`
	AwsAccessKey               *string                 `jsonapi:"attr,aws-access-key,omitempty"`
	AwsSecretKey               *string                 `jsonapi:"attr,aws-secret-key,omitempty"`
	AwsAccountType             *string                 `jsonapi:"attr,aws-account-type"`
	AwsCredentialsType         *string                 `jsonapi:"attr,aws-credentials-type"`
	AwsTrustedEntityType       *string                 `jsonapi:"attr,aws-trusted-entity-type"`
	AwsAudience                *string                 `jsonapi:"attr,aws-audience"`
	AwsRoleArn                 *string                 `jsonapi:"attr,aws-role-arn"`
	AwsExternalId              *string                 `jsonapi:"attr,aws-external-id"`
	AwsDefaultTags             *map[string]string      `jsonapi:"attr,aws-default-tags,omitempty"`
	AwsDefaultTagsStrategy     *AwsDefaultTagsStrategy `jsonapi:"attr,aws-default-tags-strategy,omitempty"`
	AzurermClientId            *string                 `jsonapi:"attr,azurerm-client-id,omitempty"`
	AzurermClientSecret        *string                 `jsonapi:"attr,azurerm-client-secret,omitempty"`
	AzurermSubscriptionId      *string                 `jsonapi:"attr,azurerm-subscription-id,omitempty"`
	AzurermTenantId            *string                 `jsonapi:"attr,azurerm-tenant-id,omitempty"`
	AzurermAuthType            *string                 `jsonapi:"attr,azurerm-auth-type,omitempty"`
	AzurermAudience            *string                 `jsonapi:"attr,azurerm-audience,omitempty"`
	GoogleAuthType             *string                 `jsonapi:"attr,google-auth-type,omitempty"`
	GoogleServiceAccountEmail  *string                 `jsonapi:"attr,google-service-account-email,omitempty"`
	GoogleWorkloadProviderName *string                 `jsonapi:"attr,google-workload-provider-name,omitempty"`
	GoogleProject              *string                 `jsonapi:"attr,google-project,omitempty"`
	GoogleCredentials          *string                 `jsonapi:"attr,google-credentials,omitempty"`
	GoogleUseDefaultProject    *bool                   `jsonapi:"attr,google-use-default-project,omitempty"`
	ScalrHostname              *string                 `jsonapi:"attr,scalr-hostname,omitempty"`
	ScalrToken                 *string                 `jsonapi:"attr,scalr-token,omitempty"`

	Account      *Account       `jsonapi:"relation,account,omitempty"`
	Environments []*Environment `jsonapi:"relation,environments,omitempty"`
	Owners       []*Team        `jsonapi:"relation,owners"`
}

// Create is used to create a new provider configuration.
func (s *providerConfigurations) Create(ctx context.Context, options ProviderConfigurationCreateOptions) (*ProviderConfiguration, error) {
	options.ID = ""

	req, err := s.client.newRequest("POST", "provider-configurations", &options)
	if err != nil {
		return nil, err
	}

	pcfg := &ProviderConfiguration{}
	err = s.client.do(ctx, req, pcfg)
	if err != nil {
		return nil, err
	}

	return pcfg, nil
}

// Read a provider configuration by configuration ID.
func (s *providerConfigurations) Read(ctx context.Context, configurationID string) (*ProviderConfiguration, error) {
	if !validStringID(&configurationID) {
		return nil, errors.New("invalid value for provider configuration ID")
	}
	options := struct {
		Include string `url:"include"`
	}{
		Include: "parameters",
	}
	url_path := fmt.Sprintf("provider-configurations/%s", url.QueryEscape(configurationID))
	req, err := s.client.newRequest("GET", url_path, options)
	if err != nil {
		return nil, err
	}

	config := &ProviderConfiguration{}
	err = s.client.do(ctx, req, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// ProviderConfigurationUpdateOptions represents the options for updating a provider configuration.
type ProviderConfigurationUpdateOptions struct {
	ID string `jsonapi:"primary,provider-configurations"`

	Name                       *string                 `jsonapi:"attr,name"`
	IsShared                   *bool                   `jsonapi:"attr,is-shared,omitempty"`
	Environments               []*Environment          `jsonapi:"relation,environments"`
	ExportShellVariables       *bool                   `jsonapi:"attr,export-shell-variables"`
	AwsAccessKey               *string                 `jsonapi:"attr,aws-access-key"`
	AwsSecretKey               *string                 `jsonapi:"attr,aws-secret-key"`
	AwsAccountType             *string                 `jsonapi:"attr,aws-account-type"`
	AwsCredentialsType         *string                 `jsonapi:"attr,aws-credentials-type"`
	AwsTrustedEntityType       *string                 `jsonapi:"attr,aws-trusted-entity-type"`
	AwsRoleArn                 *string                 `jsonapi:"attr,aws-role-arn"`
	AwsExternalId              *string                 `jsonapi:"attr,aws-external-id"`
	AwsAudience                *string                 `jsonapi:"attr,aws-audience"`
	AwsDefaultTags             *map[string]string      `jsonapi:"attr,aws-default-tags"`
	AwsDefaultTagsStrategy     *AwsDefaultTagsStrategy `jsonapi:"attr,aws-default-tags-strategy"`
	AzurermAuthType            *string                 `jsonapi:"attr,azurerm-auth-type"`
	AzurermAudience            *string                 `jsonapi:"attr,azurerm-audience"`
	AzurermClientId            *string                 `jsonapi:"attr,azurerm-client-id"`
	AzurermClientSecret        *string                 `jsonapi:"attr,azurerm-client-secret"`
	AzurermSubscriptionId      *string                 `jsonapi:"attr,azurerm-subscription-id"`
	AzurermTenantId            *string                 `jsonapi:"attr,azurerm-tenant-id"`
	GoogleAuthType             *string                 `jsonapi:"attr,google-auth-type"`
	GoogleServiceAccountEmail  *string                 `jsonapi:"attr,google-service-account-email"`
	GoogleWorkloadProviderName *string                 `jsonapi:"attr,google-workload-provider-name"`
	GoogleProject              *string                 `jsonapi:"attr,google-project"`
	GoogleCredentials          *string                 `jsonapi:"attr,google-credentials"`
	GoogleUseDefaultProject    *bool                   `jsonapi:"attr,google-use-default-project,omitempty"`
	ScalrHostname              *string                 `jsonapi:"attr,scalr-hostname"`
	ScalrToken                 *string                 `jsonapi:"attr,scalr-token"`
	Owners                     []*Team                 `jsonapi:"relation,owners"`
}

// Update an existing provider configuration.
func (s *providerConfigurations) Update(ctx context.Context, configurationID string, options ProviderConfigurationUpdateOptions) (*ProviderConfiguration, error) {
	if !validStringID(&configurationID) {
		return nil, errors.New("invalid value for provider configuration ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	url_path := fmt.Sprintf("provider-configurations/%s", url.QueryEscape(configurationID))
	req, err := s.client.newRequest("PATCH", url_path, &options)
	if err != nil {
		return nil, err
	}

	configuration := &ProviderConfiguration{}
	err = s.client.do(ctx, req, configuration)
	if err != nil {
		return nil, err
	}

	return configuration, nil
}

// Delete deletes a provider configuration by its ID.
func (s *providerConfigurations) Delete(ctx context.Context, configurationID string) error {
	if !validStringID(&configurationID) {
		return errors.New("invalid value for provider configuration ID")
	}

	url_path := fmt.Sprintf("provider-configurations/%s", url.QueryEscape(configurationID))
	req, err := s.client.newRequest("DELETE", url_path, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
