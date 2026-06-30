package product

import (
	"time"

	"github.com/gin-gonic/gin"
	"qisur-service/internal/websocket"
	"qisur-service/pkg/audit"
	"qisur-service/pkg/rabbitmq"
)

type Service interface {
	GetAll(page, pageSize int) ([]Product, int64, error)
	GetByID(id string) (*Product, error)
	Create(traceID string, prod *Product, categoryIDs []string) (*Product, error)
	Update(traceID string, id string, req UpdateProductRequest) (*Product, error)
	Delete(traceID string, id string) error
	GetHistory(productID string, start, end time.Time) ([]ProductHistory, error)
}

type service struct {
	repo Repository
	hub  *websocket.Hub
	rmq  *rabbitmq.Client
}

func NewService(repo Repository, hub *websocket.Hub, rmq *rabbitmq.Client) Service {
	return &service{
		repo: repo,
		hub:  hub,
		rmq:  rmq,
	}
}

func (s *service) GetAll(page, pageSize int) ([]Product, int64, error) {
	return s.repo.FindAll(page, pageSize)
}

func (s *service) GetByID(id string) (*Product, error) {
	return s.repo.FindByID(id)
}

func (s *service) Create(traceID string, prod *Product, categoryIDs []string) (*Product, error) {
	if err := s.repo.Create(prod, categoryIDs); err != nil {
		return nil, err
	}

	createdProd, _ := s.repo.FindByID(prod.ID)
	
	audit.Emit(s.rmq, traceID, "DB_SAVED", "product", prod.ID)

	if s.hub != nil {
		s.hub.BroadcastEvent(traceID, "product_created", createdProd)
	}

	return createdProd, nil
}

func (s *service) Update(traceID string, id string, req UpdateProductRequest) (*Product, error) {
	prod, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		prod.Name = req.Name
	}
	if req.Description != "" {
		prod.Description = req.Description
	}
	if req.Price > 0 {
		prod.Price = req.Price
	}
	if req.Stock >= 0 {
		prod.Stock = req.Stock
	}

	if err := s.repo.Update(prod, req.CategoryIDs); err != nil {
		return nil, err
	}

	updatedProd, _ := s.repo.FindByID(prod.ID)
	
	audit.Emit(s.rmq, traceID, "DB_SAVED", "product", prod.ID)

	if s.hub != nil {
		s.hub.BroadcastEvent(traceID, "product_updated", updatedProd)
	}

	return updatedProd, nil
}

func (s *service) Delete(traceID string, id string) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	audit.Emit(s.rmq, traceID, "DB_SAVED", "product", id)

	if s.hub != nil {
		s.hub.BroadcastEvent(traceID, "product_deleted", gin.H{"id": id})
	}

	return nil
}

func (s *service) GetHistory(productID string, start, end time.Time) ([]ProductHistory, error) {
	return s.repo.GetHistory(productID, start, end)
}
