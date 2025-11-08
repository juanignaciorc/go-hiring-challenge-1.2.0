package catalog

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/mytheresa/go-hiring-challenge/models"
)

// Centralized error messages.
const (
	errOffsetMustBeInt = "offset must be an integer"
	errLimitMustBeInt  = "limit must be an integer"
	errPriceMustBeNum  = "price_lt must be numeric"
	errPriceGteZero    = "price_lt must be greater than or equal to 0"
)

// ProductRepository defines the read operations needed by the catalog handler.
// It is satisfied by models.ProductsRepository and any other implementation
// providing the same behavior.
type ProductRepository interface {
	GetProducts(opts models.ListProductsOptions) ([]models.Product, int64, error)
}

type Response struct {
	Total    int64     `json:"total"`
	Products []Product `json:"products"`
}

type Category struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Product struct {
	Code     string   `json:"code"`
	Price    float64  `json:"price"`
	Category Category `json:"category"`
}

type CatalogHandler struct {
	repo ProductRepository
}

func NewCatalogHandler(r ProductRepository) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

// Defaults and limits for pagination.
const (
	defaultOffset = 0
	defaultLimit  = 10
	minLimit      = 1
	maxLimit      = 100
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func badRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, errorResponse{Error: msg})
}

// ListProducts processes GET /catalog requests by parsing and validating query parameters,
// delegating to the repository, mapping domain models to API types, and writing the JSON response.
func (h *CatalogHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	offset, ok, msg := parseOffset(q.Get("offset"))
	if !ok {
		badRequest(w, msg)
		return
	}

	limit, ok, msg := parseLimit(q.Get("limit"))
	if !ok {
		badRequest(w, msg)
		return
	}

	category := normalize(q.Get("category"))

	pricePtr, ok, msg := parsePriceLT(q.Get("price_lt"))
	if !ok {
		badRequest(w, msg)
		return
	}

	opts := models.ListProductsOptions{
		Offset:        offset,
		Limit:         limit,
		CategoryCode:  category,
		PriceLessThan: pricePtr,
	}

	res, total, err := h.repo.GetProducts(opts)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	// Map response
	products := make([]Product, len(res))
	for i, p := range res {
		products[i] = Product{
			Code:  p.Code,
			Price: p.Price.InexactFloat64(),
			Category: Category{
				Code: p.Category.Code,
				Name: p.Category.Name,
			},
		}
	}

	writeJSON(w, http.StatusOK, Response{
		Total:    total,
		Products: products,
	})
}

// Helpers extracted for clarity and reuse.

// parseOffset parses the "offset" query parameter.
// - Empty input returns the defaultOffset.
// - Non-integer input returns ok=false and a user-facing error message.
// - Negative values are clamped to 0.
func parseOffset(raw string) (int, bool, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultOffset, true, ""
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false, errOffsetMustBeInt
	}
	if v < 0 {
		v = 0
	}
	return v, true, ""
}

// parseLimit parses the "limit" query parameter.
// - Empty input returns the defaultLimit.
// - Non-integer input returns ok=false and a user-facing error message.
// - The result is clamped into the inclusive range [minLimit, maxLimit].
func parseLimit(raw string) (int, bool, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultLimit, true, ""
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false, errLimitMustBeInt
	}
	if v < minLimit {
		v = minLimit
	} else if v > maxLimit {
		v = maxLimit
	}
	return v, true, ""
}

// parsePriceLT parses the "price_lt" query parameter.
// - Empty input returns nil to indicate "no filter".
// - Non-numeric input returns ok=false and a user-facing error message.
// - Values must be >= 0; otherwise ok=false is returned.
// - On success, returns a pointer to the parsed float64 to distinguish from "not provided".
func parsePriceLT(raw string) (*float64, bool, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, true, ""
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, false, errPriceMustBeNum
	}
	if f < 0 {
		return nil, false, errPriceGteZero
	}
	return &f, true, ""
}

// normalize trims surrounding spaces and lowercases the input to build case-insensitive filters.
func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
