package scalr

import (
	"context"
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
}

// workspaces implements Workspaces.
type workspaces struct {
	client *Client
}

// WorkspaceList represents a list of workspaces.
type WorkspaceList struct {
	*Pagination
	Items []*Workspace
}

// Workspace represents a Scalr workspace.
type Workspace struct {
	ID                   string                `jsonapi:"primary,workspaces"`
	Actions              *WorkspaceActions     `jsonapi:"attr,actions"`
	AutoApply            bool                  `jsonapi:"attr,auto-apply"`
	CanQueueDestroyPlan  bool                  `jsonapi:"attr,can-queue-destroy-plan"`
	CreatedAt            time.Time             `jsonapi:"attr,created-at,iso8601"`
	FileTriggersEnabled  bool                  `jsonapi:"attr,file-triggers-enabled"`
	Locked               bool                  `jsonapi:"attr,locked"`
	MigrationEnvironment string                `jsonapi:"attr,migration-environment"`
	Name                 string                `jsonapi:"attr,name"`
	Operations           bool                  `jsonapi:"attr,operations"`
	Permissions          *WorkspacePermissions `jsonapi:"attr,permissions"`
	TerraformVersion     string                `jsonapi:"attr,terraform-version"`
	VCSRepo              *VCSRepo              `jsonapi:"attr,vcs-repo"`
	WorkingDirectory     string                `jsonapi:"attr,working-directory"`
	HasResources         bool                  `jsonapi:"attr,has-resources"`
	Hooks                *Hooks                `jsonapi:"attr,hooks"`

	// Relations
	CurrentRun    *Run                `jsonapi:"relation,current-run"`
	Environment   *Environment        `jsonapi:"relation,environment"`
	CreatedBy     *User               `jsonapi:"relation,created-by"`
	VcsProvider   *VcsProviderOptions `jsonapi:"relation,vcs-provider"`
	ModuleVersion *ModuleVersion      `jsonapi:"relation,module-version,omitempty"`
}

// Hooks contains the custom hooks field.
type Hooks struct {
	PrePlan   string `json:"pre-plan"`
	PostPlan  string `json:"post-plan"`
	PreApply  string `json:"pre-apply"`
	PostApply string `json:"post-apply"`
}

// VCSRepo contains the configuration of a VCS integration.
type VCSRepo struct {
	Branch            string   `json:"branch"`
	Identifier        string   `json:"identifier"`
	IngressSubmodules bool     `json:"ingress-submodules"`
	Path              string   `json:"path"`
	TriggerPrefixes   []string `json:"trigger-prefixes,omitempty"`
	DryRunsEnabled    bool     `json:"dry-runs-enabled"`
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

// WorkspaceListOptions represents the options for listing workspaces.
type WorkspaceListOptions struct {
	ListOptions

	Environment *string `url:"filter[environment],omitempty"`
	Name        *string `url:"filter[workspace][name],omitempty"`
	Include     string  `url:"include,omitempty"`
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

	// The name of the workspace, which can only include letters, numbers, -,
	// and _. This will be used as an identifier and must be unique in the
	// environment.
	Name *string `jsonapi:"attr,name"`

	// Whether the workspace will use remote or local execution mode.
	Operations *bool `jsonapi:"attr,operations,omitempty"`

	// The version of Terraform to use for this workspace. Upon creating a
	// workspace, the latest version is selected unless otherwise specified.
	TerraformVersion *string `jsonapi:"attr,terraform-version,omitempty"`

	// Settings for the workspace's VCS repository. If omitted, the workspace is
	// created without a VCS repo. If included, you must specify at least the
	// oauth-token-id and identifier keys below.
	VCSRepo *VCSRepoOptions `jsonapi:"attr,vcs-repo,omitempty"`

	// Contains configuration for custom hooks,
	// which can be triggered before or after plan or apply phases
	Hooks *HooksOptions `jsonapi:"attr,hooks,omitempty"`

	// A relative path that Terraform will execute within. This defaults to the
	// root of your repository and is typically set to a subdirectory matching the
	// environment when multiple environments exist within the same repository.
	WorkingDirectory *string `jsonapi:"attr,working-directory,omitempty"`

	// Specifies the VcsProvider for workspace vcs-repo. Required if vcs-repo attr passed
	VcsProvider *VcsProviderOptions `jsonapi:"relation,vcs-provider,omitempty"`

	// Specifies the Environment for workspace.
	Environment *Environment `jsonapi:"relation,environment"`
	// Specifies the Module based on create workspace
	ModuleVersion *ModuleVersion `jsonapi:"relation,module-version,omitempty"`
}

// VCSRepoOptions represents the configuration options of a VCS integration.
type VCSRepoOptions struct {
	Branch            *string   `json:"branch,omitempty"`
	Identifier        *string   `json:"identifier,omitempty"`
	IngressSubmodules *bool     `json:"ingress-submodules,omitempty"`
	Path              *string   `json:"path,omitempty"`
	TriggerPrefixes   *[]string `json:"trigger-prefixes,omitempty"`
	DryRunsEnabled    *bool     `json:"dry-runs-enabled,omitempty"`
}

type VcsProviderOptions struct {
	ID      string `jsonapi:"primary,vcs-providers"`
	VcsType string `jsonapi:"attr,vcs-type"`
	Url     string `jsonapi:"attr,url"`
}

// HooksOptions represents the WorkspaceHooks configuration.
type HooksOptions struct {
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

	options := WorkspaceListOptions{Environment: &environmentID, Name: &workspaceName, Include: "created-by"}

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
	Operations *bool `jsonapi:"attr,operations,omitempty"`

	// The version of Terraform to use for this workspace.
	TerraformVersion *string `jsonapi:"attr,terraform-version,omitempty"`

	// To delete a workspace's existing VCS repo, specify null instead of an
	// object. To modify a workspace's existing VCS repo, include whichever of
	// the keys below you wish to modify. To add a new VCS repo to a workspace
	// that didn't previously have one, include at least the oauth-token-id and
	// identifier keys.
	VCSRepo *VCSRepoOptions `jsonapi:"attr,vcs-repo,omitempty"`

	// Contains configuration for custom hooks,
	// which can be triggered before or after plan or apply phases
	Hooks *HooksOptions `jsonapi:"attr,hooks,omitempty"`

	// A relative path that Terraform will execute within. This defaults to the
	// root of your repository and is typically set to a subdirectory matching
	// the environment when multiple environments exist within the same
	// repository.
	WorkingDirectory *string `jsonapi:"attr,working-directory,omitempty"`

	// Specifies the VcsProvider for workspace vcs-repo.
	VcsProvider *VcsProviderOptions `jsonapi:"relation,vcs-provider,omitempty"`
	// Specifies the Module based on create workspace
	ModuleVersion *ModuleVersion `jsonapi:"relation,module-version"`
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
