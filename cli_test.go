package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestSetupLoggingUsesEnvVars(t *testing.T) {
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "text")
	defer func() {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FORMAT")
	}()

	cli := NewCLI()
	cli.setupLogging()

	assert.Equal(t, logrus.DebugLevel, cli.logger.GetLevel())
	_, ok := cli.logger.Formatter.(*logrus.TextFormatter)
	assert.True(t, ok)
}

func TestNewCLI(t *testing.T) {
	cli := NewCLI()

	assert.NotNil(t, cli)
	assert.NotNil(t, cli.logger)
	assert.NotNil(t, cli.rootCmd)
	assert.Equal(t, "github-autofix", cli.rootCmd.Use)
	assert.Equal(t, "1.0.0", cli.rootCmd.Version)

	// Check that all commands were added
	commands := cli.rootCmd.Commands()
	commandNames := make([]string, len(commands))
	for i, cmd := range commands {
		commandNames[i] = cmd.Name()
	}

	expectedCommands := []string{"monitor", "analyze", "fix", "validate", "status", "config", "test"}
	for _, expected := range expectedCommands {
		assert.Contains(t, commandNames, expected)
	}
}

func TestSetupLogging(t *testing.T) {
	tests := []struct {
		name           string
		logLevel       string
		logFormat      string
		verbose        bool
		expectedLevel  logrus.Level
		expectedFormat string
	}{
		{
			name:           "Debug level with JSON format",
			logLevel:       "debug",
			logFormat:      "json",
			verbose:        false,
			expectedLevel:  logrus.DebugLevel,
			expectedFormat: "json",
		},
		{
			name:           "Info level with text format",
			logLevel:       "info",
			logFormat:      "text",
			verbose:        false,
			expectedLevel:  logrus.InfoLevel,
			expectedFormat: "text",
		},
		{
			name:           "Verbose overrides log level",
			logLevel:       "error",
			logFormat:      "json",
			verbose:        true,
			expectedLevel:  logrus.DebugLevel,
			expectedFormat: "json",
		},
		{
			name:           "Invalid log level defaults to info",
			logLevel:       "invalid",
			logFormat:      "json",
			verbose:        false,
			expectedLevel:  logrus.InfoLevel,
			expectedFormat: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("LOG_LEVEL", tt.logLevel)
			os.Setenv("LOG_FORMAT", tt.logFormat)
			os.Setenv("VERBOSE", "false")
			if tt.verbose {
				os.Setenv("VERBOSE", "true")
			}
			defer func() {
				os.Unsetenv("LOG_LEVEL")
				os.Unsetenv("LOG_FORMAT")
				os.Unsetenv("VERBOSE")
			}()

			cli := NewCLI()
			cli.setupLogging()

			assert.Equal(t, tt.expectedLevel, cli.logger.GetLevel())

			if tt.expectedFormat == "text" {
				_, ok := cli.logger.Formatter.(*logrus.TextFormatter)
				assert.True(t, ok)
			} else {
				_, ok := cli.logger.Formatter.(*logrus.JSONFormatter)
				assert.True(t, ok)
			}
		})
	}
}

func TestGetCurrentConfig(t *testing.T) {
	cli := NewCLI()

	// Set up environment variables
	os.Setenv("GITHUB_TOKEN", "env-token")
	os.Setenv("LLM_PROVIDER", "anthropic")
	os.Setenv("LLM_API_KEY", "env-key")
	os.Setenv("REPO_OWNER", "env-owner")
	os.Setenv("REPO_NAME", "env-repo")
	os.Setenv("TARGET_BRANCH", "develop")
	os.Setenv("MIN_COVERAGE", "90")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("LLM_PROVIDER")
		os.Unsetenv("LLM_API_KEY")
		os.Unsetenv("REPO_OWNER")
		os.Unsetenv("REPO_NAME")
		os.Unsetenv("TARGET_BRANCH")
		os.Unsetenv("MIN_COVERAGE")
	}()

	config := cli.getCurrentConfig(cli.rootCmd)

	assert.Equal(t, "env-token", config.GitHubToken)
	assert.Equal(t, "anthropic", config.LLMProvider)
	assert.Equal(t, "env-key", config.LLMAPIKey)
	assert.Equal(t, "env-owner", config.RepoOwner)
	assert.Equal(t, "env-repo", config.RepoName)
	assert.Equal(t, "develop", config.TargetBranch)
	assert.Equal(t, 90, config.MinCoverage)
}

