# Concurrency Questions & Answers

**Q1: Why does declaring the mutex inside the function make it useless?**
A mutex provides mutual exclusion by allowing only one goroutine to hold the lock at a time. However, this only works if multiple goroutines are attempting to acquire the **same** mutex instance. By declaring `var mu sync.Mutex` inside the function, a new, independent mutex is allocated on the stack (or escaped to heap) for *every single function call*. Goroutines are essentially locking their own private doors while the main entrance remains wide open.

**Q2: If you use per-product locks, what can happen in `ReserveMultiple`? How do you prevent it?**
Using per-product locks can lead to **Deadlocks**.
- Scenario: Goroutine 1 locks Product A and waits for Product B. Goroutine 2 locks Product B and waits for Product A. Both are stuck forever (Circular Wait).
- **Prevention**: Establish a global **Lock Ordering**. Always acquire locks in a consistent, predefined order (e.g., sorted by Product ID). This breaks the circular wait condition necessary for a deadlock.

**Q3: Why is the "early unlock" fix worse than no locks?**
```go
s.mu.Lock()
product := s.products[productID]
s.mu.Unlock() // Release early
// ... check and update ...
```
This is worse because it gives a **false sense of security**. While it protects the map access itself (preventing a crash), it explicitly allows a race condition to occur between the check and the update. In a high-concurrency environment, this "check-then-act" gap is exactly where overselling happens. It's "worse" because the code looks like it's trying to be safe, making the bug harder to spot during review than a totally unprotected function.

**Q4: You run tests with `-race` flag and get no warnings. Does this mean your code is race-free?**
**No.** The Go race detector is a **dynamic analysis tool**, not a static one. It only detects races that *actually occur* during the execution of the code it is observing. If your tests don't happen to trigger a specific interleaving of goroutines that causes a race, the detector won't see it. It is not a proof of correctness; it only proves that no races were detected during that specific run.
