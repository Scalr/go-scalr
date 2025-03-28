package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentHooksCreate(t *testing.T) {
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
		hookEvents := []string{"pre-plan", "post-plan"}
		createOptions := EnvironmentHookCreateOptions{
			Events:      hookEvents,
			Environment: environment,
			Hook:        hook,
		}

		envHook, err := client.EnvironmentHooks.Create(ctx, createOptions)
		defer func() { client.EnvironmentHooks.Delete(ctx, envHook.ID) }()

		require.NoError(t, err)

		envHook, err = client.EnvironmentHooks.Read(ctx, envHook.ID)
		require.NoError(t, err)

		assert.Equal(t, createOptions.Events, envHook.Events)
		assert.Equal(t, environment.ID, envHook.Environment.ID)
		assert.Equal(t, hook.ID, envHook.Hook.ID)
		assert.Equal(t, createOptions.Events, envHook.Events)
	})
}

func TestEnvironmentHooksRead(t *testing.T) {
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
		hookEvents := []string{"pre-plan", "post-plan"}
		createOptions := EnvironmentHookCreateOptions{
			Events:      hookEvents,
			Environment: environment,
			Hook:        hook,
		}

		envHook, err := client.EnvironmentHooks.Create(ctx, createOptions)
		defer func() { client.EnvironmentHooks.Delete(ctx, envHook.ID) }()

		require.NoError(t, err)

		readEnvHook, err := client.EnvironmentHooks.Read(ctx, envHook.ID)
		require.NoError(t, err)

		assert.Equal(t, envHook.ID, readEnvHook.ID)
		assert.Equal(t, createOptions.Events, readEnvHook.Events)
		assert.Equal(t, environment.ID, readEnvHook.Environment.ID)
		assert.Equal(t, hook.ID, readEnvHook.Hook.ID)
	})
}

func TestEnvironmentHooksUpdate(t *testing.T) {
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
		createOptions := EnvironmentHookCreateOptions{
			Events:      initialEvents,
			Environment: environment,
			Hook:        hook,
		}

		envHook, err := client.EnvironmentHooks.Create(ctx, createOptions)
		defer func() { client.EnvironmentHooks.Delete(ctx, envHook.ID) }()

		require.NoError(t, err)

		updatedEvents := []string{"pre-plan", "post-plan"}
		updateOptions := EnvironmentHookUpdateOptions{
			Events: &updatedEvents,
		}

		updatedEnvHook, err := client.EnvironmentHooks.Update(ctx, envHook.ID, updateOptions)
		require.NoError(t, err)

		assert.Equal(t, envHook.ID, updatedEnvHook.ID)
		assert.Equal(t, *updateOptions.Events, updatedEnvHook.Events)
	})
}

func TestEnvironmentHooksDelete(t *testing.T) {
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
		hookEvents := []string{"pre-plan"}
		createOptions := EnvironmentHookCreateOptions{
			Events:      hookEvents,
			Environment: environment,
			Hook:        hook,
		}

		envHook, err := client.EnvironmentHooks.Create(ctx, createOptions)
		defer func() { client.EnvironmentHooks.Delete(ctx, envHook.ID) }()

		require.NoError(t, err)

		// Delete the hook
		err = client.EnvironmentHooks.Delete(ctx, envHook.ID)
		require.NoError(t, err)

		// Try loading the hook - it should fail
		_, err = client.EnvironmentHooks.Read(ctx, envHook.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("HookEnvironmentLink with ID '%s' not found or user unauthorized.", envHook.ID),
			}.Error(),
			err.Error(),
		)
	})
}

func TestEnvironmentHooksList(t *testing.T) {
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
		hook1Options := EnvironmentHookCreateOptions{
			Events:      events1,
			Environment: environment,
			Hook:        hook1,
		}

		envHook1, err := client.EnvironmentHooks.Create(ctx, hook1Options)
		defer client.EnvironmentHooks.Delete(ctx, envHook1.ID)

		require.NoError(t, err)

		events2 := []string{"post-plan"}
		hook2Options := EnvironmentHookCreateOptions{
			Events:      events2,
			Environment: environment,
			Hook:        hook2,
		}

		envHook2, err := client.EnvironmentHooks.Create(ctx, hook2Options)
		defer client.EnvironmentHooks.Delete(ctx, envHook2.ID)

		require.NoError(t, err)

		hook3Options := EnvironmentHookCreateOptions{
			Events:      events2,
			Environment: environment2,
			Hook:        hook3,
		}

		envHook3, err := client.EnvironmentHooks.Create(ctx, hook3Options)
		defer client.EnvironmentHooks.Delete(ctx, envHook3.ID)

		require.NoError(t, err)

		// List hooks with environment filter
		listOptions := EnvironmentHookListOptions{
			Environment: String(environment.ID),
		}

		hookList, err := client.EnvironmentHooks.List(ctx, listOptions)
		require.NoError(t, err)

		found1 := false
		found2 := false
		found3 := false
		for _, h := range hookList.Items {
			if h.ID == envHook1.ID {
				found1 = true
			}
			if h.ID == envHook2.ID {
				found2 = true
			}
			if h.ID == envHook3.ID {
				found3 = true
			}
		}
		assert.True(t, found1, "Hook for hook %s should be in the list", hook1.Name)
		assert.True(t, found2, "Hook for hook %s should be in the list", hook2.Name)
		assert.False(t, found3, "Hook for hook %s should not be in the list", hook3.Name)
	})
}
