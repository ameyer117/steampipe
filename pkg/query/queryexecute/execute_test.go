package queryexecute

import (
	"context"
	"testing"

	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/steampipe/v2/pkg/initialisation"
	"github.com/turbot/steampipe/v2/pkg/query"
)

// TestExecuteQueries_NilClient tests that executeQueries handles nil Client gracefully
// Related to issue #4797
func TestExecuteQueries_NilClient(t *testing.T) {
	ctx := context.Background()

	// Create initData with nil Client but with queries
	// This simulates a scenario where initialization failed but queries were still provided
	initData := &query.InitData{
		InitData: *initialisation.NewInitData(),
		Queries: []*modconfig.ResolvedQuery{
			{
				Name:        "test_query",
				ExecuteSQL:  "SELECT 1",
				RawSQL:      "SELECT 1",
			},
		},
	}
	// Explicitly set Client to nil to test the nil case
	initData.Client = nil

	// This should not panic - it should handle nil Client gracefully
	// Currently this will panic with nil pointer dereference
	failures := executeQueries(ctx, initData)

	// We expect 1 failure (the query should fail gracefully, not panic)
	if failures != 1 {
		t.Errorf("Expected 1 failure with nil client, got %d", failures)
	}
}
