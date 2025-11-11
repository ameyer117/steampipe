package connection

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/plugin"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
)

// mockPluginManager is a mock implementation of pluginManager interface for testing
type mockPluginManager struct {
	pool *pgxpool.Pool
}

func (m *mockPluginManager) Pool() *pgxpool.Pool {
	return m.pool
}

// Implement other required methods from pluginManager interface
func (m *mockPluginManager) OnConnectionConfigChanged(context.Context, ConnectionConfigMap, map[string]*plugin.Plugin) {
}

func (m *mockPluginManager) GetConnectionConfig() ConnectionConfigMap {
	return nil
}

func (m *mockPluginManager) HandlePluginLimiterChanges(PluginLimiterMap) error {
	return nil
}

func (m *mockPluginManager) ShouldFetchRateLimiterDefs() bool {
	return false
}

func (m *mockPluginManager) LoadPluginRateLimiters(map[string]string) (PluginLimiterMap, error) {
	return nil, nil
}

func (m *mockPluginManager) SendPostgresSchemaNotification(context.Context) error {
	return nil
}

func (m *mockPluginManager) SendPostgresErrorsAndWarningsNotification(context.Context, error_helpers.ErrorAndWarnings) {
}

func (m *mockPluginManager) UpdatePluginColumnsTable(context.Context, map[string]*sdkproto.Schema, []string) error {
	return nil
}

// Implement shared.PluginManager interface methods
func (m *mockPluginManager) Get(req *proto.GetRequest) (*proto.GetResponse, error) {
	return nil, nil
}

func (m *mockPluginManager) RefreshConnections(req *proto.RefreshConnectionsRequest) (*proto.RefreshConnectionsResponse, error) {
	return nil, nil
}

func (m *mockPluginManager) Shutdown(req *proto.ShutdownRequest) (*proto.ShutdownResponse, error) {
	return nil, nil
}

// TestNewRefreshConnectionState_NilPool tests that newRefreshConnectionState handles nil pool gracefully
// This test demonstrates issue #4778 - nil pool from pluginManager causes panic
func TestNewRefreshConnectionState_NilPool(t *testing.T) {
	ctx := context.Background()

	// Create a mock plugin manager that returns nil pool
	mockPM := &mockPluginManager{
		pool: nil,
	}

	// This should not panic - should return an error instead
	_, err := newRefreshConnectionState(ctx, mockPM, []string{})

	if err == nil {
		t.Error("Expected error when pool is nil, got nil")
	}
}
