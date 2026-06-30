package category

import (
	"errors"
	"net/http"

	"qisur-service/pkg/web"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

// GetAll retrieves all categories
// @Summary Get all categories
// @Description Retrieve a list of all categories.
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Category
// @Failure 500 {object} web.Response
// @Router /api/categories [get]
func (h *Handler) GetAll(c *gin.Context) {
	categories, err := h.svc.GetAll()
	if err != nil {
		web.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	web.JSON(c, http.StatusOK, categories, "Categories retrieved")
}

// Create adds a new category
// @Summary Create category
// @Description Create a new category. Admin role required.
// @Tags categories
// @Accept json
// @Produce json
// @Param request body CreateCategoryRequest true "Category data"
// @Security BearerAuth
// @Success 201 {object} Category
// @Failure 400 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /api/categories [post]
func (h *Handler) Create(c *gin.Context) {
	traceID := uuid.New().String()
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		web.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	cat := &Category{
		Name:        req.Name,
		Description: req.Description,
	}

	createdCat, err := h.svc.Create(traceID, cat)
	if err != nil {
		web.Error(c, http.StatusInternalServerError, "Failed to create category")
		return
	}

	web.JSON(c, http.StatusCreated, createdCat, "Category created")
}

// Update modifies an existing category
// @Summary Update category
// @Description Update an existing category by ID. Admin role required.
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param request body UpdateCategoryRequest true "Updated category data"
// @Security BearerAuth
// @Success 200 {object} Category
// @Failure 400 {object} web.Response
// @Failure 404 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /api/categories/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	traceID := uuid.New().String()
	id := c.Param("id")

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		web.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	updatedCat, err := h.svc.Update(traceID, id, req)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			web.Error(c, http.StatusNotFound, "Category not found")
			return
		}
		web.Error(c, http.StatusInternalServerError, "Failed to update category")
		return
	}

	web.JSON(c, http.StatusOK, updatedCat, "Category updated")
}

// Delete removes a category
// @Summary Delete category
// @Description Delete a category by ID. Admin role required.
// @Tags categories
// @Produce json
// @Param id path string true "Category ID"
// @Security BearerAuth
// @Success 200 {object} web.Response
// @Failure 404 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /api/categories/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	traceID := uuid.New().String()
	id := c.Param("id")

	if err := h.svc.Delete(traceID, id); err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			web.Error(c, http.StatusNotFound, "Category not found")
			return
		}
		web.Error(c, http.StatusInternalServerError, "Failed to delete category")
		return
	}

	web.JSON(c, http.StatusOK, nil, "Category deleted")
}
