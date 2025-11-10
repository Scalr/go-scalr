package scalr

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/go-uuid"
)

const defaultAccountID = "acc-svrcncgh453bi8g"
const defaultAccountName = "mainiacp"
const defaultModuleID = "mod-svsmkkjo8sju4o0"
const badIdentifier = "! / nope"
const policyGroupVcsRepoID = "Scalr/tf-revizor-fixtures"
const policyGroupVcsRepoPath = "policies/clouds"
const policyGroupVcsCommonFunctions = "policies/instances"

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

func createAgentPool(t *testing.T, client *Client, vcsEnabled bool) (*AgentPool, func()) {
	ctx := context.Background()
	ap, err := client.AgentPools.Create(ctx, AgentPoolCreateOptions{
		Name:       String("provider-tst-pool-" + randomString(t)),
		VcsEnabled: Bool(vcsEnabled),
	})
	if err != nil {
		t.Fatal(err)
	}

	return ap, func() {
		if err := client.AgentPools.Delete(ctx, ap.ID); err != nil {
			t.Errorf("Error destroying agent pool! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"AgentPool: %s\nError: %s", ap.ID, err)
		}
	}
}

func createAgentPoolToken(t *testing.T, client *Client, poolID string) (*AccessToken, func()) {
	ctx := context.Background()
	apt, err := client.AgentPoolTokens.Create(ctx, poolID, AccessTokenCreateOptions{Description: String("provider test token")})
	if err != nil {
		t.Fatal(err)
	}

	return apt, func() {
		if err := client.AccessTokens.Delete(ctx, apt.ID); err != nil {
			t.Errorf("Error destroying agent pool token! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Agent pool token: %s\nError: %s", apt.ID, err)
		}
	}
}

func createRole(t *testing.T, client *Client, permissions []*Permission) (*Role, func()) {
	ctx := context.Background()
	role, err := client.Roles.Create(ctx, RoleCreateOptions{
		Name:        String("tst-role-" + randomString(t)),
		Permissions: permissions,
	})
	if err != nil {
		t.Fatal(err)
	}

	return role, func() {
		if err := client.Roles.Delete(ctx, role.ID); err != nil {
			t.Errorf("Error destroying role! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Role: %s\nError: %s", role.ID, err)
		}
	}
}

func createAccessPolicy(t *testing.T, client *Client, roles []*Role, object interface{}) (*AccessPolicy, func()) {
	ctx := context.Background()
	options := AccessPolicyCreateOptions{
		Roles:   roles,
		Account: &Account{ID: defaultAccountID},
	}

	if user, ok := object.(*User); ok {
		options.User = user
	} else if team, ok := object.(*Team); ok {
		options.Team = team
	} else {
		t.Fatal("got object of undefined type")
	}

	ap, err := client.AccessPolicies.Create(ctx, options)
	if err != nil {
		t.Fatal(err)
	}

	return ap, func() {
		if err := client.AccessPolicies.Delete(ctx, ap.ID); err != nil {
			t.Errorf("Error destroying access policy! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"AccessPolicy: %s\nError: %s", ap.ID, err)
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
	var cvCleanup func()

	if ws == nil {
		ws, wsCleanup = createWorkspace(t, client, nil)
	}
	if cv == nil {
		cv, cvCleanup = createConfigurationVersion(t, client, ws)
	}

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

func createVariable(t *testing.T, client *Client, ws *Workspace, env *Environment, acc *Account) (*Variable, func()) {
	ctx := context.Background()
	v, err := client.Variables.Create(ctx, VariableCreateOptions{
		Key:         String(randomVariableKey(t)),
		Value:       String(randomString(t)),
		Category:    Category(CategoryEnv),
		Description: String("Create by go-scalr test helper."),
		Workspace:   ws,
		Environment: env,
		Account:     acc,
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
	}
}

func createVcsProvider(t *testing.T, client *Client, envs []*Environment) (*VcsProvider, func()) {
	ctx := context.Background()
	vcsProvider, err := client.VcsProviders.Create(
		ctx,
		VcsProviderCreateOptions{
			Name:     String("tst-" + randomString(t)),
			VcsType:  Github,
			AuthType: PersonalToken,
			Token:    os.Getenv("GITHUB_TOKEN"),

			Environments: envs,
			Account:      &Account{ID: defaultAccountID},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return vcsProvider, func() {
		if err := client.VcsProviders.Delete(ctx, vcsProvider.ID); err != nil {
			t.Errorf("Error deleting vcs provider! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"VCS Providder: %s\nError: %s", vcsProvider.ID, err)
		}
	}
}

func createTag(t *testing.T, client *Client) (*Tag, func()) {
	ctx := context.Background()
	tag, err := client.Tags.Create(ctx, TagCreateOptions{
		Name:    String("tst-" + randomString(t)),
		Account: &Account{ID: defaultAccountID},
	})
	if err != nil {
		t.Fatal(err)
	}

	return tag, func() {
		if err := client.Tags.Delete(ctx, tag.ID); err != nil {
			t.Errorf("Error destroying tag! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Tag: %s\nError: %s", tag.ID, err)
		}
	}
}

func createWorkloadIdentityProvider(t *testing.T, client *Client) (*WorkloadIdentityProvider, func()) {
	ctx := context.Background()

	provider, err := client.WorkloadIdentityProviders.Create(ctx, WorkloadIdentityProviderCreateOptions{
		Name:             String("wip-" + randomString(t)),
		URL:              String("https://" + strings.ReplaceAll(randomString(t), "-", "")),
		AllowedAudiences: []string{"audience1", "audience2"},
	})
	if err != nil {
		t.Fatal(err)
	}

	return provider, func() {
		if err := client.WorkloadIdentityProviders.Delete(ctx, provider.ID); err != nil {
			t.Errorf("Error destroying provider! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Provider: %s\nError: %s", provider.ID, err)
		}
	}
}

func createTeam(t *testing.T, client *Client, users []*User) (*Team, func()) {
	ctx := context.Background()
	team, err := client.Teams.Create(
		ctx,
		TeamCreateOptions{
			Name:    String("tst-" + randomString(t)),
			Account: &Account{ID: defaultAccountID},
			Users:   users,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return team, func() {
		if err := client.Teams.Delete(ctx, team.ID); err != nil {
			t.Errorf("Error deleting team! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"VCS Providder: %s\nError: %s", team.ID, err)
		}
	}
}

func createPolicyGroup(t *testing.T, client *Client, vcsProvider *VcsProvider) (*PolicyGroup, func()) {
	var vcsCleanup func()

	if vcsProvider == nil {
		vcsProvider, vcsCleanup = createVcsProvider(t, client, nil)
	}

	ctx := context.Background()
	pg, err := client.PolicyGroups.Create(ctx, PolicyGroupCreateOptions{
		Name:        String("tst-" + randomString(t)),
		Account:     &Account{ID: defaultAccountID},
		VcsProvider: vcsProvider,
		VCSRepo: &PolicyGroupVCSRepoOptions{
			Identifier: String(policyGroupVcsRepoID),
			Path:       String(policyGroupVcsRepoPath),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	return pg, func() {
		if err := client.PolicyGroups.Delete(ctx, pg.ID); err != nil {
			t.Errorf("Error destroying policy group! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Policy group: %s\nError: %s", pg.ID, err)
		}

		if vcsCleanup != nil {
			vcsCleanup()
		}
	}
}

func createHook(t *testing.T, client *Client, vcsProvider *VcsProvider) (*Hook, func()) {
	var vcsCleanup func()

	if vcsProvider == nil {
		vcsProvider, vcsCleanup = createVcsProvider(t, client, nil)
	}

	ctx := context.Background()

	hookName := "test-hook-" + randomString(t)
	hookInterpreter := "bash"
	hookScriptfilePath := "pre-plan.sh"
	hookVcsRepo := &HookVcsRepo{
		Identifier: "Scalr/tf-revizor-fixtures",
		Branch:     "master",
	}

	options := HookCreateOptions{
		Name:           hookName,
		Interpreter:    hookInterpreter,
		ScriptfilePath: hookScriptfilePath,
		VcsRepo:        hookVcsRepo,
		VcsProvider:    vcsProvider,
	}

	hook, err := client.Hooks.Create(ctx, options)

	if err != nil {
		t.Fatal(err)
	}

	return hook, func() {
		if err := client.Hooks.Delete(ctx, hook.ID); err != nil {
			t.Errorf("Error destroying hook! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Hook: %s\nError: %s", hook.ID, err)
		}

		if vcsCleanup != nil {
			vcsCleanup()
		}
	}
}

func linkPolicyGroupToEnvironment(t *testing.T, client *Client, policyGroup *PolicyGroup, environment *Environment) func() {
	ctx := context.Background()
	options := PolicyGroupEnvironmentsCreateOptions{
		PolicyGroupID:           policyGroup.ID,
		PolicyGroupEnvironments: []*PolicyGroupEnvironment{{ID: environment.ID}},
	}

	err := client.PolicyGroupEnvironments.Create(ctx, options)

	if err != nil {
		t.Fatal(err)
	}

	return func() {
		err := client.PolicyGroupEnvironments.Delete(
			ctx,
			PolicyGroupEnvironmentDeleteOptions{
				PolicyGroupID: policyGroup.ID,
				EnvironmentID: environment.ID,
			},
		)

		if err != nil {
			t.Errorf("Error destroying policy group environment linkage! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Policy group: %s\nError: %s", policyGroup.ID, err)
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

func createProviderConfiguration(t *testing.T, client *Client, providerName string, configurationName string) (*ProviderConfiguration, func()) {
	ctx := context.Background()
	config, err := client.ProviderConfigurations.Create(
		ctx,
		ProviderConfigurationCreateOptions{
			Account:      &Account{ID: defaultAccountID},
			Name:         String(configurationName),
			ProviderName: String(providerName),
			IsShared:     Bool(true),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return config, func() {
		if err := client.ProviderConfigurations.Delete(ctx, config.ID); err != nil {
			t.Errorf("Error destroying provider configuration ! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Provider configuration: %s\nError: %s", config.ID, err)
		}
	}
}

func createServiceAccount(
	t *testing.T,
	client *Client,
	account *Account,
	status *ServiceAccountStatus) (*ServiceAccount, func()) {
	ctx := context.Background()
	sa, err := client.ServiceAccounts.Create(ctx, ServiceAccountCreateOptions{
		Name:        String("tst-" + randomString(t)),
		Description: String("tst-description-" + randomString(t)),
		Status:      status,
		Account:     account,
	})
	if err != nil {
		t.Fatal(err)
	}

	return sa, func() {
		if err := client.ServiceAccounts.Delete(ctx, sa.ID); err != nil {
			t.Errorf("Error destroying service account! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Service account: %s\nError: %s", sa.ID, err)
		}
	}
}

func createServiceAccountToken(t *testing.T, client *Client, serviceAccountID string) (*AccessToken, func()) {
	ctx := context.Background()
	sat, err := client.ServiceAccountTokens.Create(
		ctx, serviceAccountID, AccessTokenCreateOptions{Description: String("tst-description-" + randomString(t))},
	)
	if err != nil {
		t.Fatal(err)
	}

	return sat, func() {
		if err := client.AccessTokens.Delete(ctx, sat.ID); err != nil {
			t.Errorf("Error destroying service account token! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Service account token: %s\nError: %s", sat.ID, err)
		}
	}
}

func assignTagsToWorkspace(t *testing.T, client *Client, workspace *Workspace, tags []*Tag) {
	ctx := context.Background()
	tagRels := make([]*TagRelation, len(tags))
	for i, tag := range tags {
		tagRels[i] = &TagRelation{ID: tag.ID}
	}
	err := client.WorkspaceTags.Add(ctx, workspace.ID, tagRels)

	if err != nil {
		t.Fatal(err)
	}
}

func assignTagsToEnvironment(t *testing.T, client *Client, environment *Environment, tags []*Tag) {
	ctx := context.Background()
	tagRels := make([]*TagRelation, len(tags))
	for i, tag := range tags {
		tagRels[i] = &TagRelation{ID: tag.ID}
	}
	err := client.EnvironmentTags.Add(ctx, environment.ID, tagRels)

	if err != nil {
		t.Fatal(err)
	}
}

func createWebhookIntegration(
	t *testing.T, client *Client, isShared bool, envs []*Environment,
) (*WebhookIntegration, func()) {
	ctx := context.Background()
	opts := WebhookIntegrationCreateOptions{
		Name:         String("tst-" + randomString(t)),
		Enabled:      Bool(true),
		IsShared:     Bool(isShared),
		Url:          String("https://example.com"),
		Account:      &Account{ID: defaultAccountID},
		Events:       []*EventDefinition{{ID: "run:completed"}},
		Environments: envs,
	}
	w, err := client.WebhookIntegrations.Create(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}

	return w, func() {
		if err := client.WebhookIntegrations.Delete(ctx, w.ID); err != nil {
			t.Errorf("Error destroying webhook integration! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Webhook: %s\nError: %s", w.ID, err)
		}
	}
}

func createSlackIntegration(
	t *testing.T, client *Client, slackConnection *SlackConnection, environment *Environment,
) (*SlackIntegration, func()) {
	ctx := context.Background()
	options := SlackIntegrationCreateOptions{
		Name: String("test-" + randomString(t)),
		Events: []string{
			SlackIntegrationEventRunApprovalRequired,
			SlackIntegrationEventRunSuccess,
			SlackIntegrationEventRunErrored,
		},
		ChannelId:    String("C123"),
		Account:      &Account{ID: defaultAccountID},
		Connection:   slackConnection,
		Environments: []*Environment{environment},
	}
	si, err := client.SlackIntegrations.Create(ctx, options)
	if err != nil {
		t.Fatal(err)
	}

	return si, func() {
		if err := client.SlackIntegrations.Delete(ctx, si.ID); err != nil {
			t.Errorf("Error deleting slack integration! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Webhook: %s\nError: %s", si.ID, err)
		}
	}
}

func createRunScheduleRule(t *testing.T, client *Client, workspace *Workspace, mode ScheduleMode) (*RunScheduleRule, func()) {
	ctx := context.Background()
	options := RunScheduleRuleCreateOptions{
		Schedule:     "0 0 * * *",
		ScheduleMode: mode,
		Workspace:    workspace,
	}
	rule, err := client.RunScheduleRules.Create(ctx, options)

	if err != nil {
		t.Fatal(err)
	}

	return rule, func() {
		if err := client.RunScheduleRules.Delete(ctx, rule.ID); err != nil {
			t.Errorf("Error destroying run schedule rule! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Run schedule rule: %s\nError: %s", rule.ID, err)
		}
	}
}

func createSSHKey(t *testing.T, client *Client, name string, isShared bool, privateKey string) (*SSHKey, func()) {
	ctx := context.Background()

	if privateKey == "" {
		privateKey = `-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIBvMDyNaYtWK2TmJIfFhmPZeGxK0bWnNDhjlTZ+V6e4x
-----END PRIVATE KEY-----`
	}

	opts := SSHKeyCreateOptions{
		Name:       String(name),
		PrivateKey: String(privateKey),
		IsShared:   Bool(isShared),
		Account:    &Account{ID: defaultAccountID},
	}

	sshKey, err := client.SSHKeys.Create(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}

	return sshKey, func() {
		if err := client.SSHKeys.Delete(ctx, sshKey.ID); err != nil {
			t.Errorf("Error destroying SSH key! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"SSH Key: %s\nError: %s", sshKey.ID, err)
		}
	}
}

func addRemoteStateConsumersToWorkspace(t *testing.T, client *Client, workspace *Workspace, consumers []*Workspace) {
	ctx := context.Background()
	wsRels := make([]*WorkspaceRelation, len(consumers))
	for i, ws := range consumers {
		wsRels[i] = &WorkspaceRelation{ID: ws.ID}
	}
	err := client.RemoteStateConsumers.Add(ctx, workspace.ID, wsRels)

	if err != nil {
		t.Fatal(err)
	}
}

func addFederatedEnvironments(t *testing.T, client *Client, environment *Environment, federated []*Environment) {
	ctx := context.Background()
	envRels := make([]*EnvironmentRelation, len(federated))
	for i, env := range federated {
		envRels[i] = &EnvironmentRelation{ID: env.ID}
	}
	err := client.FederatedEnvironments.Add(ctx, environment.ID, envRels)

	if err != nil {
		t.Fatal(err)
	}
}

func createDriftDetection(t *testing.T, client *Client, environment *Environment, period DriftDetectionSchedulePeriod) (*DriftDetection, func()) {
	ctx := context.Background()
	driftDetection, err := client.DriftDetections.Create(ctx, DriftDetectionCreateOptions{
		Schedule:    period,
		Environment: environment,
	})
	if err != nil {
		t.Fatal(err)
	}

	return driftDetection, func() {
		if err := client.DriftDetections.Delete(ctx, driftDetection.ID); err != nil {
			t.Errorf("Error destroying drift tetection schedule! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"DriftDetection: %s\nError: %s", driftDetection.ID, err)
		}
	}
}
