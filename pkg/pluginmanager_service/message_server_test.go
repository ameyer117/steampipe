package pluginmanager_service

import (
	"context"
	"testing"

	"github.com/hashicorp/go-hclog"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/connection"
)

// TestPluginManager_UpdateConnectionSchema_NilPool tests that updateConnectionSchema handles nil pool gracefully
// Related to bug #4783
func TestPluginManager_UpdateConnectionSchema_NilPool(t *testing.T) {
	// Create a PluginManager without initializing the pool
	// This simulates a scenario where pool initialization failed but connections were refreshed
	pm := &PluginManager{
		logger:                    hclog.NewNullLogger(),
		runningPluginMap:          make(map[string]*runningPlugin),
		connectionConfigMap:       make(connection.ConnectionConfigMap),
		plugins:                   make(connection.PluginMap),
		pluginConnectionConfigMap: make(map[string][]*sdkproto.ConnectionConfig),
		// Note: pool is not initialized, will be nil
	}

	// Calling updateConnectionSchema should not panic even with nil pool
	// This will panic at line 799: m.pool.Acquire(ctx) because pool is nil
	ctx := context.Background()
	pm.updateConnectionSchema(ctx, "test-connection")

	// If we reach here without panic, the test passes
}
