package category

import (
	"github.com/gin-gonic/gin"
	"qisur-service/internal/websocket"
	"qisur-service/pkg/audit"
	"qisur-service/pkg/rabbitmq"
)

type Service interface {
	GetAll() ([]Category, error)
	Create(traceID string, cat *Category) (*Category, error)
	Update(traceID string, id string, req UpdateCategoryRequest) (*Category, error)
	Delete(traceID string, id string) error
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

func (s *service) GetAll() ([]Category, error) {
	return s.repo.FindAll()
}

func (s *service) Create(traceID string, cat *Category) (*Category, error) {
	if err := s.repo.Create(cat); err != nil {
		return nil, err
	}

	audit.Emit(s.rmq, traceID, "DB_SAVED", "category", cat.ID)

	if s.hub != nil {
		s.hub.BroadcastEvent(traceID, "category_created", cat)
	}
	return cat, nil
}

func (s *service) Update(traceID string, id string, req UpdateCategoryRequest) (*Category, error) {
	cat, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		cat.Name = req.Name
	}
	if req.Description != "" {
		cat.Description = req.Description
	}

	if err := s.repo.Update(cat); err != nil {
		return nil, err
	}

	audit.Emit(s.rmq, traceID, "DB_SAVED", "category", cat.ID)

	if s.hub != nil {
		s.hub.BroadcastEvent(traceID, "category_updated", cat)
	}
	return cat, nil
}

func (s *service) Delete(traceID string, id string) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	audit.Emit(s.rmq, traceID, "DB_SAVED", "category", id)

	if s.hub != nil {
		s.hub.BroadcastEvent(traceID, "category_deleted", gin.H{"id": id})
	}
	return nil
}
