package pluginmanager_service

import (
	"testing"
)

// TestPluginMessageServer_LogReceiveError_NilError tests that logReceiveError
// handles nil error gracefully without panicking
func TestPluginMessageServer_LogReceiveError_NilError(t *testing.T) {
	// Create a message server
	pm := &PluginManager{}
	server := &PluginMessageServer{
		pluginManager: pm,
	}

	// This should not panic - calling logReceiveError with nil error
	server.logReceiveError(nil, "test-connection")
}
