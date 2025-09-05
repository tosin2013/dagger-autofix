package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// mock implementations

type mockGitHub struct {
	// Workflow operations
	getWorkflowRunFunc        func(ctx context.Context, runID int64) (*WorkflowRun, error)
	getWorkflowLogsFunc       func(ctx context.Context, runID int64) (*WorkflowLogs, error)
	getFailedWorkflowRunsFunc func(ctx context.Context) ([]*WorkflowRun, error)
	createTestBranchFunc      func(ctx context.Context, branchName string, changes []CodeChange) (func(), error)
	
	// Pull request operations
	createPullRequestFunc    func(ctx context.Context, options *PRCreationOptions) (*PullRequest, error)
	updatePullRequestFunc    func(ctx context.Context, prNumber int, updates *PRUpdateOptions) (*PullRequest, error)
	getPullRequestFunc       func(ctx context.Context, prNumber int) (*PullRequest, error)
	closePullRequestFunc     func(ctx context.Context, prNumber int) error
	addPullRequestCommentFunc func(ctx context.Context, prNumber int, comment string) error
	
	// Repository information
	repoOwner string
	repoName  string
}

func (m *mockGitHub) GetWorkflowRun(ctx context.Context, runID int64) (*WorkflowRun, error) {
	if m.getWorkflowRunFunc != nil {
		return m.getWorkflowRunFunc(ctx, runID)
	}
	return nil, nil
}

func (m *mockGitHub) GetWorkflowLogs(ctx context.Context, runID int64) (*WorkflowLogs, error) {
	if m.getWorkflowLogsFunc != nil {
		return m.getWorkflowLogsFunc(ctx, runID)
	}
	return nil, nil
}

func (m *mockGitHub) GetFailedWorkflowRuns(ctx context.Context) ([]*WorkflowRun, error) {
	if m.getFailedWorkflowRunsFunc != nil {
		return m.getFailedWorkflowRunsFunc(ctx)
	}
	return nil, nil
}

func (m *mockGitHub) CreateTestBranch(ctx context.Context, branchName string, changes []CodeChange) (func(), error) {
	if m.createTestBranchFunc != nil {
		return m.createTestBranchFunc(ctx, branchName, changes)
	}
	return func() {}, nil
}

// Pull request operations
func (m *mockGitHub) CreatePullRequest(ctx context.Context, options *PRCreationOptions) (*PullRequest, error) {
	if m.createPullRequestFunc != nil {
		return m.createPullRequestFunc(ctx, options)
	}
	return nil, nil
}

func (m *mockGitHub) UpdatePullRequest(ctx context.Context, prNumber int, updates *PRUpdateOptions) (*PullRequest, error) {
	if m.updatePullRequestFunc != nil {
		return m.updatePullRequestFunc(ctx, prNumber, updates)
	}
	return nil, nil
}

func (m *mockGitHub) GetPullRequest(ctx context.Context, prNumber int) (*PullRequest, error) {
	if m.getPullRequestFunc != nil {
		return m.getPullRequestFunc(ctx, prNumber)
	}
	return nil, nil
}

func (m *mockGitHub) ClosePullRequest(ctx context.Context, prNumber int) error {
	if m.closePullRequestFunc != nil {
		return m.closePullRequestFunc(ctx, prNumber)
	}
	return nil
}

func (m *mockGitHub) AddPullRequestComment(ctx context.Context, prNumber int, comment string) error {
	if m.addPullRequestCommentFunc != nil {
		return m.addPullRequestCommentFunc(ctx, prNumber, comment)
	}
	return nil
}

// Repository information
func (m *mockGitHub) GetRepoOwner() string {
	return m.repoOwner
}

func (m *mockGitHub) GetRepoName() string {
	return m.repoName
}

type mockFailureAnalysisEngine struct {
	analyzeFunc       func(ctx context.Context, fc FailureContext) (*FailureAnalysisResult, error)
	generateFixesFunc func(ctx context.Context, analysis *FailureAnalysisResult) ([]*ProposedFix, error)
}

func (m *mockFailureAnalysisEngine) AnalyzeFailure(ctx context.Context, fc FailureContext) (*FailureAnalysisResult, error) {
	if m.analyzeFunc != nil {
		return m.analyzeFunc(ctx, fc)
	}
	return nil, nil
}

func (m *mockFailureAnalysisEngine) GenerateFixes(ctx context.Context, analysis *FailureAnalysisResult) ([]*ProposedFix, error) {
	if m.generateFixesFunc != nil {
		return m.generateFixesFunc(ctx, analysis)
	}
	return nil, nil
}

type mockTestEngine struct {
	runTestsFunc func(ctx context.Context, owner, repo, branch string) (*TestResult, error)
}

