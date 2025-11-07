package models

import (
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

// GetProducts retrieves a paginated list of products along with the total count.
func (r *ProductsRepository) GetProducts(offset, limit int) ([]Product, int64, error) {
	var (
		products []Product
		total    int64
	)

	// Count total products available (without pagination)
	if err := r.db.Model(&Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and preload associations
	if err := r.db.Preload("Category").Preload("Variants").
		Offset(offset).Limit(limit).
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
