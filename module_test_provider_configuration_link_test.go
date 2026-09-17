package scalr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createModuleTestConfiguration enables the test configuration for the default test module
// and returns it together with a cleanup function that disables it again.
func createModuleTestConfiguration(t *testing.T, client *Client) (*ModuleTestConfiguration, func()) {
	ctx := context.Background()
	tc, err := client.ModuleTestConfigurations.Update(
		ctx, defaultModuleID, ModuleTestConfigurationUpdateOptions{
			Enabled: Bool(true),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return tc, func() {
		if _, err := client.ModuleTestConfigurations.Update(
			ctx, defaultModuleID, ModuleTestConfigurationUpdateOptions{Enabled: Bool(false)},
		); err != nil {
			t.Errorf("Error disabling module test configuration! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"ModuleTestConfiguration: %s\nError: %s", tc.ID, err)
		}
	}
}

func createAllowedProviderConfiguration(t *testing.T, client *Client, name string) (*ProviderConfiguration, func()) {
	ctx := context.Background()
	pcfg, err := client.ProviderConfigurations.Create(
		ctx,
		ProviderConfigurationCreateOptions{
			Account:               &Account{ID: defaultAccountID},
			Name:                  String(name),
			ProviderName:          String("aws"),
			IsShared:              Bool(true),
			IsAllowedInModuleTest: Bool(true),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return pcfg, func() {
		if err := client.ProviderConfigurations.Delete(ctx, pcfg.ID); err != nil {
			t.Errorf("Error destroying provider configuration! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Provider configuration: %s\nError: %s", pcfg.ID, err)
		}
	}
}

func TestModuleTestProviderConfigurationLinksCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	tc, removeTestConfiguration := createModuleTestConfiguration(t, client)
	defer removeTestConfiguration()

	t.Run("success", func(t *testing.T) {
		pcfg, removePcfg := createAllowedProviderConfiguration(t, client, "test-mtpcl-create")
		defer removePcfg()

		link, err := client.ModuleTestProviderConfigurationLinks.Create(
			ctx, tc.ID, ModuleTestProviderConfigurationLinkCreateOptions{
				ProviderConfiguration: &ProviderConfiguration{ID: pcfg.ID},
			},
		)
		require.NoError(t, err)
		defer client.ModuleTestProviderConfigurationLinks.Delete(ctx, link.ID)

		assert.NotEmpty(t, link.ID)
		assert.Equal(t, pcfg.ID, link.ProviderConfiguration.ID)
	})

	t.Run("when the test configuration id is invalid", func(t *testing.T) {
		_, err := client.ModuleTestProviderConfigurationLinks.Create(
			ctx, badIdentifier, ModuleTestProviderConfigurationLinkCreateOptions{},
		)
		assert.EqualError(t, err, "invalid value for test configuration ID")
	})
}

func TestModuleTestProviderConfigurationLinksList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	tc, removeTestConfiguration := createModuleTestConfiguration(t, client)
	defer removeTestConfiguration()

	pcfg, removePcfg := createAllowedProviderConfiguration(t, client, "test-mtpcl-list")
	defer removePcfg()

	link, err := client.ModuleTestProviderConfigurationLinks.Create(
		ctx, tc.ID, ModuleTestProviderConfigurationLinkCreateOptions{
			ProviderConfiguration: &ProviderConfiguration{ID: pcfg.ID},
		},
	)
	require.NoError(t, err)
	defer client.ModuleTestProviderConfigurationLinks.Delete(ctx, link.ID)

	t.Run("success", func(t *testing.T) {
		list, err := client.ModuleTestProviderConfigurationLinks.List(ctx, tc.ID, ListOptions{})
		require.NoError(t, err)

		var ids []string
		for _, item := range list.Items {
			ids = append(ids, item.ID)
		}
		assert.Contains(t, ids, link.ID)
	})
}

func TestModuleTestProviderConfigurationLinksRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	tc, removeTestConfiguration := createModuleTestConfiguration(t, client)
	defer removeTestConfiguration()

	pcfg, removePcfg := createAllowedProviderConfiguration(t, client, "test-mtpcl-read")
	defer removePcfg()

	link, err := client.ModuleTestProviderConfigurationLinks.Create(
		ctx, tc.ID, ModuleTestProviderConfigurationLinkCreateOptions{
			ProviderConfiguration: &ProviderConfiguration{ID: pcfg.ID},
		},
	)
	require.NoError(t, err)
	defer client.ModuleTestProviderConfigurationLinks.Delete(ctx, link.ID)

	t.Run("success", func(t *testing.T) {
		readLink, err := client.ModuleTestProviderConfigurationLinks.Read(ctx, link.ID)
		require.NoError(t, err)
		assert.Equal(t, link.ID, readLink.ID)
		assert.Equal(t, pcfg.ID, readLink.ProviderConfiguration.ID)
	})

	t.Run("when the link id is invalid", func(t *testing.T) {
		_, err := client.ModuleTestProviderConfigurationLinks.Read(ctx, badIdentifier)
		assert.EqualError(t, err, "invalid value for module test provider configuration link ID")
	})
}

func TestModuleTestProviderConfigurationLinksUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	tc, removeTestConfiguration := createModuleTestConfiguration(t, client)
	defer removeTestConfiguration()

	pcfg1, removePcfg1 := createAllowedProviderConfiguration(t, client, "test-mtpcl-update-1")
	defer removePcfg1()

	pcfg2, removePcfg2 := createAllowedProviderConfiguration(t, client, "test-mtpcl-update-2")
	defer removePcfg2()

	link, err := client.ModuleTestProviderConfigurationLinks.Create(
		ctx, tc.ID, ModuleTestProviderConfigurationLinkCreateOptions{
			ProviderConfiguration: &ProviderConfiguration{ID: pcfg1.ID},
		},
	)
	require.NoError(t, err)
	defer client.ModuleTestProviderConfigurationLinks.Delete(ctx, link.ID)

	t.Run("success", func(t *testing.T) {
		updated, err := client.ModuleTestProviderConfigurationLinks.Update(
			ctx, link.ID, ModuleTestProviderConfigurationLinkUpdateOptions{
				ProviderConfiguration: &ProviderConfiguration{ID: pcfg2.ID},
			},
		)
		require.NoError(t, err)
		assert.Equal(t, pcfg2.ID, updated.ProviderConfiguration.ID)
	})
}

func TestModuleTestProviderConfigurationLinksDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	tc, removeTestConfiguration := createModuleTestConfiguration(t, client)
	defer removeTestConfiguration()

	pcfg, removePcfg := createAllowedProviderConfiguration(t, client, "test-mtpcl-delete")
	defer removePcfg()

	link, err := client.ModuleTestProviderConfigurationLinks.Create(
		ctx, tc.ID, ModuleTestProviderConfigurationLinkCreateOptions{
			ProviderConfiguration: &ProviderConfiguration{ID: pcfg.ID},
		},
	)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		err := client.ModuleTestProviderConfigurationLinks.Delete(ctx, link.ID)
		require.NoError(t, err)

		_, err = client.ModuleTestProviderConfigurationLinks.Read(ctx, link.ID)
		assert.Error(t, err)
	})
}
