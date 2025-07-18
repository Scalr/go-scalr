package scalr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Compile-time proof of interface implementation.
var _ Workspaces = (*workspaces)(nil)

// Workspaces describes all the workspace related methods that the Scalr API supports.
type Workspaces interface {
	// List all the workspaces within an environment.
	List(ctx context.Context, options WorkspaceListOptions) (*WorkspaceList, error)

	// Create is used to create a new workspace.
	Create(ctx context.Context, options WorkspaceCreateOptions) (*Workspace, error)

	// Read a workspace by its environment ID and name.
	Read(ctx context.Context, environmentID, workspaceName string) (*Workspace, error)

	// ReadByID reads a workspace by its ID.
	ReadByID(ctx context.Context, workspaceID string) (*Workspace, error)

	// Update settings of an existing workspace.
	Update(ctx context.Context, workspaceID string, options WorkspaceUpdateOptions) (*Workspace, error)

	// Delete deletes a workspace by its ID.
	Delete(ctx context.Context, workspaceID string) error

	// SetSchedule sets run schedules for workspace.
	SetSchedule(ctx context.Context, workspaceID string, options WorkspaceRunScheduleOptions) (*Workspace, error)

	// Read outputs
	ReadOutputs(ctx context.Context, workspaceID string) ([]*Output, error)
}

// workspaces implements Workspaces.
type workspaces struct {
	client *Client
}

// WorkspaceExecutionMode represents an execution mode setting of the workspace.
type WorkspaceExecutionMode string

// Available execution modes
const (
	WorkspaceExecutionModeRemote WorkspaceExecutionMode = "remote"
	WorkspaceExecutionModeLocal  WorkspaceExecutionMode = "local"
)

// WorkspaceAutoQueueRuns represents run triggering modes
type WorkspaceAutoQueueRuns string

// Available auto queue modes
const (
	AutoQueueRunsModeSkipFirst    WorkspaceAutoQueueRuns = "skip_first"
	AutoQueueRunsModeAlways       WorkspaceAutoQueueRuns = "always"
	AutoQueueRunsModeNever        WorkspaceAutoQueueRuns = "never"
	AutoQueueRunsModeOnCreateOnly WorkspaceAutoQueueRuns = "on_create_only"
)

// WorkspaceIaCPlatform represents an IaC platform used in this workspace.
type WorkspaceIaCPlatform string

// Available IaC platforms
const (
	WorkspaceIaCPlatformTerraform WorkspaceIaCPlatform = "terraform"
	WorkspaceIaCPlatformOpenTofu  WorkspaceIaCPlatform = "opentofu"
)

// WorkspaceEnvironmentType represents the type of workspace environment.
type WorkspaceEnvironmentType string

// Available workspace environment types
const (
	WorkspaceEnvironmentTypeProduction  WorkspaceEnvironmentType = "production"
	WorkspaceEnvironmentTypeStaging     WorkspaceEnvironmentType = "staging"
	WorkspaceEnvironmentTypeTesting     WorkspaceEnvironmentType = "testing"
	WorkspaceEnvironmentTypeDevelopment WorkspaceEnvironmentType = "development"
	WorkspaceEnvironmentTypeUnmapped    WorkspaceEnvironmentType = "unmapped"
)

// WorkspaceList represents a list of workspaces.
type WorkspaceList struct {
	*Pagination
	Items []*Workspace
}

