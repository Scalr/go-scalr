package scalr

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/go-uuid"
)

const defaultAccountID = "acc-svrcncgh453bi8g"
const badIdentifier = "! / nope"

func testClient(t *testing.T) *Client {
	client, err := NewClient(nil)
	if err != nil {
		t.Fatal(err)
	}

	return client
}

func createEnvironment(t *testing.T, client *Client) (*Environment, func()) {
	ctx := context.Background()
	env, err := client.Environments.Create(ctx, EnvironmentCreateOptions{
		Name:    String("tst-" + randomString(t)),
		Account: &Account{ID: defaultAccountID},
	})
	if err != nil {
		t.Fatal(err)
	}

	return env, func() {
		if err := client.Environments.Delete(ctx, env.ID); err != nil {
			t.Errorf("Error destroying environment! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Environment: %s\nError: %s", env.ID, err)
		}
	}
}

func createRole(t *testing.T, client *Client, permissions []*Permission) (*Role, func()) {
	ctx := context.Background()
	role, err := client.Roles.Create(ctx, RoleCreateOptions{
		Name:        String("tst-role-" + randomString(t)),
		Permissions: permissions,
		Account:     &Account{ID: defaultAccountID},
	})
	if err != nil {
		t.Fatal(err)
	}

	return role, func() {
		if err := client.Roles.Delete(ctx, role.ID); err != nil {
			t.Errorf("Error destroying role! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Environment: %s\nError: %s", role.ID, err)
		}
	}
}

func createWorkspace(t *testing.T, client *Client, env *Environment) (*Workspace, func()) {
	var envCleanup func()

	if env == nil {
		env, envCleanup = createEnvironment(t, client)
	}
	ctx := context.Background()
	ws, err := client.Workspaces.Create(
		ctx,
		WorkspaceCreateOptions{Name: String("tst-" + randomString(t)), Environment: env},
	)
	if err != nil {
		t.Fatal(err)
	}

	return ws, func() {
		if err := client.Workspaces.Delete(ctx, ws.ID); err != nil {
			t.Errorf("Error destroying workspace! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Workspace: %s\nError: %s", ws.ID, err)
		}
		if envCleanup != nil {
			envCleanup()
		}
	}
}

func createConfigurationVersion(t *testing.T, client *Client, ws *Workspace) (*ConfigurationVersion, func()) {
	var wsCleanup func()

	if ws == nil {
		ws, wsCleanup = createWorkspace(t, client, nil)
	}
	ctx := context.Background()
	cv, err := client.ConfigurationVersions.Create(ctx, ConfigurationVersionCreateOptions{Workspace: ws})
	if err != nil {
		t.Fatal(err)
	}
	return cv, func() {
		if wsCleanup != nil {
			wsCleanup()
		}
	}
}

func createRun(t *testing.T, client *Client, ws *Workspace, cv *ConfigurationVersion) (*Run, func()) {
	var wsCleanup func()

	if ws == nil {
		ws, wsCleanup = createWorkspace(t, client, nil)
	}
	cv, cvCleanup := createConfigurationVersion(t, client, ws)

	ctx := context.Background()
	run, err := client.Runs.Create(ctx, RunCreateOptions{
		Workspace:            ws,
		ConfigurationVersion: cv,
	})
	if err != nil {
		t.Fatal(err)
	}

	return run, func() {
		if wsCleanup != nil {
			wsCleanup()
		} else {
			cvCleanup()
		}
	}
}

func createVariable(t *testing.T, client *Client, ws *Workspace) (*Variable, func()) {
	var wsCleanup func()

	if ws == nil {
		ws, wsCleanup = createWorkspace(t, client, nil)
	}

	ctx := context.Background()
	v, err := client.Variables.Create(ctx, VariableCreateOptions{
		Key:       String(randomString(t)),
		Value:     String(randomString(t)),
		Category:  Category(CategoryTerraform),
		Workspace: ws,
	})
	if err != nil {
		t.Fatal(err)
	}

	return v, func() {
		if err := client.Variables.Delete(ctx, v.ID); err != nil {
			t.Errorf("Error destroying variable! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Variable: %s\nError: %s", v.Key, err)
		}

		if wsCleanup != nil {
			wsCleanup()
		}
	}
}

func randomString(t *testing.T) string {
	v, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func randomVariableKey(t *testing.T) string {
	return "_" + strings.ReplaceAll(randomString(t), "-", "")
}
