package model

import (
	"time"

	"github.com/emzola/issuetracker/pkg/validator"
)

// Issue defines issue data.
type Issue struct {
	ID                   int64      `json:"id"`
	Title                string     `json:"title"`
	Description          string     `json:"description,omitempty"`
	ReporterID           int64      `json:"reporter_id"`
	ReportedDate         time.Time  `json:"reported_date"`
	ProjectID            int64      `json:"project_id"`
	AssignedTo           *int64     `json:"assigned_to,omitempty"`
	Status               string     `json:"status"`
	Priority             string     `json:"priority"`
	TargetResolutionDate time.Time  `json:"target_resolution_date"`
	Progress             string     `json:"progress,omitempty"`
	ActualResolutionDate *time.Time `json:"actual_resolution_date,omitempty"`
	ResolutionSummary    string     `json:"resolution_summary,omitempty"`
	CreatedOn            time.Time  `json:"created_on"`
	CreatedBy            string     `json:"created_by"`
	ModifiedOn           time.Time  `json:"modified_on"`
	ModifiedBy           string     `json:"modified_by"`
	Version              int64      `json:"-"`
}

// Validate issue data.
func (i Issue) Validate(v *validator.Validator) {
	v.Check(i.Title != "", "title", "must be provided")
	v.Check(len(i.Title) >= 5, "title", "must not be less than 5 bytes")
	v.Check(len(i.Title) <= 500, "iitle", "must not be more than 500 bytes")
	v.Check(len(i.Description) >= 5, "description", "must not be less than 5 bytes long")
	v.Check(len(i.Description) <= 5000, "description", "must not be more than 5000 bytes long")
	if !i.ReportedDate.IsZero() {
		v.Check(!i.ReportedDate.IsZero(), "reported date", "must be provided")
	}
	v.Check(!i.TargetResolutionDate.IsZero(), "target resolution date", "must be provided")
	v.Check(i.TargetResolutionDate.After(i.ReportedDate), "target resolution date", "must not be before reported date")
	if i.Progress != "" {
		v.Check(len(i.Progress) >= 5, "progress", "must not be less than 5 bytes long")
		v.Check(len(i.Progress) <= 1000, "progress", "must not be more than 1000 bytes long")
	}
	if i.ResolutionSummary != "" {
		v.Check(len(i.ResolutionSummary) >= 5, "resolution summary", "must not be less than 5 bytes long")
		v.Check(len(i.ResolutionSummary) <= 1000, "resolution summary", "must not be more than 1000 bytes long")
	}
	if i.ActualResolutionDate != nil {
		v.Check(i.ActualResolutionDate.After(i.ReportedDate), "actual resolution date", "must not be before reported date")
	}
}
