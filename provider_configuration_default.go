package scalr

import (
	"context"
	"errors"
)

// Compile-time proof of interface implementation.
var _ ProviderConfigurationDefaults = (*providerConfigurationDefault)(nil)

// ProviderConfigurationDefaults describes all the provider configuration defaults related methods that the Scalr API supports.
type ProviderConfigurationDefaults interface {
	Create(ctx context.Context, options ProviderConfigurationDefaultsCreateOptions) error
	Delete(ctx context.Context, options ProviderConfigurationDefaultsDeleteOptions) error
}

// providerConfigurationDefault implements ProviderConfigurationDefaults.
type providerConfigurationDefault struct {
	client *Client
}

// ProviderConfigurationDefault represents a single provider configuration default relation.
type ProviderConfigurationDefault struct {
	ID string `jsonapi:"primary,provider-configurations"`
}

// ProviderConfigurationDefaultsCreateOptions represents options for creating new provider configuration default linkage
type ProviderConfigurationDefaultsCreateOptions struct {
	EnvironmentID           string
	ProviderConfigurationID string
}

type ProviderConfigurationDefaultsDeleteOptions struct {
	EnvironmentID           string
	ProviderConfigurationID string
}

func (o ProviderConfigurationDefaultsCreateOptions) valid() error {
	if !validStringID(&o.EnvironmentID) {
		return errors.New("invalid value for environment ID")
	}
	if !validStringID(&o.ProviderConfigurationID) {
		return errors.New("invalid value for provider configuration ID")
	}
	return nil
}

func (o ProviderConfigurationDefaultsDeleteOptions) valid() error {
	if !validStringID(&o.EnvironmentID) {
		return errors.New("invalid value for environment ID")
	}
	if !validStringID(&o.ProviderConfigurationID) {
		return errors.New("invalid value for provider configuration ID")
	}
	return nil
}

// Create a new provider configuration default linkage.
func (s *providerConfigurationDefault) Create(ctx context.Context, options ProviderConfigurationDefaultsCreateOptions) error {
	if !validStringID(&options.EnvironmentID) {
		return errors.New("invalid value for environment ID")
	}

	if !validStringID(&options.ProviderConfigurationID) {
		return errors.New("invalid value for provider configuration ID")
	}

	environment, err := s.client.Environments.Read(ctx, options.EnvironmentID)
	if err != nil {
		return err
	}

	for _, pc := range environment.DefaultProviderConfigurations {
		if pc.ID == options.ProviderConfigurationID {
			return errors.New("provider configuration with ID " + options.ProviderConfigurationID + " is already default for environment with ID" + options.EnvironmentID)
		}
	}

	environment.DefaultProviderConfigurations = append(environment.DefaultProviderConfigurations, &ProviderConfiguration{ID: options.ProviderConfigurationID})

	updateOpts := EnvironmentUpdateOptions{
		DefaultProviderConfigurations: environment.DefaultProviderConfigurations,
	}
	_, err = s.client.Environments.Update(ctx, environment.ID, updateOpts)
	if err != nil {
		return err
	}

	return nil
}

// Delete a provider configuration default linkage.
func (s *providerConfigurationDefault) Delete(ctx context.Context, options ProviderConfigurationDefaultsDeleteOptions) error {

	if !validStringID(&options.EnvironmentID) {
		return errors.New("invalid value for environment ID")
	}

	if !validStringID(&options.ProviderConfigurationID) {
		return errors.New("invalid value for provider configuration ID")
	}

	environment, err := s.client.Environments.Read(ctx, options.EnvironmentID)
	if err != nil {
		return err
	}

	found := false
	for i, pc := range environment.DefaultProviderConfigurations {
		if pc.ID == options.ProviderConfigurationID {
			environment.DefaultProviderConfigurations = append(environment.DefaultProviderConfigurations[:i], environment.DefaultProviderConfigurations[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return errors.New("provider configuration is not in the list of default provider configurations")
	}

	updateOpts := EnvironmentUpdateOptions{
		DefaultProviderConfigurations: environment.DefaultProviderConfigurations,
	}
	_, err = s.client.Environments.Update(ctx, environment.ID, updateOpts)
	if err != nil {
		return err
	}

	return nil
}
