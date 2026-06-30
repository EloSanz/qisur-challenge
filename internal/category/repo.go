package category

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
)

type Repository interface {
	FindAll(ctx context.Context) ([]Category, error)
	FindByID(ctx context.Context, id string) (*Category, error)
	Create(ctx context.Context, cat *Category) error
	Update(ctx context.Context, cat *Category) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context) ([]Category, error) {
	var categories []Category
	err := r.db.WithContext(ctx).Find(&categories).Error
	return categories, err
}

func (r *repository) FindByID(ctx context.Context, id string) (*Category, error) {
	var cat Category
	err := r.db.WithContext(ctx).First(&cat, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return &cat, nil
}

func (r *repository) Create(ctx context.Context, cat *Category) error {
	return r.db.WithContext(ctx).Create(cat).Error
}

func (r *repository) Update(ctx context.Context, cat *Category) error {
	return r.db.WithContext(ctx).Save(cat).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&Category{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCategoryNotFound
	}
	return nil
}
