package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAndSanitizeInput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr bool
		expected    string
	}{
		{
			name:        "valid input",
			input:       "hello world",
			expectedErr: false,
			expected:    "hello world",
		},
		{
			name:        "empty input",
			input:       "",
			expectedErr: true,
		},
		{
			name:        "input with special characters",
			input:       "test@#$%^&*()",
			expectedErr: false,
			expected:    "test@#$%^&*()",
		},
		{
			name:        "very long input",
			input:       strings.Repeat("a", MaxLogSize+1),
			expectedErr: true,
		},
		{
			name:        "input with newlines gets sanitized",
			input:       "line1\nline2\nline3",
			expectedErr: false,
			expected:    "line1\\nline2\\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateAndSanitizeInput(tt.input)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	context := map[string]interface{}{
		"operation": "test_operation",
		"user_id":   123,
	}

	wrappedErr := wrapError(originalErr, "test operation", context)
	
	assert.Error(t, wrappedErr)
	assert.Contains(t, wrappedErr.Error(), "test operation")
	assert.Contains(t, wrappedErr.Error(), "original error")
}

func TestProcessLogsInChunks(t *testing.T) {
	testData := "this is a test log that should be processed in chunks"
	reader := strings.NewReader(testData)
	
	var processedChunks []string
	processor := func(chunk []byte) error {
		processedChunks = append(processedChunks, string(chunk))
		return nil
	}

	err := processLogsInChunks(reader, 10, processor)
	assert.NoError(t, err)
	assert.True(t, len(processedChunks) > 0)

	// Test with processor that returns error
	reader2 := strings.NewReader(testData)
	errorProcessor := func(chunk []byte) error {
		return errors.New("processing failed")
	}

	err = processLogsInChunks(reader2, 10, errorProcessor)
	assert.Error(t, err)
}

func TestRetryWithBackoff(t *testing.T) {
	ctx := context.Background()

	t.Run("successful operation", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return nil
		}

		err := retryWithBackoff(ctx, operation, 3, time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("operation succeeds after retries", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			if callCount < 3 {
				return errors.New("temporary failure")
			}
			return nil
		}

		err := retryWithBackoff(ctx, operation, 5, time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 3, callCount)
	})

	t.Run("operation fails after max retries", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return errors.New("persistent failure")
		}

		err := retryWithBackoff(ctx, operation, 3, time.Millisecond)
		assert.Error(t, err)
		assert.Equal(t, 3, callCount)
	})

	t.Run("context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		operation := func() error {
			return errors.New("should not retry when context is cancelled")
		}

		err := retryWithBackoff(cancelCtx, operation, 3, time.Millisecond)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context")
	})
}

func TestAnalysisCache(t *testing.T) {
	cache := NewAnalysisCache(3, time.Hour)
	require.NotNil(t, cache)

	t.Run("cache operations", func(t *testing.T) {
		// Test cache miss
		result, found := cache.Get("missing-key")
		assert.False(t, found)
		assert.Nil(t, result)

		// Test cache set and get
		testResult := &FailureAnalysisResult{
			ID:          "test-1",
			RootCause:   "Test failure",
			Description: "Test description",
			Classification: FailureClassification{
				Type:       BuildFailure,
				Confidence: 0.8,
			},
		}

		cache.Set("test-key", testResult, time.Hour)
		
		result, found = cache.Get("test-key")
		assert.True(t, found)
		assert.NotNil(t, result)
		assert.Equal(t, BuildFailure, result.Classification.Type)
		assert.Equal(t, 0.8, result.Classification.Confidence)
	})

	t.Run("cache eviction", func(t *testing.T) {
		// Fill cache beyond capacity
		for i := 0; i < 5; i++ {
			result := &FailureAnalysisResult{
				ID:          fmt.Sprintf("test-%d", i),
				Description: fmt.Sprintf("Result %d", i),
			}
			cache.Set(fmt.Sprintf("key-%d", i), result, time.Hour)
		}

		// Check that oldest entries were evicted
		_, found := cache.Get("key-0")
		assert.False(t, found)
		
		_, found = cache.Get("key-4")
		assert.True(t, found)
	})

	t.Run("cache expiration", func(t *testing.T) {
		cache := NewAnalysisCache(10, time.Nanosecond)
		
		result := &FailureAnalysisResult{
			ID:          "expiring-1",
			Description: "Expiring result",
		}
		cache.Set("expiring-key", result, time.Nanosecond)
		
		// Wait for expiration
		time.Sleep(time.Millisecond)
		
		_, found := cache.Get("expiring-key")
		assert.False(t, found)
	})
}

