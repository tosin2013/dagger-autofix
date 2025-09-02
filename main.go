// Package main provides a comprehensive GitHub Actions auto-fix agent using Dagger.io
// with multi-LLM support for intelligent failure analysis and automated resolution.
package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dagger.io/dagger"
	"github.com/sirupsen/logrus"
)

// Interfaces for dependency injection
type GitHubClient interface {
	GetWorkflowRun(ctx context.Context, runID int64) (*WorkflowRun, error)
	GetWorkflowLogs(ctx context.Context, runID int64) (*WorkflowLogs, error)
	GetFailedWorkflowRuns(ctx context.Context) ([]*WorkflowRun, error)
	CreateTestBranch(ctx context.Context, branchName string, changes []CodeChange) (func(), error)
}

type FailureEngine interface {
	AnalyzeFailure(ctx context.Context, fc FailureContext) (*FailureAnalysisResult, error)
	GenerateFixes(ctx context.Context, analysis *FailureAnalysisResult) ([]*ProposedFix, error)
}

type TestRunner interface {
	RunTests(ctx context.Context, owner, repo, branch string) (*TestResult, error)
}

type PREngine interface {
	CreateFixPR(ctx context.Context, analysis *FailureAnalysisResult, fix *FixValidationResult) (*PullRequest, error)
}

// DaggerAutofix represents the main Dagger module for GitHub Actions auto-fixing
type DaggerAutofix struct {
	// Source directory for the project
	Source *dagger.Directory

	// Configuration
	GitHubToken  *dagger.Secret
	LLMProvider  LLMProvider
	LLMAPIKey    *dagger.Secret
	RepoOwner    string
	RepoName     string
	TargetBranch string
	MinCoverage  int

	// Internal state
	logger        *logrus.Logger
	githubClient  GitHubClient
	llmClient     *LLMClient
	failureEngine FailureEngine
	testEngine    TestRunner
	prEngine      PREngine
}

var (
	newGitHubIntegration     = NewGitHubIntegration
	newLLMClient             = NewLLMClient
	newFailureAnalysisEngine = NewFailureAnalysisEngine
	newTestEngine            = NewTestEngine
	newPullRequestEngine     = NewPullRequestEngine
	newTicker                = time.NewTicker
)

// New creates a new DaggerAutofix instance with default configuration
// Optionally accepts a source directory; if none provided, uses current directory
func New(source ...*dagger.Directory) *DaggerAutofix {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	var sourceDir *dagger.Directory
	if len(source) > 0 && source[0] != nil {
		sourceDir = source[0]
	} else if dag != nil {
		// Use host directory as default source (only if dag is available)
		sourceDir = dag.Host().Directory(".")
	}

	return &DaggerAutofix{
		Source:       sourceDir,
		LLMProvider:  OpenAI, // default provider
		TargetBranch: "main",
		MinCoverage:  85,
		logger:       logger,
	}
}

// WithSource configures the source directory
func (m *DaggerAutofix) WithSource(source *dagger.Directory) *DaggerAutofix {
	m.Source = source
	return m
}

// WithGitHubToken configures GitHub authentication
func (m *DaggerAutofix) WithGitHubToken(token *dagger.Secret) *DaggerAutofix {
	m.GitHubToken = token
	return m
}

// WithLLMProvider configures the LLM provider and API key
func (m *DaggerAutofix) WithLLMProvider(provider string, apiKey *dagger.Secret) *DaggerAutofix {
	m.LLMProvider = LLMProvider(strings.ToLower(provider))
	m.LLMAPIKey = apiKey
	return m
}

// WithRepository configures the target GitHub repository
func (m *DaggerAutofix) WithRepository(owner, name string) *DaggerAutofix {
	m.RepoOwner = owner
	m.RepoName = name
	return m
}

// WithTargetBranch configures the target branch (default: main)
func (m *DaggerAutofix) WithTargetBranch(branch string) *DaggerAutofix {
	m.TargetBranch = branch
	return m
}

// WithMinCoverage configures minimum test coverage requirement (default: 85%)
func (m *DaggerAutofix) WithMinCoverage(coverage int) *DaggerAutofix {
	m.MinCoverage = coverage
	return m
}

