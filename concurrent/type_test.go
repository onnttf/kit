package concurrent

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorAction_String(t *testing.T) {
	tests := []struct {
		action   ErrorAction
		expected string
	}{
		{ActionContinue, "Continue"},
		{ActionRetry, "Retry"},
		{ActionAbort, "Abort"},
		{ErrorAction(100), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.action.String())
		})
	}
}

func TestHandler(t *testing.T) {
	var called bool
	handler := func(ctx context.Context, item int) error {
		called = true
		return nil
	}
	err := handler(context.Background(), 1)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestHandlerWithError(t *testing.T) {
	expectedErr := errors.New("handler error")
	handler := func(ctx context.Context, item int) error {
		return expectedErr
	}
	err := handler(context.Background(), 1)
	assert.ErrorIs(t, err, expectedErr)
}

func TestWorkItem(t *testing.T) {
	item := workItem[int]{
		id:      42,
		data:    123,
		attempt: 3,
	}
	assert.Equal(t, 42, item.id)
	assert.Equal(t, 123, item.data)
	assert.Equal(t, 3, item.attempt)
}