package pluginmanager_service

import (
	"context"
	"testing"

	"github.com/turbot/pipe-fittings/v2/plugin"
)

// TestPluginManager_OnConnectionConfigChanged_EmptyToNonEmpty tests the scenario where
// a PluginManager with no pool (e.g., in a testing environment) receives a configuration change.
// This test demonstrates bug #4784 - a nil pointer panic when m.pool is nil.
func TestPluginManager_OnConnectionConfigChanged_EmptyToNonEmpty(t *testing.T) {
	// Create a minimal PluginManager without pool initialization
	// This simulates a testing scenario or edge case where the pool might not be initialized
	m := &PluginManager{
		plugins: make(map[string]*plugin.Plugin),
		// Note: pool is intentionally nil to demonstrate the bug
	}

	// Create a new plugin map with one plugin
	newPlugins := map[string]*plugin.Plugin{
		"aws": {
			Plugin:   "hub.steampipe.io/plugins/turbot/aws@latest",
			Instance: "aws",
		},
	}

	ctx := context.Background()

	// This should panic with nil pointer dereference when trying to use m.pool
	err := m.handlePluginInstanceChanges(ctx, newPlugins)

	// If we get here without panic, the fix is working
	if err != nil {
		t.Logf("Expected error when pool is nil: %v", err)
	}
}