// Initialize sets up all internal components
func (m *DaggerAutofix) Initialize(ctx context.Context) (*DaggerAutofix, error) {
	if err := m.validateConfiguration(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Initialize GitHub client
	ghClient, err := newGitHubIntegration(ctx, m.GitHubToken, m.RepoOwner, m.RepoName)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GitHub client: %w", err)
	}
	m.githubClient = ghClient

	// Initialize LLM client
	llmClient, err := newLLMClient(ctx, m.LLMProvider, m.LLMAPIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM client: %w", err)
	}
	m.llmClient = llmClient

	// Initialize failure analysis engine
	m.failureEngine = newFailureAnalysisEngine(m.llmClient, m.logger)

	// Initialize test engine
	m.testEngine = newTestEngine(m.MinCoverage, m.logger)

	// Initialize PR engine
	m.prEngine = newPullRequestEngine(ghClient, m.logger)

	m.logger.Info("DaggerAutofix initialized successfully")
	return m, nil
}

// MonitorWorkflows continuously monitors GitHub Actions workflows for failures
func (m *DaggerAutofix) MonitorWorkflows(ctx context.Context) error {
	if m.githubClient == nil {
		return fmt.Errorf("module not initialized, call Initialize first")
	}

	m.logger.Info("Starting workflow monitoring")

	ticker := newTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Monitoring stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := m.checkForFailures(ctx); err != nil {
				m.logger.WithError(err).Error("Failed to check for workflow failures")
			}
		}
	}
}

