package scalr

// Bool returns a pointer to the given bool
func Bool(v bool) *bool {
	return &v
}

// Category returns a pointer to the given category type.
func Category(v CategoryType) *CategoryType {
	return &v
}

// Int returns a pointer to the given int.
func Int(v int) *int {
	return &v
}

// Int64 returns a pointer to the given int64.
func Int64(v int64) *int64 {
	return &v
}

// String returns a pointer to the given string.
func String(v string) *string {
	return &v
}

// WorkspaceExecutionModePtr returns a pointer to the given execution mode
func WorkspaceExecutionModePtr(v WorkspaceExecutionMode) *WorkspaceExecutionMode {
	return &v
}

// AutoQueueRunsModePtr returns a pointer to the given auto queue runs mode
func AutoQueueRunsModePtr(v WorkspaceAutoQueueRuns) *WorkspaceAutoQueueRuns {
	return &v
}

// ServiceAccountStatusPtr returns a pointer to the given service account status value.
func ServiceAccountStatusPtr(v ServiceAccountStatus) *ServiceAccountStatus {
	return &v
}

// WorkspaceIaCPlatformPtr returns a pointer to the given execution mode
func WorkspaceIaCPlatformPtr(v WorkspaceIaCPlatform) *WorkspaceIaCPlatform {
	return &v
}
