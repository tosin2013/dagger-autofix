package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/sirupsen/logrus"
)

// CLI represents the command-line interface for the GitHub Auto-Fix Agent
type CLI struct {
	logger *logrus.Logger
	rootCmd *cobra.Command
	agent  *DaggerAutofix
}

// CLIConfig holds CLI configuration
type CLIConfig struct {
	GitHubToken   string `json:"github_token"`
	LLMProvider   string `json:"llm_provider"`
	LLMAPIKey     string `json:"llm_api_key"`
	RepoOwner     string `json:"repo_owner"`
	RepoName      string `json:"repo_name"`
	TargetBranch  string `json:"target_branch"`
	MinCoverage   int    `json:"min_coverage"`
	ConfigFile    string `json:"config_file"`
	Verbose       bool   `json:"verbose"`
	DryRun        bool   `json:"dry_run"`
	LogLevel      string `json:"log_level"`
	LogFormat     string `json:"log_format"`
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	cli := &CLI{
		logger: logger,
	}

	cli.setupRootCommand()
	cli.setupCommands()

	return cli
}

// Execute runs the CLI
func (c *CLI) Execute() error {
	return c.rootCmd.Execute()
}

func (c *CLI) setupRootCommand() {
	c.rootCmd = &cobra.Command{
		Use:   "github-autofix",
		Short: "GitHub Actions Auto-Fix Agent",
		Long: `A comprehensive Dagger.io agent for automatically resolving GitHub Actions pipeline failures.

This tool monitors GitHub Actions workflows, analyzes failures using LLM-powered intelligence,
generates and validates fixes, and creates pull requests automatically.

Supports multiple LLM providers: OpenAI, Anthropic, Gemini, DeepSeek, and LiteLLM proxy.`,
		Version: "1.0.0",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			c.setupLogging()
			c.loadConfiguration()
		},
	}

	// Global flags
	c.rootCmd.PersistentFlags().String("config", ".github-autofix.env", "Configuration file path")
	c.rootCmd.PersistentFlags().String("github-token", "", "GitHub personal access token")
	c.rootCmd.PersistentFlags().String("llm-provider", "openai", "LLM provider (openai, anthropic, gemini, deepseek, litellm)")
	c.rootCmd.PersistentFlags().String("llm-api-key", "", "LLM API key")
	c.rootCmd.PersistentFlags().String("repo-owner", "", "GitHub repository owner")
	c.rootCmd.PersistentFlags().String("repo-name", "", "GitHub repository name")
	c.rootCmd.PersistentFlags().String("target-branch", "main", "Target branch for fixes")
	c.rootCmd.PersistentFlags().Int("min-coverage", 85, "Minimum test coverage percentage")
	c.rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose logging")
	c.rootCmd.PersistentFlags().Bool("dry-run", false, "Dry run mode (no actual changes)")
	c.rootCmd.PersistentFlags().String("log-level", "info", "Log level (trace, debug, info, warn, error)")
	c.rootCmd.PersistentFlags().String("log-format", "json", "Log format (json, text)")
}

