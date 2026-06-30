package product

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// MockService is a manual mock for the Service interface
type MockService struct {
	GetAllFunc     func(ctx context.Context, page, pageSize int) ([]Product, int64, error)
	GetByIDFunc    func(ctx context.Context, id string) (*Product, error)
	CreateFunc     func(ctx context.Context, traceID string, prod *Product, categoryIDs []string) (*Product, error)
	UpdateFunc     func(ctx context.Context, traceID string, id string, req UpdateProductRequest) (*Product, error)
	DeleteFunc     func(ctx context.Context, traceID string, id string) error
	GetHistoryFunc func(ctx context.Context, productID string, start, end time.Time) ([]ProductHistory, error)
}

func (m *MockService) GetAll(ctx context.Context, page, pageSize int) ([]Product, int64, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockService) GetByID(ctx context.Context, id string) (*Product, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, ErrProductNotFound
}

func (m *MockService) Create(ctx context.Context, traceID string, prod *Product, categoryIDs []string) (*Product, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, traceID, prod, categoryIDs)
	}
	prod.ID = "mock-id-123"
	return prod, nil
}

func (m *MockService) Update(ctx context.Context, traceID string, id string, req UpdateProductRequest) (*Product, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, traceID, id, req)
	}
	return &Product{ID: id}, nil
}

func (m *MockService) Delete(ctx context.Context, traceID string, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, traceID, id)
	}
	return nil
}

func (m *MockService) GetHistory(ctx context.Context, productID string, start, end time.Time) ([]ProductHistory, error) {
	if m.GetHistoryFunc != nil {
		return m.GetHistoryFunc(ctx, productID, start, end)
	}
	return nil, nil
}

func setupTestRouter(svc Service) (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := NewHandler(svc)
	return r, h
}

func TestHandler_GetByID(t *testing.T) {
	mockSvc := &MockService{
		GetByIDFunc: func(ctx context.Context, id string) (*Product, error) {
			if id == "123" {
				return &Product{ID: "123", Name: "Mocked Product"}, nil
			}
			return nil, ErrProductNotFound
		},
	}

	r, h := setupTestRouter(mockSvc)
	r.GET("/products/:id", h.GetByID)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/products/123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/products/404", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestHandler_Create(t *testing.T) {
	mockSvc := &MockService{}

	r, h := setupTestRouter(mockSvc)
	r.POST("/products", h.Create)

	t.Run("Valid Payload", func(t *testing.T) {
		payload := CreateProductRequest{
			Name:  "Test Item",
			Price: 100,
			Stock: 50,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})

	t.Run("Invalid Payload", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer([]byte("bad json")))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}
