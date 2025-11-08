package catalog

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
)

// stubCategoriesRepo is a test double implementing CategoriesRepository.
type stubCategoriesRepo struct {
	items []models.Category
	err   error
}

func (s stubCategoriesRepo) ListCategories() ([]models.Category, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.items, nil
}

func TestCategoriesHandler_ListCategories_Success(t *testing.T) {
	repo := stubCategoriesRepo{items: []models.Category{
		{Code: "clothing", Name: "Clothing"},
		{Code: "shoes", Name: "Shoes"},
		{Code: "accessories", Name: "Accessories"},
	}}
	h := NewCategoriesHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	rr := httptest.NewRecorder()

	h.ListCategories(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var payload []CategoryItem
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	assert.Len(t, payload, 3)
	assert.Equal(t, CategoryItem{Code: "clothing", Name: "Clothing"}, payload[0])
	assert.Equal(t, CategoryItem{Code: "shoes", Name: "Shoes"}, payload[1])
	assert.Equal(t, CategoryItem{Code: "accessories", Name: "Accessories"}, payload[2])
}

func TestCategoriesHandler_ListCategories_Empty(t *testing.T) {
	repo := stubCategoriesRepo{items: []models.Category{}}
	h := NewCategoriesHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	rr := httptest.NewRecorder()

	h.ListCategories(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var payload []CategoryItem
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	assert.Len(t, payload, 0)
}

func TestCategoriesHandler_ListCategories_Error(t *testing.T) {
	repo := stubCategoriesRepo{err: errors.New("db failed")}
	h := NewCategoriesHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	rr := httptest.NewRecorder()

	h.ListCategories(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	var payload struct {
		Error string `json:"error"`
	}
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	assert.Equal(t, "db failed", payload.Error)
}
