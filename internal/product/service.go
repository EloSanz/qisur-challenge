package product

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"qisur-service/internal/websocket"
	"qisur-service/pkg/audit"
	"qisur-service/pkg/rabbitmq"
)

type Service interface {
	GetAll(ctx context.Context, page, pageSize int) ([]Product, int64, error)
	GetByID(ctx context.Context, id string) (*Product, error)
	Create(ctx context.Context, traceID string, prod *Product, categoryIDs []string) (*Product, error)
	Update(ctx context.Context, traceID string, id string, req UpdateProductRequest) (*Product, error)
	Delete(ctx context.Context, traceID string, id string) error
	GetHistory(ctx context.Context, productID string, start, end time.Time) ([]ProductHistory, error)
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

func (s *service) GetAll(ctx context.Context, page, pageSize int) ([]Product, int64, error) {
	return s.repo.FindAll(ctx, page, pageSize)
}

func (s *service) GetByID(ctx context.Context, id string) (*Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *service) Create(ctx context.Context, traceID string, prod *Product, categoryIDs []string) (*Product, error) {
	if err := s.repo.Create(ctx, prod, categoryIDs); err != nil {
		return nil, err
	}

	createdProd, _ := s.repo.FindByID(ctx, prod.ID)
	
	audit.Emit(s.rmq, traceID, "DB_SAVED", "product", prod.ID)

	if s.hub != nil {
		s.hub.BroadcastEvent(traceID, "product_created", createdProd)
	}

	return createdProd, nil
}

func (s *service) Update(ctx context.Context, traceID string, id string, req UpdateProductRequest) (*Product, error) {
	prod, err := s.repo.FindByID(ctx, id)
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

	if err := s.repo.Update(ctx, prod, req.CategoryIDs); err != nil {
		return nil, err
	}

	updatedProd, _ := s.repo.FindByID(ctx, prod.ID)
	
	audit.Emit(s.rmq, traceID, "DB_SAVED", "product", prod.ID)

	if s.hub != nil {
		s.hub.BroadcastEvent(traceID, "product_updated", updatedProd)
	}

	return updatedProd, nil
}

func (s *service) Delete(ctx context.Context, traceID string, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	audit.Emit(s.rmq, traceID, "DB_SAVED", "product", id)

	if s.hub != nil {
		s.hub.BroadcastEvent(traceID, "product_deleted", gin.H{"id": id})
	}

	return nil
}

func (s *service) GetHistory(ctx context.Context, productID string, start, end time.Time) ([]ProductHistory, error) {
	return s.repo.GetHistory(ctx, productID, start, end)
}
