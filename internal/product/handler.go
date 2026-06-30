package product

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"qisur-service/pkg/web"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaginationQuery struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

// GetAll retrieves all products
// @Summary Get all products
// @Description Retrieve a paginated list of all products.
// @Tags products
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Security BearerAuth
// @Success 200 {object} web.PaginatedResponse
// @Failure 400 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /api/products [get]
func (h *Handler) GetAll(c *gin.Context) {
	var query PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		web.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}

	products, total, err := h.svc.GetAll(c.Request.Context(), query.Page, query.PageSize)
	if err != nil {
		web.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	web.PaginatedJSON(c, http.StatusOK, web.PaginatedResponse{
		Items:      products,
		TotalCount: total,
		Page:       query.Page,
		PageSize:   query.PageSize,
	}, "Products retrieved")
}

// GetByID retrieves a single product by ID
// @Summary Get product by ID
// @Description Retrieve a specific product by its ID.
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Security BearerAuth
// @Success 200 {object} Product
// @Failure 404 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /api/products/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	prod, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			web.Error(c, http.StatusNotFound, "Product not found")
			return
		}
		web.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	web.JSON(c, http.StatusOK, prod, "Product retrieved")
}

// Create godoc
// @Summary Create a product
// @Description Create a new product
// @Tags products
// @Accept json
// @Produce json
// @Param request body CreateProductRequest true "Product data"
// @Security BearerAuth
// @Success 201 {object} Product
// @Failure 400 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /api/products [post]
func (h *Handler) Create(c *gin.Context) {
	traceID := uuid.New().String()
	// audit emit should be done in service or we can emit API_RECEIVED here
	// We need rabbitmq client to emit from handler, but handler doesn't have it.
	// We'll just pass traceID down for now, and ideally log API_RECEIVED.

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		web.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	prod := &Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}

	createdProd, err := h.svc.Create(c.Request.Context(), traceID, prod, req.CategoryIDs)
	if err != nil {
		web.Error(c, http.StatusInternalServerError, "Failed to create product")
		return
	}

	web.JSON(c, http.StatusCreated, createdProd, "Product created successfully")
}

// Update modifies an existing product
// @Summary Update product
// @Description Update fields of an existing product. Admin role required.
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param request body UpdateProductRequest true "Updated product data"
// @Security BearerAuth
// @Success 200 {object} Product
// @Failure 400 {object} web.Response
// @Failure 404 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /api/products/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	traceID := uuid.New().String()
	id := c.Param("id")

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		web.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	updatedProd, err := h.svc.Update(c.Request.Context(), traceID, id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			web.Error(c, http.StatusNotFound, "Product not found")
			return
		}
		web.Error(c, http.StatusInternalServerError, "Failed to update product")
		return
	}

	web.JSON(c, http.StatusOK, updatedProd, "Product updated")
}

// Delete removes a product
// @Summary Delete product
// @Description Delete a product by ID. Admin role required.
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Security BearerAuth
// @Success 200 {object} web.Response
// @Failure 404 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /api/products/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	traceID := uuid.New().String()
	id := c.Param("id")

	if err := h.svc.Delete(c.Request.Context(), traceID, id); err != nil {
		if errors.Is(err, ErrProductNotFound) {
			web.Error(c, http.StatusNotFound, "Product not found")
			return
		}
		web.Error(c, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	web.JSON(c, http.StatusOK, nil, "Product deleted")
}

// GetHistory retrieves the audit history of a product
// @Summary Get product history
// @Description Retrieve the price and stock change history of a product.
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param start query string false "Start date (RFC3339)"
// @Param end query string false "End date (RFC3339)"
// @Security BearerAuth
// @Success 200 {array} ProductHistory
// @Failure 500 {object} web.Response
// @Router /api/products/{id}/history [get]
func (h *Handler) GetHistory(c *gin.Context) {
	id := c.Param("id")
	startStr := c.Query("start")
	endStr := c.Query("end")

	var start, end time.Time
	if startStr != "" {
		start, _ = time.Parse(time.RFC3339, startStr)
	}
	if endStr != "" {
		end, _ = time.Parse(time.RFC3339, endStr)
	}

	history, err := h.svc.GetHistory(c.Request.Context(), id, start, end)
	if err != nil {
		web.Error(c, http.StatusInternalServerError, "Failed to retrieve history")
		return
	}

	web.JSON(c, http.StatusOK, history, "Product history retrieved")
}
