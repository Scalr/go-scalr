package scalr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspacesList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	wsTest1, wsTest1Cleanup := createWorkspace(t, client, envTest)
	defer wsTest1Cleanup()
	wsTest2, wsTest2Cleanup := createWorkspace(t, client, envTest)
	defer wsTest2Cleanup()

	t.Run("without list options", func(t *testing.T) {
		wsl, err := client.Workspaces.List(ctx, WorkspaceListOptions{Environment: &envTest.ID})
		require.NoError(t, err)
		wslIDs := make([]string, len(wsl.Items))
		for _, ws := range wsl.Items {
			wslIDs = append(wslIDs, ws.ID)
		}
		assert.Contains(t, wslIDs, wsTest1.ID)
		assert.Contains(t, wslIDs, wsTest2.ID)
		assert.Equal(t, 1, wsl.CurrentPage)
		assert.Equal(t, 2, wsl.TotalCount)
	})

	t.Run("with list options", func(t *testing.T) {
		// Request a page number which is out of range. The result should
		// be successful, but return no results if the paging options are
		// properly passed along.
		wl, err := client.Workspaces.List(ctx, WorkspaceListOptions{
			ListOptions: ListOptions{
				PageNumber: 999,
				PageSize:   100,
			},
			Environment: &envTest.ID,
		})
		require.NoError(t, err)
		assert.Empty(t, wl.Items)
		assert.Equal(t, 999, wl.CurrentPage)
		assert.Equal(t, 2, wl.TotalCount)
	})
	t.Run("without a valid environment", func(t *testing.T) {
		wl, err := client.Workspaces.List(ctx, WorkspaceListOptions{Environment: String(badIdentifier)})
		assert.Len(t, wl.Items, 0)
		assert.NoError(t, err)
	})
}

func TestWorkspacesCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	t.Run("with valid options", func(t *testing.T) {
		options := WorkspaceCreateOptions{
			Environment:      envTest,
			Name:             String("foo"),
			AutoApply:        Bool(true),
			Operations:       Bool(true),
			TerraformVersion: String("0.12.25"),
			WorkingDirectory: String("bar/"),
		}

		ws, err := client.Workspaces.Create(ctx, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.Workspaces.ReadByID(ctx, ws.ID)
		require.NoError(t, err)

		for _, item := range []*Workspace{
			ws,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, *options.AutoApply, item.AutoApply)
			assert.Equal(t, false, item.HasActiveState)
			assert.Equal(t, *options.Operations, item.Operations)
			assert.Equal(t, *options.TerraformVersion, item.TerraformVersion)
			assert.Equal(t, *options.WorkingDirectory, item.WorkingDirectory)
		}
	})

	t.Run("when options is missing name", func(t *testing.T) {
		w, err := client.Workspaces.Create(ctx, WorkspaceCreateOptions{Environment: envTest})
		assert.Nil(t, w)
		assert.EqualError(t, err, "name is required")
	})

	t.Run("when options has an invalid name", func(t *testing.T) {
		w, err := client.Workspaces.Create(ctx, WorkspaceCreateOptions{
			Name:        String(badIdentifier),
			Environment: envTest,
		})
		assert.Nil(t, w)
		assert.EqualError(t, err, "invalid value for name")
	})

	t.Run("when options has an invalid environment", func(t *testing.T) {
		_, err := client.Workspaces.Create(ctx, WorkspaceCreateOptions{
			Name:        String("foo"),
			Environment: &Environment{ID: badIdentifier},
		})
		assert.Equal(t, err, ErrResourceNotFound)
	})

	t.Run("when an error is returned from the api", func(t *testing.T) {
		ws, err := client.Workspaces.Create(ctx, WorkspaceCreateOptions{
			Name:             String("bar"),
			TerraformVersion: String("nonexisting"),
			Environment:      envTest,
		})
		assert.Nil(t, ws)
		assert.Error(t, err)
	})
}

func TestWorkspacesRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	wsTest, wsTestCleanup := createWorkspace(t, client, envTest)
	defer wsTestCleanup()

	t.Run("when the workspace exists", func(t *testing.T) {
		ws, err := client.Workspaces.Read(ctx, envTest.ID, wsTest.Name)
		require.NoError(t, err)
		assert.Equal(t, wsTest.ID, ws.ID)

		t.Run("relationships are properly decoded", func(t *testing.T) {
			assert.Equal(t, envTest.ID, ws.Environment.ID)
		})

		t.Run("timestamps are properly decoded", func(t *testing.T) {
			assert.NotEmpty(t, ws.CreatedAt)
		})
	})

	t.Run("when the workspace does not exist", func(t *testing.T) {
		_, err := client.Workspaces.Read(ctx, envTest.ID, "nonexisting")
		assert.Error(t, err)
	})

	t.Run("when the environment does not exist", func(t *testing.T) {
		_, err := client.Workspaces.Read(ctx, "nonexisting", "nonexisting")
		assert.Error(t, err)
	})

	t.Run("without a valid environment", func(t *testing.T) {
		_, err := client.Workspaces.Read(ctx, badIdentifier, wsTest.Name)
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value for environment")
	})

	t.Run("without a valid workspace", func(t *testing.T) {
		ws, err := client.Workspaces.Read(ctx, envTest.Name, badIdentifier)
		assert.Nil(t, ws)
		assert.EqualError(t, err, "invalid value for workspace")
	})
}

func TestWorkspacesReadByID(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	wsTest, wsTestCleanup := createWorkspace(t, client, envTest)
	defer wsTestCleanup()

	t.Run("when the workspace exists", func(t *testing.T) {
		ws, err := client.Workspaces.ReadByID(ctx, wsTest.ID)
		require.NoError(t, err)
		assert.Equal(t, wsTest.ID, ws.ID)

		t.Run("relationships are properly decoded", func(t *testing.T) {
			assert.Equal(t, envTest.ID, ws.Environment.ID)
		})

		t.Run("timestamps are properly decoded", func(t *testing.T) {
			assert.NotEmpty(t, ws.CreatedAt)
		})
	})

	t.Run("when the workspace does not exist", func(t *testing.T) {
		ws, err := client.Workspaces.ReadByID(ctx, "nonexisting")
		assert.Nil(t, ws)
		assert.Error(t, err)
	})

	t.Run("without a valid workspace ID", func(t *testing.T) {
		ws, err := client.Workspaces.ReadByID(ctx, badIdentifier)
		assert.Nil(t, ws)
		assert.EqualError(t, err, "invalid value for workspace ID")
	})
}

func TestWorkspacesUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	wsTest, _ := createWorkspace(t, client, envTest)

	t.Run("when updating a subset of values", func(t *testing.T) {
		options := WorkspaceUpdateOptions{
			Name:             String(wsTest.Name),
			AutoApply:        Bool(true),
			Operations:       Bool(true),
			TerraformVersion: String("0.12.25"),
		}

		wsAfter, err := client.Workspaces.Update(ctx, wsTest.ID, options)
		require.NoError(t, err)

		assert.Equal(t, wsTest.Name, wsAfter.Name)
		assert.NotEqual(t, wsTest.AutoApply, wsAfter.AutoApply)
		assert.NotEqual(t, wsTest.TerraformVersion, wsAfter.TerraformVersion)
		assert.Equal(t, wsTest.WorkingDirectory, wsAfter.WorkingDirectory)
	})

	t.Run("with valid options", func(t *testing.T) {
		options := WorkspaceUpdateOptions{
			Name:             String(randomString(t)),
			AutoApply:        Bool(false),
			Operations:       Bool(false),
			TerraformVersion: String("0.12.25"),
			WorkingDirectory: String("baz/"),
		}

		w, err := client.Workspaces.Update(ctx, wsTest.ID, options)
		require.NoError(t, err)

		// Get a refreshed view of the workspace from the API
		refreshed, err := client.Workspaces.Read(ctx, envTest.ID, *options.Name)
		require.NoError(t, err)

		for _, item := range []*Workspace{
			w,
			refreshed,
		} {
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, *options.AutoApply, item.AutoApply)
			assert.Equal(t, *options.Operations, item.Operations)
			assert.Equal(t, *options.TerraformVersion, item.TerraformVersion)
			assert.Equal(t, *options.WorkingDirectory, item.WorkingDirectory)
		}
	})

	t.Run("when an error is returned from the api", func(t *testing.T) {
		w, err := client.Workspaces.Update(ctx, wsTest.ID, WorkspaceUpdateOptions{
			TerraformVersion: String("nonexisting"),
		})
		assert.Nil(t, w)
		assert.Error(t, err)
	})

	t.Run("when options has an invalid name", func(t *testing.T) {
		w, err := client.Workspaces.Update(ctx, badIdentifier, WorkspaceUpdateOptions{})
		assert.Nil(t, w)
		assert.EqualError(t, err, "invalid value for workspace ID")
	})

}

