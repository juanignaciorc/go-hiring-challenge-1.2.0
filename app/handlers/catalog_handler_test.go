package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// stubProductsRepo is a test double implementing ProductRepository.
// It records the last options and can return items, total, or an error.
type stubProductsRepo struct {
	items    []models.Product
	total    int64
	err      error
	lastOpts models.ListProductsOptions
	calls    int

	byCode      models.Product
	byCodeErr   error
	lastCodeArg string
}

func (s *stubProductsRepo) GetProducts(_ context.Context, opts models.ListProductsOptions) ([]models.Product, int64, error) {
	s.lastOpts = opts
	s.calls++
	if s.err != nil {
		return nil, 0, s.err
	}
	return s.items, s.total, nil
}

func (s *stubProductsRepo) GetProductByCode(_ context.Context, code string) (models.Product, error) {
	s.lastCodeArg = code
	if s.byCodeErr != nil {
		return models.Product{}, s.byCodeErr
	}
	return s.byCode, nil
}

func TestCatalogHandler_ListProducts_Success(t *testing.T) {
	repo := &stubProductsRepo{
		items: []models.Product{
			{Code: "P1", Price: decimal.NewFromInt(100), Category: models.Category{Code: "clothing", Name: "Clothing"}},
			{Code: "P2", Price: decimal.RequireFromString("29.95"), Category: models.Category{Code: "shoes", Name: "Shoes"}},
		},
		total: 42,
	}
	h := NewCatalogHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
	rr := httptest.NewRecorder()

	h.ListProducts(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var payload api.Response
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	assert.Equal(t, int64(42), payload.Total)
	if assert.Len(t, payload.Products, 2) {
		assert.Equal(t, "P1", payload.Products[0].Code)
		assert.InDelta(t, 100.0, payload.Products[0].Price, 0.0001)
		assert.Equal(t, api.Category{Code: "clothing", Name: "Clothing"}, payload.Products[0].Category)

		assert.Equal(t, "P2", payload.Products[1].Code)
		assert.InDelta(t, 29.95, payload.Products[1].Price, 0.0001)
		assert.Equal(t, api.Category{Code: "shoes", Name: "Shoes"}, payload.Products[1].Category)
	}

	// Defaults: offset=0, limit=10, no filters
	assert.Equal(t, models.ListProductsOptions{Offset: api.DefaultOffset, Limit: api.DefaultLimit}, repo.lastOpts)
}

func TestCatalogHandler_ListProducts_InvalidOffset(t *testing.T) {
	repo := &stubProductsRepo{}
	h := NewCatalogHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/catalog?"+url.Values{"offset": {"abc"}}.Encode(), nil)
	rr := httptest.NewRecorder()

	h.ListProducts(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	var payload struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(res.Body).Decode(&payload)
	assert.Equal(t, "offset must be an integer", payload.Error)
	assert.Equal(t, 0, repo.calls)
}

func TestCatalogHandler_ListProducts_InvalidLimit(t *testing.T) {
	repo := &stubProductsRepo{}
	h := NewCatalogHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/catalog?"+url.Values{"limit": {"x"}}.Encode(), nil)
	rr := httptest.NewRecorder()

	h.ListProducts(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	var payload struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(res.Body).Decode(&payload)
	assert.Equal(t, "limit must be an integer", payload.Error)
	assert.Equal(t, 0, repo.calls)
}

func TestCatalogHandler_ListProducts_ClampOffsetAndLimit(t *testing.T) {
	repo := &stubProductsRepo{}
	h := NewCatalogHandler(repo)

	// negative offset should clamp to 0, limit less than MinLimit clamps to 1
	req := httptest.NewRequest(http.MethodGet, "/catalog?"+url.Values{"offset": {"-5"}, "limit": {"0"}}.Encode(), nil)
	rr := httptest.NewRecorder()
	h.ListProducts(rr, req)
	res := rr.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, 1, repo.lastOpts.Limit)
	assert.Equal(t, 0, repo.lastOpts.Offset)

	// limit above MaxLimit should clamp to MaxLimit
	req = httptest.NewRequest(http.MethodGet, "/catalog?"+url.Values{"limit": {"9999"}}.Encode(), nil)
	rr = httptest.NewRecorder()
	h.ListProducts(rr, req)
	res = rr.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, api.MaxLimit, repo.lastOpts.Limit)
}

func TestCatalogHandler_ListProducts_CategoryNormalization(t *testing.T) {
	repo := &stubProductsRepo{}
	h := NewCatalogHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/catalog?"+url.Values{"category": {"  CLOThing  "}}.Encode(), nil)
	rr := httptest.NewRecorder()
	h.ListProducts(rr, req)
	res := rr.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "clothing", repo.lastOpts.CategoryCode)
}

