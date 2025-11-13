package repositories

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
)

func TestProductsRepository_GetProducts_NoFilters_EmptyResult(t *testing.T) {
	db, mock, cleanup := newGormWithMock(t)
	defer cleanup()

	r := NewProductsRepository(db)

	// Expect count query without filters (no joins needed)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Expect main select without joins and no results (projection may vary by GORM)
	mock.ExpectQuery(`SELECT .* FROM "products"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "price", "category_id"}))

	ctx := context.Background()
	items, total, err := r.GetProducts(ctx, models.ListProductsOptions{})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, items, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductsRepository_GetProducts_FilterCategoryAndPrice_WithPagination_Success(t *testing.T) {
	db, mock, cleanup := newGormWithMock(t)
	defer cleanup()

	r := NewProductsRepository(db)

	price := 20.0
	opts := models.ListProductsOptions{
		CategoryCode:  "shoes",
		PriceLessThan: &price,
		Offset:        2,
		Limit:         3,
	}

	// Count with filters (LEFT JOIN categories + WHERE on table-qualified columns)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products" LEFT JOIN "categories" ON "categories"."id" = "products"."category_id" WHERE categories.code = $1 AND products.price < $2`)).
		WithArgs("shoes", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Main select with same filters and pagination
	mock.ExpectQuery(`SELECT .* FROM "products" LEFT JOIN "categories" ON "categories"\."id" = "products"\."category_id" WHERE categories\.code = \$1 AND products\.price < \$2 LIMIT \$3 OFFSET \$4`).
		WithArgs("shoes", sqlmock.AnyArg(), 3, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "price", "category_id"}).
			AddRow(10, "PROD010", "12.00", 5).
			AddRow(11, "PROD011", "19.99", 5))

	// Preload Category for found product category IDs (may collapse to single equality by GORM)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categories" WHERE "categories"."id" = $1`)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name"}).
			AddRow(5, "shoes", "Shoes"))

	// Preload Variants for found product IDs
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "product_variants" WHERE "product_variants"."product_id" IN ($1,$2)`)).
		WithArgs(10, 11).
		WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "name", "sku", "price"}).
			AddRow(100, 10, "Variant A", "SKU010A", "11.00").
			AddRow(101, 11, "Variant A", "SKU011A", nil))

	ctx := context.Background()
	items, total, err := r.GetProducts(ctx, opts)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
	assert.Equal(t, uint(10), items[0].ID)
	assert.Equal(t, "PROD010", items[0].Code)
	assert.Equal(t, uint(5), items[0].CategoryID)
	assert.Equal(t, "shoes", items[0].Category.Code)
	assert.Len(t, items[0].Variants, 1)
	assert.Equal(t, uint(100), items[0].Variants[0].ID)
	assert.Equal(t, uint(11), items[1].ID)
	assert.Equal(t, "PROD011", items[1].Code)
	assert.Equal(t, uint(5), items[1].CategoryID)
	assert.Equal(t, "shoes", items[1].Category.Code)
	assert.Len(t, items[1].Variants, 1)
	assert.Equal(t, uint(101), items[1].Variants[0].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductsRepository_GetProducts_CountError(t *testing.T) {
	db, mock, cleanup := newGormWithMock(t)
	defer cleanup()

	r := NewProductsRepository(db)

	// Force error on count
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
		WillReturnError(assert.AnError)

	ctx := context.Background()
	items, total, err := r.GetProducts(ctx, models.ListProductsOptions{})
	assert.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, items)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductsRepository_GetProducts_SelectError(t *testing.T) {
	db, mock, cleanup := newGormWithMock(t)
	defer cleanup()

	r := NewProductsRepository(db)

	// Count succeeds (no joins needed)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Main select fails
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products"`)).
		WillReturnError(assert.AnError)

	ctx := context.Background()
	items, total, err := r.GetProducts(ctx, models.ListProductsOptions{})
	assert.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, items)
	assert.NoError(t, mock.ExpectationsWereMet())
}
