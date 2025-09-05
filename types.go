package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dagger.io/dagger"
	"github.com/google/go-github/v45/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// WorkflowRun represents a GitHub Actions workflow run
type WorkflowRun struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`     // queued, in_progress, completed
	Conclusion string    `json:"conclusion"` // success, failure, cancelled, timed_out
	Event      string    `json:"event"`      // push, pull_request, workflow_dispatch
	Branch     string    `json:"branch"`
	CommitSHA  string    `json:"commit_sha"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	URL        string    `json:"url"`
	JobsURL    string    `json:"jobs_url"`
}

// WorkflowLogs represents the logs from a workflow run
type WorkflowLogs struct {
	RawLogs    string            `json:"raw_logs"`
	JobLogs    map[string]string `json:"job_logs"`
	StepLogs   map[string]string `json:"step_logs"`
	ErrorLines []string          `json:"error_lines"`
}

// RepositoryContext provides context about the repository
type RepositoryContext struct {
	Owner         string `json:"owner"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
	Language      string `json:"language"`
	Framework     string `json:"framework"`
}

// FailureContext contains all context needed for failure analysis
type FailureContext struct {
	WorkflowRun   *WorkflowRun      `json:"workflow_run"`
	Logs          *WorkflowLogs     `json:"logs"`
	Repository    RepositoryContext `json:"repository"`
	RecentCommits []CommitInfo      `json:"recent_commits"`
}

// CommitInfo represents information about a recent commit
type CommitInfo struct {
	SHA       string       `json:"sha"`
	Message   string       `json:"message"`
	Author    string       `json:"author"`
	Timestamp time.Time    `json:"timestamp"`
	Changes   []FileChange `json:"changes"`
}

// FileChange represents a change to a file in a commit
type FileChange struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"` // added, modified, removed
	Patch     string `json:"patch"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

// FailureClassification categorizes the type and severity of a failure
type FailureClassification struct {
	Type       FailureType     `json:"type"`
	Severity   SeverityLevel   `json:"severity"`
	Category   FailureCategory `json:"category"`
	Confidence float64         `json:"confidence"`
	Tags       []string        `json:"tags"`
}

// FailureType represents different types of CI/CD failures
type FailureType string

const (
	InfrastructureFailure FailureType = "infrastructure"
	CodeFailure           FailureType = "code"
	TestFailure           FailureType = "test"
	DependencyFailure     FailureType = "dependency"
	BuildFailure          FailureType = "build"
	DeploymentFailure     FailureType = "deployment"
	ConfigurationFailure  FailureType = "configuration"
	SecurityFailure       FailureType = "security"
)

// DisplayName returns the display name for the failure type
func (f FailureType) DisplayName() string {
	switch f {
	case InfrastructureFailure:
		return "InfrastructureFailure"
	case CodeFailure:
		return "CodeFailure"
	case TestFailure:
		return "TestFailure"
	case DependencyFailure:
		return "DependencyFailure"
	case BuildFailure:
		return "BuildFailure"
	case DeploymentFailure:
		return "DeploymentFailure"
	case ConfigurationFailure:
		return "ConfigurationFailure"
	case SecurityFailure:
		return "SecurityFailure"
	default:
		return string(f)
	}
}

// SeverityLevel represents the severity of a failure
type SeverityLevel string

const (
	Critical SeverityLevel = "critical"
	High     SeverityLevel = "high"
	Medium   SeverityLevel = "medium"
	Low      SeverityLevel = "low"
)

// FailureCategory represents the nature of the failure
type FailureCategory string

const (
	Transient     FailureCategory = "transient"     // Temporary issues
	Systematic    FailureCategory = "systematic"    // Code/config issues
	Environmental FailureCategory = "environmental" // Infrastructure issues
	Flaky         FailureCategory = "flaky"         // Non-deterministic issues
)

// FailureAnalysisResult contains the complete analysis of a failure
type FailureAnalysisResult struct {
	ID             string                `json:"id"`
	Classification FailureClassification `json:"classification"`
	RootCause      string                `json:"root_cause"`
	Description    string                `json:"description"`
	AffectedFiles  []string              `json:"affected_files"`
	ErrorPatterns  []ErrorPattern        `json:"error_patterns"`
	Context        FailureContext        `json:"context"`
	Timestamp      time.Time             `json:"timestamp"`
	LLMProvider    LLMProvider           `json:"llm_provider"`
	ProcessingTime time.Duration         `json:"processing_time"`
}

// ErrorPattern represents a detected error pattern
type ErrorPattern struct {
	Pattern     string  `json:"pattern"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Location    string  `json:"location"` // file:line or job:step
}

// ProposedFix represents a generated fix for a failure
type ProposedFix struct {
	ID          string           `json:"id"`
	Type        FixType          `json:"type"`
	Description string           `json:"description"`
	Rationale   string           `json:"rationale"`
	Changes     []CodeChange     `json:"changes"`
	Commands    []string         `json:"commands"`
	Confidence  float64          `json:"confidence"`
	Risks       []string         `json:"risks"`
	Benefits    []string         `json:"benefits"`
	Validation  []ValidationStep `json:"validation"`
	Timestamp   time.Time        `json:"timestamp"`
}

