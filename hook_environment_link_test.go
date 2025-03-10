package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHookEnvironmentLinksCreate(t *testing.T) {
	t.Skip("Works with personal token but does not work with github action token.")

	client := testClient(t)
	ctx := context.Background()

	environment, removeEnvironment := createEnvironment(t, client)
	defer removeEnvironment()

	vcsProvider, removeVcsProvider := createVcsProvider(t, client, []*Environment{environment})
	defer removeVcsProvider()

	hook, removeHook := createHook(t, client, vcsProvider)
	defer removeHook()

	t.Run("with valid options", func(t *testing.T) {
		linkEvents := []string{"pre-plan", "post-plan"}
		createOptions := HookEnvironmentLinkCreateOptions{
			Events:      &linkEvents,
			Environment: environment,
			Hook:        hook,
		}

		link, err := client.HookEnvironmentLinks.Create(ctx, createOptions)
		defer func() { client.HookEnvironmentLinks.Delete(ctx, link.ID) }()

		require.NoError(t, err)

		link, err = client.HookEnvironmentLinks.Read(ctx, link.ID)
		require.NoError(t, err)

		assert.Equal(t, *createOptions.Events, link.Events)
		assert.Equal(t, environment.ID, link.Environment.ID)
		assert.Equal(t, hook.ID, link.Hook.ID)
		assert.Equal(t, *createOptions.Events, link.Events)
	})
}

func TestHookEnvironmentLinksRead(t *testing.T) {
	t.Skip("Works with personal token but does not work with github action token.")

	client := testClient(t)
	ctx := context.Background()

	environment, removeEnvironment := createEnvironment(t, client)
	defer removeEnvironment()

	vcsProvider, removeVcsProvider := createVcsProvider(t, client, []*Environment{environment})
	defer removeVcsProvider()

	hook, removeHook := createHook(t, client, vcsProvider)
	defer removeHook()

	t.Run("with valid ID", func(t *testing.T) {
		linkEvents := []string{"pre-plan", "post-plan"}
		createOptions := HookEnvironmentLinkCreateOptions{
			Events:      &linkEvents,
			Environment: environment,
			Hook:        hook,
		}

		link, err := client.HookEnvironmentLinks.Create(ctx, createOptions)
		defer func() { client.HookEnvironmentLinks.Delete(ctx, link.ID) }()

		require.NoError(t, err)

		readLink, err := client.HookEnvironmentLinks.Read(ctx, link.ID)
		require.NoError(t, err)

		assert.Equal(t, link.ID, readLink.ID)
		assert.Equal(t, *createOptions.Events, readLink.Events)
		assert.Equal(t, environment.ID, readLink.Environment.ID)
		assert.Equal(t, hook.ID, readLink.Hook.ID)
	})
}

func TestHookEnvironmentLinksUpdate(t *testing.T) {
	t.Skip("Works with personal token but does not work with github action token.")

	client := testClient(t)
	ctx := context.Background()

	environment, removeEnvironment := createEnvironment(t, client)
	defer removeEnvironment()

	vcsProvider, removeVcsProvider := createVcsProvider(t, client, []*Environment{environment})
	defer removeVcsProvider()

	hook, removeHook := createHook(t, client, vcsProvider)
	defer removeHook()

	t.Run("with valid options", func(t *testing.T) {
		initialEvents := []string{"pre-plan"}
		createOptions := HookEnvironmentLinkCreateOptions{
			Events:      &initialEvents,
			Environment: environment,
			Hook:        hook,
		}

		link, err := client.HookEnvironmentLinks.Create(ctx, createOptions)
		defer func() { client.HookEnvironmentLinks.Delete(ctx, link.ID) }()

		require.NoError(t, err)

		updatedEvents := []string{"pre-plan", "post-plan"}
		updateOptions := HookEnvironmentLinkUpdateOptions{
			Events: &updatedEvents,
		}

		updatedLink, err := client.HookEnvironmentLinks.Update(ctx, link.ID, updateOptions)
		require.NoError(t, err)

		assert.Equal(t, link.ID, updatedLink.ID)
		assert.Equal(t, *updateOptions.Events, updatedLink.Events)
	})
}

