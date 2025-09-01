package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Constants for improved code quality
const (
	DefaultTimeout         = 30 * time.Second
	MaxLogSize            = 10 * 1024 * 1024 // 10MB
	MaxRetries           = 3
	RetryBackoffDuration = 2 * time.Second
	MaxConcurrentOps     = 10
)

//nolint:unused // Context key types for future infrastructure use
type contextKey string

//nolint:unused // Context keys for future infrastructure use
const (
	operationContextKey contextKey = "operation"
	timeoutContextKey   contextKey = "timeout"
)

//nolint:unused // Infrastructure code for future use
type EnhancedFailureAnalysisEngine struct {
	llmClient      *LLMClient      //nolint:unused
	logger         *logrus.Logger  //nolint:unused
	patterns       *ErrorPatternDatabase //nolint:unused
	prompts        *PromptTemplates //nolint:unused
	cache          *AnalysisCache  //nolint:unused
	rateLimiter    *RateLimiter    //nolint:unused
	metrics        *EngineMetrics  //nolint:unused
}

// AnalysisCache provides caching for repeated analysis
type AnalysisCache struct {
	mu      sync.RWMutex
	cache   map[string]*CacheEntry
	maxSize int
}

type CacheEntry struct {
	result    *FailureAnalysisResult
	timestamp time.Time
	ttl       time.Duration
}

// RateLimiter provides API rate limiting
type RateLimiter struct {
	mu       sync.Mutex
	tokens   int
	maxTokens int
	refillRate time.Duration
	lastRefill time.Time
}

//nolint:unused // Infrastructure code for future use
type EngineMetrics struct {
	mu                    sync.RWMutex //nolint:unused
	totalAnalyses        int64        //nolint:unused
	successfulAnalyses   int64        //nolint:unused
	failedAnalyses       int64        //nolint:unused
	averageResponseTime  time.Duration //nolint:unused
	cacheHits            int64        //nolint:unused
	cacheMisses          int64        //nolint:unused
}

//nolint:unused // Infrastructure function for future use
func validateAndSanitizeInput(input string) (string, error) {
	if len(input) == 0 {
		return "", fmt.Errorf("input cannot be empty")
	}
	
	if len(input) > MaxLogSize {
		return "", fmt.Errorf("input too large: %d bytes exceeds maximum %d bytes", len(input), MaxLogSize)
	}
	
	// Sanitize input to prevent log injection
	sanitized := strings.ReplaceAll(input, "\r", "\\r")
	sanitized = strings.ReplaceAll(sanitized, "\n", "\\n")
	
	// Remove control characters except newlines and tabs
	var result strings.Builder
	for _, r := range sanitized {
		if r >= 32 || r == '\n' || r == '\t' {
			result.WriteRune(r)
		}
	}
	
	return result.String(), nil
}

//nolint:unused // Infrastructure function for future use
func wrapError(err error, operation string, context map[string]interface{}) error {
	if err == nil {
		return nil
	}
	
	var contextStr strings.Builder
	contextStr.WriteString(operation)
	
	if len(context) > 0 {
		contextStr.WriteString(" with context: ")
		for k, v := range context {
			contextStr.WriteString(fmt.Sprintf("%s=%v ", k, v))
		}
	}
	
	return fmt.Errorf("%s: %w", contextStr.String(), err)
}

//nolint:unused // Infrastructure function for future use
func processLogsInChunks(logs io.Reader, chunkSize int, processor func([]byte) error) error {
	if chunkSize <= 0 {
		chunkSize = 64 * 1024 // Default 64KB chunks
	}
	
	buffer := make([]byte, chunkSize)
	for {
		n, err := logs.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return wrapError(err, "reading log chunk", map[string]interface{}{
				"chunk_size": chunkSize,
			})
		}
		
		if err := processor(buffer[:n]); err != nil {
			return wrapError(err, "processing log chunk", map[string]interface{}{
				"chunk_size": n,
			})
		}
	}
	
	return nil
}

//nolint:unused // Infrastructure function for future use
func retryWithBackoff(ctx context.Context, operation func() error, maxRetries int, baseDelay time.Duration) error {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		if err := operation(); err != nil {
			lastErr = err
			
			if attempt < maxRetries {
				delay := time.Duration(1<<uint(attempt)) * baseDelay
				if delay > 30*time.Second {
					delay = 30 * time.Second
				}
				
				select {
				case <-time.After(delay):
					continue
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		} else {
			return nil
		}
	}
	
	return wrapError(lastErr, "operation failed after retries", map[string]interface{}{
		"max_retries": maxRetries,
		"attempts": maxRetries + 1,
	})
}

// NewAnalysisCache creates a new analysis cache
func NewAnalysisCache(maxSize int, defaultTTL time.Duration) *AnalysisCache {
	return &AnalysisCache{
		cache:   make(map[string]*CacheEntry),
		maxSize: maxSize,
	}
}

// Get retrieves an item from cache
func (c *AnalysisCache) Get(key string) (*FailureAnalysisResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	entry, exists := c.cache[key]
	if !exists {
		return nil, false
	}
	
	// Check TTL
	if time.Since(entry.timestamp) > entry.ttl {
		// Entry expired, remove it
		delete(c.cache, key)
		return nil, false
	}
	
	return entry.result, true
}

// Set stores an item in cache
func (c *AnalysisCache) Set(key string, result *FailureAnalysisResult, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// If cache is full, remove oldest entry
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}
	
	c.cache[key] = &CacheEntry{
		result:    result,
		timestamp: time.Now(),
		ttl:       ttl,
	}
}

