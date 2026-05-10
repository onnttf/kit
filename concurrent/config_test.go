package concurrent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config[int]
		expectErr bool
	}{
		{"valid config", Config[int]{Concurrency: 1}, false},
		{"zero concurrency", Config[int]{Concurrency: 0}, true},
		{"negative concurrency", Config[int]{Concurrency: -1}, true},
		{"negative max retry", Config[int]{Concurrency: 1, MaxRetry: -1}, true},
		{"negative timeout", Config[int]{Concurrency: 1, Timeout: -1}, true},
		{"valid with timeout", Config[int]{Concurrency: 1, Timeout: time.Second}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_SetDefaults(t *testing.T) {
	config := Config[int]{Concurrency: 1}
	config.SetDefaults()
	assert.Equal(t, "executor", config.Name)
	assert.NotNil(t, config.ErrorPolicy)
	assert.NotNil(t, config.PanicPolicy)
	assert.Equal(t, 100, config.MaxErrorSamples)
}

func TestConfig_Callbacks(t *testing.T) {
	var beginCalled, endCalled bool
	config := Config[int]{
		Concurrency: 1,
		OnBegin: func(context.Context, int) {
			beginCalled = true
		},
		OnEnd: func(context.Context, *Result) {
			endCalled = true
		},
	}
	exec, err := New(config)
	assert.NoError(t, err)
	result, err := exec.Run(context.Background(), []int{1, 2, 3}, func(context.Context, int) error {
		return nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, beginCalled)
	assert.True(t, endCalled)
}
