package model

// IssueStatus holds data for issues status report.
type IssuesStatus struct {
	Status string
	Count  int
}

// IssuesAssignee holds data for issues assignee report.
type IssuesAssignee struct {
	Name  string
	Count string
}
