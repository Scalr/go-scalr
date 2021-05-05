package scalr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersList(t *testing.T) {
	client := testClient(t)
	client.headers.Set("Prefer", "profile=internal")
	ctx := context.Background()

	usrl, err := client.Users.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	totalCount := usrl.TotalCount
	usrTest1, usrTest1Cleanup := createUser(t, client)
	defer usrTest1Cleanup()

	t.Run("with no list options", func(t *testing.T) {
		usrl, err := client.Users.List(ctx)
		usrlIDs := make([]string, len(usrl.Items))
		for _, usr := range usrl.Items {
			usrlIDs = append(usrlIDs, usr.ID)
		}
		require.NoError(t, err)
		assert.Contains(t, usrlIDs, usrTest1.ID)

		assert.Equal(t, 1, usrl.CurrentPage)
		assert.Equal(t, 1+totalCount, usrl.TotalCount)
	})

}

func TestUsersCreate(t *testing.T) {
	client := testClient(t)
	client.headers.Set("Prefer", "profile=internal")
	ctx := context.Background()
	t.Run("when no email is provided", func(t *testing.T) {
		_, err := client.Users.Create(ctx, UserCreateOptions{
			IdentityProvider: &IdentityProvider{ID: defaultIdentityProviderID},
			Username:         String("tst-" + randomString(t)),
			Status:           UserStatusActive,
		})
		assert.EqualError(t, err, "email is required")
	})
	t.Run("when invalid email is provided", func(t *testing.T) {
		_, err := client.Users.Create(ctx, UserCreateOptions{
			Email:            String("go-scalr-test&scalr.com"),
			IdentityProvider: &IdentityProvider{ID: defaultIdentityProviderID},
			Username:         String("tst-" + randomString(t)),
			Status:           UserStatusActive,
		})
		assert.EqualError(t, err, "invalid value for email")
	})
	t.Run("when no identity provider is provided", func(t *testing.T) {
		_, err := client.Users.Create(ctx, UserCreateOptions{
			Email:    String("go-scalr-test@scalr.com"),
			Username: String("tst-" + randomString(t)),
			Status:   UserStatusActive,
		})
		assert.EqualError(t, err, "identity provider is required")
	})
	t.Run("with invalid identity-provider id", func(t *testing.T) {
		usr, err := client.Users.Create(ctx, UserCreateOptions{
			Email:            String("go-scalr-test@scalr.com"),
			IdentityProvider: &IdentityProvider{ID: badIdentifier},
			Username:         String("tst-" + randomString(t)),
			Status:           UserStatusActive,
		})
		assert.Nil(t, usr)
		assert.EqualError(t, err, "invalid value for identity provider ID")
	})
	t.Run("with valid options", func(t *testing.T) {
		options := UserCreateOptions{
			IdentityProvider: &IdentityProvider{ID: defaultIdentityProviderID},
			Email:            String("go-scalr-test@scalr.com"),
			Status:           UserStatusActive,
		}

		usr, err := client.Users.Create(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		// Get a refreshed view of the user
		_, err = client.Users.Read(ctx, usr.ID)
		require.NoError(t, err)

		defer client.Users.Delete(ctx, usr.ID)

		assert.Equal(t, *options.Email, usr.Email)
		assert.Equal(t, options.Status, usr.Status)
		assert.Equal(t, (*options.IdentityProvider).ID, (*usr.IdentityProvider).ID)
	})

}

func TestUsersRead(t *testing.T) {
	client := testClient(t)
	client.headers.Set("Prefer", "profile=internal")
	ctx := context.Background()

	usrTest, usrTestCleanup := createUser(t, client)
	defer usrTestCleanup()
	t.Run("when the user exists", func(t *testing.T) {
		_, err := client.Users.Read(ctx, usrTest.ID)
		require.NoError(t, err)
	})

	t.Run("when the user does not exist", func(t *testing.T) {
		_, err := client.Users.Read(ctx, "notexisting")
		assert.Equal(t, err, ErrResourceNotFound)
	})

	t.Run("with invalid usr ID", func(t *testing.T) {
		r, err := client.Users.Read(ctx, badIdentifier)
		assert.Nil(t, r)
		assert.EqualError(t, err, "invalid value for user ID")
	})
}

func TestUsersUpdate(t *testing.T) {
	client := testClient(t)
	client.headers.Set("Prefer", "profile=internal")
	ctx := context.Background()

	t.Run("with valid options", func(t *testing.T) {
		usrTest, usrTestCleanup := createUser(t, client)

		options := UserUpdateOptions{
			FullName: String("Leto Atreides"),
			Status:   UserStatusInactive,
		}

		usr, err := client.Users.Update(ctx, usrTest.ID, options)
		if err != nil {
			usrTestCleanup()
		}
		require.NoError(t, err)

		// Make sure we clean up the updated usr.
		defer client.Users.Delete(ctx, usr.ID)

		// Also get a fresh result from the API to ensure we get the
		// expected values back.
		refreshed, err := client.Users.Read(ctx, usr.ID)
		require.NoError(t, err)

		for _, item := range []*User{
			usr,
			refreshed,
		} {
			assert.Equal(t, *options.FullName, item.FullName)
			assert.Equal(t, options.Status, item.Status)
		}
	})

	t.Run("when only updating a subset of fields", func(t *testing.T) {
		usrTest, usrTestCleanup := createUser(t, client)
		defer usrTestCleanup()

		usr, err := client.Users.Update(ctx, usrTest.ID, UserUpdateOptions{})
		require.NoError(t, err)
		assert.Equal(t, usrTest.FullName, usr.FullName)
	})
}

func TestUsersDelete(t *testing.T) {
	client := testClient(t)
	client.headers.Set("Prefer", "profile=internal")
	ctx := context.Background()

	t.Run("with valid options", func(t *testing.T) {
		usrTest, _ := createUser(t, client)

		err := client.Users.Delete(ctx, usrTest.ID)
		require.NoError(t, err)

		// Try fetching the usr again - it should error.
		_, err = client.Users.Read(ctx, usrTest.ID)
		assert.Equal(t, err, ErrResourceNotFound)
	})

	t.Run("when the usr does not exist", func(t *testing.T) {
		err := client.Users.Delete(ctx, randomString(t))
		assert.Equal(t, err, ErrResourceNotFound)
	})
}
