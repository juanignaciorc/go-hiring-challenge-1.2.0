package models

import "gorm.io/gorm"

// CategoriesRepository provides read operations for categories.
type CategoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) *CategoriesRepository {
	return &CategoriesRepository{db: db}
}

// ListCategories returns all categories.
func (r *CategoriesRepository) ListCategories() ([]Category, error) {
	var categories []Category
	if err := r.db.Order("id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}
