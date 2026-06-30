package category

import (
	"context"
	"github.com/gin-gonic/gin"
	"qisur-service/internal/websocket"
	"qisur-service/pkg/audit"
	"qisur-service/pkg/rabbitmq"
)

type Service interface {
	GetAll(ctx context.Context) ([]Category, error)
	GetByID(ctx context.Context, id string) (*Category, error)
	Create(ctx context.Context, traceID string, cat *Category) (*Category, error)
	Update(ctx context.Context, traceID string, id string, req UpdateCategoryRequest) (*Category, error)
	Delete(ctx context.Context, traceID string, id string) error
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

func (s *service) GetAll(ctx context.Context) ([]Category, error) {
	return s.repo.FindAll(ctx)
}

func (s *service) GetByID(ctx context.Context, id string) (*Category, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *service) Create(ctx context.Context, traceID string, cat *Category) (*Category, error) {
	if err := s.repo.Create(ctx, cat); err != nil {
		return nil, err
	}

	audit.Emit(s.rmq, traceID, "DB_SAVED", "category", cat.ID)

	if s.hub != nil {
		s.hub.BroadcastEvent(traceID, "category_created", cat)
	}
	return cat, nil
}

func (s *service) Update(ctx context.Context, traceID string, id string, req UpdateCategoryRequest) (*Category, error) {
	cat, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		cat.Name = req.Name
	}
	if req.Description != "" {
		cat.Description = req.Description
	}

	if err := s.repo.Update(ctx, cat); err != nil {
		return nil, err
	}

	audit.Emit(s.rmq, traceID, "DB_SAVED", "category", cat.ID)

	if s.hub != nil {
		s.hub.BroadcastEvent(traceID, "category_updated", cat)
	}
	return cat, nil
}

func (s *service) Delete(ctx context.Context, traceID string, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	audit.Emit(s.rmq, traceID, "DB_SAVED", "category", id)

	if s.hub != nil {
		s.hub.BroadcastEvent(traceID, "category_deleted", gin.H{"id": id})
	}
	return nil
}