func (c *CLI) setupCommands() {
	// Monitor command
	monitorCmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor GitHub Actions workflows for failures",
		Long:  "Continuously monitor GitHub Actions workflows and automatically fix failures when detected.",
		RunE:  c.runMonitor,
	}

	// Analyze command
	analyzeCmd := &cobra.Command{
		Use:   "analyze [workflow-run-id]",
		Short: "Analyze a specific workflow failure",
		Long:  "Analyze a specific GitHub Actions workflow run failure and provide detailed insights.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.runAnalyze,
	}

	// Fix command
	fixCmd := &cobra.Command{
		Use:   "fix [workflow-run-id]",
		Short: "Generate and apply fixes for a workflow failure",
		Long:  "Generate fixes for a specific workflow failure, validate them, and create a pull request.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.runFix,
	}

	// Validate command
	validateCmd := &cobra.Command{
		Use:   "validate [branch]",
		Short: "Validate fixes by running tests",
		Long:  "Run tests and validation checks on a specific branch to verify fixes.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.runValidate,
	}

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show agent status and metrics",
		Long:  "Display current status, metrics, and operational information about the auto-fix agent.",
		RunE:  c.runStatus,
	}

	// Config command
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long:  "Manage configuration settings for the auto-fix agent.",
	}

	// Config subcommands
	configInitCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration file",
		RunE:  c.runConfigInit,
	}

	configShowCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE:  c.runConfigShow,
	}

	configValidateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		RunE:  c.runConfigValidate,
	}

	// Test command
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test agent functionality",
		Long:  "Run tests to verify agent functionality and integration.",
	}

	// Test subcommands
	testConnectionCmd := &cobra.Command{
		Use:   "connection",
		Short: "Test connections to GitHub and LLM providers",
		RunE:  c.runTestConnection,
	}

	testLLMCmd := &cobra.Command{
		Use:   "llm",
		Short: "Test LLM provider functionality",
		RunE:  c.runTestLLM,
	}

	// Add subcommands
	configCmd.AddCommand(configInitCmd, configShowCmd, configValidateCmd)
	testCmd.AddCommand(testConnectionCmd, testLLMCmd)
	c.rootCmd.AddCommand(monitorCmd, analyzeCmd, fixCmd, validateCmd, statusCmd, configCmd, testCmd)
}

// Command implementations