// Workspace represents a Scalr workspace.
type Workspace struct {
	ID                        string                   `jsonapi:"primary,workspaces"`
	Actions                   *WorkspaceActions        `jsonapi:"attr,actions"`
	AutoApply                 bool                     `jsonapi:"attr,auto-apply"`
	ForceLatestRun            bool                     `jsonapi:"attr,force-latest-run"`
	DeletionProtectionEnabled bool                     `jsonapi:"attr,deletion-protection-enabled"`
	CanQueueDestroyPlan       bool                     `jsonapi:"attr,can-queue-destroy-plan"`
	CreatedAt                 time.Time                `jsonapi:"attr,created-at,iso8601"`
	FileTriggersEnabled       bool                     `jsonapi:"attr,file-triggers-enabled"`
	Locked                    bool                     `jsonapi:"attr,locked"`
	MigrationEnvironment      string                   `jsonapi:"attr,migration-environment"`
	Name                      string                   `jsonapi:"attr,name"`
	Operations                bool                     `jsonapi:"attr,operations"`
	ExecutionMode             WorkspaceExecutionMode   `jsonapi:"attr,execution-mode"`
	Permissions               *WorkspacePermissions    `jsonapi:"attr,permissions"`
	TerraformVersion          string                   `jsonapi:"attr,terraform-version"`
	IaCPlatform               WorkspaceIaCPlatform     `jsonapi:"attr,iac-platform"`
	VCSRepo                   *WorkspaceVCSRepo        `jsonapi:"attr,vcs-repo"`
	Terragrunt                *WorkspaceTerragrunt     `jsonapi:"attr,terragrunt"`
	WorkingDirectory          string                   `jsonapi:"attr,working-directory"`
	ApplySchedule             string                   `jsonapi:"attr,apply-schedule"`
	DestroySchedule           string                   `jsonapi:"attr,destroy-schedule"`
	HasResources              bool                     `jsonapi:"attr,has-resources"`
	AutoQueueRuns             WorkspaceAutoQueueRuns   `jsonapi:"attr,auto-queue-runs"`
	Hooks                     *WorkspaceHooks          `jsonapi:"attr,hooks"`
	RunOperationTimeout       *int                     `jsonapi:"attr,run-operation-timeout"`
	VarFiles                  []string                 `jsonapi:"attr,var-files"`
	EnvironmentType           WorkspaceEnvironmentType `jsonapi:"attr,environment-type"`
	RemoteStateSharing        bool                     `jsonapi:"attr,remote-state-sharing"`

	// Relations
	CurrentRun           *Run                  `jsonapi:"relation,current-run"`
	LatestRun            *Run                  `jsonapi:"relation,latest-run"`
	Environment          *Environment          `jsonapi:"relation,environment"`
	CreatedBy            *User                 `jsonapi:"relation,created-by"`
	VcsProvider          *VcsProvider          `jsonapi:"relation,vcs-provider"`
	AgentPool            *AgentPool            `jsonapi:"relation,agent-pool"`
	ModuleVersion        *ModuleVersion        `jsonapi:"relation,module-version,omitempty"`
	Tags                 []*Tag                `jsonapi:"relation,tags"`
	ConfigurationVersion *ConfigurationVersion `jsonapi:"relation,configuration-version"`
	SSHKey               *SSHKey               `jsonapi:"relation,ssh-key"`
}

type WorkspaceRelation struct {
	ID string `jsonapi:"primary,workspaces"`
}

// WorkspaceHooks contains the custom hooks field.
type WorkspaceHooks struct {
	PreInit   string `json:"pre-init"`
	PrePlan   string `json:"pre-plan"`
	PostPlan  string `json:"post-plan"`
	PreApply  string `json:"pre-apply"`
	PostApply string `json:"post-apply"`
}

// WorkspaceVCSRepo contains the configuration of a VCS integration.
type WorkspaceVCSRepo struct {
	Branch            string   `json:"branch"`
	Identifier        string   `json:"identifier"`
	IngressSubmodules bool     `json:"ingress-submodules"`
	Path              string   `json:"path"`
	TriggerPrefixes   []string `json:"trigger-prefixes,omitempty"`
	TriggerPatterns   string   `json:"trigger-patterns,omitempty"`
	DryRunsEnabled    bool     `json:"dry-runs-enabled"`
	VersionConstraint string   `json:"version-constraint"`
}

