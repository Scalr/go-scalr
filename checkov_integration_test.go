package scalr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckovIntegrationsCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("with valid options", func(t *testing.T) {

		options := CheckovIntegrationCreateOptions{
			Name: String("test-" + randomString(t)),
		}

		ci, err := client.CheckovIntegrations.Create(ctx, options)
		require.NoError(t, err)

		refreshed, err := client.CheckovIntegrations.Read(ctx, ci.ID)
		require.NoError(t, err)

		for _, item := range []*CheckovIntegration{
			ci,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
			assert.NotEmpty(t, item.Version)
		}

		err = client.CheckovIntegrations.Delete(ctx, ci.ID)
		require.NoError(t, err)
	})
}

func TestCheckovIntegrationsUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	ci, deleteCI := createCheckovIntegration(t, client)
	defer deleteCI()

	t.Run("with valid options", func(t *testing.T) {

		options := CheckovIntegrationUpdateOptions{
			Name: String("test-" + randomString(t)),
		}

		updated, err := client.CheckovIntegrations.Update(ctx, ci.ID, options)
		require.NoError(t, err)

		refreshed, err := client.CheckovIntegrations.Read(ctx, updated.ID)
		require.NoError(t, err)

		for _, item := range []*CheckovIntegration{
			updated,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
		}
	})
}

func TestCheckovIntegrationsList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	ci, deleteCI := createCheckovIntegration(t, client)
	defer deleteCI()

	t.Run("with valid options", func(t *testing.T) {

		options := CheckovIntegrationListOptions{}

		cil, err := client.CheckovIntegrations.List(ctx, options)
		require.NoError(t, err)

		actualIDs := make([]string, len(cil.Items))
		for i, s := range cil.Items {
			actualIDs[i] = s.ID
		}
		assert.Contains(t, actualIDs, ci.ID)
	})
}

func createCheckovIntegration(t *testing.T, client *Client) (*CheckovIntegration, func()) {
	ctx := context.Background()

	ci, err := client.CheckovIntegrations.Create(ctx, CheckovIntegrationCreateOptions{
		Name: String("test-" + randomString(t)),
	})

	if err != nil {
		t.Fatal(err)
	}

	return ci, func() {
		if err := client.CheckovIntegrations.Delete(ctx, ci.ID); err != nil {
			t.Errorf("Error destroying Checkov integration! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Checkov integration: %s\nError: %s", ci.ID, err)
		}
	}
}