func (c *CLI) runMonitor(cmd *cobra.Command, args []string) error {
	c.logger.Info("Starting workflow monitoring")

	ctx := context.Background()
	agent, err := c.initializeAgent(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	return agent.MonitorWorkflows(ctx)
}

func (c *CLI) runAnalyze(cmd *cobra.Command, args []string) error {
	runIDStr := args[0]
	runID, err := strconv.ParseInt(runIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid workflow run ID: %w", err)
	}

	c.logger.WithField("run_id", runID).Info("Analyzing workflow failure")

	ctx := context.Background()
	agent, err := c.initializeAgent(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	analysis, err := agent.AnalyzeFailure(ctx, runID)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	c.printAnalysisResult(analysis)
	return nil
}

func (c *CLI) runFix(cmd *cobra.Command, args []string) error {
	runIDStr := args[0]
	runID, err := strconv.ParseInt(runIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid workflow run ID: %w", err)
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	c.logger.WithFields(logrus.Fields{
		"run_id":  runID,
		"dry_run": dryRun,
	}).Info("Generating and applying fix")

	ctx := context.Background()
	agent, err := c.initializeAgent(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	if dryRun {
		// In dry-run mode, only analyze and generate fixes
		analysis, err := agent.AnalyzeFailure(ctx, runID)
		if err != nil {
			return fmt.Errorf("analysis failed: %w", err)
		}

		fixes, err := agent.failureEngine.GenerateFixes(ctx, analysis)
		if err != nil {
			return fmt.Errorf("fix generation failed: %w", err)
		}

		c.printAnalysisResult(analysis)
		c.printGeneratedFixes(fixes)
		return nil
	}

	result, err := agent.AutoFix(ctx, runID)
	if err != nil {
		return fmt.Errorf("auto-fix failed: %w", err)
	}

	c.printAutoFixResult(result)
	return nil
}

func (c *CLI) runValidate(cmd *cobra.Command, args []string) error {
	branch := args[0]
	c.logger.WithField("branch", branch).Info("Running validation")

	ctx := context.Background()
	agent, err := c.initializeAgent(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	testResult, err := agent.testEngine.RunTests(ctx, agent.RepoOwner, agent.RepoName, branch)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	c.printTestResult(testResult)
	return nil
}

func (c *CLI) runStatus(cmd *cobra.Command, args []string) error {
	c.logger.Info("Getting agent status")

	ctx := context.Background()
	agent, err := c.initializeAgent(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	metrics, err := agent.GetMetrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	c.printMetrics(metrics)
	return nil
}

func (c *CLI) runConfigInit(cmd *cobra.Command, args []string) error {
	c.logger.Info("Initializing configuration")

	configFile, _ := cmd.Parent().Parent().PersistentFlags().GetString("config")
	return c.createDefaultConfig(configFile)
}

func (c *CLI) runConfigShow(cmd *cobra.Command, args []string) error {
	c.logger.Info("Showing configuration")

	config := c.getCurrentConfig(cmd)
	c.printConfig(config)
	return nil
}

func (c *CLI) runConfigValidate(cmd *cobra.Command, args []string) error {
	c.logger.Info("Validating configuration")

	ctx := context.Background()
	_, err := c.initializeAgent(ctx)
	if err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	c.logger.Info("Configuration is valid")
	return nil
}

func (c *CLI) runTestConnection(cmd *cobra.Command, args []string) error {
	c.logger.Info("Testing connections")

	ctx := context.Background()
	_, err := c.initializeAgent(ctx)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	c.logger.Info("All connections successful")
	return nil
}

func (c *CLI) runTestLLM(cmd *cobra.Command, args []string) error {
	c.logger.Info("Testing LLM provider")

	config := c.getCurrentConfig(cmd)
	ctx := context.Background()

	// Create LLM client for testing
	apiKey := dag.SetSecret("llm-api-key", config.LLMAPIKey)
	llmClient, err := NewLLMClient(ctx, LLMProvider(config.LLMProvider), apiKey)
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Test with simple request
	req := &LLMRequest{
		Prompt:    "Hello, please respond with 'LLM test successful'",
		SystemMsg: "You are a helpful assistant for testing purposes.",
	}

	response, err := llmClient.Chat(ctx, req)
	if err != nil {
		return fmt.Errorf("LLM test failed: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"provider": config.LLMProvider,
		"response": response.Content,
	}).Info("LLM test successful")

	return nil
}

// Helper methods

func (c *CLI) setupLogging() {
	verbose, _ := c.rootCmd.PersistentFlags().GetBool("verbose")
	logLevel, _ := c.rootCmd.PersistentFlags().GetString("log-level")
	logFormat, _ := c.rootCmd.PersistentFlags().GetString("log-format")

	// Set log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	if verbose {
		level = logrus.DebugLevel
	}
	c.logger.SetLevel(level)

	// Set log format
	switch logFormat {
	case "text":
		c.logger.SetFormatter(&logrus.TextFormatter{})
	default:
		c.logger.SetFormatter(&logrus.JSONFormatter{})
	}
}

func (c *CLI) loadConfiguration() {
	configFile, _ := c.rootCmd.PersistentFlags().GetString("config")
	if configFile != "" {
		if err := godotenv.Load(configFile); err != nil {
			c.logger.WithError(err).Debug("Could not load config file, using environment variables")
		}
	}
}

func (c *CLI) initializeAgent(ctx context.Context) (*DaggerAutofix, error) {
	config := c.getCurrentConfig(c.rootCmd)

	// Validate required configuration
	if config.GitHubToken == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}
	if config.LLMAPIKey == "" {
		return nil, fmt.Errorf("LLM API key is required")
	}
	if config.RepoOwner == "" || config.RepoName == "" {
		return nil, fmt.Errorf("repository owner and name are required")
	}

	// Create agent
	agent := New().
		WithGitHubToken(dag.SetSecret("github-token", config.GitHubToken)).
		WithLLMProvider(config.LLMProvider, dag.SetSecret("llm-api-key", config.LLMAPIKey)).
		WithRepository(config.RepoOwner, config.RepoName).
		WithTargetBranch(config.TargetBranch).
		WithMinCoverage(config.MinCoverage)

	// Initialize agent
	return agent.Initialize(ctx)
}

func (c *CLI) getCurrentConfig(cmd *cobra.Command) *CLIConfig {
	config := &CLIConfig{}

	// Get values from flags or environment variables
	config.GitHubToken = c.getStringValue(cmd, "github-token", "GITHUB_TOKEN")
	config.LLMProvider = c.getStringValue(cmd, "llm-provider", "LLM_PROVIDER")
	config.LLMAPIKey = c.getStringValue(cmd, "llm-api-key", "LLM_API_KEY")
	config.RepoOwner = c.getStringValue(cmd, "repo-owner", "REPO_OWNER")
	config.RepoName = c.getStringValue(cmd, "repo-name", "REPO_NAME")
	config.TargetBranch = c.getStringValue(cmd, "target-branch", "TARGET_BRANCH")
	config.MinCoverage = c.getIntValue(cmd, "min-coverage", "MIN_COVERAGE")

	config.Verbose, _ = cmd.PersistentFlags().GetBool("verbose")
	config.DryRun, _ = cmd.PersistentFlags().GetBool("dry-run")
	config.LogLevel, _ = cmd.PersistentFlags().GetString("log-level")
	config.LogFormat, _ = cmd.PersistentFlags().GetString("log-format")
	config.ConfigFile, _ = cmd.PersistentFlags().GetString("config")

	return config
}

func (c *CLI) getStringValue(cmd *cobra.Command, flagName, envName string) string {
	if val, _ := cmd.PersistentFlags().GetString(flagName); val != "" {
		return val
	}
	return os.Getenv(envName)
}

func (c *CLI) getIntValue(cmd *cobra.Command, flagName, envName string) int {
	if val, _ := cmd.PersistentFlags().GetInt(flagName); val != 0 {
		return val
	}
	if envVal := os.Getenv(envName); envVal != "" {
		if intVal, err := strconv.Atoi(envVal); err == nil {
			return intVal
		}
	}
	return 85 // default
}

func (c *CLI) createDefaultConfig(filename string) error {
	configContent := `# GitHub Actions Auto-Fix Agent Configuration

# GitHub Settings
GITHUB_TOKEN=your_github_token_here
REPO_OWNER=your_repo_owner
REPO_NAME=your_repo_name
TARGET_BRANCH=main

# LLM Settings
LLM_PROVIDER=openai
LLM_API_KEY=your_llm_api_key_here

# Agent Settings
MIN_COVERAGE=85

# Logging Settings
LOG_LEVEL=info
LOG_FORMAT=json

# Optional: LiteLLM Proxy (if using)
# LITELLM_BASE_URL=http://localhost:4000

# Optional: Advanced Settings
# DRY_RUN=false
# VERBOSE=false
`

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(configContent)
	if err != nil {
		return err
	}

	c.logger.WithField("file", filename).Info("Configuration file created")
	return nil
}

func (c *CLI) printAnalysisResult(analysis *FailureAnalysisResult) {
	fmt.Printf("\n=== Failure Analysis Result ===\n")
	fmt.Printf("ID: %s\n", analysis.ID)
	fmt.Printf("Type: %s\n", analysis.Classification.Type)
	fmt.Printf("Severity: %s\n", analysis.Classification.Severity)
	fmt.Printf("Category: %s\n", analysis.Classification.Category)
	fmt.Printf("Confidence: %.1f%%\n", analysis.Classification.Confidence*100)
	fmt.Printf("Root Cause: %s\n", analysis.RootCause)
	fmt.Printf("Description: %s\n", analysis.Description)
	fmt.Printf("Processing Time: %v\n", analysis.ProcessingTime)

	if len(analysis.AffectedFiles) > 0 {
		fmt.Printf("\nAffected Files:\n")
		for _, file := range analysis.AffectedFiles {
			fmt.Printf("  - %s\n", file)
		}
	}

	if len(analysis.ErrorPatterns) > 0 {
		fmt.Printf("\nError Patterns:\n")
		for _, pattern := range analysis.ErrorPatterns {
			fmt.Printf("  - %s (%.1f%% confidence)\n", pattern.Description, pattern.Confidence*100)
		}
	}
	fmt.Println()
}

func (c *CLI) printGeneratedFixes(fixes []*ProposedFix) {
	fmt.Printf("\n=== Generated Fixes ===\n")
	for i, fix := range fixes {
		fmt.Printf("\nFix %d:\n", i+1)
		fmt.Printf("  ID: %s\n", fix.ID)
		fmt.Printf("  Type: %s\n", fix.Type)
		fmt.Printf("  Confidence: %.1f%%\n", fix.Confidence*100)
		fmt.Printf("  Description: %s\n", fix.Description)
		fmt.Printf("  Rationale: %s\n", fix.Rationale)

		if len(fix.Changes) > 0 {
			fmt.Printf("  Changes:\n")
			for _, change := range fix.Changes {
				fmt.Printf("    - %s: %s\n", change.Operation, change.FilePath)
			}
		}

		if len(fix.Risks) > 0 {
			fmt.Printf("  Risks:\n")
			for _, risk := range fix.Risks {
				fmt.Printf("    - %s\n", risk)
			}
		}
	}
	fmt.Println()
}

func (c *CLI) printAutoFixResult(result *AutoFixResult) {
	fmt.Printf("\n=== Auto-Fix Result ===\n")
	fmt.Printf("Success: %t\n", result.Success)
	fmt.Printf("Duration: %v\n", result.Duration)

	if result.PullRequest != nil {
		fmt.Printf("\nPull Request Created:\n")
		fmt.Printf("  Number: #%d\n", result.PullRequest.Number)
		fmt.Printf("  Title: %s\n", result.PullRequest.Title)
		fmt.Printf("  URL: %s\n", result.PullRequest.URL)
		fmt.Printf("  Branch: %s\n", result.PullRequest.Branch)
	}

	if result.Fix != nil {
		fmt.Printf("\nFix Validation:\n")
		fmt.Printf("  Valid: %t\n", result.Fix.Valid)
		fmt.Printf("  Tests Passed: %t\n", result.Fix.TestResult.Success)
		fmt.Printf("  Coverage: %.1f%%\n", result.Fix.TestResult.Coverage)
	}
	fmt.Println()
}

func (c *CLI) printTestResult(result *TestResult) {
	fmt.Printf("\n=== Test Results ===\n")
	fmt.Printf("Success: %t\n", result.Success)
	fmt.Printf("Total Tests: %d\n", result.TotalTests)
	fmt.Printf("Passed: %d\n", result.PassedTests)
	fmt.Printf("Failed: %d\n", result.FailedTests)
	fmt.Printf("Skipped: %d\n", result.SkippedTests)
	fmt.Printf("Coverage: %.1f%%\n", result.Coverage)
	fmt.Printf("Duration: %v\n", result.Duration)

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors:\n")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}
	fmt.Println()
}

func (c *CLI) printMetrics(metrics *OperationalMetrics) {
	fmt.Printf("\n=== Agent Metrics ===\n")
	fmt.Printf("Total Failures Detected: %d\n", metrics.TotalFailuresDetected)
	fmt.Printf("Successful Fixes: %d\n", metrics.SuccessfulFixes)
	fmt.Printf("Failed Fixes: %d\n", metrics.FailedFixes)
	fmt.Printf("Average Fix Time: %v\n", metrics.AverageFixTime)
	fmt.Printf("Test Coverage: %.1f%%\n", metrics.TestCoverage)
	fmt.Printf("Last Updated: %v\n", metrics.LastUpdated)
	fmt.Println()
}

func (c *CLI) printConfig(config *CLIConfig) {
	fmt.Printf("\n=== Current Configuration ===\n")
	fmt.Printf("GitHub Token: %s\n", c.maskToken(config.GitHubToken))
	fmt.Printf("LLM Provider: %s\n", config.LLMProvider)
	fmt.Printf("LLM API Key: %s\n", c.maskToken(config.LLMAPIKey))
	fmt.Printf("Repository: %s/%s\n", config.RepoOwner, config.RepoName)
	fmt.Printf("Target Branch: %s\n", config.TargetBranch)
	fmt.Printf("Min Coverage: %d%%\n", config.MinCoverage)
	fmt.Printf("Config File: %s\n", config.ConfigFile)
	fmt.Printf("Log Level: %s\n", config.LogLevel)
	fmt.Printf("Log Format: %s\n", config.LogFormat)
	fmt.Printf("Verbose: %t\n", config.Verbose)
	fmt.Printf("Dry Run: %t\n", config.DryRun)
	fmt.Println()
}

func (c *CLI) maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "***" + token[len(token)-4:]
}

// Main entry point for CLI
func main() {
	cli := NewCLI()
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
