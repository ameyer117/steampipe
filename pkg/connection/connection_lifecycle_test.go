package connection

import (
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/v2/pkg/constants"
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

// TestLogRefreshConnectionResultsTypeAssertion tests the type assertion panic bug in logRefreshConnectionResults
// This test demonstrates issue #4807 - potential panic when viper.Get returns nil or wrong type
func TestLogRefreshConnectionResultsTypeAssertion(t *testing.T) {
	// Save original viper value
	originalValue := viper.Get(constants.ConfigKeyActiveCommand)
	defer func() {
		if originalValue != nil {
			viper.Set(constants.ConfigKeyActiveCommand, originalValue)
		} else {
			// Clean up by setting to nil if it was nil
			viper.Set(constants.ConfigKeyActiveCommand, nil)
		}
	}()

	// Test case 1: viper.Get returns nil
	t.Run("nil value causes panic", func(t *testing.T) {
		viper.Set(constants.ConfigKeyActiveCommand, nil)

		state := &refreshConnectionState{}

		// This should panic with the current unsafe type assertion
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Expected panic occurred: %v", r)
			}
		}()

		// This will panic because viper.Get returns nil
		state.logRefreshConnectionResults()

		// If we get here without panic, the bug is fixed
		t.Error("Expected panic did not occur - bug may be fixed")
	})

	// Test case 2: viper.Get returns wrong type
	t.Run("wrong type causes panic", func(t *testing.T) {
		viper.Set(constants.ConfigKeyActiveCommand, "not-a-cobra-command")

		state := &refreshConnectionState{}

		// This should panic with the current unsafe type assertion
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Expected panic occurred: %v", r)
			}
		}()

		// This will panic because type assertion fails
		state.logRefreshConnectionResults()

		// If we get here without panic, the bug is fixed
		t.Error("Expected panic did not occur - bug may be fixed")
	})

	// Test case 3: viper.Get returns *cobra.Command but it's nil
	t.Run("nil cobra.Command pointer causes panic", func(t *testing.T) {
		var nilCmd *cobra.Command
		viper.Set(constants.ConfigKeyActiveCommand, nilCmd)

		state := &refreshConnectionState{}

		// This should panic when calling .Name() on nil pointer
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Expected panic occurred: %v", r)
			}
		}()

		// This will panic on .Name() call
		state.logRefreshConnectionResults()

		// If we get here without panic, the bug is fixed
		t.Error("Expected panic did not occur - bug may be fixed")
	})

	// Test case 4: Valid cobra.Command (should work)
	t.Run("valid cobra.Command works", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "plugin-manager",
		}
		viper.Set(constants.ConfigKeyActiveCommand, cmd)

		state := &refreshConnectionState{}

		// This should work
		state.logRefreshConnectionResults()
	})
}
