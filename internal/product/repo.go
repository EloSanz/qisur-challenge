package product

import (
	"context"
	"errors"
	"time"

	"qisur-service/internal/category"

	"gorm.io/gorm"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type Repository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]Product, int64, error)
	FindByID(ctx context.Context, id string) (*Product, error)
	Create(ctx context.Context, prod *Product, categoryIDs []string) error
	Update(ctx context.Context, prod *Product, categoryIDs []string) error
	Delete(ctx context.Context, id string) error
	GetHistory(ctx context.Context, productID string, start, end time.Time) ([]ProductHistory, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, page, pageSize int) ([]Product, int64, error) {
	var products []Product
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.WithContext(ctx).Model(&Product{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Preload("Categories").Limit(pageSize).Offset(offset).Find(&products).Error
	return products, total, err
}

func (r *repository) FindByID(ctx context.Context, id string) (*Product, error) {
	var prod Product
	err := r.db.WithContext(ctx).Preload("Categories").First(&prod, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return &prod, nil
}

func (r *repository) Create(ctx context.Context, prod *Product, categoryIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(prod).Error; err != nil {
			return err
		}


		if len(categoryIDs) > 0 {
			var categories []category.Category
			if err := tx.Where("id IN ?", categoryIDs).Find(&categories).Error; err != nil {
				return err
			}
			if err := tx.Model(prod).Association("Categories").Append(categories); err != nil {
				return err
			}
		}


		history := ProductHistory{
			ProductID: prod.ID,
			Price:     prod.Price,
			Stock:     prod.Stock,
		}
		return tx.Create(&history).Error
	})
}

func (r *repository) Update(ctx context.Context, prod *Product, categoryIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		var oldProd Product
		if err := tx.First(&oldProd, "id = ?", prod.ID).Error; err != nil {
			return err
		}

		priceChanged := oldProd.Price != prod.Price
		stockChanged := oldProd.Stock != prod.Stock


		if err := tx.Save(prod).Error; err != nil {
			return err
		}


		if categoryIDs != nil {
			var categories []category.Category
			if len(categoryIDs) > 0 {
				if err := tx.Where("id IN ?", categoryIDs).Find(&categories).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(prod).Association("Categories").Replace(categories); err != nil {
				return err
			}
		}


		if priceChanged || stockChanged {
			history := ProductHistory{
				ProductID: prod.ID,
				Price:     prod.Price,
				Stock:     prod.Stock,
			}
			if err := tx.Create(&history).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *repository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&Product{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (r *repository) GetHistory(ctx context.Context, productID string, start, end time.Time) ([]ProductHistory, error) {
	var history []ProductHistory
	query := r.db.WithContext(ctx).Where("product_id = ?", productID)

	if !start.IsZero() {
		query = query.Where("changed_at >= ?", start)
	}
	if !end.IsZero() {
		query = query.Where("changed_at <= ?", end)
	}

	err := query.Order("changed_at DESC").Find(&history).Error
	return history, err
}
