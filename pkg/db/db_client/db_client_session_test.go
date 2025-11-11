package db_client

import (
	"context"
	"sync"
	"testing"
)

// TestDbClient_SessionSearchPathUpdatesThreadSafe verifies that concurrent access
// to customSearchPath does not cause data races.
// Reference: https://github.com/turbot/steampipe/issues/4792
//
// This test simulates concurrent goroutines accessing and modifying the customSearchPath
// slice. Without proper synchronization, this causes a data race.
//
// Run with: go test -race -run TestDbClient_SessionSearchPathUpdatesThreadSafe
func TestDbClient_SessionSearchPathUpdatesThreadSafe(t *testing.T) {
	// Create a DbClient with the fields we need for testing
	client := &DbClient{
		customSearchPath: []string{"public", "internal"},
		userSearchPath:   []string{"public"},
	}

	// Number of concurrent operations to test
	const numGoroutines = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3)

	// Simulate concurrent readers calling GetRequiredSessionSearchPath
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = client.GetRequiredSessionSearchPath()
		}()
	}

	// Simulate concurrent readers calling GetCustomSearchPath
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = client.GetCustomSearchPath()
		}()
	}

	// Simulate concurrent writers calling SetRequiredSessionSearchPath
	// This is the most dangerous operation as it modifies the slice
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			ctx := context.Background()
			// This will write to customSearchPath
			_ = client.SetRequiredSessionSearchPath(ctx)
		}()
	}

	wg.Wait()
}
