# Inventory Service Race Condition Analysis

## Race Condition 1: Concurrent Map Access
- **Location**: `GetStock`, `Reserve`, `ReserveMultiple`, and `SafeReserve`.
- **Code**: `product := s.products[productID]`
- **What happens**: Go maps are not thread-safe for concurrent read and write operations. If one goroutine is writing to the map (e.g., adding a new product) while another is reading from it, the program will crash with a fatal error: `fatal error: concurrent map read and map write`.
- **Production scenario**: A high-traffic e-commerce site where many users are checking stock levels (`GetStock`) while an administrator is adding or updating products in the catalog.
- **Fix approach**: Wrap map access in `sync.RWMutex`. Use `RLock()` for reading and `Lock()` for writing.

## Race Condition 2: Non-Atomic Update (Read-Modify-Write)
- **Location**: `Reserve` function.
- **Code**:
  ```go
  if product.Stock < quantity {
      return ErrInsufficientStock
  }
  product.Stock -= quantity
  ```
- **What happens**: This is a classic "Time of Check to Time of Use" (TOCTOU) bug. Between the time the stock is checked and the time it is decremented, another goroutine could have already decremented the stock.
- **Production scenario**: An item has 1 unit of stock left. Two users click "Buy" at the exact same millisecond. Both goroutines see `Stock = 1`, both pass the check, and both decrement the stock. Final stock is -1, and you've oversold an item you don't have.
- **Fix approach**: The check and the update must be performed inside a single critical section protected by a `sync.Mutex`.

## Race Condition 3: Lack of Atomicity in Batch Operations
- **Location**: `ReserveMultiple` function.
- **Code**: The function iterates through all items once to check stock, then iterates again to decrement.
- **What happens**: Similar to Race Condition 2, but on a larger scale. The "check all" loop provides no guarantee that the stock will still be sufficient when the "update all" loop runs. Furthermore, if one update fails or if a partial update occurs, the system is left in an inconsistent state.
- **Production scenario**: A user tries to buy a "Bundle" of Product A and B. The system checks both, sees they are available. Before it can reserve them, another user buys the last Unit of B. The first user's reservation still proceeds for A, but B might go negative or the system might fail inconsistently.
- **Fix approach**: Use a global lock (or ordered per-product locks) to ensure the entire batch operation is atomic. Ensure consistency by either locking the whole map or using a transaction-like approach.

## Race Condition 4: Useless Local Mutex
- **Location**: `SafeReserve` function.
- **Code**: `var mu sync.Mutex` inside the function body.
- **What happens**: A mutex only works if all competing goroutines share the same instance of the mutex. By declaring it locally inside the function, every goroutine creates its own private mutex instance. Locking a private mutex does not block any other goroutine from entering the same section of code.
- **Production scenario**: Even with `SafeReserve`, the program will still experience the exact same data races and overselling issues as the unprotected `Reserve` function under load.
- **Fix approach**: Move the mutex to the `InventoryService` struct so it is shared across all method calls on that instance.
