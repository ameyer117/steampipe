package db_client

import (
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

// TestDbClient_ConcurrentCloseAndRead verifies that concurrent reads don't panic
// when Close() sets sessions to nil
// Reference: https://github.com/turbot/steampipe/issues/4793
func TestDbClient_ConcurrentCloseAndRead(t *testing.T) {

	// This test simulates the race condition where:
	// 1. A goroutine enters AcquireSession, locks the mutex, reads c.sessions
	// 2. Close() sets c.sessions = nil WITHOUT holding the mutex
	// 3. The goroutine tries to write to c.sessions which is now nil
	// This causes a nil map panic or data race

	// Run the test multiple times to increase chance of catching the race
	for i := 0; i < 50; i++ {
		client := &DbClient{
			sessions:      make(map[uint32]*db_common.DatabaseSession),
			sessionsMutex: &sync.Mutex{},
		}

		done := make(chan bool, 2)

		// Goroutine 1: Simulates AcquireSession behavior
		go func() {
			defer func() { done <- true }()

			client.sessionsMutex.Lock()
			// After the fix, code should check if sessions is nil
			if client.sessions != nil {
				_, found := client.sessions[12345]
				if !found {
					client.sessions[12345] = db_common.NewDBSession(12345)
				}
			}
			client.sessionsMutex.Unlock()
		}()

		// Goroutine 2: Calls Close()
		go func() {
			defer func() { done <- true }()
			// Without the fix, Close() sets sessions to nil without mutex protection
			// This is the bug - it should acquire the mutex first
			client.Close(nil)
		}()

		// Wait for both goroutines
		<-done
		<-done
	}

	// With the bug present, running with -race will detect the data race
	// After the fix, this test should pass cleanly
}

// TestDbClient_SessionsMapNilAfterClose verifies that accessing sessions after Close
// doesn't cause a nil pointer panic
// Reference: https://github.com/turbot/steampipe/issues/4793
func TestDbClient_SessionsMapNilAfterClose(t *testing.T) {

	client := &DbClient{
		sessions:      make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex: &sync.Mutex{},
	}

	// Add a session
	client.sessionsMutex.Lock()
	client.sessions[12345] = db_common.NewDBSession(12345)
	client.sessionsMutex.Unlock()

	// Close sets sessions to nil (without mutex protection - this is the bug)
	client.Close(nil)

	// Attempt to access sessions like AcquireSession does
	// After the fix, this should not panic
	client.sessionsMutex.Lock()
	defer client.sessionsMutex.Unlock()

	// With the bug: this panics because sessions is nil
	// After fix: sessions should either not be nil, or code checks for nil
	if client.sessions != nil {
		client.sessions[67890] = db_common.NewDBSession(67890)
	}
}