type WorkspaceTerragrunt struct {
	Version                     string `json:"version"`
	UseRunAll                   bool   `json:"use-run-all"`
	IncludeExternalDependencies bool   `json:"include-external-dependencies"`
}

// WorkspaceActions represents the workspace actions.
type WorkspaceActions struct {
	IsDestroyable bool `json:"is-destroyable"`
}

// WorkspacePermissions represents the workspace permissions.
type WorkspacePermissions struct {
	CanDestroy        bool `json:"can-destroy"`
	CanForceUnlock    bool `json:"can-force-unlock"`
	CanLock           bool `json:"can-lock"`
	CanQueueApply     bool `json:"can-queue-apply"`
	CanQueueDestroy   bool `json:"can-queue-destroy"`
	CanQueueRun       bool `json:"can-queue-run"`
	CanReadSettings   bool `json:"can-read-settings"`
	CanUnlock         bool `json:"can-unlock"`
	CanUpdate         bool `json:"can-update"`
	CanUpdateVariable bool `json:"can-update-variable"`
}

// WorkspaceSpareseFields is used to request only specific attributes or relationships from a resource,
// rather than the entire set of data.
type WorkspaceSparseFields struct {
	Workspaces string `url:"workspaces,omitempty"`
}

// WorkspaceListOptions represents the options for listing workspaces.
type WorkspaceListOptions struct {
	ListOptions
	Include string                 `url:"include,omitempty"`
	Filter  *WorkspaceFilter       `url:"filter,omitempty"`
	Fields  *WorkspaceSparseFields `url:"fields,omitempty"`
}

// WorkspaceFilter represents the options for filtering workspaces.
type WorkspaceFilter struct {
	Id          *string `url:"workspace,omitempty"`
	Account     *string `url:"account,omitempty"`
	Environment *string `url:"environment,omitempty"`
	Name        *string `url:"name,omitempty"`
	Tag         *string `url:"tag,omitempty"`
	AgentPool   *string `url:"agent-pool,omitempty"`
}

// WorkspaceRunScheduleOptions represents option for setting run schedules for workspace
type WorkspaceRunScheduleOptions struct {
	ApplySchedule   *string `json:"apply-schedule"`
	DestroySchedule *string `json:"destroy-schedule"`
}

type Output struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Sensitive bool   `json:"sensitive"`
}

type WorkspaceTerragruntOptions struct {
	Version                     string `json:"version"`
	UseRunAll                   *bool  `json:"use-run-all,omitempty"`
	IncludeExternalDependencies *bool  `json:"include-external-dependencies,omitempty"`
}

// List all the workspaces within an environment.
func (s *workspaces) List(ctx context.Context, options WorkspaceListOptions) (*WorkspaceList, error) {
	req, err := s.client.newRequest("GET", "workspaces", &options)
	if err != nil {
		return nil, err
	}

	wl := &WorkspaceList{}
	err = s.client.do(ctx, req, wl)
	if err != nil {
		return nil, err
	}

	return wl, nil
}

