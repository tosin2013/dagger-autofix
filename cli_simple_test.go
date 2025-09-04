package main

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCLIHelperFunctions(t *testing.T) {
	// Test functions that don't require full CLI setup
	logger := logrus.New()
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	cli := NewCLI()
	cli.logger = logger

	t.Run("loadConfiguration", func(t *testing.T) {
		// Save original environment
		originalEnv := map[string]string{
			"GITHUB_TOKEN": os.Getenv("GITHUB_TOKEN"),
			"LLM_PROVIDER": os.Getenv("LLM_PROVIDER"),
			"LLM_API_KEY":  os.Getenv("LLM_API_KEY"),
			"REPO_OWNER":   os.Getenv("REPO_OWNER"),
			"REPO_NAME":    os.Getenv("REPO_NAME"),
		}

		defer func() {
			for key, value := range originalEnv {
				if value == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}
		}()

		// Set test environment
		os.Setenv("GITHUB_TOKEN", "test_token")
		os.Setenv("LLM_PROVIDER", "openai")
		os.Setenv("LLM_API_KEY", "test_key")
		os.Setenv("REPO_OWNER", "test_owner")
		os.Setenv("REPO_NAME", "test_repo")

		cli.loadConfiguration()
		
		// Test that the function executes without error (it doesn't return anything)
		// Configuration loading is tested by the other CLI functions
	})
}

func TestCLIExecuteFunction(t *testing.T) {
	// Test the main Execute function
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	
	// Test help command
	os.Args = []string{"github-autofix", "--help"}
	
	// Capture output
	var buf bytes.Buffer
	cli := NewCLI()
	cli.rootCmd.SetOut(&buf)
	cli.rootCmd.SetErr(&buf)
	
	err := cli.Execute()
	assert.NoError(t, err)
}

func TestPrintFunctions(t *testing.T) {
	logger := logrus.New()
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	cli := NewCLI()
	cli.logger = logger

	t.Run("printAnalysisResult", func(t *testing.T) {
		result := &FailureAnalysisResult{
			ID:             "test-analysis-1",
			RootCause:      "Test failure",
			Description:    "Test description",
			Classification: FailureClassification{
				Type:       BuildFailure,
				Confidence: 0.9,
			},
		}
		cli.printAnalysisResult(result)
		// Test passes if no panic occurs
	})

	t.Run("printGeneratedFixes", func(t *testing.T) {
		fixes := []*ProposedFix{
			{
				ID:          "test-fix-1",
				Description: "Test fix description",
				Changes: []CodeChange{
					{
						FilePath:  "test.go",
						Operation: "modify",
						NewContent: "test content",
					},
				},
				Confidence: 0.8,
			},
		}
		cli.printGeneratedFixes(fixes)
		// Test passes if no panic occurs
	})

	t.Run("printAutoFixResult", func(t *testing.T) {
		result := &AutoFixResult{
			ID:      "test-fix-1",
			Success: true,
			PullRequest: &PullRequest{
				Number: 123,
				URL:    "https://github.com/test/test/pull/123",
			},
		}
		cli.printAutoFixResult(result)
		// Test passes if no panic occurs
	})

	t.Run("printTestResult", func(t *testing.T) {
		result := &TestResult{
			Success:     true,
			Coverage:    85.5,
			PassedTests: 10,
			FailedTests: 0,
			Details:     map[string]interface{}{"message": "All tests passed"},
		}
		cli.printTestResult(result)
		// Test passes if no panic occurs
	})

	t.Run("printMetrics", func(t *testing.T) {
		metrics := &OperationalMetrics{
			TotalFailuresDetected: 10,
			SuccessfulFixes:       8,
			FailedFixes:           2,
			AverageFixTime:        time.Minute * 5,
			TestCoverage:          85.5,
		}
		cli.printMetrics(metrics)
		// Test passes if no panic occurs
	})
}

func TestMainFunction(t *testing.T) {
	// Test that main function exists and can be called
	// We can't easily test main() directly, but we can test the CLI creation
	cli := NewCLI()
	assert.NotNil(t, cli)
	assert.NotNil(t, cli.rootCmd)
	assert.NotNil(t, cli.logger)
}