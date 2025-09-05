package main

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCheckForFailures(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	runs := []*WorkflowRun{
		{ID: 1, CreatedAt: now, Event: "push"},
		{ID: 2, CreatedAt: now.Add(-25 * time.Hour), Event: "push"},
		{ID: 3, CreatedAt: now, Event: "workflow_dispatch"},
		{ID: 4, CreatedAt: now, Event: "push"},
	}

	autoFixCalled := make(chan int64, 4)

	gh := &mockGitHub{
		getFailedWorkflowRunsFunc: func(ctx context.Context) ([]*WorkflowRun, error) {
			return runs, nil
		},
		getWorkflowRunFunc: func(ctx context.Context, runID int64) (*WorkflowRun, error) {
			return &WorkflowRun{ID: runID}, nil
		},
		getWorkflowLogsFunc: func(ctx context.Context, runID int64) (*WorkflowLogs, error) {
			return &WorkflowLogs{}, nil
		},
	}

	fe := &mockFailureAnalysisEngine{
		analyzeFunc: func(ctx context.Context, fc FailureContext) (*FailureAnalysisResult, error) {
			autoFixCalled <- fc.WorkflowRun.ID
			return &FailureAnalysisResult{Classification: FailureClassification{Type: BuildFailure, Confidence: 0.9}}, nil
		},
		generateFixesFunc: func(ctx context.Context, analysis *FailureAnalysisResult) ([]*ProposedFix, error) {
			return []*ProposedFix{{ID: "fix1", Confidence: 0.8}}, nil
		},
	}

	te := &mockTestEngine{
		runTestsFunc: func(ctx context.Context, owner, repo, branch string) (*TestResult, error) {
			return &TestResult{Success: true, Coverage: 100}, nil
		},
	}

	pr := &mockPullRequestEngine{
		createFunc: func(ctx context.Context, analysis *FailureAnalysisResult, fix *FixValidationResult) (*PullRequest, error) {
			return &PullRequest{Number: 1, URL: "url"}, nil
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
		processedRuns: map[int64]struct{}{4: {}},
	}

	err := m.checkForFailures(ctx)
	assert.NoError(t, err)

	select {
	case id := <-autoFixCalled:
		assert.Equal(t, int64(1), id)
	case <-time.After(time.Second):
		t.Fatal("AutoFix not invoked")
	}

	select {
	case id := <-autoFixCalled:
		t.Fatalf("unexpected AutoFix call for run %d", id)
	case <-time.After(100 * time.Millisecond):
		// no additional calls
	}
}

func TestShouldProcessRun(t *testing.T) {
	now := time.Now()
	m := &DaggerAutofix{
		logger:        logrus.New(),
		processedRuns: map[int64]struct{}{1: {}},
	}

	assert.False(t, m.shouldProcessRun(&WorkflowRun{ID: 1, CreatedAt: now, Event: "push"}))
	assert.False(t, m.shouldProcessRun(&WorkflowRun{ID: 2, CreatedAt: now.Add(-25 * time.Hour), Event: "push"}))
	assert.False(t, m.shouldProcessRun(&WorkflowRun{ID: 3, CreatedAt: now, Event: "workflow_dispatch"}))
	assert.True(t, m.shouldProcessRun(&WorkflowRun{ID: 4, CreatedAt: now, Event: "push"}))

	_, exists := m.processedRuns[4]
	assert.True(t, exists)
}

func TestSelectBestFix(t *testing.T) {
	m := &DaggerAutofix{}

	t.Run("selects highest confidence valid fix", func(t *testing.T) {
		fixes := []*FixValidationResult{
			{Fix: &ProposedFix{ID: "1", Confidence: 0.5}, Valid: true},
			{Fix: &ProposedFix{ID: "2", Confidence: 0.9}, Valid: true},
			{Fix: &ProposedFix{ID: "3", Confidence: 1.0}, Valid: false},
		}
		best := m.selectBestFix(fixes)
		if assert.NotNil(t, best) {
			assert.Equal(t, "2", best.Fix.ID)
		}
	})

	t.Run("returns nil when no valid fixes", func(t *testing.T) {
		fixes := []*FixValidationResult{
			{Fix: &ProposedFix{ID: "1", Confidence: 0.5}, Valid: false},
		}
		best := m.selectBestFix(fixes)
		assert.Nil(t, best)
	})
}