func (m *mockTestEngine) RunTests(ctx context.Context, owner, repo, branch string) (*TestResult, error) {
	if m.runTestsFunc != nil {
		return m.runTestsFunc(ctx, owner, repo, branch)
	}
	return nil, nil
}

type mockPullRequestEngine struct {
	createFunc func(ctx context.Context, analysis *FailureAnalysisResult, fix *FixValidationResult) (*PullRequest, error)
}

func (m *mockPullRequestEngine) CreateFixPR(ctx context.Context, analysis *FailureAnalysisResult, fix *FixValidationResult) (*PullRequest, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, analysis, fix)
	}
	return nil, nil
}

// Tests

func TestWorkflowMonitorWorkflows(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := make(chan struct{})
	gh := &mockGitHub{
		getFailedWorkflowRunsFunc: func(ctx context.Context) ([]*WorkflowRun, error) {
			close(called)
			return []*WorkflowRun{}, nil
		},
	}

	m := &DaggerAutofix{githubClient: gh, logger: logrus.New()}

	oldTicker := newTicker
	newTicker = func(d time.Duration) *time.Ticker {
		return time.NewTicker(time.Millisecond)
	}
	defer func() { newTicker = oldTicker }()

	errCh := make(chan error)
	go func() {
		errCh <- m.MonitorWorkflows(ctx)
	}()

	select {
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for checkForFailures")
	case <-called:
		cancel()
	}

	select {
	case <-time.After(100 * time.Millisecond):
		t.Fatal("monitor did not exit")
	case err := <-errCh:
		assert.Equal(t, context.Canceled, err)
	}
}

func TestWorkflowAnalyzeFailure(t *testing.T) {
	ctx := context.Background()
	run := &WorkflowRun{ID: 123}
	logs := &WorkflowLogs{ErrorLines: []string{"error"}}
	expected := &FailureAnalysisResult{Classification: FailureClassification{Type: BuildFailure, Confidence: 0.9}}

	gh := &mockGitHub{
		getWorkflowRunFunc: func(ctx context.Context, runID int64) (*WorkflowRun, error) {
			assert.Equal(t, int64(123), runID)
			return run, nil
		},
		getWorkflowLogsFunc: func(ctx context.Context, runID int64) (*WorkflowLogs, error) {
			assert.Equal(t, int64(123), runID)
			return logs, nil
		},
	}

	fe := &mockFailureAnalysisEngine{
		analyzeFunc: func(ctx context.Context, fc FailureContext) (*FailureAnalysisResult, error) {
			assert.Equal(t, run, fc.WorkflowRun)
			assert.Equal(t, logs, fc.Logs)
			assert.Equal(t, "owner", fc.Repository.Owner)
			assert.Equal(t, "repo", fc.Repository.Name)
			return expected, nil
		},
	}

	m := &DaggerAutofix{githubClient: gh, failureEngine: fe, RepoOwner: "owner", RepoName: "repo", logger: logrus.New()}

	res, err := m.AnalyzeFailure(ctx, 123)
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}

