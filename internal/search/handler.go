package search

import (
	"net/http"
	"strings"

	"qisur-service/internal/category"
	"qisur-service/internal/product"
	"qisur-service/pkg/web"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// Search handles GET /api/search?type={product/category}&q={query}
// Search searches for products
// @Summary Search products
// @Description Search products by query string. Includes associated categories.
// @Tags search
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Security BearerAuth
// @Success 200 {array} product.Product
// @Failure 400 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /api/search [get]
func (h *Handler) Search(c *gin.Context) {
	searchType := strings.ToLower(c.Query("type"))
	query := c.Query("q")

	if query == "" {
		web.Error(c, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	searchPattern := "%" + query + "%"

	if searchType == "category" {
		var categories []category.Category
		if err := h.db.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern).Find(&categories).Error; err != nil {
			web.Error(c, http.StatusInternalServerError, "Failed to search categories")
			return
		}
		web.JSON(c, http.StatusOK, categories, "Categories found")
		return
	}

	// Default to product search
	var products []product.Product
	if err := h.db.Preload("Categories").Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern).Find(&products).Error; err != nil {
		web.Error(c, http.StatusInternalServerError, "Failed to search products")
		return
	}
	web.JSON(c, http.StatusOK, products, "Products found")
}