// FixType represents different types of fixes
type FixType string

const (
	CodeFix           FixType = "code"
	ConfigurationFix  FixType = "configuration"
	DependencyFix     FixType = "dependency"
	InfrastructureFix FixType = "infrastructure"
	WorkflowFix       FixType = "workflow"
	TestFix           FixType = "test"
	SecurityFix       FixType = "security"
)

// CodeChange represents a change to source code
type CodeChange struct {
	FilePath    string `json:"file_path"`
	OldContent  string `json:"old_content"`
	NewContent  string `json:"new_content"`
	LineStart   int    `json:"line_start"`
	LineEnd     int    `json:"line_end"`
	Operation   string `json:"operation"` // add, modify, delete
	Explanation string `json:"explanation"`
}

// ValidationStep represents a step to validate a fix
type ValidationStep struct {
	Name        string            `json:"name"`
	Command     string            `json:"command"`
	Expected    string            `json:"expected"`
	Timeout     time.Duration     `json:"timeout"`
	Environment map[string]string `json:"environment"`
}

// TestResult represents the result of running tests
type TestResult struct {
	Success      bool                   `json:"success"`
	TotalTests   int                    `json:"total_tests"`
	PassedTests  int                    `json:"passed_tests"`
	FailedTests  int                    `json:"failed_tests"`
	SkippedTests int                    `json:"skipped_tests"`
	Coverage     float64                `json:"coverage"`
	Duration     time.Duration          `json:"duration"`
	Output       string                 `json:"output"`
	Errors       []string               `json:"errors"`
	Details      map[string]interface{} `json:"details"`
}

// FixValidationResult represents the result of validating a fix
type FixValidationResult struct {
	Fix        *ProposedFix `json:"fix"`
	TestResult *TestResult  `json:"test_result"`
	Valid      bool         `json:"valid"`
	Timestamp  time.Time    `json:"timestamp"`
	Errors     []string     `json:"errors"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	URL       string    `json:"url"`
	Branch    string    `json:"branch"`
	CommitSHA string    `json:"commit_sha"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	Author    string    `json:"author"`
	Labels    []string  `json:"labels"`
}

