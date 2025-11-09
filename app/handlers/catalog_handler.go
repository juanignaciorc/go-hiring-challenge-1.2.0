package handlers

import (
	"context"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

// ProductRepository defines the read operations needed by the catalog handler.
// It is satisfied by models.ProductsRepository and any other implementation
// providing the same behavior.
type ProductRepository interface {
	GetProducts(ctx context.Context, opts models.ListProductsOptions) ([]models.Product, int64, error)
}

type CatalogHandler struct {
	repo ProductRepository
}

func NewCatalogHandler(r ProductRepository) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

// ListProducts processes GET /catalog requests by parsing and validating query parameters,
// delegating to the repository, mapping domain models to API types, and writing the JSON response.
func (h *CatalogHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	offset, ok, msg := api.ParseOffset(q.Get("offset"))
	if !ok {
		api.BadRequest(w, msg)
		return
	}

	limit, ok, msg := api.ParseLimit(q.Get("limit"))
	if !ok {
		api.BadRequest(w, msg)
		return
	}

	category := api.Normalize(q.Get("category"))

	pricePtr, ok, msg := api.ParsePriceLT(q.Get("price_lt"))
	if !ok {
		api.BadRequest(w, msg)
		return
	}

	opts := models.ListProductsOptions{
		Offset:        offset,
		Limit:         limit,
		CategoryCode:  category,
		PriceLessThan: pricePtr,
	}

	res, total, err := h.repo.GetProducts(r.Context(), opts)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Map response
	products := make([]api.Product, len(res))
	for i, p := range res {
		products[i] = api.Product{
			Code:  p.Code,
			Price: p.Price.InexactFloat64(),
			Category: api.Category{
				Code: p.Category.Code,
				Name: p.Category.Name,
			},
		}
	}

	api.OKResponse(w, api.Response{
		Total:    total,
		Products: products,
	})
}
