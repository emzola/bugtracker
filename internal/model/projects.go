package model

import (
	"time"

	"github.com/emzola/issuetracker/pkg/validator"
)

// Project defines project data.
type Project struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description,omitempty"`
	AssignedTo    *int64     `json:"assigned_to,omitempty"`
	StartDate     time.Time  `json:"start_date"`
	TargetEndDate time.Time  `json:"target_end_date"`
	ActualEndDate *time.Time `json:"actual_end_date,omitempty"`
	CreatedOn     time.Time  `json:"created_on"`
	CreatedBy     string     `json:"created_by"`
	ModifiedOn    time.Time  `json:"modified_on"`
	ModifiedBy    string     `json:"modified_by"`
	Version       int64      `json:"-"`
}

// Validate project data.
func (p Project) Validate(v *validator.Validator) {
	v.Check(p.Name != "", "name", "must be provided")
	v.Check(len(p.Name) >= 5, "name", "must not be less than 5 bytes long")
	v.Check(len(p.Name) <= 500, "name", "must not be more than 500 bytes long")
	v.Check(len(p.Description) >= 5, "description", "must not be less than 5 bytes long")
	v.Check(len(p.Description) <= 5000, "description", "must not be more than 5000 bytes long")
	v.Check(!p.StartDate.IsZero(), "start date", "must be provided")
	v.Check(!p.TargetEndDate.IsZero(), "target end date", "must be provided")
	v.Check(p.StartDate.Before(p.TargetEndDate), "target end date", "must not be before start date")
	if p.ActualEndDate != nil {
		v.Check(p.StartDate.Before(*p.ActualEndDate), "actual end date", "must not be before start date")
	}
}
