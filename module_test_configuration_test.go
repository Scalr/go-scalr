package scalr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleTestConfigurationsUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("success basic", func(t *testing.T) {
		tc, err := client.ModuleTestConfigurations.Update(
			ctx, defaultModuleID, ModuleTestConfigurationUpdateOptions{
				Enabled: Bool(true),
			},
		)
		require.NoError(t, err)

		assert.True(t, tc.Enabled)
		assert.Equal(t, ModuleTestFailureBehaviorNotify, tc.FailureBehavior)
		assert.False(t, tc.TriggerOnPrActivityEnabled)
		assert.False(t, tc.TriggerOnNewVersionEnabled)
	})

	t.Run("success with triggers and failure behavior", func(t *testing.T) {
		options := ModuleTestConfigurationUpdateOptions{
			Enabled:                    Bool(true),
			FailureBehavior:            ModuleTestFailureBehaviorPtr(ModuleTestFailureBehaviorFailure),
			TriggerOnPrActivityEnabled: Bool(true),
			TriggerOnNewVersionEnabled: Bool(true),
		}
		tc, err := client.ModuleTestConfigurations.Update(ctx, defaultModuleID, options)
		require.NoError(t, err)

		assert.True(t, tc.Enabled)
		assert.Equal(t, ModuleTestFailureBehaviorFailure, tc.FailureBehavior)
		assert.True(t, tc.TriggerOnPrActivityEnabled)
		assert.True(t, tc.TriggerOnNewVersionEnabled)
	})

	t.Run("success disable", func(t *testing.T) {
		tc, err := client.ModuleTestConfigurations.Update(
			ctx, defaultModuleID, ModuleTestConfigurationUpdateOptions{
				Enabled: Bool(false),
			},
		)
		require.NoError(t, err)

		assert.False(t, tc.Enabled)
	})

	t.Run("when module id is invalid", func(t *testing.T) {
		_, err := client.ModuleTestConfigurations.Update(ctx, badIdentifier, ModuleTestConfigurationUpdateOptions{})
		assert.EqualError(t, err, "invalid value for module ID")
	})
}

func TestModuleTestConfigurationsRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		created, err := client.ModuleTestConfigurations.Update(
			ctx, defaultModuleID, ModuleTestConfigurationUpdateOptions{
				Enabled:         Bool(true),
				FailureBehavior: ModuleTestFailureBehaviorPtr(ModuleTestFailureBehaviorNotify),
			},
		)
		require.NoError(t, err)

		tc, err := client.ModuleTestConfigurations.Read(ctx, created.ID)
		require.NoError(t, err)

		assert.Equal(t, created.ID, tc.ID)
		assert.Equal(t, created.Enabled, tc.Enabled)
		assert.Equal(t, created.FailureBehavior, tc.FailureBehavior)
	})

	t.Run("when the test configuration id is invalid", func(t *testing.T) {
		_, err := client.ModuleTestConfigurations.Read(ctx, badIdentifier)
		assert.EqualError(t, err, "invalid value for test configuration ID")
	})
}
