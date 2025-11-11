package interactive

import (
	"testing"
)

// TestClosePromptNilCancelPanic tests that ClosePrompt doesn't panic
// when cancelPrompt is nil.
//
// This can happen if ClosePrompt is called before the prompt is fully
// initialized or after manual nil assignment.
//
// Bug: #4788
func TestClosePromptNilCancelPanic(t *testing.T) {
	// Create an InteractiveClient with nil cancelPrompt
	c := &InteractiveClient{
		cancelPrompt: nil,
	}

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ClosePrompt() panicked with nil cancelPrompt: %v", r)
		}
	}()

	// Call ClosePrompt with nil cancelPrompt
	// This will panic without the fix
	c.ClosePrompt(AfterPromptCloseExit)
}
