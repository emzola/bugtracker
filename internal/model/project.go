package model

import (
	"time"

	"github.com/emzola/bugtracker/pkg/validator"
)

// Project defines the project data.
type Project struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Description  string     `json:"description,omitempty"`
	Owner        string     `json:"owner"`
	Status       string     `json:"status"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	CompletedOn  *time.Time `json:"completed_on,omitempty"`
	CreatedOn    time.Time  `json:"created_on"`
	LastModified time.Time  `json:"last_modified"`
	CreatedBy    string     `json:"created_by"`
	ModifiedBy   string     `json:"modified_by"`
	PublicAccess bool       `json:"public_access"`
	Version      int64      `json:"-"`
}

// Validate project data
func (p Project) Validate(v *validator.Validator) {
	v.Check(p.Name != "", "name", "must be provided")
	v.Check(len(p.Name) >= 5, "name", "must not be less than 5 bytes long")
	v.Check(len(p.Name) <= 500, "name", "must not be more than 500 bytes long")
	v.Check(len(p.Description) >= 5, "description", "must not be less than 5 bytes long")
	v.Check(len(p.Description) <= 1000, "description", "must not be more than 1000 bytes long")
	if p.EndDate != nil {
		v.Check(p.StartDate != nil, "start date", "must be provided")
		if p.StartDate != nil {
			v.Check(p.StartDate.Before(*p.EndDate), "end date", "must not be before start date")
		}
	}
	if p.CompletedOn != nil {
		if p.StartDate != nil {
			v.Check(p.StartDate.Before(*p.CompletedOn), "completed on", "must not be before start date")
		}
	}
}