func TestWorkflowAutoFix(t *testing.T) {
	ctx := context.Background()

	t.Run("successful autofix", func(t *testing.T) {
		calls := []string{}
		gh := &mockGitHub{
			getWorkflowRunFunc: func(ctx context.Context, runID int64) (*WorkflowRun, error) {
				return &WorkflowRun{ID: runID}, nil
			},
			getWorkflowLogsFunc: func(ctx context.Context, runID int64) (*WorkflowLogs, error) {
				return &WorkflowLogs{}, nil
			},
			createTestBranchFunc: func(ctx context.Context, branch string, changes []CodeChange) (func(), error) {
				calls = append(calls, "validate")
				return func() {}, nil
			},
		}

		fe := &mockFailureAnalysisEngine{
			analyzeFunc: func(ctx context.Context, fc FailureContext) (*FailureAnalysisResult, error) {
				calls = append(calls, "analyze")
				return &FailureAnalysisResult{Classification: FailureClassification{Type: BuildFailure, Confidence: 0.9}}, nil
			},
			generateFixesFunc: func(ctx context.Context, analysis *FailureAnalysisResult) ([]*ProposedFix, error) {
				calls = append(calls, "generate")
				return []*ProposedFix{{ID: "1", Confidence: 0.8}}, nil
			},
		}

		te := &mockTestEngine{
			runTestsFunc: func(ctx context.Context, owner, repo, branch string) (*TestResult, error) {
				return &TestResult{Success: true, Coverage: 90}, nil
			},
		}

		pr := &mockPullRequestEngine{
			createFunc: func(ctx context.Context, analysis *FailureAnalysisResult, fix *FixValidationResult) (*PullRequest, error) {
				calls = append(calls, "pr")
				return &PullRequest{Number: 1, URL: "http://example"}, nil
			},
		}

		m := &DaggerAutofix{
			githubClient:  gh,
			failureEngine: fe,
			testEngine:    te,
			prEngine:      pr,
			llmClient:     &LLMClient{},
			logger:        logrus.New(),
			RepoOwner:     "o",
			RepoName:      "r",
			MinCoverage:   80,
		}

		res, err := m.AutoFix(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, []string{"analyze", "generate", "validate", "pr"}, calls)
	})

	t.Run("no valid fix", func(t *testing.T) {
		calls := []string{}
		gh := &mockGitHub{
			getWorkflowRunFunc: func(ctx context.Context, runID int64) (*WorkflowRun, error) {
				return &WorkflowRun{ID: runID}, nil
			},
			getWorkflowLogsFunc: func(ctx context.Context, runID int64) (*WorkflowLogs, error) {
				return &WorkflowLogs{}, nil
			},
			createTestBranchFunc: func(ctx context.Context, branch string, changes []CodeChange) (func(), error) {
				calls = append(calls, "validate")
				return func() {}, nil
			},
		}

		fe := &mockFailureAnalysisEngine{
			analyzeFunc: func(ctx context.Context, fc FailureContext) (*FailureAnalysisResult, error) {
				calls = append(calls, "analyze")
				return &FailureAnalysisResult{Classification: FailureClassification{Type: BuildFailure, Confidence: 0.9}}, nil
			},
			generateFixesFunc: func(ctx context.Context, analysis *FailureAnalysisResult) ([]*ProposedFix, error) {
				calls = append(calls, "generate")
				return []*ProposedFix{{ID: "1", Confidence: 0.8}}, nil
			},
		}

		te := &mockTestEngine{
			runTestsFunc: func(ctx context.Context, owner, repo, branch string) (*TestResult, error) {
				return nil, errors.New("tests failed")
			},
		}

		pr := &mockPullRequestEngine{}

		m := &DaggerAutofix{
			githubClient:  gh,
			failureEngine: fe,
			testEngine:    te,
			prEngine:      pr,
			llmClient:     &LLMClient{},
			logger:        logrus.New(),
			RepoOwner:     "o",
			RepoName:      "r",
			MinCoverage:   80,
		}

		res, err := m.AutoFix(ctx, 1)
		assert.Nil(t, res)
		assert.Error(t, err)
		assert.Equal(t, []string{"analyze", "generate", "validate"}, calls)
	})
}

func TestWorkflowValidateFix(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		cleanupCalled := false
		branchName := ""
		gh := &mockGitHub{
			createTestBranchFunc: func(ctx context.Context, branch string, changes []CodeChange) (func(), error) {
				branchName = branch
				return func() { cleanupCalled = true }, nil
			},
		}

		te := &mockTestEngine{
			runTestsFunc: func(ctx context.Context, owner, repo, branch string) (*TestResult, error) {
				assert.Equal(t, branchName, branch)
				return &TestResult{Success: true, Coverage: 90}, nil
			},
		}

		m := &DaggerAutofix{
			githubClient: gh,
			testEngine:   te,
			logger:       logrus.New(),
			RepoOwner:    "o",
			RepoName:     "r",
			MinCoverage:  80,
		}

		fix := &ProposedFix{ID: "1"}
		res, err := m.ValidateFix(ctx, fix)
		assert.NoError(t, err)
		assert.True(t, res.Valid)
		assert.True(t, cleanupCalled)
	})

	t.Run("invalid due to coverage", func(t *testing.T) {
		cleanupCalled := false
		gh := &mockGitHub{
			createTestBranchFunc: func(ctx context.Context, branch string, changes []CodeChange) (func(), error) {
				return func() { cleanupCalled = true }, nil
			},
		}

		te := &mockTestEngine{
			runTestsFunc: func(ctx context.Context, owner, repo, branch string) (*TestResult, error) {
				return &TestResult{Success: true, Coverage: 50}, nil
			},
		}

		m := &DaggerAutofix{
			githubClient: gh,
			testEngine:   te,
			logger:       logrus.New(),
			RepoOwner:    "o",
			RepoName:     "r",
			MinCoverage:  80,
		}

		fix := &ProposedFix{ID: "1"}
		res, err := m.ValidateFix(ctx, fix)
		assert.NoError(t, err)
		assert.False(t, res.Valid)
		assert.True(t, cleanupCalled)
	})

	t.Run("branch creation error", func(t *testing.T) {
		gh := &mockGitHub{
			createTestBranchFunc: func(ctx context.Context, branch string, changes []CodeChange) (func(), error) {
				return nil, errors.New("branch failure")
			},
		}

		m := &DaggerAutofix{
			githubClient: gh,
			testEngine:   &mockTestEngine{},
			logger:       logrus.New(),
		}

		_, err := m.ValidateFix(ctx, &ProposedFix{ID: "1"})
		assert.Error(t, err)
	})
}
