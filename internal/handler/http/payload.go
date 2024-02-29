package http

type createIssuePayload struct {
	Title                string `json:"title"`
	Description          string `json:"description"`
	ProjectID            int64  `json:"project_id"`
	AssignedTo           *int64 `json:"assigned_to"`
	Priority             string `json:"priority"`
	TargetResolutionDate string `json:"target_resolution_date"`
}

type updateIsssuePayload struct {
	Title                *string `json:"title"`
	Description          *string `json:"description"`
	AssignedTo           *int64  `json:"assigned_to"`
	Status               *string `json:"status"`
	Priority             *string `json:"priority"`
	TargetResolutionDate *string `json:"target_resolution_date"`
	Progress             *string `json:"progress"`
	ActualResolutionDate *string `json:"actual_resolution_date"`
	ResolutionSummary    *string `json:"resolution_summary"`
}

type createProjectPayload struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	AssignedTo    *int64 `json:"assigned_to"`
	StartDate     string `json:"start_date"`
	TargetEndDate string `json:"target_end_date"`
}

type updateProjectPayload struct {
	Name          *string `json:"name"`
	Description   *string `json:"description"`
	AssignedTo    *int64  `json:"assigned_to"`
	StartDate     *string `json:"start_date"`
	TargetEndDate *string `json:"target_end_date"`
	ActualEndDate *string `json:"actual_end_date"`
}

type createUserPayload struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type activateUserPayload struct {
	Token string `json:"token"`
}

type updateUserPayload struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
	Role  *string `json:"role"`
}

type assignUserToProjectPayload struct {
	ProjectID int64 `json:"project_id"`
}

type createActivationTokenPayload struct {
	Email string `json:"email"`
}

type createAuthenticationTokenPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
