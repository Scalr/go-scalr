package scalr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHooksCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	vcsProvider, vcsProviderCleanup := createVcsProvider(t, client, nil)
	defer vcsProviderCleanup()

	t.Run("with valid options", func(t *testing.T) {
		hookName := "test-hook"
		hookInterpreter := "bash"
		hookScriptfilePath := "pre-plan.sh"
		hookVcsRepo := &HookVcsRepo{
			Identifier: "Scalr/tf-revizor-fixtures",
			Branch:     "master",
		}

		options := HookCreateOptions{
			Name:           String(hookName),
			Interpreter:    String(hookInterpreter),
			ScriptfilePath: String(hookScriptfilePath),
			VcsRepo:        hookVcsRepo,
			VcsProvider:    vcsProvider,
			Account:        &Account{ID: defaultAccountID},
		}

		hook, err := client.Hooks.Create(ctx, options)
		defer func() { client.Hooks.Delete(ctx, hook.ID) }()

		require.NoError(t, err)

		// Read the hook to verify it was created correctly
		refreshed, err := client.Hooks.Read(ctx, hook.ID)
		require.NoError(t, err)

		assert.Equal(t, *options.Name, refreshed.Name)
		assert.Equal(t, *options.Interpreter, refreshed.Interpreter)
		assert.Equal(t, *options.ScriptfilePath, refreshed.ScriptfilePath)
		assert.Equal(t, options.VcsRepo.Identifier, refreshed.VcsRepo.Identifier)
		assert.Equal(t, options.VcsRepo.Branch, refreshed.VcsRepo.Branch)
		assert.Equal(t, options.VcsProvider.ID, refreshed.VcsProvider.ID)
	})

	t.Run("without vcs repo options", func(t *testing.T) {
		hook, err := client.Hooks.Create(ctx, HookCreateOptions{
			Name:           String("test-hook"),
			Interpreter:    String("bash"),
			ScriptfilePath: String("pre-plan.sh"),
			VcsProvider:    vcsProvider,
			Account:        &Account{ID: defaultAccountID},
		})
		assert.Nil(t, hook)
		assert.EqualError(t, err, "vcs repo is required")
	})

	t.Run("without vcs provider options", func(t *testing.T) {
		hook, err := client.Hooks.Create(ctx, HookCreateOptions{
			Name:           String("test-hook"),
			Interpreter:    String("bash"),
			ScriptfilePath: String("pre-plan.sh"),
			VcsRepo: &HookVcsRepo{
				Identifier: "Scalr/tf-revizor-fixtures",
				Branch:     "master",
			},
			Account: &Account{ID: defaultAccountID},
		})
		assert.Nil(t, hook)
		assert.EqualError(t, err, "vcs provider is required")
	})

	t.Run("without account options", func(t *testing.T) {
		hook, err := client.Hooks.Create(ctx, HookCreateOptions{
			Name:           String("test-hook"),
			Interpreter:    String("bash"),
			ScriptfilePath: String("pre-plan.sh"),
			VcsRepo: &HookVcsRepo{
				Identifier: "Scalr/tf-revizor-fixtures",
				Branch:     "master",
			},
			VcsProvider: vcsProvider,
		})
		assert.Nil(t, hook)
		assert.EqualError(t, err, "account is required")
	})

	t.Run("without interpreter options", func(t *testing.T) {
		hook, err := client.Hooks.Create(ctx, HookCreateOptions{
			Name:           String("test-hook"),
			ScriptfilePath: String("pre-plan.sh"),
			VcsRepo: &HookVcsRepo{
				Identifier: "Scalr/tf-revizor-fixtures",
				Branch:     "master",
			},
			VcsProvider: vcsProvider,
			Account:     &Account{ID: defaultAccountID},
		})
		assert.Nil(t, hook)
		assert.EqualError(t, err, "interpreter is required")
	})

	t.Run("without scriptfile path options", func(t *testing.T) {
		hook, err := client.Hooks.Create(ctx, HookCreateOptions{
			Name:        String("test-hook"),
			Interpreter: String("bash"),
			VcsRepo: &HookVcsRepo{
				Identifier: "Scalr/tf-revizor-fixtures",
				Branch:     "master",
			},
			VcsProvider: vcsProvider,
			Account:     &Account{ID: defaultAccountID},
		})
		assert.Nil(t, hook)
		assert.EqualError(t, err, "scriptfile path is required")
	})

	t.Run("without name options", func(t *testing.T) {
		hook, err := client.Hooks.Create(ctx, HookCreateOptions{
			Interpreter:    String("bash"),
			ScriptfilePath: String("pre-plan.sh"),
			VcsRepo: &HookVcsRepo{
				Identifier: "Scalr/tf-revizor-fixtures",
				Branch:     "master",
			},
			VcsProvider: vcsProvider,
			Account:     &Account{ID: defaultAccountID},
		})
		assert.Nil(t, hook)
		assert.EqualError(t, err, "name is required")
	})
}

func TestHooksRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	vcsProvider, vcsProviderCleanup := createVcsProvider(t, client, nil)
	defer vcsProviderCleanup()

	hook, hookCleanup := createHook(t, client, vcsProvider)
	defer hookCleanup()

	t.Run("with existing hook", func(t *testing.T) {
		readHook, err := client.Hooks.Read(ctx, hook.ID)
		require.NoError(t, err)

		assert.Equal(t, hook.ID, readHook.ID)
		assert.Equal(t, hook.Name, readHook.Name)
		assert.Equal(t, hook.Interpreter, readHook.Interpreter)
		assert.Equal(t, hook.ScriptfilePath, readHook.ScriptfilePath)
	})
}

func TestHooksUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	vcsProvider, vcsProviderCleanup := createVcsProvider(t, client, nil)
	defer vcsProviderCleanup()

	hook, hookCleanup := createHook(t, client, vcsProvider)
	defer hookCleanup()

	t.Run("with valid options", func(t *testing.T) {
		updatedName := "updated-hook"
		updatedInterpreter := "python"
		updatedScriptfilePath := "updated-script.py"

		updateOptions := HookUpdateOptions{
			Name:           String(updatedName),
			Interpreter:    String(updatedInterpreter),
			ScriptfilePath: String(updatedScriptfilePath),
		}

		updatedHook, err := client.Hooks.Update(ctx, hook.ID, updateOptions)
		require.NoError(t, err)

		assert.Equal(t, hook.ID, updatedHook.ID)
		assert.Equal(t, *updateOptions.Name, updatedHook.Name)
		assert.Equal(t, *updateOptions.Interpreter, updatedHook.Interpreter)
		assert.Equal(t, *updateOptions.ScriptfilePath, updatedHook.ScriptfilePath)
	})
}

func TestHooksDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	vcsProvider, vcsProviderCleanup := createVcsProvider(t, client, nil)
	defer vcsProviderCleanup()

	hook, _ := createHook(t, client, vcsProvider)

	t.Run("success", func(t *testing.T) {
		// Delete the hook
		err := client.Hooks.Delete(ctx, hook.ID)
		require.NoError(t, err)

		// Try loading the hook - it should fail
		_, err = client.Hooks.Read(ctx, hook.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestHooksList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	vcsProvider, vcsProviderCleanup := createVcsProvider(t, client, nil)
	defer vcsProviderCleanup()

	hook1, hook1Cleanup := createHook(t, client, vcsProvider)
	defer hook1Cleanup()

	hook2, hook2Cleanup := createHook(t, client, vcsProvider)
	defer hook2Cleanup()

	t.Run("without list options", func(t *testing.T) {
		hookList, err := client.Hooks.List(ctx, HookListOptions{})
		require.NoError(t, err)
		hookIDs := make([]string, len(hookList.Items))
		for _, h := range hookList.Items {
			hookIDs = append(hookIDs, h.ID)
		}
		assert.Contains(t, hookIDs, hook1.ID)
		assert.Contains(t, hookIDs, hook2.ID)
		assert.Equal(t, 1, hookList.CurrentPage)

		assert.True(t, hookList.TotalCount >= 2)
	})

	t.Run("with name and account options", func(t *testing.T) {
		hookList, err := client.Hooks.List(ctx, HookListOptions{
			Account: defaultAccountID,
			Name:    hook1.Name,
		})
		require.NoError(t, err)

		assert.Equal(t, 1, hookList.CurrentPage)
		assert.Equal(t, 1, hookList.TotalCount)
		assert.Equal(t, hook1.ID, hookList.Items[0].ID)
	})

	t.Run("with list options", func(t *testing.T) {
		hookList, err := client.Hooks.List(ctx, HookListOptions{
			ListOptions: ListOptions{
				PageNumber: 999,
				PageSize:   100,
			},
		})
		require.NoError(t, err)

		assert.Empty(t, hookList.Items)
		assert.Equal(t, 999, hookList.CurrentPage)
		assert.True(t, hookList.TotalCount >= 2)
	})
}
