package scalr

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProviderConfigurationDefaultCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	environment, removeEnvironment := createEnvironment(t, client)
	defer removeEnvironment()

	configuration, err := client.ProviderConfigurations.Create(
		ctx,
		ProviderConfigurationCreateOptions{
			Account:      &Account{ID: defaultAccountID},
			Name:         String("kubernetes_dev"),
			ProviderName: String("kubernetes"),
			IsShared:     Bool(false),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	configuration, err = client.ProviderConfigurations.Update(ctx, configuration.ID, ProviderConfigurationUpdateOptions{
		Name:                 String("scalr"),
		ExportShellVariables: Bool(false),
		Environments:         []*Environment{environment},
	})
	require.NoError(t, err)

	t.Run("with valid options", func(t *testing.T) {
		options := ProviderConfigurationDefaultsCreateOptions{
			EnvironmentID:           environment.ID,
			ProviderConfigurationID: configuration.ID,
		}

		err := client.ProviderConfigurationDefaults.Create(ctx, options)

		require.NoError(t, err)

		// Get a refreshed view from the API.
		environment, err := client.Environments.Read(ctx, environment.ID)
		require.NoError(t, err)

		var found bool
		for _, pc := range environment.DefaultProviderConfigurations {
			if pc.ID == configuration.ID {
				found = true
			}
		}

		assert.True(t, found)
	})

	t.Run("with invalid environment ID", func(t *testing.T) {
		options := ProviderConfigurationDefaultsCreateOptions{
			EnvironmentID:           "invalid",
			ProviderConfigurationID: configuration.ID,
		}

		err := client.ProviderConfigurationDefaults.Create(ctx, options)

		require.Error(t, err)
	})

	t.Run("with invalid provider configuration ID", func(t *testing.T) {
		options := ProviderConfigurationDefaultsCreateOptions{
			EnvironmentID:           environment.ID,
			ProviderConfigurationID: "invalid",
		}

		err := client.ProviderConfigurationDefaults.Create(ctx, options)

		require.Error(t, err)
	})
}

func TestProviderConfigurationDefaultDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	environment, removeEnvironment := createEnvironment(t, client)
	defer removeEnvironment()

	configuration, deleteConfiguration := createProviderConfiguration(
		t, client, "kubernetes", "kubernetes_dev",
	)
	defer deleteConfiguration()

	configuration, err := client.ProviderConfigurations.Update(ctx, configuration.ID, ProviderConfigurationUpdateOptions{
		Environments: []*Environment{environment},
	})
	require.NoError(t, err)

	t.Run("with valid options", func(t *testing.T) {
		options := ProviderConfigurationDefaultsDeleteOptions{
			EnvironmentID:           environment.ID,
			ProviderConfigurationID: configuration.ID,
		}

		err := client.ProviderConfigurationDefaults.Delete(ctx, options)

		require.NoError(t, err)

		// Get a refreshed view from the API.
		environment, err := client.Environments.Read(ctx, environment.ID)
		require.NoError(t, err)

		var found bool
		for _, pc := range environment.DefaultProviderConfigurations {
			if pc.ID == configuration.ID {
				found = true
			}
		}

		assert.False(t, found)
	})

	t.Run("with invalid options", func(t *testing.T) {
		options := ProviderConfigurationDefaultsDeleteOptions{
			EnvironmentID:           environment.ID,
			ProviderConfigurationID: configuration.ID,
		}

		err := client.ProviderConfigurationDefaults.Delete(ctx, options)

		require.Error(t, err)
	})

	t.Run("with invalid environment ID", func(t *testing.T) {
		options := ProviderConfigurationDefaultsDeleteOptions{
			EnvironmentID:           "invalid",
			ProviderConfigurationID: configuration.ID,
		}

		err := client.ProviderConfigurationDefaults.Delete(ctx, options)

		require.Error(t, err)
	})

	t.Run("with invalid provider configuration ID", func(t *testing.T) {
		options := ProviderConfigurationDefaultsDeleteOptions{
			EnvironmentID:           environment.ID,
			ProviderConfigurationID: "invalid",
		}

		err := client.ProviderConfigurationDefaults.Delete(ctx, options)

		require.Error(t, err)
	})
}
