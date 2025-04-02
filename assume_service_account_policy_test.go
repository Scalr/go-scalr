package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssumeServiceAccountPolicyCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("create policy", func(t *testing.T) {
		serviceAccount, removeServiceAccount := createServiceAccount(
			t, client, &Account{ID: defaultAccountID}, ServiceAccountStatusPtr(ServiceAccountStatusActive),
		)
		defer removeServiceAccount()

		provider, removeProvider := createWorkloadIdentityProvider(t, client)
		defer removeProvider()

		createOptions := AssumeServiceAccountPolicyCreateOptions{
			Name:                   String("test-policy"),
			Provider:               provider,
			MaximumSessionDuration: Int(4000),
			ClaimConditions: []ClaimCondition{
				{Claim: "sub", Value: "test@example.com", Operator: String("eq")},
			},
		}

		policy, err := client.AssumeServiceAccountPolicies.Create(
			ctx, serviceAccount.ID, createOptions,
		)
		require.NoError(t, err)

		policy, err = client.AssumeServiceAccountPolicies.Read(ctx, serviceAccount.ID, policy.ID)
		require.NoError(t, err)
		assert.NotEqual(t, policy.ServiceAccount, nil)

		assert.Equal(t, *createOptions.Name, policy.Name)
		assert.Equal(t, *createOptions.MaximumSessionDuration, policy.MaximumSessionDuration)
		assert.Equal(t, createOptions.ClaimConditions, policy.ClaimConditions)
	})
}

func TestAssumeServiceAccountPolicyUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	serviceAccount, removeServiceAccount := createServiceAccount(
		t, client, &Account{ID: defaultAccountID}, ServiceAccountStatusPtr(ServiceAccountStatusActive),
	)
	defer removeServiceAccount()

	provider, removeProvider := createWorkloadIdentityProvider(t, client)
	defer removeProvider()

	t.Run("update policy", func(t *testing.T) {
		createOptions := AssumeServiceAccountPolicyCreateOptions{
			Name:                   String("test-policy"),
			Provider:               provider,
			MaximumSessionDuration: Int(3600),
			ClaimConditions: []ClaimCondition{
				{Claim: "sub", Value: "test@example.com", Operator: String("eq")},
			},
		}

		policy, err := client.AssumeServiceAccountPolicies.Create(
			ctx, serviceAccount.ID, createOptions,
		)
		require.NoError(t, err)

		updateOptions := AssumeServiceAccountPolicyUpdateOptions{
			Name: String("updated-policy"),
			ClaimConditions: &[]ClaimCondition{
				{Claim: "sub", Value: "admin@example.com", Operator: String("eq")},
			},
		}

		policy, err = client.AssumeServiceAccountPolicies.Update(ctx, serviceAccount.ID, policy.ID, updateOptions)
		require.NoError(t, err)

		assert.Equal(t, *updateOptions.Name, policy.Name)
		assert.Equal(t, *updateOptions.ClaimConditions, policy.ClaimConditions)
	})
}

func TestAssumeServiceAccountPolicyDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	serviceAccount, removeServiceAccount := createServiceAccount(
		t, client, &Account{ID: defaultAccountID}, ServiceAccountStatusPtr(ServiceAccountStatusActive),
	)
	defer removeServiceAccount()

	provider, removeProvider := createWorkloadIdentityProvider(t, client)
	defer removeProvider()

	t.Run("delete policy", func(t *testing.T) {
		createOptions := AssumeServiceAccountPolicyCreateOptions{
			Name:                   String("test-policy"),
			Provider:               provider,
			MaximumSessionDuration: Int(3600),
			ClaimConditions: []ClaimCondition{
				{Claim: "sub", Value: "test@example.com", Operator: String("eq")},
			},
		}

		policy, err := client.AssumeServiceAccountPolicies.Create(
			ctx, serviceAccount.ID, createOptions,
		)
		require.NoError(t, err)

		err = client.AssumeServiceAccountPolicies.Delete(ctx, serviceAccount.ID, policy.ID)
		require.NoError(t, err)

		_, err = client.AssumeServiceAccountPolicies.Read(ctx, serviceAccount.ID, policy.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("AssumeServiceAccountPolicy with ID '%s' not found or user unauthorized.", policy.ID),
			}.Error(),
			err.Error(),
		)
	})
}

func TestAssumeServiceAccountPoliciesList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	serviceAccount, removeServiceAccount := createServiceAccount(
		t, client, &Account{ID: defaultAccountID}, ServiceAccountStatusPtr(ServiceAccountStatusActive),
	)
	defer removeServiceAccount()

	provider, removeProvider := createWorkloadIdentityProvider(t, client)
	defer removeProvider()

	// Create test policy
	policy, err := client.AssumeServiceAccountPolicies.Create(ctx, serviceAccount.ID, AssumeServiceAccountPolicyCreateOptions{
		Name:                   String("test-policy"),
		Provider:               provider,
		MaximumSessionDuration: Int(3600),
		ClaimConditions: []ClaimCondition{
			{Claim: "sub", Value: "test@example.com", Operator: String("eq")},
		},
	})
	require.NoError(t, err)

	t.Run("list all policies for service account", func(t *testing.T) {
		policies, err := client.AssumeServiceAccountPolicies.List(ctx, AssumeServiceAccountPoliciesListOptions{
			Filter: &AssumeServiceAccountPolicyFilter{
				ServiceAccount: serviceAccount.ID,
			},
		})
		require.NoError(t, err)

		found := false
		for _, p := range policies.Items {
			if p.ID == policy.ID {
				found = true
				assert.Equal(t, "test-policy", p.Name)
				assert.Equal(t, 3600, p.MaximumSessionDuration)
				assert.Equal(t, policy.ClaimConditions, p.ClaimConditions)
				break
			}
		}
		assert.True(t, found, "policy not found in list")
	})

	t.Run("list with name filter", func(t *testing.T) {
		policies, err := client.AssumeServiceAccountPolicies.List(ctx, AssumeServiceAccountPoliciesListOptions{
			Filter: &AssumeServiceAccountPolicyFilter{
				ServiceAccount: serviceAccount.ID,
				Name:           "test-policy",
			},
		})
		require.NoError(t, err)

		assert.Equal(t, 1, len(policies.Items))
		assert.Equal(t, policy.ID, policies.Items[0].ID)
	})

	t.Run("empty list when no matches", func(t *testing.T) {
		policies, err := client.AssumeServiceAccountPolicies.List(ctx, AssumeServiceAccountPoliciesListOptions{
			Filter: &AssumeServiceAccountPolicyFilter{
				ServiceAccount: serviceAccount.ID,
				Name:           "non-existent-policy",
			},
		})
		require.NoError(t, err)
		assert.Empty(t, policies.Items)
	})
}
