package db_client

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
)

// TestSessionMapCleanupImplemented verifies that the session map memory leak is fixed
// Reference: https://github.com/turbot/steampipe/issues/3737
//
// This test verifies that a BeforeClose callback is registered to clean up
// session map entries when connections are dropped by pgx.
//
// Without this fix, sessions accumulate indefinitely causing a memory leak.
func TestSessionMapCleanupImplemented(t *testing.T) {
	// Read the db_client_connect.go file to verify BeforeClose callback exists
	content, err := os.ReadFile("db_client_connect.go")
	require.NoError(t, err, "should be able to read db_client_connect.go")

	sourceCode := string(content)

	// Verify BeforeClose callback is registered
	assert.Contains(t, sourceCode, "config.BeforeClose",
		"BeforeClose callback must be registered to clean up sessions when connections close")

	// Verify the callback deletes from sessions map
	assert.Contains(t, sourceCode, "delete(c.sessions, backendPid)",
		"BeforeClose callback must delete session entries to prevent memory leak")

	// Verify the comment in db_client.go documents automatic cleanup
	clientContent, err := os.ReadFile("db_client.go")
	require.NoError(t, err, "should be able to read db_client.go")

	clientCode := string(clientContent)

	// The comment should document automatic cleanup, not a TODO
	assert.NotContains(t, clientCode, "TODO: there's no code which cleans up this map",
		"TODO comment should be removed after implementing the fix")

	// Should document the automatic cleanup mechanism
	hasCleanupComment := strings.Contains(clientCode, "automatically cleaned up") ||
		strings.Contains(clientCode, "automatic cleanup") ||
		strings.Contains(clientCode, "BeforeClose")
	assert.True(t, hasCleanupComment,
		"Comment should document automatic cleanup mechanism")
}

// TestDbClient_ConcurrentClose tests concurrent Close() calls
// BUG FOUND: Race condition in Close() - c.sessions = nil at line 171 is not protected by mutex
func TestDbClient_ConcurrentClose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	ctx := context.Background()

	client := &DbClient{
		sessions:      make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex: &sync.Mutex{},
	}

	var wg sync.WaitGroup
	numGoroutines := 10

	// Call Close() from multiple goroutines simultaneously
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.Close(ctx)
		}()
	}

	wg.Wait()

	// Should not panic and sessions should be nil
	assert.Nil(t, client.sessions)
}