// WorkspaceCreateOptions represents the options for creating a new workspace.
type WorkspaceCreateOptions struct {
	// For internal use only!
	ID string `jsonapi:"primary,workspaces"`

	// Whether to automatically apply changes when a Terraform plan is successful.
	AutoApply *bool `jsonapi:"attr,auto-apply,omitempty"`

	// Whether to automatically raise the priority of the latest new run.
	ForceLatestRun *bool `jsonapi:"attr,force-latest-run,omitempty"`

	// Whether to prevent deletion when the workspace has resources.
	DeletionProtectionEnabled *bool `jsonapi:"attr,deletion-protection-enabled,omitempty"`

	// The name of the workspace, which can only include letters, numbers, -,
	// and _. This will be used as an identifier and must be unique in the
	// environment.
	Name *string `jsonapi:"attr,name"`

	// Whether the workspace will use remote or local execution mode.
	Operations    *bool                   `jsonapi:"attr,operations,omitempty"`
	ExecutionMode *WorkspaceExecutionMode `jsonapi:"attr,execution-mode,omitempty"`

	// The version of Terraform to use for this workspace. Upon creating a
	// workspace, the latest version is selected unless otherwise specified.
	TerraformVersion *string `jsonapi:"attr,terraform-version,omitempty"`
	// Settings for the workspace terragrunt configuration
	Terragrunt *WorkspaceTerragruntOptions `jsonapi:"attr,terragrunt,omitempty"`

	// The IaC platform to use for this workspace.
	IacPlatform *WorkspaceIaCPlatform `jsonapi:"attr,iac-platform,omitempty"`

	// Settings for the workspace's VCS repository. If omitted, the workspace is
	// created without a VCS repo. If included, you must specify at least the
	// oauth-token-id and identifier keys below.
	VCSRepo *WorkspaceVCSRepoOptions `jsonapi:"attr,vcs-repo,omitempty"`

	// Contains configuration for custom hooks,
	// which can be triggered before or after plan or apply phases
	Hooks *HooksOptions `jsonapi:"attr,hooks,omitempty"`

	// A relative path that Terraform will execute within. This defaults to the
	// root of your repository and is typically set to a subdirectory matching the
	// environment when multiple environments exist within the same repository.
	WorkingDirectory *string `jsonapi:"attr,working-directory,omitempty"`

	// Indicates if runs have to be queued automatically when a new configuration version is uploaded.
	AutoQueueRuns *WorkspaceAutoQueueRuns `jsonapi:"attr,auto-queue-runs,omitempty"`

	// Specifies the VcsProvider for workspace vcs-repo. Required if vcs-repo attr passed
	VcsProvider *VcsProvider `jsonapi:"relation,vcs-provider,omitempty"`

	// Specifies the Environment for workspace.
	Environment *Environment `jsonapi:"relation,environment"`

	// Specifies the AgentPool for workspace.
	AgentPool *AgentPool `jsonapi:"relation,agent-pool,omitempty"`

	// Specifies the VarFiles for workspace.
	VarFiles []string `jsonapi:"attr,var-files"`

	// The type of the Scalr Workspace environment.
	EnvironmentType *WorkspaceEnvironmentType `jsonapi:"attr,environment-type,omitempty"`

	// Specifies the ModuleVersion based on create workspace
	ModuleVersion *ModuleVersion `jsonapi:"relation,module-version,omitempty"`

	// Specifies the number of minutes run operation can be executed before termination.
	RunOperationTimeout *int `jsonapi:"attr,run-operation-timeout"`

	// Specifies tags assigned to the workspace
	Tags []*Tag `jsonapi:"relation,tags,omitempty"`

	RemoteStateSharing *bool `jsonapi:"attr,remote-state-sharing,omitempty"`
}

// WorkspaceVCSRepoOptions represents the configuration options of a VCS integration.
type WorkspaceVCSRepoOptions struct {
	Branch            *string   `json:"branch"`
	Identifier        *string   `json:"identifier,omitempty"`
	IngressSubmodules *bool     `json:"ingress-submodules,omitempty"`
	Path              *string   `json:"path,omitempty"`
	TriggerPrefixes   *[]string `json:"trigger-prefixes,omitempty"`
	TriggerPatterns   *string   `json:"trigger-patterns,omitempty"`
	DryRunsEnabled    *bool     `json:"dry-runs-enabled,omitempty"`
	VersionConstraint *string   `json:"version-constraint"`
}

// HooksOptions represents the WorkspaceHooks configuration.
type HooksOptions struct {
	PreInit   *string `json:"pre-init,omitempty"`
	PrePlan   *string `json:"pre-plan,omitempty"`
	PostPlan  *string `json:"post-plan,omitempty"`
	PreApply  *string `json:"pre-apply,omitempty"`
	PostApply *string `json:"post-apply,omitempty"`
}

