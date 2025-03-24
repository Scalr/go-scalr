package scalr

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkloadIdentityProviderCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("create provider", func(t *testing.T) {
		createOptions := WorkloadIdentityProviderCreateOptions{
			Name:             String("test-" + randomString(t)),
			URL:              String("https://" + strings.ReplaceAll(randomString(t), "-", "")),
			AllowedAudiences: []string{"aud1", "aud2"},
		}

		provider, err := client.WorkloadIdentityProviders.Create(ctx, createOptions)
		require.NoError(t, err)

		provider, err = client.WorkloadIdentityProviders.Read(ctx, provider.ID)
		require.NoError(t, err)

		assert.Equal(t, *createOptions.Name, provider.Name)
		assert.Equal(t, *createOptions.URL, provider.URL)
		assert.Equal(t, createOptions.AllowedAudiences, provider.AllowedAudiences)
	})
}

func TestWorkloadIdentityProviderUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	provider, removeProvider := createWorkloadIdentityProvider(t, client)
	defer removeProvider()

	t.Run("update provider", func(t *testing.T) {
		updateOptions := WorkloadIdentityProviderUpdateOptions{
			Name:             String("updated-provider"),
			AllowedAudiences: []string{"new-aud1", "new-aud2"},
		}

		provider, err := client.WorkloadIdentityProviders.Update(ctx, provider.ID, updateOptions)
		require.NoError(t, err)

		assert.Equal(t, *updateOptions.Name, provider.Name)
		assert.Equal(t, updateOptions.AllowedAudiences, provider.AllowedAudiences)
	})
}

func TestWorkloadIdentityProviderDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	provider, _ := createWorkloadIdentityProvider(t, client)

	t.Run("delete provider", func(t *testing.T) {
		err := client.WorkloadIdentityProviders.Delete(ctx, provider.ID)
		require.NoError(t, err)
	})
}

func TestWorkloadIdentityProviderList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	_, removeProvider := createWorkloadIdentityProvider(t, client)
	defer removeProvider()

	listOptions := WorkloadIdentityProvidersListOptions{
		ListOptions: ListOptions{
			PageNumber: 1,
			PageSize:   2,
		},
	}

	providers, err := client.WorkloadIdentityProviders.List(ctx, listOptions)
	require.NoError(t, err)
	assert.NotEmpty(t, providers)
}
