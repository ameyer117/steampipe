package db_client

import (
	"context"
	"testing"
)

// TestResetPools verifies that ResetPools handles nil pools gracefully without panicking.
// This test addresses bug #4698 where ResetPools panics when called on a DbClient with nil pools.
func TestResetPools(t *testing.T) {
	// Create a DbClient with nil pools (simulating a partially initialized or closed client)
	client := &DbClient{
		userPool:       nil,
		managementPool: nil,
	}

	// ResetPools should NOT panic even with nil pools
	// This is the expected correct behavior
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ResetPools panicked with nil pools: %v", r)
		}
	}()

	ctx := context.Background()
	client.ResetPools(ctx)
}
