package scalr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteStateConsumersList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	env, deleteEnv := createEnvironment(t, client)
	defer deleteEnv()

	workspace, deleteWorkspace := createWorkspace(t, client, env)
	defer deleteWorkspace()

	_, err := client.Workspaces.Update(ctx, workspace.ID, WorkspaceUpdateOptions{RemoteStateSharing: Bool(false)})
	require.NoError(t, err)

	consumer1, deleteConsumer1 := createWorkspace(t, client, env)
	defer deleteConsumer1()
	consumer2, deleteConsumer2 := createWorkspace(t, client, env)
	defer deleteConsumer2()
	consumer3, deleteConsumer3 := createWorkspace(t, client, env)
	defer deleteConsumer3()

	addRemoteStateConsumersToWorkspace(t, client, workspace, []*Workspace{consumer1, consumer2, consumer3})

	t.Run("list all", func(t *testing.T) {
		consumers, err := client.RemoteStateConsumers.List(ctx, workspace.ID, ListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 3, consumers.TotalCount)

		wsIDs := make([]string, len(consumers.Items))
		for _, ws := range consumers.Items {
			wsIDs = append(wsIDs, ws.ID)
		}
		assert.Contains(t, wsIDs, consumer1.ID)
		assert.Contains(t, wsIDs, consumer2.ID)
		assert.Contains(t, wsIDs, consumer3.ID)
	})
}

func TestRemoteStateConsumersAdd(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	env, deleteEnv := createEnvironment(t, client)
	defer deleteEnv()

	workspace, deleteWorkspace := createWorkspace(t, client, env)
	defer deleteWorkspace()

	_, err := client.Workspaces.Update(ctx, workspace.ID, WorkspaceUpdateOptions{RemoteStateSharing: Bool(false)})
	require.NoError(t, err)

	consumer1, deleteConsumer1 := createWorkspace(t, client, env)
	defer deleteConsumer1()
	consumer2, deleteConsumer2 := createWorkspace(t, client, env)
	defer deleteConsumer2()
	consumer3, deleteConsumer3 := createWorkspace(t, client, env)
	defer deleteConsumer3()

	t.Run("add consumer", func(t *testing.T) {
		err := client.RemoteStateConsumers.Add(ctx, workspace.ID,
			[]*WorkspaceRelation{
				{ID: consumer1.ID},
				{ID: consumer2.ID},
			},
		)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.RemoteStateConsumers.List(ctx, workspace.ID, ListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 2, refreshed.TotalCount)

		wsIDs := make([]string, len(refreshed.Items))
		for _, ws := range refreshed.Items {
			wsIDs = append(wsIDs, ws.ID)
		}
		assert.Contains(t, wsIDs, consumer1.ID)
		assert.Contains(t, wsIDs, consumer2.ID)
	})

	t.Run("add another one", func(t *testing.T) {
		err := client.RemoteStateConsumers.Add(ctx, workspace.ID, []*WorkspaceRelation{{ID: consumer3.ID}})
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.RemoteStateConsumers.List(ctx, workspace.ID, ListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 3, refreshed.TotalCount)

		wsIDs := make([]string, len(refreshed.Items))
		for _, ws := range refreshed.Items {
			wsIDs = append(wsIDs, ws.ID)
		}
		assert.Contains(t, wsIDs, consumer1.ID)
		assert.Contains(t, wsIDs, consumer2.ID)
		assert.Contains(t, wsIDs, consumer3.ID)
	})
}

func TestRemoteStateConsumersReplace(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	env, deleteEnv := createEnvironment(t, client)
	defer deleteEnv()

	workspace, deleteWorkspace := createWorkspace(t, client, env)
	defer deleteWorkspace()

	_, err := client.Workspaces.Update(ctx, workspace.ID, WorkspaceUpdateOptions{RemoteStateSharing: Bool(false)})
	require.NoError(t, err)

	consumer1, deleteConsumer1 := createWorkspace(t, client, env)
	defer deleteConsumer1()
	consumer2, deleteConsumer2 := createWorkspace(t, client, env)
	defer deleteConsumer2()
	consumer3, deleteConsumer3 := createWorkspace(t, client, env)
	defer deleteConsumer3()

	addRemoteStateConsumersToWorkspace(t, client, workspace, []*Workspace{consumer1})

	t.Run("replace consumers", func(t *testing.T) {
		err := client.RemoteStateConsumers.Replace(ctx, workspace.ID,
			[]*WorkspaceRelation{
				{ID: consumer2.ID},
				{ID: consumer3.ID},
			},
		)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.RemoteStateConsumers.List(ctx, workspace.ID, ListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 2, refreshed.TotalCount)

		wsIDs := make([]string, len(refreshed.Items))
		for _, ws := range refreshed.Items {
			wsIDs = append(wsIDs, ws.ID)
		}
		assert.Contains(t, wsIDs, consumer2.ID)
		assert.Contains(t, wsIDs, consumer3.ID)
	})

	t.Run("when all consumers should be removed", func(t *testing.T) {
		err := client.RemoteStateConsumers.Replace(ctx, workspace.ID, make([]*WorkspaceRelation, 0))
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.RemoteStateConsumers.List(ctx, workspace.ID, ListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 0, refreshed.TotalCount)
	})
}

func TestRemoteStateConsumersDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	env, deleteEnv := createEnvironment(t, client)
	defer deleteEnv()

	workspace, deleteWorkspace := createWorkspace(t, client, env)
	defer deleteWorkspace()

	_, err := client.Workspaces.Update(ctx, workspace.ID, WorkspaceUpdateOptions{RemoteStateSharing: Bool(false)})
	require.NoError(t, err)

	consumer1, deleteConsumer1 := createWorkspace(t, client, env)
	defer deleteConsumer1()
	consumer2, deleteConsumer2 := createWorkspace(t, client, env)
	defer deleteConsumer2()
	consumer3, deleteConsumer3 := createWorkspace(t, client, env)
	defer deleteConsumer3()

	addRemoteStateConsumersToWorkspace(t, client, workspace, []*Workspace{consumer1, consumer2, consumer3})

	t.Run("delete consumers", func(t *testing.T) {
		err := client.RemoteStateConsumers.Delete(ctx, workspace.ID,
			[]*WorkspaceRelation{
				{ID: consumer1.ID},
				{ID: consumer2.ID},
			},
		)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.RemoteStateConsumers.List(ctx, workspace.ID, ListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 1, refreshed.TotalCount)
		assert.Equal(t, consumer3.ID, refreshed.Items[0].ID)
	})
}
