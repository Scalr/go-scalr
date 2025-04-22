package scalr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFederatedEnvironmentsList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	environment, deleteEnv := createEnvironment(t, client)
	defer deleteEnv()

	federated1, deleteFederated1 := createEnvironment(t, client)
	defer deleteFederated1()
	federated2, deleteFederated2 := createEnvironment(t, client)
	defer deleteFederated2()
	federated3, deleteFederated3 := createEnvironment(t, client)
	defer deleteFederated3()

	addFederatedEnvironments(t, client, environment, []*Environment{federated1, federated2, federated3})

	t.Run("list all", func(t *testing.T) {
		federated, err := client.FederatedEnvironments.List(ctx, environment.ID, ListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 3, federated.TotalCount)

		envIDs := make([]string, len(federated.Items))
		for _, env := range federated.Items {
			envIDs = append(envIDs, env.ID)
		}
		assert.Contains(t, envIDs, federated1.ID)
		assert.Contains(t, envIDs, federated2.ID)
		assert.Contains(t, envIDs, federated3.ID)
	})
}

func TestFederatedEnvironmentsAdd(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	environment, deleteEnv := createEnvironment(t, client)
	defer deleteEnv()

	federated1, deleteFederated1 := createEnvironment(t, client)
	defer deleteFederated1()
	federated2, deleteFederated2 := createEnvironment(t, client)
	defer deleteFederated2()
	federated3, deleteFederated3 := createEnvironment(t, client)
	defer deleteFederated3()

	t.Run("add federated environments", func(t *testing.T) {
		err := client.FederatedEnvironments.Add(ctx, environment.ID,
			[]*EnvironmentRelation{
				{ID: federated1.ID},
				{ID: federated2.ID},
			},
		)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.FederatedEnvironments.List(ctx, environment.ID, ListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 2, refreshed.TotalCount)

		envIDs := make([]string, len(refreshed.Items))
		for _, env := range refreshed.Items {
			envIDs = append(envIDs, env.ID)
		}
		assert.Contains(t, envIDs, federated1.ID)
		assert.Contains(t, envIDs, federated2.ID)
	})

	t.Run("add another one", func(t *testing.T) {
		err := client.FederatedEnvironments.Add(ctx, environment.ID, []*EnvironmentRelation{{ID: federated3.ID}})
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.FederatedEnvironments.List(ctx, environment.ID, ListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 3, refreshed.TotalCount)

		envIDs := make([]string, len(refreshed.Items))
		for _, env := range refreshed.Items {
			envIDs = append(envIDs, env.ID)
		}
		assert.Contains(t, envIDs, federated1.ID)
		assert.Contains(t, envIDs, federated2.ID)
		assert.Contains(t, envIDs, federated3.ID)
	})
}

func TestFederatedEnvironmentsReplace(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	environment, deleteEnv := createEnvironment(t, client)
	defer deleteEnv()

	federated1, deleteFederated1 := createEnvironment(t, client)
	defer deleteFederated1()
	federated2, deleteFederated2 := createEnvironment(t, client)
	defer deleteFederated2()
	federated3, deleteFederated3 := createEnvironment(t, client)
	defer deleteFederated3()

	addFederatedEnvironments(t, client, environment, []*Environment{federated1})

	t.Run("replace federated environments", func(t *testing.T) {
		err := client.FederatedEnvironments.Replace(ctx, environment.ID,
			[]*EnvironmentRelation{
				{ID: federated2.ID},
				{ID: federated3.ID},
			},
		)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.FederatedEnvironments.List(ctx, environment.ID, ListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 2, refreshed.TotalCount)

		envIDs := make([]string, len(refreshed.Items))
		for _, env := range refreshed.Items {
			envIDs = append(envIDs, env.ID)
		}
		assert.Contains(t, envIDs, federated2.ID)
		assert.Contains(t, envIDs, federated3.ID)
	})

	t.Run("when all federated environments should be removed", func(t *testing.T) {
		err := client.FederatedEnvironments.Replace(ctx, environment.ID, make([]*EnvironmentRelation, 0))
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.FederatedEnvironments.List(ctx, environment.ID, ListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 0, refreshed.TotalCount)
	})
}

func TestFederatedEnvironmentsDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	environment, deleteEnv := createEnvironment(t, client)
	defer deleteEnv()

	federated1, deleteFederated1 := createEnvironment(t, client)
	defer deleteFederated1()
	federated2, deleteFederated2 := createEnvironment(t, client)
	defer deleteFederated2()
	federated3, deleteFederated3 := createEnvironment(t, client)
	defer deleteFederated3()

	addFederatedEnvironments(t, client, environment, []*Environment{federated1, federated2, federated3})

	t.Run("delete federated environments", func(t *testing.T) {
		err := client.FederatedEnvironments.Delete(ctx, environment.ID,
			[]*EnvironmentRelation{
				{ID: federated1.ID},
				{ID: federated2.ID},
			},
		)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.FederatedEnvironments.List(ctx, environment.ID, ListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 1, refreshed.TotalCount)
		assert.Equal(t, federated3.ID, refreshed.Items[0].ID)
	})
}
