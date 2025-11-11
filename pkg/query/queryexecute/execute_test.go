package queryexecute

import (
	"context"
	"testing"

	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/initialisation"
	"github.com/turbot/steampipe/v2/pkg/query"
)

// TestRunBatchSession_NilClient tests that RunBatchSession handles nil Client gracefully
func TestRunBatchSession_NilClient(t *testing.T) {
	// Create initData with nil Client
	initData := &query.InitData{
		InitData: initialisation.InitData{
			Result: &db_common.InitResult{},
			Client: nil, // nil Client should be handled gracefully
		},
		Loaded: make(chan struct{}),
	}

	// Signal that init is complete
	close(initData.Loaded)

	// This should not panic - it should handle nil Client gracefully
	_, err := RunBatchSession(context.Background(), initData)

	// We expect an error indicating that Client is required, not a panic
	if err == nil {
		t.Error("Expected error when Client is nil, got nil")
	}
}