func TestWorkspacesUpdateByID(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	wTest, _ := createWorkspace(t, client, envTest)

	t.Run("when updating a subset of values", func(t *testing.T) {
		options := WorkspaceUpdateOptions{
			Name:             String(wTest.Name),
			AutoApply:        Bool(true),
			Operations:       Bool(true),
			TerraformVersion: String("0.12.25"),
		}

		wAfter, err := client.Workspaces.Update(ctx, wTest.ID, options)
		require.NoError(t, err)

		assert.Equal(t, wTest.Name, wAfter.Name)
		assert.NotEqual(t, wTest.AutoApply, wAfter.AutoApply)
		assert.NotEqual(t, wTest.TerraformVersion, wAfter.TerraformVersion)
		assert.Equal(t, wTest.WorkingDirectory, wAfter.WorkingDirectory)
	})

	t.Run("with valid options", func(t *testing.T) {
		options := WorkspaceUpdateOptions{
			Name:             String(randomString(t)),
			AutoApply:        Bool(false),
			Operations:       Bool(false),
			TerraformVersion: String("0.12.25"),
			WorkingDirectory: String("baz/"),
		}

		w, err := client.Workspaces.Update(ctx, wTest.ID, options)
		require.NoError(t, err)

		// Get a refreshed view of the workspace from the API
		refreshed, err := client.Workspaces.Read(ctx, envTest.ID, *options.Name)
		require.NoError(t, err)

		for _, item := range []*Workspace{
			w,
			refreshed,
		} {
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, *options.AutoApply, item.AutoApply)
			assert.Equal(t, *options.Operations, item.Operations)
			assert.Equal(t, *options.TerraformVersion, item.TerraformVersion)
			assert.Equal(t, *options.WorkingDirectory, item.WorkingDirectory)
		}
	})

	t.Run("when an error is returned from the api", func(t *testing.T) {
		w, err := client.Workspaces.Update(ctx, wTest.ID, WorkspaceUpdateOptions{
			TerraformVersion: String("nonexisting"),
		})
		assert.Nil(t, w)
		assert.Error(t, err)
	})

	t.Run("without a valid workspace ID", func(t *testing.T) {
		w, err := client.Workspaces.Update(ctx, badIdentifier, WorkspaceUpdateOptions{})
		assert.Nil(t, w)
		assert.EqualError(t, err, "invalid value for workspace ID")
	})
}

func TestWorkspacesDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	wTest, _ := createWorkspace(t, client, envTest)

	t.Run("with valid options", func(t *testing.T) {
		err := client.Workspaces.Delete(ctx, wTest.ID)
		require.NoError(t, err)

		// Try loading the workspace - it should fail.
		_, err = client.Workspaces.ReadByID(ctx, wTest.ID)
		assert.Equal(t, ErrResourceNotFound, err)
	})

	t.Run("without a valid workspace ID", func(t *testing.T) {
		err := client.Workspaces.Delete(ctx, badIdentifier)
		assert.EqualError(t, err, "invalid value for workspace ID")
	})
}
