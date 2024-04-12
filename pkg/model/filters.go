package model

import (
	"strings"

	"github.com/emzola/issuetracker/pkg/validator"
)

// Filters defines sorting and pagination data.
type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

// Validate Filters.
func (f Filters) Validate(v *validator.Validator) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")
	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

// SortColumn sorts
func (f Filters) SortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter:" + f.Sort)
}

// SortDirection sorts by ascending or descending order.
func (f Filters) SortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

// Limit returns the page size for pagination.
func (f Filters) Limit() int {
	return f.PageSize
}

// Offset returns the number of rows to skip for pagination.
func (f Filters) Offset() int {
	return (f.Page - 1) * f.PageSize
}
