package model

// IssueStatus holds data for issues status report.
type IssuesStatus struct {
	Status string
	Count  int64
}

// IssuesAssignee holds data for issues assignee report.
type IssuesAssignee struct {
	AssigneeID   int64
	AssigneeName string
	Count        int64
}