func TestRateLimiter(t *testing.T) {
	t.Run("rate limiting", func(t *testing.T) {
		limiter := NewRateLimiter(2, time.Second)
		require.NotNil(t, limiter)

		// First two requests should be allowed
		assert.True(t, limiter.Allow())
		assert.True(t, limiter.Allow())

		// Third request should be blocked
		assert.False(t, limiter.Allow())
	})

	t.Run("rate limiter refill", func(t *testing.T) {
		limiter := NewRateLimiter(1, time.Millisecond*10)
		
		// Use up the token
		assert.True(t, limiter.Allow())
		assert.False(t, limiter.Allow())
		
		// Wait for refill
		time.Sleep(time.Millisecond * 15)
		
		// Should be allowed again
		assert.True(t, limiter.Allow())
	})
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("min function", func(t *testing.T) {
		assert.Equal(t, 1, min(1, 2))
		assert.Equal(t, 1, min(2, 1))
		assert.Equal(t, 5, min(5, 5))
		assert.Equal(t, -1, min(-1, 0))
	})

	t.Run("validateRunID", func(t *testing.T) {
		assert.NoError(t, validateRunID(1))
		assert.NoError(t, validateRunID(999999))
		assert.Error(t, validateRunID(0))
		assert.Error(t, validateRunID(-1))
	})

	t.Run("validateRepositoryName", func(t *testing.T) {
		assert.NoError(t, validateRepositoryName("owner", "valid-repo"))
		assert.NoError(t, validateRepositoryName("user", "repo"))
		assert.NoError(t, validateRepositoryName("org", "project-name"))
		assert.Error(t, validateRepositoryName("", "repo"))
		assert.Error(t, validateRepositoryName("owner", ""))
		assert.Error(t, validateRepositoryName("owner", "invalid repo name"))
	})

	t.Run("validateLLMProvider", func(t *testing.T) {
		assert.NoError(t, validateLLMProvider("openai"))
		assert.NoError(t, validateLLMProvider("anthropic"))
		assert.NoError(t, validateLLMProvider("gemini"))
		assert.NoError(t, validateLLMProvider("deepseek"))
		assert.NoError(t, validateLLMProvider("litellm"))
		assert.Error(t, validateLLMProvider(""))
		assert.Error(t, validateLLMProvider("invalid-provider"))
	})

	t.Run("createTimeoutContext", func(t *testing.T) {
		ctx, cancel := createTimeoutContext(context.Background(), "test-operation", time.Millisecond*100)
		defer cancel()
		
		assert.NotNil(t, ctx)
		
		// Test that context times out
		select {
		case <-ctx.Done():
			// Context should timeout
		case <-time.After(time.Millisecond * 200):
			t.Fatal("Context should have timed out")
		}
	})
}

// Mock closer for testing
type mockCloser struct {
	closed bool
}

func (m *mockCloser) Close() error {
	m.closed = true
	return nil
}

func TestResourceManager(t *testing.T) {
	t.Run("add and cleanup resources", func(t *testing.T) {
		rm := NewResourceManager()
		require.NotNil(t, rm)
		
		mock := &mockCloser{}
		rm.Add(mock)
		rm.Cleanup()
		
		assert.True(t, mock.closed)
	})

	t.Run("multiple resources cleanup", func(t *testing.T) {
		rm := NewResourceManager()
		mocks := make([]*mockCloser, 3)
		
		for i := range mocks {
			mocks[i] = &mockCloser{}
			rm.Add(mocks[i])
		}
		
		rm.Cleanup()
		
		for _, mock := range mocks {
			assert.True(t, mock.closed)
		}
	})
}

func TestCircuitBreaker(t *testing.T) {
	t.Run("successful operations", func(t *testing.T) {
		cb := NewCircuitBreaker(3, time.Minute)
		require.NotNil(t, cb)

		operation := func() error {
			return nil
		}

		for i := 0; i < 5; i++ {
			err := cb.Execute(operation)
			assert.NoError(t, err)
		}
	})

	t.Run("circuit breaker opens after failures", func(t *testing.T) {
		cb := NewCircuitBreaker(2, time.Minute)
		
		failingOperation := func() error {
			return errors.New("operation failed")
		}

		// First two failures should execute
		err := cb.Execute(failingOperation)
		assert.Error(t, err)
		
		err = cb.Execute(failingOperation)
		assert.Error(t, err)

		// Third attempt should be rejected by circuit breaker
		err = cb.Execute(failingOperation)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circuit breaker")
	})

	t.Run("circuit breaker resets after timeout", func(t *testing.T) {
		cb := NewCircuitBreaker(1, time.Millisecond*10)
		
		failingOperation := func() error {
			return errors.New("operation failed")
		}

		// Trigger circuit breaker
		cb.Execute(failingOperation)
		
		// Should be rejected
		err := cb.Execute(failingOperation)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circuit breaker")
		
		// Wait for reset
		time.Sleep(time.Millisecond * 15)
		
		// Should execute again (but still fail)
		err = cb.Execute(failingOperation)
		assert.Error(t, err)
		assert.NotContains(t, err.Error(), "circuit breaker")
	})
}

func TestGracefulShutdown(t *testing.T) {
	gs := NewGracefulShutdown()
	require.NotNil(t, gs)

	t.Run("shutdown functions execute", func(t *testing.T) {
		shutdownCalled := false
		
		gs.AddShutdownFunc(func() error {
			shutdownCalled = true
			return nil
		})
		
		gs.Shutdown()
		assert.True(t, shutdownCalled)
	})

	t.Run("multiple shutdown functions", func(t *testing.T) {
		shutdownCount := 0
		
		for i := 0; i < 3; i++ {
			gs.AddShutdownFunc(func() error {
				shutdownCount++
				return nil
			})
		}
		
		gs.Shutdown()
		assert.Equal(t, 3, shutdownCount)
	})
}