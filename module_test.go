package main

import (
	"testing"
)

// TestModuleBasics ensures the module is properly configured
func TestModuleBasics(t *testing.T) {
	t.Run("Module loads correctly", func(t *testing.T) {
		// This test ensures the module can be imported and basic functionality works
		if testing.Short() {
			t.Skip("Skipping module test in short mode")
		}
		
		// Test that we can create a new DaggerAutofix instance
		module := New()
		if module == nil {
			t.Fatal("Failed to create new DaggerAutofix instance")
		}
		
		// Test basic configuration
		if module.LLMProvider != "openai" {
			t.Errorf("Expected default LLM provider to be 'openai', got '%s'", module.LLMProvider)
		}
		
		if module.TargetBranch != "main" {
			t.Errorf("Expected default target branch to be 'main', got '%s'", module.TargetBranch)
		}
		
		if module.MinCoverage != 85 {
			t.Errorf("Expected default min coverage to be 85, got %d", module.MinCoverage)
		}
	})
}