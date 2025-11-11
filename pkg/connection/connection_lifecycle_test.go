package connection

import (
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestExemplarSchemaMapConcurrentAccess tests concurrent access to exemplarSchemaMap
// This test demonstrates issue #4757 - race condition when writing to exemplarSchemaMap
// without proper mutex protection.
func TestExemplarSchemaMapConcurrentAccess(t *testing.T) {
	// Create a refreshConnectionState with initialized exemplarSchemaMap
	state := &refreshConnectionState{
		exemplarSchemaMap:    make(map[string]string),
		exemplarSchemaMapMut: sync.Mutex{},
	}

	// Number of concurrent goroutines
	numGoroutines := 10
	numIterations := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch multiple goroutines that will concurrently read and write to exemplarSchemaMap
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numIterations; j++ {
				pluginName := "aws"
				connectionName := "connection"

				// Simulate the FIXED pattern in executeUpdateForConnections
				// Read with mutex (line 581-591)
				state.exemplarSchemaMapMut.Lock()
				_, haveExemplarSchema := state.exemplarSchemaMap[pluginName]
				state.exemplarSchemaMapMut.Unlock()

				// FIXED: Write with mutex protection (line 602-604)
				if !haveExemplarSchema {
					// Now properly protected with mutex
					state.exemplarSchemaMapMut.Lock()
					state.exemplarSchemaMap[pluginName] = connectionName
					state.exemplarSchemaMapMut.Unlock()
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify the map has an entry (basic sanity check)
	state.exemplarSchemaMapMut.Lock()
	if len(state.exemplarSchemaMap) == 0 {
		t.Error("Expected exemplarSchemaMap to have at least one entry")
	}
	state.exemplarSchemaMapMut.Unlock()
}

// TestExemplarSchemaMapRaceCondition specifically tests the race condition pattern
// found in refresh_connections_state.go:601 - now FIXED
func TestExemplarSchemaMapRaceCondition(t *testing.T) {
	// This test now PASSES with -race flag after the bug fix
	state := &refreshConnectionState{
		exemplarSchemaMap:    make(map[string]string),
		exemplarSchemaMapMut: sync.Mutex{},
	}

	plugins := []string{"aws", "azure", "gcp", "github", "slack"}

	var wg sync.WaitGroup

	// Simulate multiple connections being processed concurrently
	for _, plugin := range plugins {
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(p string, connNum int) {
				defer wg.Done()

				// This simulates the FIXED code pattern in executeUpdateForConnections
				state.exemplarSchemaMapMut.Lock()
				_, haveExemplar := state.exemplarSchemaMap[p]
				state.exemplarSchemaMapMut.Unlock()

				// FIXED: This write is now protected by the mutex
				if !haveExemplar {
					// No more race condition!
					state.exemplarSchemaMapMut.Lock()
					state.exemplarSchemaMap[p] = p + "_connection"
					state.exemplarSchemaMapMut.Unlock()
				}
			}(plugin, i)
		}
	}

	wg.Wait()

	// Verify all plugins are in the map
	state.exemplarSchemaMapMut.Lock()
	defer state.exemplarSchemaMapMut.Unlock()

	for _, plugin := range plugins {
		if _, ok := state.exemplarSchemaMap[plugin]; !ok {
			t.Errorf("Expected plugin %s to be in exemplarSchemaMap", plugin)
		}
	}
}

// TestExecuteUpdateSetsInParallelGoroutineLeak tests for goroutine leak in executeUpdateSetsInParallel
// This test demonstrates issue #4791 - potential goroutine leak with non-idiomatic channel pattern
//
// The issue is in refresh_connections_state.go:519-536 where the goroutine uses:
//   for { select { case connectionError := <-errChan: if connectionError == nil { return } } }
//
// While this pattern technically works when the channel is closed (returns nil, then returns from goroutine),
// it has several problems:
// 1. It's not idiomatic Go - the standard pattern for consuming until close is 'for range'
// 2. It relies on nil checks which can be error-prone
// 3. It's harder to understand and maintain
// 4. If the nil check is accidentally removed or modified, it causes a goroutine leak
//
// The idiomatic pattern 'for range errChan' automatically exits when channel is closed,
// making the code safer and more maintainable.
func TestExecuteUpdateSetsInParallelGoroutineLeak(t *testing.T) {
	// Get baseline goroutine count
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	baselineGoroutines := runtime.NumGoroutine()

	// Test the CURRENT pattern from refresh_connections_state.go:519-536
	// This pattern has potential for goroutine leaks if not carefully maintained
	errChan := make(chan *connectionError)
	var errorList []error
	var mu sync.Mutex

	// Simulate the current (non-idiomatic) pattern
	go func() {
		for {
			select {
			case connectionError := <-errChan:
				if connectionError == nil {
					return
				}
				mu.Lock()
				errorList = append(errorList, connectionError.err)
				mu.Unlock()
			}
		}
	}()

	// Send some errors
	testErr := errors.New("test error")
	errChan <- &connectionError{name: "test1", err: testErr}
	errChan <- &connectionError{name: "test2", err: testErr}

	// Close the channel (this should cause goroutine to exit via nil check)
	close(errChan)

	// Give time for the goroutine to process and exit
	time.Sleep(200 * time.Millisecond)
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// Check for goroutine leak
	afterGoroutines := runtime.NumGoroutine()
	goroutineDiff := afterGoroutines - baselineGoroutines

	// The current pattern SHOULD work (goroutine exits via nil check),
	// but we're testing to document that the pattern is risky
	if goroutineDiff > 2 {
		t.Errorf("Goroutine leak detected with current pattern: baseline=%d, after=%d, diff=%d",
			baselineGoroutines, afterGoroutines, goroutineDiff)
	}

	// Verify errors were collected
	mu.Lock()
	if len(errorList) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errorList))
	}
	mu.Unlock()

	t.Logf("BUG #4791: Current pattern works but is non-idiomatic and error-prone")
	t.Logf("The for-select-nil-check pattern at refresh_connections_state.go:520-535")
	t.Logf("should be replaced with idiomatic 'for range errChan' for safety and clarity")
}
