package api

// Category represents the public API shape of a category in product responses.
type Category struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Product represents the public API shape of a product in catalog endpoints.
type Product struct {
	Code     string   `json:"code"`
	Price    float64  `json:"price"`
	Category Category `json:"category"`
}

// Response represents the catalog response payload.
// It contains the total number of matched items and the current page of products.
type Response struct {
	Total    int64     `json:"total"`
	Products []Product `json:"products"`
}
