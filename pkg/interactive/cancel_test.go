package interactive

import (
	"context"
	"sync"
	"testing"
)

// TestConcurrentCancellation tests that cancelActiveQuery can be accessed
// concurrently without triggering data races.
// This test reproduces the race condition reported in issue #4802.
func TestConcurrentCancellation(t *testing.T) {
	// Create a minimal InteractiveClient
	client := &InteractiveClient{}

	// Simulate concurrent access to cancelActiveQuery from multiple goroutines
	// This mirrors real-world usage where:
	// - createQueryContext() sets cancelActiveQuery
	// - cancelActiveQueryIfAny() reads and clears it
	// - signal handlers may also call cancelActiveQueryIfAny()
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Simulate creating a query context (writes cancelActiveQuery)
			ctx := client.createQueryContext(context.Background())
			_ = ctx
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			// Simulate cancelling the active query (reads and writes cancelActiveQuery)
			client.cancelActiveQueryIfAny()
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// If we get here without panicking or race detector errors, the test passes
	// Note: This test will fail when run with -race flag if cancelActiveQuery access is not synchronized
}

// TestMultipleConcurrentCancellations tests rapid concurrent cancellations
// to stress test the synchronization.
func TestMultipleConcurrentCancellations(t *testing.T) {
	client := &InteractiveClient{}

	var wg sync.WaitGroup
	numIterations := 100

	// Create a query context first
	_ = client.createQueryContext(context.Background())

	// Now try to cancel it from multiple goroutines simultaneously
	for i := 0; i < numIterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.cancelActiveQueryIfAny()
		}()
	}

	wg.Wait()

	// Verify the client is in a consistent state
	if client.cancelActiveQuery != nil {
		t.Error("Expected cancelActiveQuery to be nil after all cancellations")
	}
}
