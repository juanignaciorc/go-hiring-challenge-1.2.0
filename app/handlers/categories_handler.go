package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
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

// ListCategories handles GET /categories and returns all categories.
func (h *CategoriesHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.repo.ListCategories()
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]CategoryItem, len(cats))
	for i, c := range cats {
		out[i] = CategoryItem{Code: c.Code, Name: c.Name}
	}
	api.WriteJSON(w, http.StatusOK, out)
}

// CreateCategory handles POST /categories and creates a new category.
func (h *CategoriesHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var in CategoryItem
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := dec.Decode(&in); err != nil {
		api.BadRequest(w, "invalid JSON body")
		return
	}
	// Basic validation
	if in.Code == "" || in.Name == "" {
		api.BadRequest(w, "code and name are required")
		return
	}

	m := models.Category{Code: in.Code, Name: in.Name}
	if err := h.repo.CreateCategory(m); err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the created entity (without internal ID)
	api.WriteJSON(w, http.StatusCreated, CategoryItem{Code: m.Code, Name: m.Name})
}
