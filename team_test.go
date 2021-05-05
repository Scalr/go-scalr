package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamsList(t *testing.T) {
	client := testClient(t)
	client.headers.Set("Prefer", "profile=internal")
	ctx := context.Background()

	teaml, err := client.Teams.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	totalCount := teaml.TotalCount
	teamTest1, teamTest1Cleanup := createTeam(t, client)
	defer teamTest1Cleanup()

	t.Run("with no list options", func(t *testing.T) {
		teaml, err := client.Teams.List(ctx)
		teamlIDs := make([]string, len(teaml.Items))
		for _, team := range teaml.Items {
			teamlIDs = append(teamlIDs, team.ID)
		}
		require.NoError(t, err)
		assert.Contains(t, teamlIDs, teamTest1.ID)

		assert.Equal(t, 1, teaml.CurrentPage)
		assert.Equal(t, 1+totalCount, teaml.TotalCount)
	})

}

func TestTeamsCreate(t *testing.T) {
	client := testClient(t)
	client.headers.Set("Prefer", "profile=internal")
	ctx := context.Background()
	t.Run("when no name is provided", func(t *testing.T) {
		_, err := client.Teams.Create(ctx, TeamCreateOptions{
			IdentityProvider: &IdentityProvider{ID: defaultIdentityProviderID},
			Account:          &Account{ID: defaultAccountID},
		})
		assert.EqualError(t, err, "name is required")
	})
	t.Run("when no identity provider is provided", func(t *testing.T) {
		_, err := client.Teams.Create(ctx, TeamCreateOptions{
			Account: &Account{ID: defaultAccountID},
			Name:    String("tst-" + randomString(t)),
		})
		assert.EqualError(t, err, "identity provider is required")
	})
	t.Run("with invalid identity-provider id", func(t *testing.T) {
		team, err := client.Teams.Create(ctx, TeamCreateOptions{
			Account:          &Account{ID: defaultAccountID},
			Name:             String("tst-" + randomString(t)),
			IdentityProvider: &IdentityProvider{ID: badIdentifier},
		})
		assert.Nil(t, team)
		assert.EqualError(t, err, "invalid value for identity provider ID")
	})
	t.Run("when no account is provided", func(t *testing.T) {
		_, err := client.Teams.Create(ctx, TeamCreateOptions{
			IdentityProvider: &IdentityProvider{ID: defaultIdentityProviderID},
			Name:             String("tst-" + randomString(t)),
		})
		assert.EqualError(t, err, "account is required")
	})
	t.Run("with invalid account id", func(t *testing.T) {
		team, err := client.Teams.Create(ctx, TeamCreateOptions{
			Account:          &Account{ID: badIdentifier},
			Name:             String("tst-" + randomString(t)),
			IdentityProvider: &IdentityProvider{ID: defaultIdentityProviderID},
		})
		assert.Nil(t, team)
		assert.EqualError(t, err, "invalid value for account ID")
	})
	t.Run("with valid options", func(t *testing.T) {
		options := TeamCreateOptions{
			Account:          &Account{ID: defaultAccountID},
			Name:             String("tst-" + randomString(t)),
			IdentityProvider: &IdentityProvider{ID: defaultIdentityProviderID},
			Description:      String("Team created by go-scalr tests."),
			Users:            []*User{{ID: defaultUserID}},
		}

		team, err := client.Teams.Create(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		// Get a refreshed view of the team
		_, err = client.Teams.Read(ctx, team.ID)
		require.NoError(t, err)

		defer client.Teams.Delete(ctx, team.ID)

		assert.Equal(t, *options.Name, team.Name)
		assert.Equal(t, *options.Description, team.Description)
		assert.Equal(t, options.IdentityProvider.ID, team.IdentityProvider.ID)
		assert.Equal(t, options.Account.ID, team.Account.ID)
		assert.Equal(t, options.Users[0].ID, team.Users[0].ID)
	})

}

func TestTeamsRead(t *testing.T) {
	client := testClient(t)
	client.headers.Set("Prefer", "profile=internal")
	ctx := context.Background()

	teamTest, teamTestCleanup := createTeam(t, client)
	defer teamTestCleanup()
	t.Run("when the team exists", func(t *testing.T) {
		_, err := client.Teams.Read(ctx, teamTest.ID)
		require.NoError(t, err)
	})

	t.Run("when the team does not exist", func(t *testing.T) {
		_, err := client.Teams.Read(ctx, "notexisting")
		assert.Equal(t, err, ErrResourceNotFound)
	})

	t.Run("with invalid team ID", func(t *testing.T) {
		r, err := client.Teams.Read(ctx, badIdentifier)
		assert.Nil(t, r)
		assert.EqualError(t, err, "invalid value for team ID")
	})
}

func TestTeamsUpdate(t *testing.T) {
	client := testClient(t)
	client.headers.Set("Prefer", "profile=internal")
	ctx := context.Background()

	t.Run("with valid options", func(t *testing.T) {
		teamTest, teamTestCleanup := createTeam(t, client)

		options := TeamUpdateOptions{
			Name:        String(fmt.Sprintf("Updated name for team %s", teamTest.ID)),
			Description: String(fmt.Sprintf("Updated description for team %s", teamTest.ID)),
			Users:       []*User{},
		}

		team, err := client.Teams.Update(ctx, teamTest.ID, options)
		if err != nil {
			teamTestCleanup()
		}
		require.NoError(t, err)

		// Make sure we clean up the updated team.
		defer client.Teams.Delete(ctx, team.ID)

		// Also get a fresh result from the API to ensure we get the
		// expected values back.
		refreshed, err := client.Teams.Read(ctx, team.ID)
		require.NoError(t, err)

		for _, item := range []*Team{
			team,
			refreshed,
		} {
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, *options.Description, item.Description)
			//			assert.Equal(t, *options.Users[0], *item.Users[0])
		}
	})

	// this one is broken on server
	//	t.Run("when only updating a subset of fields", func(t *testing.T) {
	//		teamTest, teamTestCleanup := createTeam(t, client)
	//		defer teamTestCleanup()
	//
	//		team, err := client.Teams.Update(ctx, teamTest.ID, TeamUpdateOptions{Description: String("blah")})
	//		require.NoError(t, err)
	//		assert.Equal(t, teamTest.Name, team.Name)
	//		assert.Equal(t, "blah", team.Description)
	//	})
}

func TestTeamsDelete(t *testing.T) {
	client := testClient(t)
	client.headers.Set("Prefer", "profile=internal")
	ctx := context.Background()

	t.Run("with valid options", func(t *testing.T) {
		teamTest, _ := createTeam(t, client)

		err := client.Teams.Delete(ctx, teamTest.ID)
		require.NoError(t, err)

		// Try fetching the team again - it should error.
		_, err = client.Teams.Read(ctx, teamTest.ID)
		assert.Equal(t, err, ErrResourceNotFound)
	})

	t.Run("when the team does not exist", func(t *testing.T) {
		err := client.Teams.Delete(ctx, randomString(t))
		assert.Equal(t, err, ErrResourceNotFound)
	})
}
