package scalr

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurationVersionsCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	wsTest, wsTestCleanup := createWorkspace(t, client, nil)
	defer wsTestCleanup()

	t.Run("with valid options", func(t *testing.T) {
		cv, err := client.ConfigurationVersions.Create(ctx,
			ConfigurationVersionCreateOptions{Workspace: wsTest},
		)
		require.NoError(t, err)
		assert.Equal(t, ConfigurationPending, cv.Status)
		assert.Contains(t, cv.Links["upload"], "blobs/")

		// Get a refreshed view of the configuration version.
		refreshed, err := client.ConfigurationVersions.Read(ctx, cv.ID)
		require.NoError(t, err)
		assert.Equal(t, cv.ID, refreshed.ID)
		assert.Equal(t, ConfigurationPending, refreshed.Status)
		assert.Equal(t, cv.Workspace, refreshed.Workspace)
		assert.Equal(t, cv.Links["self"], refreshed.Links["self"])
		assert.Equal(t, "", refreshed.Links["upload"])
	})
	t.Run("when no workspace is provided", func(t *testing.T) {
		_, err := client.ConfigurationVersions.Create(ctx, ConfigurationVersionCreateOptions{})
		assert.EqualError(t, err, "workspace is required")
	})

	t.Run("with invalid workspace id", func(t *testing.T) {
		cv, err := client.ConfigurationVersions.Create(
			ctx,
			ConfigurationVersionCreateOptions{Workspace: &Workspace{ID: badIdentifier}},
		)
		assert.Nil(t, cv)
		assert.EqualError(t, err, "invalid value for workspace ID")
	})

	t.Run("with uploaded blob", func(t *testing.T) {
		cv, err := client.ConfigurationVersions.Create(ctx,
			ConfigurationVersionCreateOptions{Workspace: wsTest},
		)
		require.NoError(t, err)
		uploadBlob(cv.Links["upload"])

		i := 1
		for ; i <= 10; i++ {
			refreshed, err := client.ConfigurationVersions.Read(ctx, cv.ID)
			require.NoError(t, err)
			assert.Equal(t, cv.ID, refreshed.ID)
			if refreshed.Status == ConfigurationUploaded {
				return
			}
			time.Sleep(time.Second)
		}
		assert.NotEqual(t, 10, i)
	})
}

func TestConfigurationVersionsRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	cvTest, cvCleanup := createConfigurationVersion(t, client, nil)
	defer cvCleanup()

	t.Run("when the configuration version exists", func(t *testing.T) {
		cv, err := client.ConfigurationVersions.Read(ctx, cvTest.ID)
		require.NoError(t, err)
		assert.Equal(t, cvTest, cv)
	})

	t.Run("when the configuration version does not exist", func(t *testing.T) {
		var cvName = "nonexisting"
		cv, err := client.ConfigurationVersions.Read(ctx, cvName)
		assert.Nil(t, cv)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("ConfigurationVersion with ID '%s' not found or user unauthorized.", cvName),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("with invalid configuration version id", func(t *testing.T) {
		cv, err := client.ConfigurationVersions.Read(ctx, badIdentifier)
		assert.Nil(t, cv)
		assert.EqualError(t, err, "invalid value for configuration version ID")
	})
}
