package queryexecute

import (
	"context"
	"testing"
)

func TestRunBatchSession_NilInitData(t *testing.T) {
	ctx := context.Background()

	// This should not panic - function should validate initData is non-nil
	failures, err := RunBatchSession(ctx, nil)

	if err == nil {
		t.Fatal("Expected error when initData is nil, got nil")
	}

	if failures != 0 {
		t.Errorf("Expected 0 failures when initData is nil, got %d", failures)
	}
}
