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

// GetProducts retrieves a filtered and paginated list of products along with the total count after filters.
func (r *ProductsRepository) GetProducts(ctx context.Context, opts models.ListProductsOptions) ([]models.Product, int64, error) {
	var (
		products []models.Product
		total    int64
	)

	// Start a base query joining category to allow filtering by its code
	base := r.db.WithContext(ctx).Model(&models.Product{}).Joins("Category")

	// Apply filters
	if opts.CategoryCode != "" {
		// Use the JOIN alias that GORM creates for the Category association to ensure the filter applies correctly in real DBs.
		base = base.Where("\"Category\".\"code\" = ?", opts.CategoryCode)
	}
	if opts.PriceLessThan != nil {
		// Use decimal for exact comparison
		price := decimal.NewFromFloat(*opts.PriceLessThan)
		base = base.Where("products.price < ?", price)
	}

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
