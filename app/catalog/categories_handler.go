package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/models"
)

// CategoriesRepository defines the read operations needed by the categories handler.
type CategoriesRepository interface {
	ListCategories() ([]models.Category, error)
}

// CategoryItem is the API representation of a category.
type CategoryItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// CategoriesHandler serves requests related to categories.
type CategoriesHandler struct {
	repo CategoriesRepository
}

func NewCategoriesHandler(r CategoriesRepository) *CategoriesHandler {
	return &CategoriesHandler{repo: r}
}

// writeJSON is a minimal helper for JSON responses specific to this package.
func writeJSONCategories(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// errorResponse mirrors the one used by the catalog handler for consistency.
type errorResponseCategories struct {
	Error string `json:"error"`
}

// ListCategories handles GET /categories and returns all categories.
func (h *CategoriesHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.repo.ListCategories()
	if err != nil {
		writeJSONCategories(w, http.StatusInternalServerError, errorResponseCategories{Error: err.Error()})
		return
	}

	out := make([]CategoryItem, len(cats))
	for i, c := range cats {
		out[i] = CategoryItem{Code: c.Code, Name: c.Name}
	}
	writeJSONCategories(w, http.StatusOK, out)
}
