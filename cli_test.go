package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLIConfigEnvVars(t *testing.T) {
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("LLM_API_KEY", "llm-key")
	os.Setenv("REPO_OWNER", "owner")
	os.Setenv("REPO_NAME", "repo")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "text")
	os.Setenv("DRY_RUN", "true")
	os.Setenv("VERBOSE", "true")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("LLM_API_KEY")
		os.Unsetenv("REPO_OWNER")
		os.Unsetenv("REPO_NAME")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FORMAT")
		os.Unsetenv("DRY_RUN")
		os.Unsetenv("VERBOSE")
	}()

	cli := NewCLI()
	cfg := cli.getCurrentConfig(cli.rootCmd)

	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "text", cfg.LogFormat)
	assert.True(t, cfg.DryRun)
	assert.True(t, cfg.Verbose)
}
