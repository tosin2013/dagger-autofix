package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test the core main.go functions that can be tested in isolation
func TestMainFunctions(t *testing.T) {
	t.Run("shouldProcessRun", func(t *testing.T) {
		// Create a minimal DaggerAutofix instance for testing
		autofix := &DaggerAutofix{}
		
		// Create test workflow run
		run := &WorkflowRun{
			ID:     12345,
			Status: "failed",
		}
		
		// Test processing run (currently always returns true)
		assert.True(t, autofix.shouldProcessRun(run))
	})

	t.Run("selectBestFix", func(t *testing.T) {
		autofix := &DaggerAutofix{}
		
		// Test with empty validations
		result := autofix.selectBestFix([]*FixValidationResult{})
		assert.Nil(t, result)
		
		// Test with invalid validations
		invalidValidations := []*FixValidationResult{
			{Valid: false},
			{Valid: false},
		}
		result = autofix.selectBestFix(invalidValidations)
		assert.Nil(t, result)
		
		// Test with valid validations
		validValidations := []*FixValidationResult{
			{
				Valid: true,
				Fix: &ProposedFix{
					Confidence: 0.7,
				},
			},
			{
				Valid: true,
				Fix: &ProposedFix{
					Confidence: 0.9,
				},
			},
		}
		result = autofix.selectBestFix(validValidations)
		assert.NotNil(t, result)
		assert.Equal(t, 0.9, result.Fix.Confidence)
	})
}

// Test simple configuration setters that don't require dagger.Secret
func TestConfigurationBuilders(t *testing.T) {
	autofix := &DaggerAutofix{}
	
	t.Run("WithRepository", func(t *testing.T) {
		result := autofix.WithRepository("owner", "repo")
		assert.Equal(t, autofix, result)
		assert.Equal(t, "owner", autofix.RepoOwner)
		assert.Equal(t, "repo", autofix.RepoName)
	})
	
	t.Run("WithTargetBranch", func(t *testing.T) {
		result := autofix.WithTargetBranch("main")
		assert.Equal(t, autofix, result)
		assert.Equal(t, "main", autofix.TargetBranch)
	})
	
	t.Run("WithMinCoverage", func(t *testing.T) {
		result := autofix.WithMinCoverage(85)
		assert.Equal(t, autofix, result)
		assert.Equal(t, 85, autofix.MinCoverage)
	})
}