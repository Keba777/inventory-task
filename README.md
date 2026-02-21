# Thread-Safe Inventory Service

A robust inventory management implementation in Go, focusing on high concurrency and thread safety.

## Overview

This project provides a `SafeInventoryService` designed to handle concurrent stock reservations from multiple goroutines without race conditions or data corruption.

## Features

*   **RWMutex Implementation**: Optimized for high read-to-write ratios using `sync.RWMutex`.
*   **Atomic Operations**: Combines "Check" and "Act" phases within single critical sections to prevent double-reservation (overselling).
*   **All-or-Nothing Batching**: `ReserveMultiple` ensures that either all items in a batch are reserved or none are, maintaining inventory consistency.
*   **Race-Free**: Verified with Go's `-race` detector under high contention.

## Core Logic

### Atomic Reservation
Each reservation is wrapped in a write lock. The service first validates that all requested items are in stock and found in the catalog before performing any decrements.

### Deadlock Prevention
For multi-item operations, we recommend sorting item IDs before locking (though this implementation uses a global lock for simplicity and absolute correctness in this context).

## Testing

The test suite includes high-load concurrent scenarios:

*   **Oversell Prevention**: 200 goroutines competing for limited stock; exactly the correct number succeed.
*   **Atomicity Verification**: Ensuring batch reservations don't leave partial state when one item is out of stock.
*   **Contention Stress Test**: Simultaneous readers and writers to verify RWMutex behavior.

```bash
# Run tests with race detector
go test -v -race .
```

## Analysis Documentation

*   `REVIEW.md`: Detailed analysis of the 4 primary race conditions found in the original buggy implementation.
*   `ANSWERS.md`: Theoretical explanations for common concurrency pitfalls in Go.