func TestCatalogHandler_ListProducts_PriceLtParsing(t *testing.T) {
	// valid price_lt
	repo := &stubProductsRepo{}
	h := NewCatalogHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/catalog?"+url.Values{"price_lt": {"19.99"}}.Encode(), nil)
	rr := httptest.NewRecorder()
	h.ListProducts(rr, req)
	res := rr.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	if assert.NotNil(t, repo.lastOpts.PriceLessThan) {
		assert.InDelta(t, 19.99, *repo.lastOpts.PriceLessThan, 0.0001)
	}

	// empty price_lt -> nil pointer
	req = httptest.NewRequest(http.MethodGet, "/catalog?"+url.Values{"price_lt": {"   "}}.Encode(), nil)
	rr = httptest.NewRecorder()
	h.ListProducts(rr, req)
	res = rr.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Nil(t, repo.lastOpts.PriceLessThan)

	// non-numeric -> 400
	req = httptest.NewRequest(http.MethodGet, "/catalog?"+url.Values{"price_lt": {"abc"}}.Encode(), nil)
	rr = httptest.NewRecorder()
	h.ListProducts(rr, req)
	res = rr.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	var payload struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(res.Body).Decode(&payload)
	assert.Equal(t, "price_lt must be numeric", payload.Error)

	// negative -> 400
	req = httptest.NewRequest(http.MethodGet, "/catalog?"+url.Values{"price_lt": {"-1"}}.Encode(), nil)
	rr = httptest.NewRecorder()
	h.ListProducts(rr, req)
	res = rr.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	payload = struct {
		Error string `json:"error"`
	}{}
	_ = json.NewDecoder(res.Body).Decode(&payload)
	assert.Equal(t, "price_lt must be greater than or equal to 0", payload.Error)
}

func TestCatalogHandler_ListProducts_RepositoryError(t *testing.T) {
	repo := &stubProductsRepo{err: assert.AnError}
	h := NewCatalogHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
	rr := httptest.NewRecorder()

	h.ListProducts(rr, req)

	res := rr.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	var payload struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(res.Body).Decode(&payload)
	assert.Equal(t, assert.AnError.Error(), payload.Error)
}

func TestCatalogHandler_ProductDetails_Success(t *testing.T) {
	repo := &stubProductsRepo{}
	h := NewCatalogHandler(repo)

	// product with two variants: one priced, one inherits from product
	repo.byCode = models.Product{
		Code:     "P1",
		Price:    decimal.NewFromInt(100),
		Category: models.Category{Code: "clothing", Name: "Clothing"},
		Variants: []models.Variant{
			{Name: "Red", SKU: "SKU1", Price: decimal.RequireFromString("19.99")},
			{Name: "Blue", SKU: "SKU2"}, // zero price -> inherit 100
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/catalog/P1", nil)
	req.SetPathValue("code", "P1")
	rr := httptest.NewRecorder()

	h.ProductDetails(rr, req)

	res := rr.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var payload api.Product
	_ = json.NewDecoder(res.Body).Decode(&payload)
	assert.Equal(t, "P1", payload.Code)
	assert.InDelta(t, 100.0, payload.Price, 0.0001)
	assert.Equal(t, api.Category{Code: "clothing", Name: "Clothing"}, payload.Category)
	if assert.Len(t, payload.Variants, 2) {
		assert.Equal(t, api.Variant{Name: "Red", SKU: "SKU1", Price: 19.99}, payload.Variants[0])
		assert.Equal(t, api.Variant{Name: "Blue", SKU: "SKU2", Price: 100}, payload.Variants[1])
	}
}

func TestCatalogHandler_ProductDetails_NotFound(t *testing.T) {
	repo := &stubProductsRepo{byCodeErr: gorm.ErrRecordNotFound}
	h := NewCatalogHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/catalog/NOPE", nil)
	req.SetPathValue("code", "NOPE")
	rr := httptest.NewRecorder()

	h.ProductDetails(rr, req)

	res := rr.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	var body struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	_ = json.NewDecoder(res.Body).Decode(&body)
	assert.Equal(t, "product not found", body.Error)
	assert.Equal(t, "not_found", body.Code)
	assert.Equal(t, "NOPE", repo.lastCodeArg)
}
