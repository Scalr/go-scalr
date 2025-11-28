package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDriftDetectionCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	t.Run("with valid options", func(t *testing.T) {
		options := DriftDetectionCreateOptions{
			Schedule:    DriftDetectionSchedulePeriodDaily,
			Environment: envTest,
		}

		dds, err := client.DriftDetections.Create(ctx, options)
		require.NoError(t, err)

		refreshed, err := client.DriftDetections.Read(ctx, dds.ID)
		require.NoError(t, err)

		assert.Equal(t, options.Schedule, refreshed.Schedule)
		assert.Equal(t, options.Environment.ID, refreshed.Environment.ID)
		assert.Equal(t, refreshed.WorkspaceFilters, DriftDetectionWorkspaceFilter{})

		err = client.DriftDetections.Delete(ctx, dds.ID)
		require.NoError(t, err)
	})

	t.Run("with valid options with workspace filters (name pattern) and run mode", func(t *testing.T) {
		runMode := DriftDetectionScheduleRunModePlan
		workspaceFilters := DriftDetectionWorkspaceFilter{
			NamePatterns: &[]string{"*foo", "bar*"},
		}
		options := DriftDetectionCreateOptions{
			Schedule:         DriftDetectionSchedulePeriodDaily,
			Environment:      envTest,
			WorkspaceFilters: workspaceFilters,
			RunMode:          &runMode,
		}

		dds, err := client.DriftDetections.Create(ctx, options)
		require.NoError(t, err)

		refreshed, err := client.DriftDetections.Read(ctx, dds.ID)
		require.NoError(t, err)

		assert.Equal(t, options.Schedule, refreshed.Schedule)
		assert.Equal(t, options.Environment.ID, refreshed.Environment.ID)
		assert.Equal(t, runMode, refreshed.RunMode)
		assert.Equal(t, workspaceFilters.NamePatterns, refreshed.WorkspaceFilters.NamePatterns)
		assert.Nil(t, refreshed.WorkspaceFilters.Tags)
		assert.Nil(t, refreshed.WorkspaceFilters.EnvironmentTypes)

		err = client.DriftDetections.Delete(ctx, dds.ID)
		require.NoError(t, err)
	})

	t.Run("with valid options with workspace filters (tags) and run mode", func(t *testing.T) {
		runMode := DriftDetectionScheduleRunModeRefreshOnly
		workspaceFilters := DriftDetectionWorkspaceFilter{
			Tags: &[]string{"foo", "bar"},
		}
		options := DriftDetectionCreateOptions{
			Schedule:         DriftDetectionSchedulePeriodDaily,
			Environment:      envTest,
			WorkspaceFilters: workspaceFilters,
			RunMode:          &runMode,
		}

		dds, err := client.DriftDetections.Create(ctx, options)
		require.NoError(t, err)

		refreshed, err := client.DriftDetections.Read(ctx, dds.ID)
		require.NoError(t, err)

		assert.Equal(t, options.Schedule, refreshed.Schedule)
		assert.Equal(t, options.Environment.ID, refreshed.Environment.ID)
		assert.Equal(t, runMode, refreshed.RunMode)
		assert.Equal(t, workspaceFilters.Tags, refreshed.WorkspaceFilters.Tags)
		assert.Nil(t, refreshed.WorkspaceFilters.NamePatterns)
		assert.Nil(t, refreshed.WorkspaceFilters.EnvironmentTypes)

		err = client.DriftDetections.Delete(ctx, dds.ID)
		require.NoError(t, err)
	})

	t.Run("with valid options with workspace filters (env types) and run mode", func(t *testing.T) {
		runMode := DriftDetectionScheduleRunModePlan
		workspaceFilters := DriftDetectionWorkspaceFilter{
			EnvironmentTypes: &[]WorkspaceEnvironmentType{
				WorkspaceEnvironmentTypeProduction, WorkspaceEnvironmentTypeStaging,
			},
		}
		options := DriftDetectionCreateOptions{
			Schedule:         DriftDetectionSchedulePeriodDaily,
			Environment:      envTest,
			WorkspaceFilters: workspaceFilters,
			RunMode:          &runMode,
		}

		dds, err := client.DriftDetections.Create(ctx, options)
		require.NoError(t, err)

		refreshed, err := client.DriftDetections.Read(ctx, dds.ID)
		require.NoError(t, err)

		assert.Equal(t, options.Schedule, refreshed.Schedule)
		assert.Equal(t, options.Environment.ID, refreshed.Environment.ID)
		assert.Equal(t, runMode, refreshed.RunMode)
		assert.Equal(t, workspaceFilters.EnvironmentTypes, refreshed.WorkspaceFilters.EnvironmentTypes)
		assert.Nil(t, refreshed.WorkspaceFilters.NamePatterns)
		assert.Nil(t, refreshed.WorkspaceFilters.Tags)

		err = client.DriftDetections.Delete(ctx, dds.ID)
		require.NoError(t, err)
	})

	t.Run("when options has invalid schedule", func(t *testing.T) {

		dds, err := client.DriftDetections.Create(ctx, DriftDetectionCreateOptions{
			Schedule:    "badvalue",
			Environment: envTest,
		})
		assert.Nil(t, dds)
		assert.EqualError(t, err, "invalid value for schedule")
	})

	t.Run("when options has an empty environment", func(t *testing.T) {
		dds, err := client.DriftDetections.Create(ctx, DriftDetectionCreateOptions{
			Schedule:    DriftDetectionSchedulePeriodDaily,
			Environment: nil,
		})
		assert.Nil(t, dds)
		assert.EqualError(t, err, "environment is required")
	})

	t.Run("when options has invalid environment id", func(t *testing.T) {
		dds, err := client.DriftDetections.Create(ctx, DriftDetectionCreateOptions{
			Schedule:    DriftDetectionSchedulePeriodDaily,
			Environment: &Environment{ID: badIdentifier},
		})
		assert.Nil(t, dds)
		assert.EqualError(t, err, "invalid value for environment ID")
	})
}

func TestDriftDetectionRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	ddsTest, ddsTestCleanup := createDriftDetection(
		t, client, envTest, DriftDetectionSchedulePeriodDaily, DriftDetectionWorkspaceFilter{}, nil,
	)
	defer ddsTestCleanup()

	t.Run("by ID when the drift detection schedule exists", func(t *testing.T) {
		dds, err := client.DriftDetections.Read(ctx, ddsTest.ID)
		require.NoError(t, err)
		assert.Equal(t, ddsTest.ID, dds.ID)
	})

	t.Run("by ID when the drift detection schedule does not exist", func(t *testing.T) {
		dds, err := client.DriftDetections.Read(ctx, "dds-nonexisting")
		assert.Nil(t, dds)
		assert.Error(t, err)
	})

	t.Run("by ID without a valid drift detection schedule ID", func(t *testing.T) {
		dds, err := client.DriftDetections.Read(ctx, badIdentifier)
		assert.Nil(t, dds)
		assert.EqualError(t, err, "invalid value for drift detection schedule ID")
	})
}

func TestDriftDetectionUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	t.Run("with valid options", func(t *testing.T) {
		runModeRefreshOnly := DriftDetectionScheduleRunModeRefreshOnly
		runModePlan := DriftDetectionScheduleRunModePlan
		workspaceFilters := DriftDetectionWorkspaceFilter{
			EnvironmentTypes: &[]WorkspaceEnvironmentType{WorkspaceEnvironmentTypeProduction},
		}
		ddsTest, ddsTestCleanup := createDriftDetection(
			t, client, envTest, DriftDetectionSchedulePeriodDaily, workspaceFilters, &runModeRefreshOnly,
		)
		defer ddsTestCleanup()

		newWorkspaceFilters := DriftDetectionWorkspaceFilter{
			Tags: &[]string{"foo", "bar"},
		}
		options := DriftDetectionUpdateOptions{
			Schedule:         DriftDetectionSchedulePeriodWeekly,
			Environment:      &Environment{ID: envTest.ID},
			WorkspaceFilters: newWorkspaceFilters,
			RunMode:          &runModePlan,
		}

		updatedDds, err := client.DriftDetections.Update(ctx, ddsTest.ID, options)

		require.NoError(t, err)
		assert.Equal(t, options.Schedule, updatedDds.Schedule)

		refreshed, err := client.DriftDetections.Read(ctx, ddsTest.ID)
		require.NoError(t, err)

		assert.Equal(t, updatedDds.Schedule, refreshed.Schedule)
		assert.Equal(t, runModePlan, refreshed.RunMode)
		assert.Equal(t, newWorkspaceFilters.Tags, refreshed.WorkspaceFilters.Tags)
		assert.Nil(t, refreshed.WorkspaceFilters.EnvironmentTypes)
		assert.Nil(t, refreshed.WorkspaceFilters.NamePatterns)
	})

	t.Run("with invalid ID", func(t *testing.T) {
		dds, err := client.DriftDetections.Update(ctx, badIdentifier, DriftDetectionUpdateOptions{
			Schedule:    DriftDetectionSchedulePeriodWeekly,
			Environment: &Environment{ID: envTest.ID},
		})
		assert.Nil(t, dds)
		assert.EqualError(t, err, "invalid value for drift detection schedule ID")
	})

	t.Run("when options has invalid schedule", func(t *testing.T) {
		dds, err := client.DriftDetections.Update(ctx, "fakeId", DriftDetectionUpdateOptions{
			Schedule:    "badvalue",
			Environment: &Environment{ID: envTest.ID},
		})
		assert.Nil(t, dds)
		assert.EqualError(t, err, "invalid value for schedule")
	})

	t.Run("when options has an empty environment", func(t *testing.T) {
		dds, err := client.DriftDetections.Update(ctx, "fakeId", DriftDetectionUpdateOptions{
			Schedule:    DriftDetectionSchedulePeriodWeekly,
			Environment: nil,
		})
		assert.Nil(t, dds)
		assert.EqualError(t, err, "environment is required")
	})

	t.Run("when options has invalid environment id", func(t *testing.T) {
		dds, err := client.DriftDetections.Update(ctx, "fakeId", DriftDetectionUpdateOptions{
			Schedule:    DriftDetectionSchedulePeriodWeekly,
			Environment: &Environment{ID: badIdentifier},
		})
		assert.Nil(t, dds)
		assert.EqualError(t, err, "invalid value for environment ID")
	})
}

func TestDriftDetectionDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	ddsTest, _ := createDriftDetection(
		t, client, envTest, DriftDetectionSchedulePeriodDaily, DriftDetectionWorkspaceFilter{}, nil,
	)

	t.Run("with valid options", func(t *testing.T) {
		err := client.DriftDetections.Delete(ctx, ddsTest.ID)
		require.NoError(t, err)

		_, err = client.DriftDetections.Read(ctx, ddsTest.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("DriftDetectionSchedule with ID '%s' not found or user unauthorized.", ddsTest.ID),
			}.Error(),
			err.Error(),
		)
	})
}
