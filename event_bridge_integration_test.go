package scalr

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventBridgeIntegrationsCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	accountId := os.Getenv("AWS_EVENT_BRIDGE_ACCOUNT_ID")
	region := os.Getenv("AWS_EVENT_BRIDGE_REGION")
	if len(accountId) == 0 || len(region) == 0 {
		t.Skip("Please set AWS_EVENT_BRIDGE_ACCOUNT_ID, AWS_EVENT_BRIDGE_REGION env variables to run this test.")
	}

	t.Run("with valid options", func(t *testing.T) {

		options := EventBridgeIntegrationCreateOptions{
			Name:         String("test-" + randomString(t)),
			AWSAccountId: &accountId,
			Region:       &region,
		}

		si, err := client.EventBridgeIntegrations.Create(ctx, options)
		require.NoError(t, err)

		refreshed, err := client.EventBridgeIntegrations.Read(ctx, si.ID)
		require.NoError(t, err)

		for _, item := range []*EventBridgeIntegration{
			si,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, *options.AWSAccountId, item.AWSAccountId)
			assert.Equal(t, *options.Region, item.Region)
		}

		err = client.EventBridgeIntegrations.Delete(ctx, si.ID)
		require.NoError(t, err)
	})
}

func TestEventBridgeIntegrationsUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	accountId := os.Getenv("AWS_EVENT_BRIDGE_ACCOUNT_ID")
	region := os.Getenv("AWS_EVENT_BRIDGE_REGION")
	if len(accountId) == 0 || len(region) == 0 {
		t.Skip("Please set AWS_EVENT_BRIDGE_ACCOUNT_ID, AWS_EVENT_BRIDGE_REGION env variables to run this test.")
	}

	createOptions := EventBridgeIntegrationCreateOptions{
		Name:         String("test-" + randomString(t)),
		AWSAccountId: &accountId,
		Region:       &region,
	}

	si, err := client.EventBridgeIntegrations.Create(ctx, createOptions)
	require.NoError(t, err)

	t.Run("with valid options", func(t *testing.T) {

		options := EventBridgeIntegrationUpdateOptions{
			Status: IntegrationStatusDisabled,
		}

		si, err := client.EventBridgeIntegrations.Update(ctx, si.ID, options)
		require.NoError(t, err)

		refreshed, err := client.EventBridgeIntegrations.Read(ctx, si.ID)
		require.NoError(t, err)

		for _, item := range []*EventBridgeIntegration{
			si,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, options.Status, item.Status)
		}

		err = client.EventBridgeIntegrations.Delete(ctx, si.ID)
		require.NoError(t, err)
	})
}

func TestEventBridgeIntegrationsList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	accountId := os.Getenv("AWS_EVENT_BRIDGE_ACCOUNT_ID")
	region := os.Getenv("AWS_EVENT_BRIDGE_REGION")
	if len(accountId) == 0 || len(region) == 0 {
		t.Skip("Please set AWS_EVENT_BRIDGE_ACCOUNT_ID, AWS_EVENT_BRIDGE_REGION env variables to run this test.")
	}

	createOptions := EventBridgeIntegrationCreateOptions{
		Name:         String("test-" + randomString(t)),
		AWSAccountId: &accountId,
		Region:       &region,
	}

	si, err := client.EventBridgeIntegrations.Create(ctx, createOptions)
	require.NoError(t, err)

	t.Run("with valid options", func(t *testing.T) {

		options := EventBridgeIntegrationListOptions{}

		sil, err := client.EventBridgeIntegrations.List(ctx, options)
		require.NoError(t, err)

		assert.Equal(t, 1, sil.TotalCount)
		expectedIDs := []string{si.ID}
		actualIDs := make([]string, len(sil.Items))
		for i, s := range sil.Items {
			actualIDs[i] = s.ID
		}
		assert.ElementsMatch(t, expectedIDs, actualIDs)

		err = client.EventBridgeIntegrations.Delete(ctx, si.ID)
		require.NoError(t, err)
	})
}