func TestGetStringValue(t *testing.T) {
	cli := NewCLI()
	cmd := cli.rootCmd

	tests := []struct {
		name     string
		flagName string
		envName  string
		envValue string
		expected string
		setFlag  bool
	}{
		{
			name:     "Env used when flag not set",
			flagName: "github-token",
			envName:  "GITHUB_TOKEN",
			envValue: "env-value",
			expected: "env-value",
			setFlag:  false,
		},
		{
			name:     "Default used when neither set",
			flagName: "github-token",
			envName:  "GITHUB_TOKEN",
			envValue: "",
			expected: "",
			setFlag:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment first
			os.Unsetenv(tt.envName)

			// Set up environment
			if tt.envValue != "" {
				os.Setenv(tt.envName, tt.envValue)
				defer os.Unsetenv(tt.envName)
			}

			result := cli.getStringValue(cmd, tt.flagName, tt.envName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIntValue(t *testing.T) {
	cli := NewCLI()
	cmd := cli.rootCmd

	tests := []struct {
		name     string
		flagName string
		envName  string
		envValue string
		expected int
	}{
		{
			name:     "Valid env value",
			flagName: "min-coverage",
			envName:  "MIN_COVERAGE",
			envValue: "90",
			expected: 90,
		},
		{
			name:     "Invalid env value uses default",
			flagName: "min-coverage",
			envName:  "MIN_COVERAGE",
			envValue: "invalid",
			expected: 85,
		},
		{
			name:     "No env value uses default",
			flagName: "min-coverage",
			envName:  "MIN_COVERAGE",
			envValue: "",
			expected: 85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envName, tt.envValue)
				defer os.Unsetenv(tt.envName)
			}

			result := cli.getIntValue(cmd, tt.flagName, tt.envName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetBoolValue(t *testing.T) {
	cli := NewCLI()
	cmd := cli.rootCmd

	tests := []struct {
		name     string
		flagName string
		envName  string
		envValue string
		expected bool
	}{
		{
			name:     "True string value",
			flagName: "dry-run",
			envName:  "DRY_RUN",
			envValue: "true",
			expected: true,
		},
		{
			name:     "False string value",
			flagName: "dry-run",
			envName:  "DRY_RUN",
			envValue: "false",
			expected: false,
		},
		{
			name:     "Invalid string value defaults to false",
			flagName: "dry-run",
			envName:  "DRY_RUN",
			envValue: "invalid",
			expected: false,
		},
		{
			name:     "Empty value defaults to false",
			flagName: "dry-run",
			envName:  "DRY_RUN",
			envValue: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envName, tt.envValue)
				defer os.Unsetenv(tt.envName)
			}

			result := cli.getBoolValue(cmd, tt.flagName, tt.envName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	cli := NewCLI()

	// Create temporary file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".github-autofix.env")

	err := cli.createDefaultConfig(configFile)
	require.NoError(t, err)

	// Check that file was created
	assert.FileExists(t, configFile)

	// Read file content
	content, err := os.ReadFile(configFile)
	require.NoError(t, err)

	configStr := string(content)

	// Check that expected content is present
	assert.Contains(t, configStr, "GITHUB_TOKEN=your_github_token_here")
	assert.Contains(t, configStr, "LLM_PROVIDER=openai")
	assert.Contains(t, configStr, "MIN_COVERAGE=85")
	assert.Contains(t, configStr, "LOG_LEVEL=info")
	assert.Contains(t, configStr, "LOG_FORMAT=json")
}

func TestMaskToken(t *testing.T) {
	cli := NewCLI()

	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "GitHub token gets masked",
			token:    "ghp_1234567890abcdefghijklmnopqrstuvwxyz",
			expected: "ghp_***wxyz",
		},
		{
			name:     "GitHub PAT token gets masked",
			token:    "github_pat_1234567890abcdefghijklmnopqrstuvwxyz",
			expected: "github_pat_***wxyz",
		},
		{
			name:     "OpenAI key gets masked",
			token:    "sk-1234567890abcdefghijklmnopqrstuvwxyz",
			expected: "sk-***wxyz",
		},
		{
			name:     "Short token gets masked",
			token:    "short",
			expected: "***",
		},
		{
			name:     "Empty token",
			token:    "",
			expected: "***",
		},
		{
			name:     "Regular token with fallback",
			token:    "abcdefghijklmnopqrstuvwxyz",
			expected: "abcd***wxyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cli.maskToken(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrintConfig(t *testing.T) {
	cli := NewCLI()

	config := &CLIConfig{
		GitHubToken:  "test-token",
		LLMProvider:  "openai",
		LLMAPIKey:    "test-key",
		RepoOwner:    "test-owner",
		RepoName:     "test-repo",
		TargetBranch: "main",
		MinCoverage:  85,
		Verbose:      true,
		DryRun:       false,
		LogLevel:     "info",
		LogFormat:    "json",
	}

	// This test mainly ensures the function doesn't panic
	// Since it writes to stdout, we can't easily capture the output
	assert.NotPanics(t, func() {
		cli.printConfig(config)
	})
}