func (o WorkspaceCreateOptions) valid() error {
	if !validString(o.Name) {
		return errors.New("name is required")
	}
	if !validStringID(o.Name) {
		return errors.New("invalid value for name")
	}
	return nil
}

// Create is used to create a new workspace.
func (s *workspaces) Create(ctx context.Context, options WorkspaceCreateOptions) (*Workspace, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}
	// Make sure we don't send a user provided ID.
	options.ID = ""

	req, err := s.client.newRequest("POST", "workspaces", &options)
	if err != nil {
		return nil, err
	}

	w := &Workspace{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

// Read a workspace by environment ID and name.
func (s *workspaces) Read(ctx context.Context, environmentID, workspaceName string) (*Workspace, error) {
	if !validStringID(&environmentID) {
		return nil, errors.New("invalid value for environment")
	}
	if !validStringID(&workspaceName) {
		return nil, errors.New("invalid value for workspace")
	}

	options := WorkspaceListOptions{
		Include: "created-by",
		Filter:  &WorkspaceFilter{Environment: &environmentID, Name: &workspaceName},
	}

	req, err := s.client.newRequest("GET", "workspaces", &options)
	if err != nil {
		return nil, err
	}

	wl := &WorkspaceList{}
	err = s.client.do(ctx, req, wl)
	if err != nil {
		return nil, err
	}
	if len(wl.Items) != 1 {
		return nil, errors.New("invalid filters")
	}

	return wl.Items[0], nil
}

// ReadByID reads a workspace by its ID.
func (s *workspaces) ReadByID(ctx context.Context, workspaceID string) (*Workspace, error) {
	if !validStringID(&workspaceID) {
		return nil, errors.New("invalid value for workspace ID")
	}

	options := struct {
		Include string `url:"include"`
	}{
		Include: "created-by",
	}
	u := fmt.Sprintf("workspaces/%s", url.QueryEscape(workspaceID))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	w := &Workspace{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

// WorkspaceUpdateOptions represents the options for updating a workspace.
type WorkspaceUpdateOptions struct {
	// For internal use only!
	ID string `jsonapi:"primary,workspaces"`

	// Whether to automatically apply changes when a Terraform plan is successful.
	AutoApply *bool `jsonapi:"attr,auto-apply,omitempty"`

	// Whether to automatically raise the priority of the latest new run.
	ForceLatestRun *bool `jsonapi:"attr,force-latest-run,omitempty"`

	// Whether to prevent deletion when the workspace has resources.
	DeletionProtectionEnabled *bool `jsonapi:"attr,deletion-protection-enabled,omitempty"`

	// A new name for the workspace, which can only include letters, numbers, -,
	// and _. This will be used as an identifier and must be unique in the
	// environment. Warning: Changing a workspace's name changes its URL in the
	// API and UI.
	Name *string `jsonapi:"attr,name,omitempty"`

	// Whether to filter runs based on the changed files in a VCS push. If
	// enabled, the working directory and trigger prefixes describe a set of
	// paths which must contain changes for a VCS push to trigger a run. If
	// disabled, any push will trigger a run.
	FileTriggersEnabled *bool `jsonapi:"attr,file-triggers-enabled,omitempty"`

	// Whether the workspace will use remote or local execution mode.
	Operations    *bool                   `jsonapi:"attr,operations,omitempty"`
	ExecutionMode *WorkspaceExecutionMode `jsonapi:"attr,execution-mode,omitempty"`

	// The version of Terraform to use for this workspace.
	TerraformVersion *string `jsonapi:"attr,terraform-version,omitempty"`
	// Settings for the workspace terragrunt configuration
	Terragrunt *WorkspaceTerragruntOptions `jsonapi:"attr,terragrunt"`

	// The IaC platform to use for this workspace.
	IacPlatform *WorkspaceIaCPlatform `jsonapi:"attr,iac-platform,omitempty"`

	// To delete a workspace's existing VCS repo, specify null instead of an
	// object. To modify a workspace's existing VCS repo, include whichever of
	// the keys below you wish to modify. To add a new VCS repo to a workspace
	// that didn't previously have one, include at least the oauth-token-id and
	// identifier keys.
	VCSRepo *WorkspaceVCSRepoOptions `jsonapi:"attr,vcs-repo"`

	// Contains configuration for custom hooks,
	// which can be triggered before init, before or after plan or apply phases
	Hooks *HooksOptions `jsonapi:"attr,hooks,omitempty"`

	// A relative path that Terraform will execute within. This defaults to the
	// root of your repository and is typically set to a subdirectory matching
	// the environment when multiple environments exist within the same
	// repository.
	WorkingDirectory *string `jsonapi:"attr,working-directory,omitempty"`

	// Indicates if runs have to be queued automatically when a new configuration version is uploaded.
	AutoQueueRuns *WorkspaceAutoQueueRuns `jsonapi:"attr,auto-queue-runs,omitempty"`

	// Specifies the VcsProvider for workspace vcs-repo.
	VcsProvider *VcsProvider `jsonapi:"relation,vcs-provider"`

	// Specifies the AgentPool for workspace.
	AgentPool *AgentPool `jsonapi:"relation,agent-pool"`

	//Specifies the VarFiles for workspace.
	VarFiles []string `jsonapi:"attr,var_files"`

	// The type of the Scalr Workspace environment.
	EnvironmentType *WorkspaceEnvironmentType `jsonapi:"attr,environment-type,omitempty"`

	// Specifies the ModuleVersion based on create workspace
	ModuleVersion *ModuleVersion `jsonapi:"relation,module-version"`

	// Specifies the number of minutes run operation can be executed before termination.
	RunOperationTimeout *int `jsonapi:"attr,run-operation-timeout"`

	RemoteStateSharing *bool `jsonapi:"attr,remote-state-sharing,omitempty"`
}

// Update settings of an existing workspace.
func (s *workspaces) Update(ctx context.Context, workspaceID string, options WorkspaceUpdateOptions) (*Workspace, error) {
	if !validStringID(&workspaceID) {
		return nil, errors.New("invalid value for workspace ID")
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("workspaces/%s", url.QueryEscape(workspaceID))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	w := &Workspace{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

// Delete deletes a workspace by its ID.
func (s *workspaces) Delete(ctx context.Context, workspaceID string) error {
	if !validStringID(&workspaceID) {
		return errors.New("invalid value for workspace ID")
	}

	u := fmt.Sprintf("workspaces/%s", url.QueryEscape(workspaceID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// SetSchedule set scheduled runs
func (s *workspaces) SetSchedule(ctx context.Context, workspaceID string, options WorkspaceRunScheduleOptions) (*Workspace, error) {
	if !validStringID(&workspaceID) {
		return nil, errors.New("invalid value for workspace ID")
	}

	u := fmt.Sprintf("workspaces/%s/actions/set-schedule", url.QueryEscape(workspaceID))
	req, err := s.client.newJsonRequest("POST", u, &options)
	if err != nil {
		return nil, err
	}

	w := &Workspace{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (s *workspaces) ReadOutputs(ctx context.Context, workspaceID string) ([]*Output, error) {
	if !validStringID(&workspaceID) {
		return nil, errors.New("invalid value for workspace ID")
	}

	u := fmt.Sprintf("workspaces/%s/outputs", url.QueryEscape(workspaceID))
	req, err := s.client.newJsonRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	buffer := &bytes.Buffer{}
	if err = s.client.do(ctx, req, buffer); err != nil {
		return nil, err
	}

	outputs := struct {
		Data []*Output `json:"data"`
	}{}

	if err = json.Unmarshal(buffer.Bytes(), &outputs); err != nil {
		return nil, fmt.Errorf("error unmarshaling response body: %v", err)
	}

	return outputs.Data, nil
}
