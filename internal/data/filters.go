package data

import "moviego.madhav.net/internal/validator"

type Filters struct {
	Page     int
	PageSize int
	Sort     string
	SortSafelist []string
}

// Validate method to validate the filters struct
func ValidateFilters(v *validator.Validator, f Filters) {
	// Check that the page and page size parameters contain sensible values
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}