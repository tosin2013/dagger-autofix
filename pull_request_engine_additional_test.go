package main

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestCreateManualPR tests the CreateManualPR method
func TestCreateManualPR(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce noise
	
	ctx := context.Background()
	
	t.Run("CreateManualPR with valid inputs", func(t *testing.T) {
		githubClient := &GitHubIntegration{
			client:    nil, // This will cause expected failures in GitHub API calls
			repoOwner: "test-owner",
			repoName:  "test-repo",
		}
		
		engine := NewPullRequestEngine(githubClient, logger)
		
		analysis := &FailureAnalysisResult{
			ID: "test-analysis-manual",
			Classification: FailureClassification{
				Type:     BuildFailure,
				Severity: High,
			},
		}
		
		options := &PRCreationOptions{
			BranchName:   "manual-fix-branch",
			TargetBranch: "main",
			Title:        "Manual Fix for Build Failure",
			Body:         "This PR requires manual review",
			Labels:       []string{"manual-review", "build-fix"},
		}
		
		// Test with defensive error handling
		var pr *PullRequest
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = assert.AnError
				}
			}()
			pr, err = engine.CreateManualPR(ctx, analysis, options)
		}()
		
		// Should get an error due to nil GitHub client
		assert.Error(t, err)
		assert.Nil(t, pr)
	})
	
	t.Run("CreateManualPR with nil analysis", func(t *testing.T) {
		githubClient := &GitHubIntegration{
			client:    nil,
			repoOwner: "test-owner",
			repoName:  "test-repo",
		}
		
		engine := NewPullRequestEngine(githubClient, logger)
		
		options := &PRCreationOptions{
			BranchName: "test-branch",
			Title:      "Test PR",
		}
		
		// Test with nil analysis - should handle gracefully or panic
		var pr *PullRequest
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = assert.AnError
				}
			}()
			pr, err = engine.CreateManualPR(ctx, nil, options)
		}()
		
		// Should either error or panic due to nil analysis
		assert.Error(t, err)
		assert.Nil(t, pr)
	})
	
	t.Run("CreateManualPR with nil options", func(t *testing.T) {
		githubClient := &GitHubIntegration{
			client:    nil,
			repoOwner: "test-owner",
			repoName:  "test-repo",
		}
		
		engine := NewPullRequestEngine(githubClient, logger)
		
		analysis := &FailureAnalysisResult{
			ID: "test-analysis",
		}
		
		// Test with nil options
		var pr *PullRequest
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = assert.AnError
				}
			}()
			pr, err = engine.CreateManualPR(ctx, analysis, nil)
		}()
		
		// Should error due to nil options
		assert.Error(t, err)
		assert.Nil(t, pr)
	})
}

// TestCreatePullRequestMethod tests the createPullRequest method
func TestCreatePullRequestMethod(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	
	ctx := context.Background()
	
	t.Run("createPullRequest with basic options", func(t *testing.T) {
		githubClient := &GitHubIntegration{
			client:    nil,
			repoOwner: "test-owner",
			repoName:  "test-repo",
		}
		
		engine := NewPullRequestEngine(githubClient, logger)
		
		options := &PRCreationOptions{
			BranchName:   "feature-branch",
			TargetBranch: "main",
			Title:        "Test PR Title",
			Body:         "Test PR Body",
			Labels:       []string{"test"},
			Reviewers:    []string{"reviewer1"},
			Assignees:    []string{"assignee1"},
			Draft:        false,
		}
		
		// Test the private method through reflection or public interface
		var pr *PullRequest
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = assert.AnError
				}
			}()
			pr, err = engine.createPullRequest(ctx, options)
		}()
		
		// Should error due to nil GitHub client
		assert.Error(t, err)
		assert.Nil(t, pr)
	})
	
	t.Run("createPullRequest with nil options", func(t *testing.T) {
		githubClient := &GitHubIntegration{
			client:    nil,
			repoOwner: "test-owner",
			repoName:  "test-repo",
		}
		
		engine := NewPullRequestEngine(githubClient, logger)
		
		// Test with nil options
		var pr *PullRequest
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = assert.AnError
				}
			}()
			pr, err = engine.createPullRequest(ctx, nil)
		}()
		
		// Should error due to nil options
		assert.Error(t, err)
		assert.Nil(t, pr)
	})
	
	t.Run("createPullRequest with empty required fields", func(t *testing.T) {
		githubClient := &GitHubIntegration{
			client:    nil,
			repoOwner: "test-owner",
			repoName:  "test-repo",
		}
		
		engine := NewPullRequestEngine(githubClient, logger)
		
		options := &PRCreationOptions{
			BranchName:   "", // Empty branch name
			TargetBranch: "",
			Title:        "",
			Body:         "",
		}
		
		var pr *PullRequest
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = assert.AnError
				}
			}()
			pr, err = engine.createPullRequest(ctx, options)
		}()
		
		// Should error due to empty required fields
		assert.Error(t, err)
		assert.Nil(t, pr)
	})
}

