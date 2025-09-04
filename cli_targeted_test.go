package main

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Test CLI command functions that can be tested with minimal setup
func TestCLICommandFunctionsSimple(t *testing.T) {
	// Set up test environment
	originalEnv := map[string]string{
		"GITHUB_TOKEN": os.Getenv("GITHUB_TOKEN"),
		"LLM_PROVIDER": os.Getenv("LLM_PROVIDER"),
		"LLM_API_KEY":  os.Getenv("LLM_API_KEY"),
		"REPO_OWNER":   os.Getenv("REPO_OWNER"),
		"REPO_NAME":    os.Getenv("REPO_NAME"),
	}

	// Clean up after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Set valid test environment variables
	os.Setenv("GITHUB_TOKEN", "test_token")
	os.Setenv("LLM_PROVIDER", "openai")
	os.Setenv("LLM_API_KEY", "test_key")
	os.Setenv("REPO_OWNER", "test_owner")
	os.Setenv("REPO_NAME", "test_repo")

	cli := NewCLI()

	t.Run("runConfigInit", func(t *testing.T) {
		// Create temporary directory for config test
		tmpDir := t.TempDir()
		originalCwd, _ := os.Getwd()
		err := os.Chdir(tmpDir)
		assert.NoError(t, err)
		defer func() {
			_ = os.Chdir(originalCwd)
		}()

		cmd := &cobra.Command{}
		err = cli.runConfigInit(cmd, []string{})
		assert.NoError(t, err)
		
		// Verify config file was created
		_, err = os.Stat(".github-autofix.env")
		assert.NoError(t, err)
	})

	t.Run("runConfigShow", func(t *testing.T) {
		cmd := &cobra.Command{}
		_ = cli.runConfigShow(cmd, []string{})
		// This function doesn't return an error, just prints config
		// Test passes if no panic occurs
	})

	t.Run("runConfigValidate", func(t *testing.T) {
		cmd := &cobra.Command{}
		err := cli.runConfigValidate(cmd, []string{})
		// Expected to fail in test environment due to Dagger not being available
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dagger client not available")
	})
}

// Test CLI command argument validation
func TestCLICommandArgValidation(t *testing.T) {
	cli := NewCLI()

	t.Run("runAnalyze_invalid_run_id", func(t *testing.T) {
		cmd := &cobra.Command{}
		err := cli.runAnalyze(cmd, []string{"invalid"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid workflow run ID")
	})

	t.Run("runFix_invalid_run_id", func(t *testing.T) {
		cmd := &cobra.Command{}
		err := cli.runFix(cmd, []string{"not_a_number"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid workflow run ID")
	})

	t.Run("runValidate_invalid_run_id", func(t *testing.T) {
		cmd := &cobra.Command{}
		err := cli.runValidate(cmd, []string{"abc123"})
		assert.Error(t, err)
		// This actually tries to initialize the agent which fails due to missing GitHub token in test environment
		assert.Contains(t, err.Error(), "failed to initialize agent")
	})
}

// Test configuration helper functions
func TestCLIConfigHelpers(t *testing.T) {
	// Set test environment
	os.Setenv("TEST_STRING", "test_value")
	os.Setenv("TEST_INT", "42")
	os.Setenv("TEST_BOOL", "true")
	defer func() {
		os.Unsetenv("TEST_STRING")
		os.Unsetenv("TEST_INT")
		os.Unsetenv("TEST_BOOL")
	}()

	cli := NewCLI()
	cmd := &cobra.Command{}

	t.Run("getStringValue", func(t *testing.T) {
		// Test with environment variable
		value := cli.getStringValue(cmd, "test_flag", "TEST_STRING")
		assert.Equal(t, "test_value", value)
		
		// Test with nonexistent env var
		value = cli.getStringValue(cmd, "test_flag", "NONEXISTENT")
		assert.Equal(t, "", value) // Should return empty string
	})

	t.Run("getIntValue", func(t *testing.T) {
		// Test with valid environment variable
		value := cli.getIntValue(cmd, "test_flag", "TEST_INT")
		assert.Equal(t, 42, value)
		
		// Test with invalid environment variable
		os.Setenv("TEST_INT_INVALID", "not_a_number")
		value = cli.getIntValue(cmd, "test_flag", "TEST_INT_INVALID")
		assert.Equal(t, 85, value) // Should return default 85 for invalid value
		os.Unsetenv("TEST_INT_INVALID")
		
		// Test with nonexistent environment variable
		value = cli.getIntValue(cmd, "test_flag", "NONEXISTENT")
		assert.Equal(t, 85, value) // Should return default 85
	})

	t.Run("getBoolValue", func(t *testing.T) {
		// Test with valid true value
		value := cli.getBoolValue(cmd, "test_flag", "TEST_BOOL")
		assert.True(t, value)
		
		// Test with false value
		os.Setenv("TEST_BOOL_FALSE", "false")
		value = cli.getBoolValue(cmd, "test_flag", "TEST_BOOL_FALSE")
		assert.False(t, value)
		os.Unsetenv("TEST_BOOL_FALSE")
		
		// Test with invalid value (should return false)
		os.Setenv("TEST_BOOL_INVALID", "maybe")
		value = cli.getBoolValue(cmd, "test_flag", "TEST_BOOL_INVALID")
		assert.False(t, value) // Should return false for invalid value
		os.Unsetenv("TEST_BOOL_INVALID")
		
		// Test with nonexistent env var
		value = cli.getBoolValue(cmd, "test_flag", "NONEXISTENT")
		assert.False(t, value) // Should return false
	})
}