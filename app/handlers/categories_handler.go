package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/errs"
	"github.com/mytheresa/go-hiring-challenge/app/middleware"
	"github.com/mytheresa/go-hiring-challenge/models"
)

// CategoriesRepository defines the operations needed by the categories handler.
type CategoriesRepository interface {
	ListCategories(ctx context.Context) ([]models.Category, error)
	CreateCategory(ctx context.Context, c models.Category) error
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
	middleware.Serve(w, r, h.listCategories)
}

func (h *CategoriesHandler) listCategories(w http.ResponseWriter, r *http.Request) error {
	cats, err := h.repo.ListCategories(r.Context())
	if err != nil {
		return err
	}

	out := make([]api.CategoryItem, len(cats))
	for i, c := range cats {
		out[i] = api.CategoryItem{Code: c.Code, Name: c.Name}
	}
	api.WriteJSON(w, http.StatusOK, out)
	return nil
}

// CreateCategory handles POST /categories and creates a new category.
func (h *CategoriesHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	middleware.Serve(w, r, h.createCategory)
}

func (h *CategoriesHandler) createCategory(w http.ResponseWriter, r *http.Request) error {
	var in api.CategoryItem
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := dec.Decode(&in); err != nil {
		return errs.Invalid("invalid JSON body")
	}
	// Basic validation
	if in.Code == "" || in.Name == "" {
		return errs.Invalid("code and name are required")
	}

	m := models.Category{Code: in.Code, Name: in.Name}
	if err := h.repo.CreateCategory(r.Context(), m); err != nil {
		return err
	}

	// Return the created entity (without internal ID)
	api.WriteJSON(w, http.StatusCreated, api.CategoryItem{Code: m.Code, Name: m.Name})
	return nil
}