func (c *AnalysisCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true
	
	for key, entry := range c.cache {
		if first || entry.timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.timestamp
			first = false
		}
	}
	
	if oldestKey != "" {
		delete(c.cache, oldestKey)
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if an operation is allowed
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	
	// Refill tokens based on elapsed time
	if elapsed >= rl.refillRate {
		tokensToAdd := int(elapsed / rl.refillRate)
		rl.tokens = min(rl.maxTokens, rl.tokens+tokensToAdd)
		rl.lastRefill = now
	}
	
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//nolint:unused // Infrastructure function for future use
func validateRunID(runID int64) error {
	if runID <= 0 {
		return fmt.Errorf("invalid run ID: %d, must be positive", runID)
	}
	
	// GitHub run IDs are typically large integers
	if runID < 1000000 {
		return fmt.Errorf("run ID %d appears invalid, expected GitHub Actions run ID", runID)
	}
	
	return nil
}

//nolint:unused // Infrastructure function for future use
func validateRepositoryName(owner, name string) error {
	if owner == "" {
		return fmt.Errorf("repository owner cannot be empty")
	}
	
	if name == "" {
		return fmt.Errorf("repository name cannot be empty")
	}
	
	// GitHub repository names have specific rules
	if len(owner) > 39 || len(name) > 100 {
		return fmt.Errorf("repository owner/name too long: %s/%s", owner, name)
	}
	
	// Check for invalid characters (simplified)
	invalidChars := []string{" ", ".", "..", "~", "^", ":", "?", "*", "[", "\\"}
	for _, char := range invalidChars {
		if strings.Contains(owner, char) || strings.Contains(name, char) {
			return fmt.Errorf("repository owner/name contains invalid character '%s': %s/%s", char, owner, name)
		}
	}
	
	return nil
}

//nolint:unused // Infrastructure function for future use
func validateLLMProvider(provider LLMProvider) error {
	validProviders := []LLMProvider{OpenAI, Anthropic, Gemini, DeepSeek, LiteLLM}
	
	for _, validProvider := range validProviders {
		if provider == validProvider {
			return nil
		}
	}
	
	return fmt.Errorf("unsupported LLM provider: %s, supported providers: %v", 
		provider, validProviders)
}

//nolint:unused // Infrastructure function for future use
func createTimeoutContext(parent context.Context, operation string, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	
	ctx, cancel := context.WithTimeout(parent, timeout)
	
	// Add operation info to context for better debugging
	ctx = context.WithValue(ctx, operationContextKey, operation)
	ctx = context.WithValue(ctx, timeoutContextKey, timeout)
	
	return ctx, cancel
}

// Resource cleanup utilities
type ResourceManager struct {
	resources []io.Closer
	mu        sync.Mutex
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources: make([]io.Closer, 0),
	}
}

func (rm *ResourceManager) Add(resource io.Closer) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.resources = append(rm.resources, resource)
}

func (rm *ResourceManager) Cleanup() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	for _, resource := range rm.resources {
		if resource != nil {
			if err := resource.Close(); err != nil {
				// Log but don't return error during cleanup
				logrus.WithError(err).Warn("Failed to close resource during cleanup")
			}
		}
	}
	
	rm.resources = rm.resources[:0]
}

// Circuit breaker pattern for external services
type CircuitBreaker struct {
	mu           sync.RWMutex
	state        CircuitState
	failures     int
	maxFailures  int
	resetTimeout time.Duration
	lastFailTime time.Time
}

type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        CircuitClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

func (cb *CircuitBreaker) Execute(operation func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	switch cb.state {
	case CircuitOpen:
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	case CircuitHalfOpen:
		// Allow one request through
	case CircuitClosed:
		// Normal operation
	}
	
	err := operation()
	
	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()
		
		if cb.failures >= cb.maxFailures {
			cb.state = CircuitOpen
		}
		
		return err
	}
	
	// Success - reset failures and close circuit
	cb.failures = 0
	cb.state = CircuitClosed
	
	return nil
}

// Graceful shutdown handler
type GracefulShutdown struct {
	shutdownFuncs []func() error
	mu           sync.Mutex
}

func NewGracefulShutdown() *GracefulShutdown {
	return &GracefulShutdown{
		shutdownFuncs: make([]func() error, 0),
	}
}

func (gs *GracefulShutdown) AddShutdownFunc(f func() error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.shutdownFuncs = append(gs.shutdownFuncs, f)
}

func (gs *GracefulShutdown) Shutdown() error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	
	var errors []error
	
	for _, f := range gs.shutdownFuncs {
		if err := f(); err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}
	
	return nil
}
