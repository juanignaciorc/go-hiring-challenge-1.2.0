package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/models"
)

// CategoriesRepository defines the operations needed by the categories handler.
type CategoriesRepository interface {
	ListCategories() ([]models.Category, error)
	CreateCategory(c models.Category) error
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

// CreateCategory handles POST /categories and creates a new category.
func (h *CategoriesHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var in CategoryItem
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := dec.Decode(&in); err != nil {
		writeJSONCategories(w, http.StatusBadRequest, errorResponseCategories{Error: "invalid JSON body"})
		return
	}
	// Basic validation
	if in.Code == "" || in.Name == "" {
		writeJSONCategories(w, http.StatusBadRequest, errorResponseCategories{Error: "code and name are required"})
		return
	}

	m := models.Category{Code: in.Code, Name: in.Name}
	if err := h.repo.CreateCategory(m); err != nil {
		writeJSONCategories(w, http.StatusInternalServerError, errorResponseCategories{Error: err.Error()})
		return
	}

	// Return the created entity (without internal ID)
	writeJSONCategories(w, http.StatusCreated, CategoryItem{Code: m.Code, Name: m.Name})
}
