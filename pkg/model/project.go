package model

import (
	"time"
)

// Project defines the project data.
type Project struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Owner        string    `json:"owner"`
	Status       string    `json:"status"`
	BugCompleted int64     `json:"bug_completed"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date,omitempty"`
	CompletedOn  time.Time `json:"completed_on,omitempty"`
	CreatedOn    time.Time `json:"created_on"`
	LastModified time.Time `json:"last_modified,omitempty"`
	CreatedBy    string    `json:"created_by"`
	ModifiedBy   string    `json:"modified_by,omitempty"`
	PublicAccess bool      `json:"public_access"`
	Version      int64     `json:"version"`
}
