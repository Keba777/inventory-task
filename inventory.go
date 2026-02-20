package inventory

import (
	"errors"
	"sync"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type Product struct {
	ID    string
	Name  string
	Stock int
}

type ReserveItem struct {
	ProductID string
	Quantity  int
}

type SafeInventoryService struct {
	mu       sync.RWMutex
	products map[string]*Product
}

func NewSafeInventoryService(products map[string]*Product) *SafeInventoryService {
	if products == nil {
		products = make(map[string]*Product)
	}
	return &SafeInventoryService{
		products: products,
	}
}

func (s *SafeInventoryService) GetStock(productID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product := s.products[productID]
	if product == nil {
		return 0
	}
	return product.Stock
}

func (s *SafeInventoryService) Reserve(productID string, quantity int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	product, ok := s.products[productID]
	if !ok {
		return ErrProductNotFound
	}

	if product.Stock < quantity {
		return ErrInsufficientStock
	}

	product.Stock -= quantity
	return nil
}

// ReserveMultiple ensures all-or-nothing semantics for multiple items.
func (s *SafeInventoryService) ReserveMultiple(items []ReserveItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Validate all first (Atomic check)
	for _, item := range items {
		product, ok := s.products[item.ProductID]
		if !ok {
			return ErrProductNotFound
		}
		if product.Stock < item.Quantity {
			return ErrInsufficientStock
		}
	}

	// 2. Then reserve all (Atomic update)
	for _, item := range items {
		s.products[item.ProductID].Stock -= item.Quantity
	}

	return nil
}
