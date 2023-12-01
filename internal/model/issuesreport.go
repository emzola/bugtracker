package model

// IssueStatus holds data for issues status report.
type IssuesStatus struct {
	Status      string `json:"issue_status"`
	IssuesCount int64  `json:"issues_count"`
}

// IssuesAssignee holds data for issues assignee report.
type IssuesAssignee struct {
	AssigneeID     int64  `json:"assignee_id"`
	AssigneeName   string `json:"assignee_name"`
	IssuesAssigned int64  `json:"issues_assigned"`
}

// IssuesReporter holds data for issues reporter report.
type IssuesReporter struct {
	ReporterID     int64  `json:"reporter_id"`
	ReporterName   string `json:"reporter_name"`
	IssuesReported int64  `json:"issues_reported"`
}

// IssuesPriority holds data for issues priority report.
type IssuesPriority struct {
	Priority    string `json:"issue_priority"`
	IssuesCount int64  `json:"issues_count"`
}
