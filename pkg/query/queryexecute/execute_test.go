package queryexecute

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/initialisation"
	"github.com/turbot/steampipe/v2/pkg/query"
)

// TestRunBatchSession_LoadedTimeout demonstrates that RunBatchSession blocks forever
// if initData.Loaded never closes, even when the context is cancelled.
// References issue #4781
func TestRunBatchSession_LoadedTimeout(t *testing.T) {
	t.Skip("Test demonstrates bug #4781 - RunBatchSession blocks forever if initData.Loaded never closes")

	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Create InitData with a Loaded channel that will never close
	initData := &query.InitData{
		InitData: initialisation.InitData{
			Result: &db_common.InitResult{},
		},
		Loaded: make(chan struct{}), // This channel will never close
	}

	// This should return within the timeout, but currently blocks forever
	done := make(chan bool)
	var failures int
	var err error

	go func() {
		failures, err = RunBatchSession(ctx, initData)
		done <- true
	}()

	select{
	case <-done:
		// Function returned, check that it returned an error due to context cancellation
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
		assert.Equal(t, 0, failures)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("RunBatchSession blocked forever despite context cancellation - bug #4781")
	}
}
