package catalog

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/models"
)

// ProductRepository defines the read operations needed by the catalog handler.
// It is satisfied by models.ProductsRepository and any other implementation
// providing the same behavior.
type ProductRepository interface {
	GetProducts(offset, limit int) ([]models.Product, int64, error)
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

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Parse offset with default 0
	offset := 0
	if v := q.Get("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			if parsed >= 0 {
				offset = parsed
			}
		}
	}

	// Parse limit with default 10, clamp to [1, 100]
	limit := 10
	if v := q.Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			limit = parsed
		}
	}
	if limit < 1 {
		limit = 1
	} else if limit > 100 {
		limit = 100
	}

	res, total, err := h.repo.GetProducts(offset, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	// Return the products as a JSON response
	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Total:    total,
		Products: products,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
