package scalr

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestInfracostIntegrations_Create(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	apiKey := os.Getenv("TEST_INFRACOST_API_KEY")
	if len(apiKey) == 0 {
		t.Skip("Please set TEST_INFRACOST_API_KEY to run this test.")
	}

	t.Run("with valid options", func(t *testing.T) {
		options := InfracostIntegrationCreateOptions{
			Name:   String("test-" + randomString(t)),
			ApiKey: &apiKey,
		}

		ii, err := client.InfracostIntegrations.Create(ctx, options)
		require.NoError(t, err)
		defer func() { client.InfracostIntegrations.Delete(ctx, ii.ID) }()

		refreshed, err := client.InfracostIntegrations.Read(ctx, ii.ID)
		require.NoError(t, err)

		for _, item := range []*InfracostIntegration{
			ii,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
			assert.Empty(t, item.ApiKey)
		}
	})
}

func TestInfracostIntegrations_Update(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	apiKey := os.Getenv("TEST_INFRACOST_API_KEY")
	if len(apiKey) == 0 {
		t.Skip("Please set TEST_INFRACOST_API_KEY to run this test.")
	}

	createOptions := InfracostIntegrationCreateOptions{
		Name:   String("test-" + randomString(t)),
		ApiKey: &apiKey,
	}

	ii, err := client.InfracostIntegrations.Create(ctx, createOptions)
	require.NoError(t, err)
	defer func() { client.InfracostIntegrations.Delete(ctx, ii.ID) }()

	t.Run("with valid options", func(t *testing.T) {
		options := InfracostIntegrationUpdateOptions{
			Name:   String("test-" + randomString(t)),
			ApiKey: &apiKey,
			//Status: IntegrationStatusDisabled,
		}

		ii, err := client.InfracostIntegrations.Update(ctx, ii.ID, options)
		require.NoError(t, err)

		refreshed, err := client.InfracostIntegrations.Read(ctx, ii.ID)
		require.NoError(t, err)

		for _, item := range []*InfracostIntegration{
			ii,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
		}
	})
}

func TestInfracostIntegrations_List(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	apiKey := os.Getenv("TEST_INFRACOST_API_KEY")
	if len(apiKey) == 0 {
		t.Skip("Please set TEST_INFRACOST_API_KEY to run this test.")
	}

	createOptions := InfracostIntegrationCreateOptions{
		Name:   String("test-" + randomString(t)),
		ApiKey: &apiKey,
	}

	ii, err := client.InfracostIntegrations.Create(ctx, createOptions)
	require.NoError(t, err)
	defer func() { client.InfracostIntegrations.Delete(ctx, ii.ID) }()

	t.Run("with valid options", func(t *testing.T) {
		iil, err := client.InfracostIntegrations.List(ctx, InfracostIntegrationListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 1, iil.TotalCount)
		expectedIDs := []string{ii.ID}
		actualIDs := make([]string, len(iil.Items))
		for i, s := range iil.Items {
			actualIDs[i] = s.ID
		}
		assert.ElementsMatch(t, expectedIDs, actualIDs)
	})
}
