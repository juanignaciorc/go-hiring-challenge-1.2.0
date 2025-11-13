package repositories

import (
	"context"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

// GetProductByCode fetches a single product by its unique code with its Category and Variants preloaded.
func (r *ProductsRepository) GetProductByCode(ctx context.Context, code string) (models.Product, error) {
	var p models.Product
	if err := r.db.WithContext(ctx).Preload("Category").Preload("Variants").
		Where("code = ?", code).First(&p).Error; err != nil {
		return models.Product{}, err
	}
	return p, nil
}

// Scopes for query reuse and safer composition
func scopeJoinCategoriesIfFiltering(code string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if code == "" {
			return db
		}
		// Use explicit LEFT JOIN with quoted table/column names to be stable across drivers
		return db.Joins("LEFT JOIN \"categories\" ON \"categories\".\"id\" = \"products\".\"category_id\"")
	}
}

func scopeFilterCategory(code string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if code == "" {
			return db
		}
		return db.Where("categories.code = ?", code)
	}
}

func scopeFilterPriceLT(pricePtr *float64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if pricePtr == nil {
			return db
		}
		price := decimal.NewFromFloat(*pricePtr)
		return db.Where("products.price < ?", price)
	}
}

// GetProducts retrieves a filtered and paginated list of products along with the total count after filters.
func (r *ProductsRepository) GetProducts(ctx context.Context, opts models.ListProductsOptions) ([]models.Product, int64, error) {
	var (
		products []models.Product
		total    int64
	)

	// Build base query anchored on concrete table name for determinism across naming strategies
	base := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Table((&models.Product{}).TableName()). // ensure base table name is explicit
		Scopes(scopeJoinCategoriesIfFiltering(opts.CategoryCode)).
		Scopes(scopeFilterCategory(opts.CategoryCode)).
		Scopes(scopeFilterPriceLT(opts.PriceLessThan))

	// Count total after filters
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination to the filtered query and preload associations
	q := base.Session(&gorm.Session{})
	if opts.Offset > 0 {
		q = q.Offset(opts.Offset)
	}
	if opts.Limit > 0 {
		q = q.Limit(opts.Limit)
	}

	if err := q.Preload("Category").Preload("Variants").
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
