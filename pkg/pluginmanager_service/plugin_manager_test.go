package pluginmanager_service

import (
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/turbot/steampipe/v2/pkg/connection"
	pb "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
)

// TestPluginManager_Shutdown_NoPlugins tests that Shutdown handles nil pool gracefully
// Related to bug #4782
func TestPluginManager_Shutdown_NoPlugins(t *testing.T) {
	// Create a PluginManager without initializing the pool
	// This simulates a scenario where pool initialization failed
	pm := &PluginManager{
		logger:              hclog.NewNullLogger(),
		runningPluginMap:    make(map[string]*runningPlugin),
		connectionConfigMap: make(connection.ConnectionConfigMap),
		plugins:             make(connection.PluginMap),
		// Note: pool is not initialized, will be nil
	}

	// Calling Shutdown should not panic even with nil pool
	req := &pb.ShutdownRequest{}
	resp, err := pm.Shutdown(req)

	if err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}

	if resp == nil {
		t.Error("Shutdown returned nil response")
	}
}
