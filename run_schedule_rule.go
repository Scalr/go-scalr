package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ RunScheduleRules = (*runScheduleRules)(nil)

// RunScheduleRules describes all the run schedule rule related methods that the Scalr API supports.
type RunScheduleRules interface {
	List(ctx context.Context, options RunScheduleRuleListOptions) (*RunScheduleRulesList, error)
	Create(ctx context.Context, options RunScheduleRuleCreateOptions) (*RunScheduleRule, error)
	Read(ctx context.Context, ruleID string) (*RunScheduleRule, error)
	Delete(ctx context.Context, ruleID string) error
	Update(ctx context.Context, ruleID string, options RunScheduleRuleUpdateOptions) (*RunScheduleRule, error)
}

// runScheduleRules implements RunScheduleRules.
type runScheduleRules struct {
	client *Client
}

// RunScheduleRulesList represents a list of run schedule rules.
type RunScheduleRulesList struct {
	*Pagination
	Items []*RunScheduleRule
}

// ScheduleMode represents the run-type that will be scheduled.
type ScheduleMode string

const (
	ScheduleModeApply   ScheduleMode = "apply"
	ScheduleModeDestroy ScheduleMode = "destroy"
	ScheduleModeRefresh ScheduleMode = "refresh"
)

// RunScheduleRule represents a Scalr run schedule rule.
type RunScheduleRule struct {
	ID           string       `jsonapi:"primary,run-schedule-rules"`
	Schedule     string       `jsonapi:"attr,schedule"`
	ScheduleMode ScheduleMode `jsonapi:"attr,schedule-mode"`

	Workspace *Workspace `jsonapi:"relation,workspace,omitempty"`
}

// RunScheduleRuleListOptions represents the options for listing run schedule rules.
type RunScheduleRuleListOptions struct {
	ListOptions
	Workspace string `url:"filter[workspace],omitempty"`
	Include   string `url:"include,omitempty"`
}

// List all run schedule rules in the workspace.
func (s *runScheduleRules) List(ctx context.Context, options RunScheduleRuleListOptions) (*RunScheduleRulesList, error) {
	req, err := s.client.newRequest("GET", "run-schedule-rules", &options)
	if err != nil {
		return nil, err
	}

	runScheduleRulesList := &RunScheduleRulesList{}
	err = s.client.do(ctx, req, runScheduleRulesList)
	if err != nil {
		return nil, err
	}

	return runScheduleRulesList, nil
}

// RunScheduleRuleCreateOptions represents the options for creating a new run schedule rule.
type RunScheduleRuleCreateOptions struct {
	ID           string       `jsonapi:"primary,run-schedule-rules"`
	Schedule     string       `jsonapi:"attr,schedule"`
	ScheduleMode ScheduleMode `jsonapi:"attr,schedule-mode"`

	Workspace *Workspace `jsonapi:"relation,workspace,omitempty"`
}

// Create is used to create a new run schedule rule.
func (s *runScheduleRules) Create(ctx context.Context, options RunScheduleRuleCreateOptions) (*RunScheduleRule, error) {
	options.ID = ""

	req, err := s.client.newRequest("POST", "run-schedule-rules", &options)
	if err != nil {
		return nil, err
	}

	rule := &RunScheduleRule{}
	err = s.client.do(ctx, req, rule)

	if err != nil {
		return nil, err
	}

	return rule, nil
}

// Read a run schedule rule by ID.
func (s *runScheduleRules) Read(ctx context.Context, ruleID string) (*RunScheduleRule, error) {
	if !validStringID(&ruleID) {
		return nil, errors.New("invalid value for run schedule rule ID")
	}

	urlPath := fmt.Sprintf("run-schedule-rules/%s", url.QueryEscape(ruleID))

	options := struct {
		Include string `url:"include"`
	}{
		Include: "workspace",
	}

	req, err := s.client.newRequest("GET", urlPath, options)
	if err != nil {
		return nil, err
	}

	rule := &RunScheduleRule{}
	err = s.client.do(ctx, req, rule)
	if err != nil {
		return nil, err
	}

	return rule, nil
}

// RunScheduleRuleUpdateOptions represents the options for updating a run schedule rule.
type RunScheduleRuleUpdateOptions struct {
	ID           string        `jsonapi:"primary,run-schedule-rules"`
	Schedule     *string       `jsonapi:"attr,schedule,omitempty"`
	ScheduleMode *ScheduleMode `jsonapi:"attr,schedule-mode,omitempty"`
}

// Update an existing run schedule rule.
func (s *runScheduleRules) Update(ctx context.Context, ruleID string, options RunScheduleRuleUpdateOptions) (*RunScheduleRule, error) {
	if !validStringID(&ruleID) {
		return nil, errors.New("invalid value for run schedule rule ID")
	}

	urlPath := fmt.Sprintf("run-schedule-rules/%s", url.QueryEscape(ruleID))

	req, err := s.client.newRequest("PATCH", urlPath, &options)
	if err != nil {
		return nil, err
	}

	rule := &RunScheduleRule{}
	err = s.client.do(ctx, req, rule)
	if err != nil {
		return nil, err
	}

	return rule, nil
}

// Delete deletes a run schedule rule by its ID.
func (s *runScheduleRules) Delete(ctx context.Context, ruleID string) error {
	if !validStringID(&ruleID) {
		return errors.New("invalid value for run schedule rule ID")
	}

	urlPath := fmt.Sprintf("run-schedule-rules/%s", url.QueryEscape(ruleID))
	req, err := s.client.newRequest("DELETE", urlPath, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