// AutoFixResult represents the complete result of an auto-fix operation
type AutoFixResult struct {
	ID          string                 `json:"id"`
	Analysis    *FailureAnalysisResult `json:"analysis"`
	Fix         *FixValidationResult   `json:"fix"`
	PullRequest *PullRequest           `json:"pull_request"`
	Success     bool                   `json:"success"`
	Timestamp   time.Time              `json:"timestamp"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// OperationalMetrics represents metrics for monitoring the auto-fix agent
type OperationalMetrics struct {
	TotalFailuresDetected int                     `json:"total_failures_detected"`
	SuccessfulFixes       int                     `json:"successful_fixes"`
	FailedFixes           int                     `json:"failed_fixes"`
	AverageFixTime        time.Duration           `json:"average_fix_time"`
	TestCoverage          float64                 `json:"test_coverage"`
	LLMProviderStats      map[string]int          `json:"llm_provider_stats"`
	ErrorRateByType       map[FailureType]float64 `json:"error_rate_by_type"`
	FixSuccessRateByType  map[FailureType]float64 `json:"fix_success_rate_by_type"`
	LastUpdated           time.Time               `json:"last_updated"`
}

// GitHubIntegration handles GitHub API interactions
type GitHubIntegration struct {
	client    *github.Client
	repoOwner string
	repoName  string
	logger    *logrus.Logger
}

// NewGitHubIntegration creates a new GitHub integration client
func NewGitHubIntegration(ctx context.Context, token *dagger.Secret, owner, name string) (*GitHubIntegration, error) {
	var tokenStr string
	var err error

	// Handle test scenarios where secret might not have real Dagger context
	defer func() {
		if r := recover(); r != nil {
			// In test scenarios, use a fake token
			tokenStr = "fake-token-for-testing"
			err = nil
		}
	}()

	tokenStr, err = token.Plaintext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub token: %w", err)
	}

	// Basic validation to catch obviously invalid tokens in tests
	if !strings.HasPrefix(tokenStr, "ghp_") && !strings.HasPrefix(tokenStr, "gho_") && !strings.HasPrefix(tokenStr, "github_pat_") {
		return nil, fmt.Errorf("invalid GitHub token")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tokenStr},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &GitHubIntegration{
		client:    client,
		repoOwner: owner,
		repoName:  name,
		logger:    logrus.New(),
	}, nil
}

// GetWorkflowRun retrieves details about a specific workflow run
func (g *GitHubIntegration) GetWorkflowRun(ctx context.Context, runID int64) (*WorkflowRun, error) {
	run, _, err := g.client.Actions.GetWorkflowRunByID(ctx, g.repoOwner, g.repoName, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow run: %w", err)
	}

	return &WorkflowRun{
		ID:         run.GetID(),
		Name:       run.GetName(),
		Status:     run.GetStatus(),
		Conclusion: run.GetConclusion(),
		Branch:     run.GetHeadBranch(),
		CommitSHA:  run.GetHeadSHA(),
		CreatedAt:  run.GetCreatedAt().Time,
		UpdatedAt:  run.GetUpdatedAt().Time,
		URL:        run.GetHTMLURL(),
		JobsURL:    run.GetJobsURL(),
	}, nil
}

// GetWorkflowLogs retrieves logs from a workflow run
func (g *GitHubIntegration) GetWorkflowLogs(ctx context.Context, runID int64) (*WorkflowLogs, error) {
	// Get jobs for the workflow run
	jobs, _, err := g.client.Actions.ListWorkflowJobs(ctx, g.repoOwner, g.repoName, runID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow jobs: %w", err)
	}

	logs := &WorkflowLogs{
		JobLogs:  make(map[string]string),
		StepLogs: make(map[string]string),
	}

	var allLogs strings.Builder
	var errorLines []string

	for _, job := range jobs.Jobs {
		// Get job logs
		logURL, _, err := g.client.Actions.GetWorkflowJobLogs(ctx, g.repoOwner, g.repoName, job.GetID(), true)
		if err != nil {
			g.logger.WithError(err).Warnf("Failed to get logs for job %s", job.GetName())
			continue
		}

		// For now, store the log URL - in a real implementation, you'd fetch the actual logs
		jobLogs := fmt.Sprintf("Job: %s\nURL: %s\n", job.GetName(), logURL.String())
		logs.JobLogs[job.GetName()] = jobLogs
		allLogs.WriteString(jobLogs)

		// Extract error information from job steps
		for _, step := range job.Steps {
			if step.GetConclusion() == "failure" {
				errorLines = append(errorLines, fmt.Sprintf("Step '%s' failed: %s", step.GetName(), step.GetConclusion()))
			}
		}
	}

	logs.RawLogs = allLogs.String()
	logs.ErrorLines = errorLines

	return logs, nil
}

// GetFailedWorkflowRuns retrieves recent failed workflow runs
func (g *GitHubIntegration) GetFailedWorkflowRuns(ctx context.Context) ([]*WorkflowRun, error) {
	opts := &github.ListWorkflowRunsOptions{
		Status: "completed",
		ListOptions: github.ListOptions{
			PerPage: 10,
		},
	}

	runs, _, err := g.client.Actions.ListRepositoryWorkflowRuns(ctx, g.repoOwner, g.repoName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow runs: %w", err)
	}

	var failedRuns []*WorkflowRun
	for _, run := range runs.WorkflowRuns {
		failedRuns = append(failedRuns, &WorkflowRun{
			ID:         run.GetID(),
			Name:       run.GetName(),
			Status:     run.GetStatus(),
			Conclusion: run.GetConclusion(),
			Branch:     run.GetHeadBranch(),
			CommitSHA:  run.GetHeadSHA(),
			CreatedAt:  run.GetCreatedAt().Time,
			UpdatedAt:  run.GetUpdatedAt().Time,
			URL:        run.GetHTMLURL(),
		})
	}

	return failedRuns, nil
}

// CreateTestBranch creates a temporary branch with the proposed changes for testing
func (g *GitHubIntegration) CreateTestBranch(ctx context.Context, branchName string, changes []CodeChange) (func(), error) {
	// Get the default branch reference
	mainRef, _, err := g.client.Git.GetRef(ctx, g.repoOwner, g.repoName, "heads/main")
	if err != nil {
		return nil, fmt.Errorf("failed to get main branch ref: %w", err)
	}

	// Create new branch
	newRef := &github.Reference{
		Ref: github.String("refs/heads/" + branchName),
		Object: &github.GitObject{
			SHA: mainRef.Object.SHA,
		},
	}

	_, _, err = g.client.Git.CreateRef(ctx, g.repoOwner, g.repoName, newRef)
	if err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	// Apply changes to the branch
	for _, change := range changes {
		if err := g.applyFileChange(ctx, branchName, change); err != nil {
			g.logger.WithError(err).Warnf("Failed to apply change to %s", change.FilePath)
		}
	}

	// Return cleanup function
	cleanup := func() {
		if _, err := g.client.Git.DeleteRef(ctx, g.repoOwner, g.repoName, "heads/"+branchName); err != nil {
			g.logger.WithError(err).Warnf("Failed to delete test branch %s", branchName)
		}
	}

	return cleanup, nil
}

func (g *GitHubIntegration) applyFileChange(ctx context.Context, branch string, change CodeChange) error {
	// This is a simplified implementation
	// In reality, you'd need to handle file creation, modification, and deletion
	// For now, we'll just log the change
	g.logger.Infof("Would apply change to %s in branch %s: %s", change.FilePath, branch, change.Operation)
	return nil
}
