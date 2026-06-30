package category

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
)

type Repository interface {
	FindAll() ([]Category, error)
	FindByID(id string) (*Category, error)
	Create(cat *Category) error
	Update(cat *Category) error
	Delete(id string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll() ([]Category, error) {
	var categories []Category
	err := r.db.Find(&categories).Error
	return categories, err
}

func (r *repository) FindByID(id string) (*Category, error) {
	var cat Category
	err := r.db.First(&cat, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return &cat, nil
}

func (r *repository) Create(cat *Category) error {
	return r.db.Create(cat).Error
}

func (r *repository) Update(cat *Category) error {
	return r.db.Save(cat).Error
}

func (r *repository) Delete(id string) error {
	result := r.db.Delete(&Category{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCategoryNotFound
	}
	return nil
}
