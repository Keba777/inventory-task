package inventory

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReserve_ConcurrentOversell(t *testing.T) {
	// Setup: Product has 100 stock
	pID := "prod-1"
	service := NewSafeInventoryService(map[string]*Product{
		pID: {ID: pID, Name: "Test Product", Stock: 100},
	})

	const numGoroutines = 200
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	results := make(chan error, numGoroutines)

	// Test: 200 goroutines each try to reserve 1
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			results <- service.Reserve(pID, 1)
		}()
	}

	wg.Wait()
	close(results)

	successCount := 0
	failCount := 0
	for err := range results {
		if err == nil {
			successCount++
		} else if err == ErrInsufficientStock {
			failCount++
		}
	}

	// Verify: Exactly 100 succeed, 100 fail
	assert.Equal(t, 100, successCount, "Should have exactly 100 successes")
	assert.Equal(t, 100, failCount, "Should have exactly 100 failures")
	assert.Equal(t, 0, service.GetStock(pID), "Final stock should be 0")
}

func TestReserveMultiple_Atomicity(t *testing.T) {
	// Setup: Product A: 10 stock, Product B: 5 stock
	service := NewSafeInventoryService(map[string]*Product{
		"A": {ID: "A", Name: "Product A", Stock: 10},
		"B": {ID: "B", Name: "Product B", Stock: 5},
	})

	// Test: Try to reserve 8 of A and 8 of B (should fail because only 5 of B available)
	items := []ReserveItem{
		{ProductID: "A", Quantity: 8},
		{ProductID: "B", Quantity: 8},
	}

	err := service.ReserveMultiple(items)

	// Verify: Should fail entirely
	assert.ErrorIs(t, err, ErrInsufficientStock)

	// Verify: Both A and B should remain unchanged (All-or-nothing)
	assert.Equal(t, 10, service.GetStock("A"), "Product A stock should not change")
	assert.Equal(t, 5, service.GetStock("B"), "Product B stock should not change")
}

func TestGetStock_ConcurrentWithWrites(t *testing.T) {
	pID := "prod-1"
	service := NewSafeInventoryService(map[string]*Product{
		pID: {ID: pID, Name: "Test Product", Stock: 1000},
	})

	var wg sync.WaitGroup
	const iterations = 500
	wg.Add(2)

	// Writer
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = service.Reserve(pID, 1)
		}
	}()

	// Reader
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = service.GetStock(pID)
		}
	}()

	wg.Wait()
	assert.Equal(t, 500, service.GetStock(pID))
}
