package queryexecute

import (
	"context"
	"testing"
	"time"

	"github.com/turbot/steampipe/v2/pkg/initialisation"
	"github.com/turbot/steampipe/v2/pkg/query"
)

// TestRunBatchSession_ContextCancellation tests that RunBatchSession respects
// context cancellation before blocking on initData.Loaded
func TestRunBatchSession_ContextCancellation(t *testing.T) {
	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Create InitData manually to control the Loaded channel
	// We intentionally don't close the Loaded channel to simulate
	// a scenario where initialization is slow/blocked
	initData := &query.InitData{
		InitData: *initialisation.NewInitData(),
		Loaded:   make(chan struct{}), // Never closed - simulates slow init
	}
	// Don't call NewInitData or start the init goroutine

	// RunBatchSession should check context before blocking on Loaded
	done := make(chan struct{})
	go func() {
		defer close(done)
		RunBatchSession(ctx, initData)
	}()

	// Wait for RunBatchSession to complete
	select {
	case <-done:
		// Good - function returned
		t.Log("RunBatchSession returned (expected behavior)")
	case <-time.After(2 * time.Second):
		// BUG: Function is hanging
		t.Fatal("BUG: RunBatchSession hung for 2+ seconds despite cancelled context")
	}
}
