package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGitHubClient is a mock implementation of GitHubClient interface
type MockGitHubClient struct {
	mock.Mock
}

func (m *MockGitHubClient) GetWorkflowRun(ctx context.Context, runID int64) (*WorkflowRun, error) {
	args := m.Called(ctx, runID)
	if result := args.Get(0); result != nil {
		return result.(*WorkflowRun), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGitHubClient) GetWorkflowLogs(ctx context.Context, runID int64) (*WorkflowLogs, error) {
	args := m.Called(ctx, runID)
	if result := args.Get(0); result != nil {
		return result.(*WorkflowLogs), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGitHubClient) GetFailedWorkflowRuns(ctx context.Context) ([]*WorkflowRun, error) {
	args := m.Called(ctx)
	if result := args.Get(0); result != nil {
		return result.([]*WorkflowRun), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGitHubClient) CreateTestBranch(ctx context.Context, branchName string, changes []CodeChange) (func(), error) {
	args := m.Called(ctx, branchName, changes)
	if result := args.Get(0); result != nil {
		return result.(func()), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGitHubClient) CreatePullRequest(ctx context.Context, options *PRCreationOptions) (*PullRequest, error) {
	args := m.Called(ctx, options)
	if result := args.Get(0); result != nil {
		return result.(*PullRequest), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGitHubClient) UpdatePullRequest(ctx context.Context, prNumber int, updates *PRUpdateOptions) (*PullRequest, error) {
	args := m.Called(ctx, prNumber, updates)
	if result := args.Get(0); result != nil {
		return result.(*PullRequest), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGitHubClient) GetPullRequest(ctx context.Context, prNumber int) (*PullRequest, error) {
	args := m.Called(ctx, prNumber)
	if result := args.Get(0); result != nil {
		return result.(*PullRequest), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGitHubClient) ClosePullRequest(ctx context.Context, prNumber int) error {
	args := m.Called(ctx, prNumber)
	return args.Error(0)
}

func (m *MockGitHubClient) AddPullRequestComment(ctx context.Context, prNumber int, comment string) error {
	args := m.Called(ctx, prNumber, comment)
	return args.Error(0)
}

func (m *MockGitHubClient) GetRepoOwner() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockGitHubClient) GetRepoName() string {
	args := m.Called()
	return args.String(0)
}

func TestPullRequestEngine_WithMockClient(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*MockGitHubClient)
		operation   func(*PullRequestEngine) error
		expectError bool
	}{
		{
			name: "create fix PR with mock client",
			setupMock: func(m *MockGitHubClient) {
				m.On("CreateTestBranch", mock.Anything, mock.Anything, mock.Anything).Return(func() {}, nil)
				m.On("CreatePullRequest", mock.Anything, mock.MatchedBy(func(opts *PRCreationOptions) bool {
					return strings.HasPrefix(opts.BranchName, "autofix/code_change/") &&
						opts.TargetBranch == "main"
				})).Return(&PullRequest{
					Number: 100,
					Title:  "[AutoFix] Fix test failure",
					URL:    "https://github.com/test/repo/pull/100",
					State:  "open",
				}, nil)
				m.On("AddPullRequestComment", mock.Anything, 100, mock.AnythingOfType("string")).Return(nil)
			},
			operation: func(pr *PullRequestEngine) error {
				analysis := &FailureAnalysisResult{
					Classification: FailureClassification{
				Type:       "test_failure",
				Severity:   "low",
				Confidence: 0.8,
				Category:   "test_failure",
					},
			Context:     FailureContext{},
					RootCause:   "Missing import statement",
				}
				fix := &FixValidationResult{
					Valid: true,
					Fix: &ProposedFix{
						ID: "fix-123",
					Type:        "code_change",
					Confidence:  0.9,
						Description: "Fix missing import",
					},
				}
				result, err := pr.CreateFixPR(context.Background(), analysis, fix)
				if err != nil {
					return err
				}
				if result.Number != 100 {
					return fmt.Errorf("expected PR number 100, got %d", result.Number)
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "update PR with mock client",
			setupMock: func(m *MockGitHubClient) {
				m.On("UpdatePullRequest", mock.Anything, 123, mock.MatchedBy(func(opts *PRUpdateOptions) bool {
					return *opts.Title == "Updated Title" && *opts.Body == "Updated Body"
				})).Return(&PullRequest{
					Number: 123,
					Title:  "Updated Title",
					Body:   "Updated Body",
					State:  "open",
				}, nil)
			},
			operation: func(pr *PullRequestEngine) error {
				updates := &PRCreationOptions{
					Title: "Updated Title",
					Body:  "Updated Body",
				}
				result, err := pr.UpdatePR(context.Background(), 123, updates)
				if err != nil {
					return err
				}
				if result.Title != "Updated Title" {
					return fmt.Errorf("expected updated title, got %s", result.Title)
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "close PR with mock client",
			setupMock: func(m *MockGitHubClient) {
				m.On("AddPullRequestComment", mock.Anything, 456, mock.MatchedBy(func(comment string) bool {
					return comment == "Test completed"
				})).Return(nil)
				m.On("ClosePullRequest", mock.Anything, 456).Return(nil)
			},
			operation: func(pr *PullRequestEngine) error {
				return pr.ClosePR(context.Background(), 456, "Test completed")
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := new(MockGitHubClient)
			tt.setupMock(mockClient)

			// Create logger
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			// Create PR engine with mock client
			prEngine := NewPullRequestEngine(mockClient, logger)

			// Execute operation
			err := tt.operation(prEngine)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			mockClient.AssertExpectations(t)
		})
	}
}

func TestPullRequestEngine_InterfaceCompatibility(t *testing.T) {
	// This test ensures that both GitHubIntegration and MCPGitHubClient
	// can be used with the PullRequestEngine
	
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("with GitHubIntegration", func(t *testing.T) {
		// This would normally use a real GitHubIntegration, but we'll use a mock
		mockClient := new(MockGitHubClient)
		prEngine := NewPullRequestEngine(mockClient, logger)
		assert.NotNil(t, prEngine)
		assert.Equal(t, mockClient, prEngine.githubClient)
	})

	t.Run("with MCPGitHubClient", func(t *testing.T) {
		// This would normally use a real MCPGitHubClient, but we'll use a mock
		mockClient := new(MockGitHubClient)
		prEngine := NewPullRequestEngine(mockClient, logger)
		assert.NotNil(t, prEngine)
		assert.Equal(t, mockClient, prEngine.githubClient)
	})
}


func TestPullRequestEngine_HandleErrors(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("create PR error handling", func(t *testing.T) {
		mockClient := new(MockGitHubClient)
		mockClient.On("CreateTestBranch", mock.Anything, mock.Anything, mock.Anything).Return(func() {}, nil)

		mockClient.On("CreatePullRequest", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("API rate limit exceeded"))
		prEngine := NewPullRequestEngine(mockClient, logger)

		analysis := &FailureAnalysisResult{
			Classification: FailureClassification{
				Category: "test_failure",
			},
		}
		fix := &FixValidationResult{
			Valid: true,
			Fix: &ProposedFix{
				ID: "fix-456",
				Description: "Test fix",
			},
		}

		_, err := prEngine.CreateFixPR(context.Background(), analysis, fix)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API rate limit exceeded")

		mockClient.AssertExpectations(t)
	})

	t.Run("update PR error handling", func(t *testing.T) {
		mockClient := new(MockGitHubClient)
		mockClient.On("UpdatePullRequest", mock.Anything, 123, mock.Anything).
			Return(nil, fmt.Errorf("PR not found"))

		prEngine := NewPullRequestEngine(mockClient, logger)

		updates := &PRCreationOptions{
			Title: "Updated Title",
		}

		_, err := prEngine.UpdatePR(context.Background(), 123, updates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PR not found")

		mockClient.AssertExpectations(t)
	})
}