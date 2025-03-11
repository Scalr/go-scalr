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
