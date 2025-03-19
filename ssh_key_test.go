package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHKeysCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("success creating SSH key", func(t *testing.T) {
		options := SSHKeyCreateOptions{
			Name: String("test_ssh_key"),
			PrivateKey: String(`-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIBvMDyNaYtWK2TmJIfFhmPZeGxK0bWnNDhjlTZ+V6e4x
-----END PRIVATE KEY-----`),
			IsShared: Bool(true),
			Account:  &Account{ID: defaultAccountID},
		}

		sshKey, err := client.SSHKeys.Create(ctx, options)
		require.NoError(t, err)
		defer client.SSHKeys.Delete(ctx, sshKey.ID)

		assert.Equal(t, *options.Name, sshKey.Name)
		assert.Equal(t, *options.IsShared, sshKey.IsShared)
		assert.NotEmpty(t, sshKey.ID)
	})
}

func TestSSHKeysUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	sshKey, sshKeyCleanup := createSSHKey(t, client, "test_ssh_key_for_update", true, "")
	defer sshKeyCleanup()

	t.Run("success updating SSH key", func(t *testing.T) {
		updateOptions := SSHKeyUpdateOptions{
			Name:     String("updated_test_ssh_key"),
			IsShared: Bool(false),
		}
		updatedSSHKey, err := client.SSHKeys.Update(ctx, sshKey.ID, updateOptions)
		require.NoError(t, err)

		assert.Equal(t, *updateOptions.Name, updatedSSHKey.Name)
		assert.Equal(t, *updateOptions.IsShared, updatedSSHKey.IsShared)
		assert.Equal(t, sshKey.ID, updatedSSHKey.ID)
	})
}

func TestSSHKeysDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	sshKey, _ := createSSHKey(t, client, "test_ssh_key_for_read", true, "")

	t.Run("success deleting SSH key", func(t *testing.T) {
		err := client.SSHKeys.Delete(ctx, sshKey.ID)
		require.NoError(t, err)

		_, err = client.SSHKeys.Read(ctx, sshKey.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("AccountSSHKey with ID '%s' not found or user unauthorized.", sshKey.ID),
			}.Error(),
			err.Error(),
		)
	})
}

func TestSSHKeysList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	sshKey1, sshKey1Cleanup := createSSHKey(t, client, "test_ssh_key1", true, "")
	defer sshKey1Cleanup()

	sshKey2, sshKey2Cleanup := createSSHKey(t, client, "test_ssh_key2", false, `-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEICNioyJgilYaHbT8pgDXn3haYU0dsl6KJTIvrZm+nIU6
-----END PRIVATE KEY-----`)
	defer sshKey2Cleanup()

	t.Run("list all SSH keys", func(t *testing.T) {
		sshKeysList, err := client.SSHKeys.List(ctx, SSHKeysListOptions{})
		require.NoError(t, err)

		expectedIDs := []string{sshKey1.ID, sshKey2.ID}
		actualIDs := make([]string, len(sshKeysList.Items))
		for i, key := range sshKeysList.Items {
			actualIDs[i] = key.ID
		}

		assert.ElementsMatch(t, expectedIDs, actualIDs)
	})
}

func TestSSHKeysRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("success reading SSH key", func(t *testing.T) {
		sshKey, sshKeyCleanup := createSSHKey(t, client, "test_ssh_key_for_read", true, "")
		defer sshKeyCleanup()

		readSSHKey, err := client.SSHKeys.Read(ctx, sshKey.ID)
		require.NoError(t, err)

		assert.Equal(t, sshKey.ID, readSSHKey.ID)
		assert.Equal(t, sshKey.Name, readSSHKey.Name)
		assert.Equal(t, sshKey.IsShared, readSSHKey.IsShared)
	})
}
