package product

import (
	"time"

	"qisur-service/internal/category"

	"gorm.io/gorm"
)

// Product represents a product in the system
type Product struct {
	ID          string              `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name        string              `gorm:"type:varchar(255);not null" json:"name"`
	Description string              `gorm:"type:text" json:"description"`
	Price       float64             `gorm:"type:numeric(12,2);not null" json:"price"`
	Stock       int                 `gorm:"type:integer;not null;default:0" json:"stock"`
	Categories  []category.Category `gorm:"many2many:product_categories;" json:"categories"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	DeletedAt   gorm.DeletedAt      `gorm:"index" json:"-"`
}

// ProductHistory tracks changes to a product's price or stock
type ProductHistory struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	ProductID string    `gorm:"type:uuid;not null" json:"product_id"`
	Price     float64   `gorm:"type:numeric(12,2);not null" json:"price"`
	Stock     int       `gorm:"type:integer;not null" json:"stock"`
	ChangedAt time.Time `gorm:"autoCreateTime" json:"changed_at"`
}

type CreateProductRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Price       float64  `json:"price" binding:"required,gt=0"`
	Stock       int      `json:"stock" binding:"min=0"`
	CategoryIDs []string `json:"category_ids"`
}

type UpdateProductRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
	CategoryIDs []string `json:"category_ids"` // If provided, replaces existing categories
}