func TestHookEnvironmentLinksDelete(t *testing.T) {
	t.Skip("Works with personal token but does not work with github action token.")

	client := testClient(t)
	ctx := context.Background()

	environment, removeEnvironment := createEnvironment(t, client)
	defer removeEnvironment()

	vcsProvider, removeVcsProvider := createVcsProvider(t, client, []*Environment{environment})
	defer removeVcsProvider()

	hook, removeHook := createHook(t, client, vcsProvider)
	defer removeHook()

	t.Run("success", func(t *testing.T) {
		linkEvents := []string{"pre-plan"}
		createOptions := HookEnvironmentLinkCreateOptions{
			Events:      &linkEvents,
			Environment: environment,
			Hook:        hook,
		}

		link, err := client.HookEnvironmentLinks.Create(ctx, createOptions)
		defer func() { client.HookEnvironmentLinks.Delete(ctx, link.ID) }()

		require.NoError(t, err)

		// Delete the link
		err = client.HookEnvironmentLinks.Delete(ctx, link.ID)
		require.NoError(t, err)

		// Try loading the link - it should fail
		_, err = client.HookEnvironmentLinks.Read(ctx, link.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("HookEnvironmentLink with ID '%s' not found or user unauthorized.", link.ID),
			}.Error(),
			err.Error(),
		)
	})
}

func TestHookEnvironmentLinksList(t *testing.T) {
	t.Skip("Works with personal token but does not work with github action token.")

	client := testClient(t)
	ctx := context.Background()

	environment, removeEnvironment := createEnvironment(t, client)
	defer removeEnvironment()

	environment2, removeEnvironment2 := createEnvironment(t, client)
	defer removeEnvironment2()

	vcsProvider, removeVcsProvider := createVcsProvider(t, client, []*Environment{environment})
	defer removeVcsProvider()

	hook1, removeHook1 := createHook(t, client, vcsProvider)
	defer removeHook1()

	hook2, removeHook2 := createHook(t, client, vcsProvider)
	defer removeHook2()

	hook3, removeHook3 := createHook(t, client, vcsProvider)
	defer removeHook3()

	t.Run("with required environment filter", func(t *testing.T) {
		events1 := []string{"pre-plan"}
		link1Options := HookEnvironmentLinkCreateOptions{
			Events:      &events1,
			Environment: environment,
			Hook:        hook1,
		}

		link1, err := client.HookEnvironmentLinks.Create(ctx, link1Options)
		defer client.HookEnvironmentLinks.Delete(ctx, link1.ID)

		require.NoError(t, err)

		events2 := []string{"post-plan"}
		link2Options := HookEnvironmentLinkCreateOptions{
			Events:      &events2,
			Environment: environment,
			Hook:        hook2,
		}

		link2, err := client.HookEnvironmentLinks.Create(ctx, link2Options)
		defer client.HookEnvironmentLinks.Delete(ctx, link2.ID)

		require.NoError(t, err)

		link3Options := HookEnvironmentLinkCreateOptions{
			Events:      &events2,
			Environment: environment2,
			Hook:        hook3,
		}

		link3, err := client.HookEnvironmentLinks.Create(ctx, link3Options)
		defer client.HookEnvironmentLinks.Delete(ctx, link3.ID)

		require.NoError(t, err)

		// List links with environment filter
		listOptions := HookEnvironmentLinkListOptions{
			Environment: String(environment.ID),
		}

		linkList, err := client.HookEnvironmentLinks.List(ctx, listOptions)
		require.NoError(t, err)

		found1 := false
		found2 := false
		found3 := false
		for _, l := range linkList.Items {
			if l.ID == link1.ID {
				found1 = true
			}
			if l.ID == link2.ID {
				found2 = true
			}
			if l.ID == link3.ID {
				found3 = true
			}
		}
		assert.True(t, found1, "Link for hook %s should be in the list", hook1.Name)
		assert.True(t, found2, "Link for hook %s should be in the list", hook2.Name)
		assert.False(t, found3, "Link for hook %s should not be in the list", hook3.Name)
	})
}
