package scalr

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRunScheduleRulesList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	environment, environmentCleanup := createEnvironment(t, client)
	defer environmentCleanup()
	workspace, workspaceCleanup := createWorkspace(t, client, environment)
	defer workspaceCleanup()

	scheduleRuleA, scheduleRuleACleanup := createRunScheduleRule(t, client, workspace, ScheduleModeApply)
	defer scheduleRuleACleanup()
	scheduleRuleB, scheduleRuleBCleanup := createRunScheduleRule(t, client, workspace, ScheduleModeDestroy)
	defer scheduleRuleBCleanup()

	t.Run("without include", func(t *testing.T) {
		ruleList, err := client.RunScheduleRules.List(ctx, RunScheduleRuleListOptions{
			Workspace: workspace.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, 2, ruleList.TotalCount)

		ruleIDs := make([]string, len(ruleList.Items))
		for i, rule := range ruleList.Items {
			ruleIDs[i] = rule.ID
		}
		assert.Contains(t, ruleIDs, scheduleRuleA.ID)
		assert.Contains(t, ruleIDs, scheduleRuleB.ID)
	})

	t.Run("with include", func(t *testing.T) {
		ruleList, err := client.RunScheduleRules.List(ctx, RunScheduleRuleListOptions{
			Workspace: workspace.ID,
			Include:   "workspace",
		})
		require.NoError(t, err)
		assert.Equal(t, 2, ruleList.TotalCount)
		for _, rule := range ruleList.Items {
			assert.NotEqual(t, rule.Workspace, nil)
		}
	})
}

func TestRunScheduleRulesCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	environment, environmentCleanup := createEnvironment(t, client)
	defer environmentCleanup()
	workspace, workspaceCleanup := createWorkspace(t, client, environment)
	defer workspaceCleanup()

	t.Run("with valid options", func(t *testing.T) {
		options := RunScheduleRuleCreateOptions{
			Schedule:     "0 0 * * *",
			ScheduleMode: ScheduleModeApply,
			Workspace:    workspace,
		}
		rule, err := client.RunScheduleRules.Create(ctx, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.RunScheduleRules.Read(ctx, rule.ID)
		require.NoError(t, err)

		for _, item := range []*RunScheduleRule{
			rule,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
		}
		err = client.RunScheduleRules.Delete(ctx, rule.ID)
		require.NoError(t, err)
	})

	t.Run("when rule type collision", func(t *testing.T) {
		options := RunScheduleRuleCreateOptions{
			Schedule:     "0 0 * * *",
			ScheduleMode: ScheduleModeApply,
			Workspace:    workspace,
		}
		rule, err := client.RunScheduleRules.Create(ctx, options)
		require.NoError(t, err)

		_, err = client.RunScheduleRules.Create(ctx, options)
		require.Error(t, err)

		err = client.RunScheduleRules.Delete(ctx, rule.ID)
		require.NoError(t, err)
	})
}

func TestRunScheduleRulesRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	environment, environmentCleanup := createEnvironment(t, client)
	defer environmentCleanup()
	workspace, workspaceCleanup := createWorkspace(t, client, environment)
	defer workspaceCleanup()

	ruleTest, ruleTestCleanup := createRunScheduleRule(t, client, workspace, ScheduleModeApply)
	defer ruleTestCleanup()

	t.Run("by ID when the rule exists", func(t *testing.T) {
		rule, err := client.RunScheduleRules.Read(ctx, ruleTest.ID)
		require.NoError(t, err)
		assert.Equal(t, ruleTest.ID, rule.ID)
	})

	t.Run("by ID when the rule does not exist", func(t *testing.T) {
		rule, err := client.RunScheduleRules.Read(ctx, "rule-nonexisting")
		assert.Nil(t, rule)
		assert.Error(t, err)
	})

	t.Run("by ID without a valid rule ID", func(t *testing.T) {
		rule, err := client.RunScheduleRules.Read(ctx, badIdentifier)
		assert.Nil(t, rule)
		assert.EqualError(t, err, "invalid value for run schedule rule ID")
	})
}

func TestRunScheduleRulesUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	environment, environmentCleanup := createEnvironment(t, client)
	defer environmentCleanup()
	workspace, workspaceCleanup := createWorkspace(t, client, environment)
	defer workspaceCleanup()

	scheduleRuleA, scheduleRuleACleanup := createRunScheduleRule(t, client, workspace, ScheduleModeApply)
	defer scheduleRuleACleanup()
	_, scheduleRuleBCleanup := createRunScheduleRule(t, client, workspace, ScheduleModeRefresh)
	defer scheduleRuleBCleanup()

	t.Run("with valid options", func(t *testing.T) {
		schedule := "* * * * *"
		scheduleMode := ScheduleModeDestroy
		options := RunScheduleRuleUpdateOptions{
			Schedule:     &schedule,
			ScheduleMode: &scheduleMode,
		}

		rule, err := client.RunScheduleRules.Update(ctx, scheduleRuleA.ID, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.RunScheduleRules.Read(ctx, rule.ID)
		require.NoError(t, err)

		for _, item := range []*RunScheduleRule{
			rule,
			refreshed,
		} {
			assert.Equal(t, options.Schedule, item.Schedule)
			assert.Equal(t, options.ScheduleMode, item.ScheduleMode)
		}
	})

	t.Run("with mode collision", func(t *testing.T) {
		schedule := "* * * * *"
		scheduleMode := ScheduleModeRefresh
		rule, err := client.RunScheduleRules.Update(ctx, scheduleRuleA.ID, RunScheduleRuleUpdateOptions{
			Schedule:     &schedule,
			ScheduleMode: &scheduleMode,
		})
		assert.Nil(t, rule)
		assert.Error(t, err)
	})
}

func TestRunScheduleRulesDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	environment, environmentCleanup := createEnvironment(t, client)
	defer environmentCleanup()
	workspace, workspaceCleanup := createWorkspace(t, client, environment)
	defer workspaceCleanup()

	scheduleRule, _ := createRunScheduleRule(t, client, workspace, ScheduleModeApply)

	t.Run("with valid options", func(t *testing.T) {
		err := client.RunScheduleRules.Delete(ctx, scheduleRule.ID)
		require.NoError(t, err)

		_, err = client.RunScheduleRules.Read(ctx, scheduleRule.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("RunScheduleRule with ID '%s' not found or user unauthorized.", scheduleRule.ID),
			}.Error(),
			err.Error(),
		)
	})
}
