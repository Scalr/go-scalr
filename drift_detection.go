package scalr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ DriftDetections = (*driftDetections)(nil)

// DriftDetections describes all the drift detection schedules related methods that the Scalr API supports.
type DriftDetections interface {
	// Create creates a drift detection schedule.
	Create(ctx context.Context, options DriftDetectionCreateOptions) (*DriftDetection, error)
	// Read reads a drift detection schedule by its ID.
	Read(ctx context.Context, id string) (*DriftDetection, error)
	// Update an existing drift detection schedule by its ID.
	Update(ctx context.Context, id string, options DriftDetectionUpdateOptions) (*DriftDetection, error)
	// Delete deletes a drift detection schedule by its ID.
	Delete(ctx context.Context, id string) error
}

// driftDetections implements DriftDetections.
type driftDetections struct {
	client *Client
}

// DriftDetectionSchedulePeriod represents drift detection schedule periods (schedule attribute)
type DriftDetectionSchedulePeriod string

// Available drift detection schedule periods
const (
	DriftDetectionSchedulePeriodDaily  DriftDetectionSchedulePeriod = "daily"
	DriftDetectionSchedulePeriodWeekly DriftDetectionSchedulePeriod = "weekly"
)

func (o DriftDetectionSchedulePeriod) IsValid() bool {
	switch o {
	case DriftDetectionSchedulePeriodDaily,
		DriftDetectionSchedulePeriodWeekly:
		return true
	}
	return false
}

// DriftDetectionScheduleRunMode represents drift detection schedule run mode (run_mode attribute)
type DriftDetectionScheduleRunMode string

// Available drift detection schedule run modes
const (
	DriftDetectionScheduleRunModeRefreshOnly DriftDetectionScheduleRunMode = "refresh-only"
	DriftDetectionScheduleRunModePlan        DriftDetectionScheduleRunMode = "plan"
)

func (o DriftDetectionScheduleRunMode) IsValid() bool {
	switch o {
	case DriftDetectionScheduleRunModeRefreshOnly,
		DriftDetectionScheduleRunModePlan:
		return true
	}
	return false
}

type DriftDetectionWorkspaceFilter struct {
	NamePatterns     *[]string                   `json:"name-patterns,omitempty"`
	EnvironmentTypes *[]WorkspaceEnvironmentType `json:"environment-types,omitempty"`
	Tags             *[]string                   `json:"tags,omitempty"`
}

func (o *DriftDetectionWorkspaceFilter) IsEmpty() bool {
	return o == nil || (o.NamePatterns == nil && o.EnvironmentTypes == nil && o.Tags == nil)
}

func (o *DriftDetectionWorkspaceFilter) MarshalJSON() ([]byte, error) {
	type alias DriftDetectionWorkspaceFilter
	return json.Marshal((*alias)(o))
}

func (o *DriftDetectionWorkspaceFilter) UnmarshalJSON(data []byte) error {
	type alias DriftDetectionWorkspaceFilter
	aux := (*alias)(o)
	return json.Unmarshal(data, aux)
}

type DriftDetection struct {
	ID               string                        `jsonapi:"primary,drift-detection-schedule"`
	Schedule         DriftDetectionSchedulePeriod  `jsonapi:"attr,schedule"`
	WorkspaceFilters DriftDetectionWorkspaceFilter `jsonapi:"attr,workspace-filters"`
	RunMode          DriftDetectionScheduleRunMode `jsonapi:"attr,run-mode"`

	// Relations
	Environment *Environment `jsonapi:"relation,environment"`
}

// DriftDetectionCreateOptions represents the options for creating a new drift detection schedule.
type DriftDetectionCreateOptions struct {
	// For internal use only!
	ID string `jsonapi:"primary,drift-detection-schedule"`

	Schedule         DriftDetectionSchedulePeriod   `jsonapi:"attr,schedule"`
	WorkspaceFilters DriftDetectionWorkspaceFilter  `jsonapi:"attr,workspace-filters"`
	RunMode          *DriftDetectionScheduleRunMode `jsonapi:"attr,run-mode,omitempty"`

	// Relations
	Environment *Environment `jsonapi:"relation,environment"`
}

func (o DriftDetectionCreateOptions) valid() error {
	if o.Environment == nil {
		return errors.New("environment is required")
	}
	if !validStringID(&o.Environment.ID) {
		return errors.New("invalid value for environment ID")
	}
	if !o.Schedule.IsValid() {
		return errors.New("invalid value for schedule")
	}
	if o.RunMode != nil && !o.RunMode.IsValid() {
		return errors.New("invalid value for run_mode")
	}
	return nil
}

// DriftDetectionUpdateOptions represents the options for updating a drift detection schedule.
type DriftDetectionUpdateOptions struct {
	// For internal use only!
	ID string `jsonapi:"primary,drift-detection-schedule"`

	Schedule         DriftDetectionSchedulePeriod   `jsonapi:"attr,schedule"`
	WorkspaceFilters DriftDetectionWorkspaceFilter  `jsonapi:"attr,workspace-filters"`
	RunMode          *DriftDetectionScheduleRunMode `jsonapi:"attr,run-mode,omitempty"`

	// Relations
	Environment *Environment `jsonapi:"relation,environment"`
}

func (o DriftDetectionUpdateOptions) valid() error {
	if o.Environment == nil {
		return errors.New("environment is required")
	}
	if !validStringID(&o.Environment.ID) {
		return errors.New("invalid value for environment ID")
	}
	if !o.Schedule.IsValid() {
		return errors.New("invalid value for schedule")
	}
	if o.RunMode != nil && !o.RunMode.IsValid() {
		return errors.New("invalid value for run_mode")
	}
	return nil
}

// Create is used to create a new drift detection schedule.
func (s *driftDetections) Create(ctx context.Context, options DriftDetectionCreateOptions) (*DriftDetection, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}

	// Make sure we don't send a user-provided ID.
	options.ID = ""

	req, err := s.client.newRequest("POST", "drift-detection-schedules", &options)
	if err != nil {
		return nil, err
	}

	t := &DriftDetection{}
	err = s.client.do(ctx, req, t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// Read reads a drift detection schedule by its ID.
func (s *driftDetections) Read(ctx context.Context, id string) (*DriftDetection, error) {
	if !validStringID(&id) {
		return nil, errors.New("invalid value for drift detection schedule ID")
	}

	u := fmt.Sprintf("drift-detection-schedules/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	t := &DriftDetection{}
	err = s.client.do(ctx, req, t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// Update is used to update a drift detection schedule.
func (s *driftDetections) Update(ctx context.Context, id string, options DriftDetectionUpdateOptions) (*DriftDetection, error) {
	if !validStringID(&id) {
		return nil, errors.New("invalid value for drift detection schedule ID")
	}

	if err := options.valid(); err != nil {
		return nil, err
	}

	// Make sure we don't send a user-provided ID.
	options.ID = ""

	fmt.Printf("%s\n", id)

	u := fmt.Sprintf("drift-detection-schedules/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	t := &DriftDetection{}
	err = s.client.do(ctx, req, t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// Delete a drift detection schedule by its ID.
func (s *driftDetections) Delete(ctx context.Context, id string) error {
	if !validStringID(&id) {
		return errors.New("invalid value for drift detection schedule ID")
	}

	u := fmt.Sprintf("drift-detection-schedules/%s", url.QueryEscape(id))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
