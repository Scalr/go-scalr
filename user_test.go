package scalr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultUserID                  = "user-svrcmmpcrkmit1g"
	defaultUserEmail               = "tf-admin@scalr.com"
	defaultIdentityProviderScalrID = "idp-sohkb0o1phrdmr8"
)

func TestUsersList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("with empty options", func(t *testing.T) {
		ul, err := client.Users.List(ctx, UserListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 1, ul.CurrentPage)
		assert.True(t, ul.TotalCount >= 1)
	})

	t.Run("with page size option", func(t *testing.T) {
		ul, err := client.Users.List(ctx, UserListOptions{
			ListOptions: ListOptions{
				PageSize: 99,
			},
		})
		require.NoError(t, err)

		var uIDs []string
		for _, u := range ul.Items {
			uIDs = append(uIDs, u.ID)
		}
		assert.Equal(t, 1, ul.CurrentPage)
		assert.Contains(t, uIDs, defaultUserID)
	})

	t.Run("with user filter", func(t *testing.T) {
		ul, err := client.Users.List(ctx, UserListOptions{
			User: String(defaultUserID),
		})
		require.NoError(t, err)
		assert.Equal(t, 1, ul.CurrentPage)
		assert.Equal(t, 1, ul.TotalCount)
		assert.Equal(t, defaultUserID, ul.Items[0].ID)
	})

	t.Run("with email filter", func(t *testing.T) {
		ul, err := client.Users.List(ctx, UserListOptions{
			Email: String(defaultUserEmail),
		})
		require.NoError(t, err)
		assert.Equal(t, 1, ul.CurrentPage)
		assert.Equal(t, 1, ul.TotalCount)
		assert.Equal(t, defaultUserID, ul.Items[0].ID)
	})

	t.Run("without a valid user", func(t *testing.T) {
		ul, err := client.Users.List(ctx, UserListOptions{
			User: String(badIdentifier),
		})
		assert.NoError(t, err)
		assert.Len(t, ul.Items, 0)
	})

	t.Run("without a valid identity provider", func(t *testing.T) {
		ul, err := client.Users.List(ctx, UserListOptions{
			IdentityProvider: String(badIdentifier),
		})
		assert.NoError(t, err)
		assert.Len(t, ul.Items, 0)
	})
}

func TestUsersRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("when the user exists", func(t *testing.T) {
		u, err := client.Users.Read(ctx, defaultUserID)
		require.NoError(t, err)
		assert.Equal(t, defaultUserID, u.ID)

		t.Run("relationships are properly decoded", func(t *testing.T) {
			assert.Equal(t, u.Teams[0].ID, defaultTeamID)
		})
	})

	t.Run("when the user does not exist", func(t *testing.T) {
		u, err := client.Users.Read(ctx, "user-nonexisting")
		assert.Nil(t, u)
		assert.Error(t, err)
	})

	t.Run("without a valid user ID", func(t *testing.T) {
		u, err := client.Users.Read(ctx, badIdentifier)
		assert.Nil(t, u)
		assert.EqualError(t, err, "invalid value for user ID")
	})
}
