package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleNamespaceCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("success basic", func(t *testing.T) {
		options := ModuleNamespaceCreateOptions{
			Name: String("test-namespace"),
		}
		mn, err := client.ModuleNamespaces.Create(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ModuleNamespaces.Delete(ctx, mn.ID)

		readMn, err := client.ModuleNamespaces.Read(ctx, mn.ID)
		require.NoError(t, err)

		assert.Equal(t, *options.Name, readMn.Name)
		assert.False(t, readMn.IsShared)
	})

	t.Run("success with shared", func(t *testing.T) {
		options := ModuleNamespaceCreateOptions{
			Name:     String("shared-namespace"),
			IsShared: Bool(true),
		}
		mn, err := client.ModuleNamespaces.Create(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ModuleNamespaces.Delete(ctx, mn.ID)

		readMn, err := client.ModuleNamespaces.Read(ctx, mn.ID)
		require.NoError(t, err)

		assert.Equal(t, *options.Name, readMn.Name)
		assert.Equal(t, *options.IsShared, readMn.IsShared)
	})

	t.Run("success with environments", func(t *testing.T) {
		environment, deleteEnvironment := createEnvironment(t, client)
		defer deleteEnvironment()

		options := ModuleNamespaceCreateOptions{
			Name:         String("env-namespace"),
			Environments: []*Environment{environment},
		}
		mn, err := client.ModuleNamespaces.Create(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ModuleNamespaces.Delete(ctx, mn.ID)

		readMn, err := client.ModuleNamespaces.Read(ctx, mn.ID)
		require.NoError(t, err)

		assert.Equal(t, *options.Name, readMn.Name)
		assert.Len(t, readMn.Environments, 1)
		assert.Equal(t, environment.ID, readMn.Environments[0].ID)
	})

	t.Run("success with owners", func(t *testing.T) {
		ownerTeam, ownerTeamCleanup := createTeam(t, client, nil)
		defer ownerTeamCleanup()

		options := ModuleNamespaceCreateOptions{
			Name:   String("owned-namespace"),
			Owners: []*Team{{ID: ownerTeam.ID}},
		}
		mn, err := client.ModuleNamespaces.Create(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ModuleNamespaces.Delete(ctx, mn.ID)

		readMn, err := client.ModuleNamespaces.Read(ctx, mn.ID)
		require.NoError(t, err)

		assert.Equal(t, *options.Name, readMn.Name)
		assert.Len(t, readMn.Owners, 1)
		assert.Equal(t, ownerTeam.ID, readMn.Owners[0].ID)
	})
}

func TestModuleNamespaceRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("basic read", func(t *testing.T) {
		namespace, removeNamespace := createModuleNamespace(t, client, "test-read-namespace")
		defer removeNamespace()

		readNamespace, err := client.ModuleNamespaces.Read(ctx, namespace.ID)
		require.NoError(t, err)
		assert.Equal(t, namespace.ID, readNamespace.ID)
		assert.Equal(t, namespace.Name, readNamespace.Name)
	})
}

func TestModuleNamespaceList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("filtering", func(t *testing.T) {
		type namespaceTestingData struct {
			Name string
		}
		namespaceTestingDataSet := []namespaceTestingData{
			{Name: "namespace_prod_1"},
			{Name: "namespace_prod_2"},
			{Name: "namespace_dev_1"},
		}

		for _, namespaceData := range namespaceTestingDataSet {
			_, removeNamespace := createModuleNamespace(t, client, namespaceData.Name)
			defer removeNamespace()
		}

		requestOptions := ModuleNamespacesListOptions{
			Filter: &ModuleNamespaceFilter{
				Name: "like:_prod_",
			},
		}
		namespacesList, err := client.ModuleNamespaces.List(ctx, requestOptions)

		require.NoError(t, err)
		assert.Equal(t, 2, len(namespacesList.Items))

		var resultNames []string
		for _, namespace := range namespacesList.Items {
			resultNames = append(resultNames, namespace.Name)
		}
		assert.Contains(t, resultNames, "namespace_prod_1")
		assert.Contains(t, resultNames, "namespace_prod_2")
	})
}

func TestModuleNamespaceUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("success update shared", func(t *testing.T) {
		createOptions := ModuleNamespaceCreateOptions{
			Name: String("update-test-namespace"),
		}
		namespace, err := client.ModuleNamespaces.Create(ctx, createOptions)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ModuleNamespaces.Delete(ctx, namespace.ID)

		updateOptions := ModuleNamespaceUpdateOptions{
			IsShared: Bool(true),
		}

		updatedNamespace, err := client.ModuleNamespaces.Update(
			ctx, namespace.ID, updateOptions,
		)
		require.NoError(t, err)
		assert.Equal(t, *updateOptions.IsShared, updatedNamespace.IsShared)
	})

	t.Run("success update environments", func(t *testing.T) {
		environment1, deleteEnvironment1 := createEnvironment(t, client)
		defer deleteEnvironment1()

		environment2, deleteEnvironment2 := createEnvironment(t, client)
		defer deleteEnvironment2()

		createOptions := ModuleNamespaceCreateOptions{
			Name:         String("update-env-namespace"),
			Environments: []*Environment{environment1},
		}
		namespace, err := client.ModuleNamespaces.Create(ctx, createOptions)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ModuleNamespaces.Delete(ctx, namespace.ID)

		updateOptions := ModuleNamespaceUpdateOptions{
			Environments: []*Environment{environment2},
		}
		updatedNamespace, err := client.ModuleNamespaces.Update(
			ctx, namespace.ID, updateOptions,
		)
		require.NoError(t, err)
		assert.Len(t, updatedNamespace.Environments, 1)
		assert.Equal(t, environment2.ID, updatedNamespace.Environments[0].ID)
	})

	t.Run("success update owners", func(t *testing.T) {
		ownerTeam1, ownerTeamCleanup1 := createTeam(t, client, nil)
		defer ownerTeamCleanup1()

		ownerTeam2, ownerTeamCleanup2 := createTeam(t, client, nil)
		defer ownerTeamCleanup2()

		createOptions := ModuleNamespaceCreateOptions{
			Name:   String("update-owners-namespace"),
			Owners: []*Team{{ID: ownerTeam1.ID}},
		}
		namespace, err := client.ModuleNamespaces.Create(ctx, createOptions)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ModuleNamespaces.Delete(ctx, namespace.ID)

		updateOptions := ModuleNamespaceUpdateOptions{
			Owners: []*Team{{ID: ownerTeam2.ID}},
		}
		updatedNamespace, err := client.ModuleNamespaces.Update(
			ctx, namespace.ID, updateOptions,
		)
		require.NoError(t, err)
		assert.Len(t, updatedNamespace.Owners, 1)
		assert.Equal(t, ownerTeam2.ID, updatedNamespace.Owners[0].ID)
	})
}

func TestModuleNamespaceDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	namespace, _ := createModuleNamespace(t, client, "delete-test-namespace")

	t.Run("success", func(t *testing.T) {
		err := client.ModuleNamespaces.Delete(ctx, namespace.ID)
		require.NoError(t, err)

		// Try loading the namespace - it should fail.
		_, err = client.ModuleNamespaces.Read(ctx, namespace.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("ModuleNamespace with ID '%s' not found or user unauthorized.", namespace.ID),
			}.Error(),
			err.Error(),
		)
	})
}

// Helper function to create a module namespace for testing
func createModuleNamespace(t *testing.T, client *Client, name string) (*ModuleNamespace, func()) {
	ctx := context.Background()
	namespace, err := client.ModuleNamespaces.Create(ctx, ModuleNamespaceCreateOptions{
		Name: String(name),
	})
	if err != nil {
		t.Fatal(err)
	}

	return namespace, func() {
		if err := client.ModuleNamespaces.Delete(ctx, namespace.ID); err != nil {
			t.Errorf("Error destroying module namespace! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"ModuleNamespace: %s\nError: %s", namespace.ID, err)
		}
	}
}