// AnalyzeFailure analyzes a specific workflow failure and generates fixes
func (m *DaggerAutofix) AnalyzeFailure(ctx context.Context, runID int64) (*FailureAnalysisResult, error) {
	if m.failureEngine == nil {
		return nil, fmt.Errorf("module not initialized, call Initialize first")
	}

	m.logger.WithField("run_id", runID).Info("Analyzing workflow failure")

	// Get workflow run details
	workflowRun, err := m.githubClient.GetWorkflowRun(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow run: %w", err)
	}

	// Get failure logs
	logs, err := m.githubClient.GetWorkflowLogs(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow logs: %w", err)
	}

	// Analyze failure with LLM
	analysis, err := m.failureEngine.AnalyzeFailure(ctx, FailureContext{
		WorkflowRun: workflowRun,
		Logs:        logs,
		Repository: RepositoryContext{
			Owner: m.RepoOwner,
			Name:  m.RepoName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failure analysis failed: %w", err)
	}

	m.logger.WithFields(logrus.Fields{
		"failure_type": analysis.Classification.Type,
		"confidence":   analysis.Classification.Confidence,
	}).Info("Failure analysis completed")

	return analysis, nil
}

// AutoFix performs end-to-end automated fixing of a workflow failure
func (m *DaggerAutofix) AutoFix(ctx context.Context, runID int64) (*AutoFixResult, error) {
	if err := m.ensureInitialized(); err != nil {
		return nil, err
	}

	m.logger.WithField("run_id", runID).Info("Starting automated fix process")

	// Step 1: Analyze failure
	analysis, err := m.AnalyzeFailure(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("failure analysis failed: %w", err)
	}

	// Step 2: Generate fixes
	fixes, err := m.failureEngine.GenerateFixes(ctx, analysis)
	if err != nil {
		return nil, fmt.Errorf("fix generation failed: %w", err)
	}

	// Step 3: Validate fixes
	validationResults := make([]*FixValidationResult, 0, len(fixes))
	for _, fix := range fixes {
		validation, err := m.ValidateFix(ctx, fix)
		if err != nil {
			m.logger.WithError(err).Warn("Fix validation failed, skipping")
			continue
		}
		validationResults = append(validationResults, validation)
	}

	if len(validationResults) == 0 {
		return nil, fmt.Errorf("no valid fixes generated")
	}

	// Step 4: Select best fix (highest confidence + passes tests)
	bestFix := m.selectBestFix(validationResults)

	// Step 5: Create pull request
	pr, err := m.prEngine.CreateFixPR(ctx, analysis, bestFix)
	if err != nil {
		return nil, fmt.Errorf("PR creation failed: %w", err)
	}

	result := &AutoFixResult{
		Analysis:    analysis,
		Fix:         bestFix,
		PullRequest: pr,
		Timestamp:   time.Now(),
	}

	m.logger.WithFields(logrus.Fields{
		"pr_number": pr.Number,
		"pr_url":    pr.URL,
	}).Info("Automated fix completed successfully")

	return result, nil
}

// ValidateFix validates a proposed fix by running tests and checking coverage
func (m *DaggerAutofix) ValidateFix(ctx context.Context, fix *ProposedFix) (*FixValidationResult, error) {
	if m.testEngine == nil {
		return nil, fmt.Errorf("module not initialized, call Initialize first")
	}

	m.logger.WithField("fix_id", fix.ID).Info("Validating proposed fix")

	// Create temporary branch with fix
	testBranch := fmt.Sprintf("autofix-test-%s-%d", fix.ID, time.Now().Unix())
	cleanup, err := m.githubClient.CreateTestBranch(ctx, testBranch, fix.Changes)
	if err != nil {
		return nil, fmt.Errorf("failed to create test branch: %w", err)
	}
	defer cleanup()

	// Run tests
	testResult, err := m.testEngine.RunTests(ctx, m.RepoOwner, m.RepoName, testBranch)
	if err != nil {
		return nil, fmt.Errorf("test execution failed: %w", err)
	}

	validation := &FixValidationResult{
		Fix:        fix,
		TestResult: testResult,
		Valid:      testResult.Success && testResult.Coverage >= float64(m.MinCoverage),
		Timestamp:  time.Now(),
	}

	m.logger.WithFields(logrus.Fields{
		"tests_passed": testResult.Success,
		"coverage":     testResult.Coverage,
		"valid":        validation.Valid,
	}).Info("Fix validation completed")

	return validation, nil
}

// GetMetrics returns operational metrics for monitoring
func (m *DaggerAutofix) GetMetrics(ctx context.Context) (*OperationalMetrics, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.githubClient == nil {
		return nil, fmt.Errorf("module not initialized")
	}
	return &OperationalMetrics{
		TotalFailuresDetected: 0, // TODO: implement metrics collection
		SuccessfulFixes:       0,
		FailedFixes:           0,
		AverageFixTime:        0,
		TestCoverage:          float64(m.MinCoverage),
	}, nil
}

// CLI returns a CLI container for manual execution
func (m *DaggerAutofix) CLI() (container *dagger.Container) {
	// Use defer/recover to handle nil pointer when not in Dagger context
	defer func() {
		if r := recover(); r != nil {
			// This happens when running outside of Dagger context (e.g., unit tests)
			// Return nil in this case
			container = nil
		}
	}()

	// This will panic if dag is nil or not properly initialized
	// which happens when running outside of Dagger context
	container = dag.Container().
		From("golang:1.21-alpine").
		WithExec([]string{"apk", "add", "git", "curl"}).
		WithWorkdir("/app").
		WithDirectory("/app", m.Source).
		WithExec([]string{"go", "mod", "download"}).
		WithExec([]string{"go", "build", "-o", "github-autofix", "."})

	return container
}

// Helper methods

func (m *DaggerAutofix) validateConfiguration() error {
	if m.GitHubToken == nil {
		return fmt.Errorf("GitHub token is required")
	}
	if m.RepoOwner == "" || m.RepoName == "" {
		return fmt.Errorf("repository owner and name are required")
	}
	if m.LLMAPIKey == nil {
		return fmt.Errorf("LLM API key is required")
	}
	return nil
}

func (m *DaggerAutofix) ensureInitialized() error {
	if m.githubClient == nil || m.llmClient == nil || m.failureEngine == nil {
		return fmt.Errorf("module not initialized, call Initialize first")
	}
	return nil
}

func (m *DaggerAutofix) checkForFailures(ctx context.Context) error {
	failedRuns, err := m.githubClient.GetFailedWorkflowRuns(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workflow runs: %w", err)
	}

	for _, run := range failedRuns {
		if !m.shouldProcessRun(run) {
			continue
		}

		go func(runID int64) {
			if _, err := m.AutoFix(ctx, runID); err != nil {
				m.logger.WithError(err).WithField("run_id", runID).Error("Auto-fix failed")
			}
		}(run.ID)
	}

	return nil
}

func (m *DaggerAutofix) shouldProcessRun(run *WorkflowRun) bool {
	// Skip if already processed
	// Skip if too old
	// Skip if manual trigger
	return true // Simplified for now
}

func (m *DaggerAutofix) selectBestFix(validations []*FixValidationResult) *FixValidationResult {
	var best *FixValidationResult
	for _, validation := range validations {
		if !validation.Valid {
			continue
		}
		if best == nil || validation.Fix.Confidence > best.Fix.Confidence {
			best = validation
		}
	}
	return best
}
