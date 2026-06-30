package product

import (
	"errors"
	"testing"
	"time"
)

// MockRepository is a simple manual mock for the Repository interface
type MockRepository struct {
	FindAllFunc    func(page, pageSize int) ([]Product, int64, error)
	FindByIDFunc   func(id string) (*Product, error)
	CreateFunc     func(prod *Product, categoryIDs []string) error
	UpdateFunc     func(prod *Product, categoryIDs []string) error
	DeleteFunc     func(id string) error
	GetHistoryFunc func(productID string, start, end time.Time) ([]ProductHistory, error)
}

func (m *MockRepository) FindAll(page, pageSize int) ([]Product, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockRepository) FindByID(id string) (*Product, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(id)
	}
	return nil, ErrProductNotFound
}

func (m *MockRepository) Create(prod *Product, categoryIDs []string) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(prod, categoryIDs)
	}
	// Simulate ID generation
	prod.ID = "mock-id-123"
	return nil
}

func (m *MockRepository) Update(prod *Product, categoryIDs []string) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(prod, categoryIDs)
	}
	return nil
}

func (m *MockRepository) Delete(id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

func (m *MockRepository) GetHistory(productID string, start, end time.Time) ([]ProductHistory, error) {
	if m.GetHistoryFunc != nil {
		return m.GetHistoryFunc(productID, start, end)
	}
	return nil, nil
}

func TestService_GetByID(t *testing.T) {
	mockRepo := &MockRepository{
		FindByIDFunc: func(id string) (*Product, error) {
			if id == "found" {
				return &Product{ID: "found", Name: "Test Product"}, nil
			}
			return nil, ErrProductNotFound
		},
	}

	svc := NewService(mockRepo, nil, nil) // no hub needed for simple get

	t.Run("Found", func(t *testing.T) {
		prod, err := svc.GetByID("found")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if prod == nil || prod.Name != "Test Product" {
			t.Errorf("unexpected product: %v", prod)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		prod, err := svc.GetByID("not-found")
		if !errors.Is(err, ErrProductNotFound) {
			t.Errorf("expected ErrProductNotFound, got %v", err)
		}
		if prod != nil {
			t.Errorf("expected nil product, got %v", prod)
		}
	})
}

func TestService_Create(t *testing.T) {
	mockRepo := &MockRepository{
		CreateFunc: func(prod *Product, categoryIDs []string) error {
			if prod.Name == "Error Product" {
				return errors.New("db error")
			}
			prod.ID = "new-id"
			return nil
		},
		FindByIDFunc: func(id string) (*Product, error) {
			return &Product{ID: "new-id", Name: "New Product"}, nil
		},
	}

	svc := NewService(mockRepo, nil, nil)

	t.Run("Success", func(t *testing.T) {
		prod := &Product{Name: "New Product"}
		created, err := svc.Create("test-trace", prod, nil)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if created == nil || created.ID != "new-id" {
			t.Errorf("expected product with id new-id, got %v", created)
		}
	})

	t.Run("Error", func(t *testing.T) {
		prod := &Product{Name: "Error Product"}
		created, err := svc.Create("test-trace", prod, nil)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if created != nil {
			t.Errorf("expected nil product, got %v", created)
		}
	})
}
