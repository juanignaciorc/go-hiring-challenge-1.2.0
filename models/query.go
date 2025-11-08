package models

// ListProductsOptions holds pagination and filter options for listing products.
// Zero values mean "not set"; callers should pre-validate ranges when needed.
type ListProductsOptions struct {
	Offset int
	Limit  int
	// CategoryCode filters products that belong to the category with this code.
	// Empty string means no filter.
	CategoryCode string
	// PriceLessThan, when non-nil, filters products whose price is strictly less than this value.
	// The unit is the same as stored in the DB (e.g., EUR). Nil means no filter.
	PriceLessThan *float64
}
