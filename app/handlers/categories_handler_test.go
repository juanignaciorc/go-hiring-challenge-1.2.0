package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
)

// stubCategoriesRepo is a test double implementing CategoriesRepository.
type stubCategoriesRepo struct {
	items       []models.Category
	err         error
	createErr   error
	createdItem models.Category
}

func (s *stubCategoriesRepo) ListCategories(_ context.Context) ([]models.Category, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.items, nil
}

func (s *stubCategoriesRepo) CreateCategory(_ context.Context, c models.Category) error {
	s.createdItem = c
	if s.createErr != nil {
		return s.createErr
	}
	return nil
}

func TestCategoriesHandler_ListCategories_Success(t *testing.T) {
	repo := stubCategoriesRepo{items: []models.Category{
		{Code: "clothing", Name: "Clothing"},
		{Code: "shoes", Name: "Shoes"},
		{Code: "accessories", Name: "Accessories"},
	}}
	h := NewCategoriesHandler(&repo)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	rr := httptest.NewRecorder()

	h.ListCategories(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var payload []api.CategoryItem
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	assert.Len(t, payload, 3)
	assert.Equal(t, api.CategoryItem{Code: "clothing", Name: "Clothing"}, payload[0])
	assert.Equal(t, api.CategoryItem{Code: "shoes", Name: "Shoes"}, payload[1])
	assert.Equal(t, api.CategoryItem{Code: "accessories", Name: "Accessories"}, payload[2])
}

func TestCategoriesHandler_ListCategories_Empty(t *testing.T) {
	repo := stubCategoriesRepo{items: []models.Category{}}
	h := NewCategoriesHandler(&repo)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	rr := httptest.NewRecorder()

	h.ListCategories(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var payload []api.CategoryItem
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	assert.Len(t, payload, 0)
}

func TestCategoriesHandler_ListCategories_Error(t *testing.T) {
	repo := stubCategoriesRepo{err: errors.New("db failed")}
	h := NewCategoriesHandler(&repo)

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

func TestCategoriesHandler_CreateCategory_Success(t *testing.T) {
	repo := &stubCategoriesRepo{}
	h := NewCategoriesHandler(repo)

	body := bytes.NewBufferString(`{"code":"new-cat","name":"New Category"}`)
	req := httptest.NewRequest(http.MethodPost, "/categories", body)
	rr := httptest.NewRecorder()

	h.CreateCategory(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var payload api.CategoryItem
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	assert.Equal(t, api.CategoryItem{Code: "new-cat", Name: "New Category"}, payload)
	// Ensure repo received the item
	if assert.NotNil(t, repo.createdItem) {
		assert.Equal(t, "new-cat", repo.createdItem.Code)
		assert.Equal(t, "New Category", repo.createdItem.Name)
	}
}

func TestCategoriesHandler_CreateCategory_BadJSON(t *testing.T) {
	repo := &stubCategoriesRepo{}
	h := NewCategoriesHandler(repo)

	body := bytes.NewBufferString(`{"code":"oops"`)
	req := httptest.NewRequest(http.MethodPost, "/categories", body)
	rr := httptest.NewRecorder()

	h.CreateCategory(rr, req)

	res := rr.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	var payload struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(res.Body).Decode(&payload)
	assert.Equal(t, "invalid JSON body", payload.Error)
}

func TestCategoriesHandler_CreateCategory_Validation(t *testing.T) {
	repo := &stubCategoriesRepo{}
	h := NewCategoriesHandler(repo)

	cases := []string{
		`{}`,
		`{"code":"","name":"X"}`,
		`{"code":"x","name":""}`,
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString(c))
		rr := httptest.NewRecorder()
		h.CreateCategory(rr, req)
		res := rr.Result()
		res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	}
}

func TestCategoriesHandler_CreateCategory_RepoError(t *testing.T) {
	repo := &stubCategoriesRepo{createErr: errors.New("db failed")}
	h := NewCategoriesHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString(`{"code":"c","name":"n"}`))
	rr := httptest.NewRecorder()

	h.CreateCategory(rr, req)

	res := rr.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	var payload struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(res.Body).Decode(&payload)
	assert.Equal(t, "db failed", payload.Error)
}