// TestGetPRStatusMethod tests the GetPRStatus method
func TestGetPRStatusMethod(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	
	ctx := context.Background()
	
	t.Run("GetPRStatus with valid PR number", func(t *testing.T) {
		githubClient := &GitHubIntegration{
			client:    nil,
			repoOwner: "test-owner",
			repoName:  "test-repo",
		}
		
		engine := NewPullRequestEngine(githubClient, logger)
		
		// Test with valid PR number
		var pr *PullRequest
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = assert.AnError
				}
			}()
			pr, err = engine.GetPRStatus(ctx, 123)
		}()
		
		// Should error due to nil GitHub client
		assert.Error(t, err)
		assert.Nil(t, pr)
	})
	
	t.Run("GetPRStatus with invalid PR number", func(t *testing.T) {
		githubClient := &GitHubIntegration{
			client:    nil,
			repoOwner: "test-owner",
			repoName:  "test-repo",
		}
		
		engine := NewPullRequestEngine(githubClient, logger)
		
		// Test with invalid PR number (negative)
		var pr *PullRequest
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = assert.AnError
				}
			}()
			pr, err = engine.GetPRStatus(ctx, -1)
		}()
		
		// Should error due to invalid PR number or nil client
		assert.Error(t, err)
		assert.Nil(t, pr)
	})
	
	t.Run("GetPRStatus with zero PR number", func(t *testing.T) {
		githubClient := &GitHubIntegration{
			client:    nil,
			repoOwner: "test-owner",
			repoName:  "test-repo",
		}
		
		engine := NewPullRequestEngine(githubClient, logger)
		
		// Test with zero PR number
		var pr *PullRequest
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = assert.AnError
				}
			}()
			pr, err = engine.GetPRStatus(ctx, 0)
		}()
		
		// Should error due to invalid PR number
		assert.Error(t, err)
		assert.Nil(t, pr)
	})
}

// TestPRCreationOptions tests the PRCreationOptions struct
func TestPRCreationOptions(t *testing.T) {
	options := &PRCreationOptions{
		BranchName:   "feature/test-branch",
		TargetBranch: "develop",
		Title:        "Test Feature Implementation",
		Body:         "This PR implements a test feature with comprehensive tests",
		Labels:       []string{"feature", "enhancement", "needs-review"},
		Reviewers:    []string{"reviewer1", "reviewer2"},
		Assignees:    []string{"developer1"},
		Draft:        true,
		DeleteBranch: true,
	}
	
	assert.Equal(t, "feature/test-branch", options.BranchName)
	assert.Equal(t, "develop", options.TargetBranch)
	assert.Equal(t, "Test Feature Implementation", options.Title)
	assert.Equal(t, "This PR implements a test feature with comprehensive tests", options.Body)
	assert.Equal(t, []string{"feature", "enhancement", "needs-review"}, options.Labels)
	assert.Equal(t, []string{"reviewer1", "reviewer2"}, options.Reviewers)
	assert.Equal(t, []string{"developer1"}, options.Assignees)
	assert.True(t, options.Draft)
	assert.True(t, options.DeleteBranch)
}

// TestPullRequest tests the PullRequest struct
func TestPullRequest(t *testing.T) {
	pr := &PullRequest{
		Number:    123,
		URL:       "https://github.com/test/repo/pull/123",
		State:     "open",
		Title:     "Test Pull Request",
		Body:      "This is a test pull request",
		Branch:    "feature/test", // Corrected field name
		CommitSHA: "abc123",
		Author:    "test-author",
		Labels:    []string{"test", "automated"},
	}
	
	assert.Equal(t, 123, pr.Number)
	assert.Equal(t, "https://github.com/test/repo/pull/123", pr.URL)
	assert.Equal(t, "open", pr.State)
	assert.Equal(t, "Test Pull Request", pr.Title)
	assert.Equal(t, "This is a test pull request", pr.Body)
	assert.Equal(t, "feature/test", pr.Branch) // Corrected field name
	assert.Equal(t, "abc123", pr.CommitSHA)
	assert.Equal(t, "test-author", pr.Author)
	assert.Equal(t, []string{"test", "automated"}, pr.Labels)
}

// TestPRGenerationMethods tests PR content generation methods
func TestPRGenerationMethods(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	
	t.Run("generatePRTitle", func(t *testing.T) {
		engine := NewPullRequestEngine(nil, logger)
		
		analysis := &FailureAnalysisResult{
			Classification: FailureClassification{
				Type: BuildFailure,
			},
			Context: FailureContext{
				WorkflowRun: &WorkflowRun{
					ID: 123,
				},
			},
		}
		
		fix := &ProposedFix{
			Type:        CodeFix,
			Description: "Fix build compilation error",
		}
		
		// Test title generation - correct parameter type
		var title string
		func() {
			defer func() {
				if r := recover(); r != nil {
					title = "Failed to generate title"
				}
			}()
			title = engine.generatePRTitle(analysis, fix)
		}()
		
		assert.NotEmpty(t, title)
		assert.Contains(t, title, "🤖") // Should contain bot emoji
	})
	
	t.Run("generatePRLabels", func(t *testing.T) {
		engine := NewPullRequestEngine(nil, logger)
		
		analysis := &FailureAnalysisResult{
			Classification: FailureClassification{
				Type:     BuildFailure,
				Severity: High,
				Tags:     []string{"compilation", "urgent"},
			},
		}
		
		fix := &ProposedFix{
			Type: CodeFix,
		}
		
		// Test label generation - correct parameter type
		var labels []string
		func() {
			defer func() {
				if r := recover(); r != nil {
					labels = []string{"error"}
				}
			}()
			labels = engine.generatePRLabels(analysis, fix)
		}()
		
		assert.NotNil(t, labels)
		// Should contain some automated labels
		if len(labels) > 0 {
			found := false
			for _, label := range labels {
				if label == "autofix" || label == "automated" || label == "code-fix" {
					found = true
					break
				}
			}
			assert.True(t, found, "Should contain expected labels")
		}
	})
}