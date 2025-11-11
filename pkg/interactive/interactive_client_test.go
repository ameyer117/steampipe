package interactive

import (
	"context"
	"testing"

	"github.com/c-bata/go-prompt"
)

// TestGetTableAndConnectionSuggestions_ReturnsEmptySliceNotNil tests that
// getTableAndConnectionSuggestions returns an empty slice instead of nil
// when no matching connection is found in the schema.
//
// This is important for proper API contract - functions that return slices
// should return empty slices rather than nil to avoid unexpected nil pointer
// issues in calling code.
//
// Bug: #4710
// PR: #4734
func TestGetTableAndConnectionSuggestions_ReturnsEmptySliceNotNil(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		expected bool // true if we expect non-nil result
	}{
		{
			name:     "empty word should return non-nil",
			word:     "",
			expected: true,
		},
		{
			name:     "unqualified table should return non-nil",
			word:     "table",
			expected: true,
		},
		{
			name:     "non-existent connection should return non-nil",
			word:     "nonexistent.table",
			expected: true,
		},
		{
			name:     "qualified table with dot should return non-nil",
			word:     "aws.instances",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal InteractiveClient with empty suggestions
			c := &InteractiveClient{
				suggestions: &autoCompleteSuggestions{
					schemas:          []prompt.Suggest{},
					unqualifiedTables: []prompt.Suggest{},
					tablesBySchema:    make(map[string][]prompt.Suggest),
				},
			}

			result := c.getTableAndConnectionSuggestions(tt.word)

			if tt.expected && result == nil {
				t.Errorf("getTableAndConnectionSuggestions(%q) returned nil, expected non-nil empty slice", tt.word)
			}

			// Additional check: even if not nil, should be empty in these test cases
			if result != nil && len(result) != 0 {
				t.Errorf("getTableAndConnectionSuggestions(%q) returned non-empty slice %v, expected empty slice", tt.word, result)
			}
		})
	}
}

// TestExecuteMetaquery_NotInitialised tests that executeMetaquery returns
// an error instead of panicking when the client is not initialized.
//
// Bug: #4789
func TestExecuteMetaquery_NotInitialised(t *testing.T) {
	// Create an InteractiveClient that is not initialized
	c := &InteractiveClient{
		initialisationComplete: false,
	}

	ctx := context.Background()

	// Attempt to execute a metaquery before initialization
	// This should return an error, not panic
	err := c.executeMetaquery(ctx, ".inspect")

	// We expect an error
	if err == nil {
		t.Error("Expected error when executing metaquery before initialization, but got nil")
	}

	// The test passes if we get here without a panic
	t.Logf("Successfully received error instead of panic: %v", err)
}
