package repositories

import (
	"context"

	"github.com/mytheresa/go-hiring-challenge/models"
	"gorm.io/gorm"
)

// CategoriesRepository provides operations for categories.
type CategoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) *CategoriesRepository {
	return &CategoriesRepository{db: db}
}

// ListCategories returns all categories.
func (r *CategoriesRepository) ListCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// CreateCategory persists a new category.
func (r *CategoriesRepository) CreateCategory(ctx context.Context, c models.Category) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var maxID int
		if err := tx.Model(&models.Category{}).Select("COALESCE(MAX(id), 0)").Scan(&maxID).Error; err != nil {
			return err
		}

		c.ID = uint(maxID + 1)
		return tx.Create(&c).Error
	})
}
