package main

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestMCPPRIntegration tests that both GitHubIntegration and MCPGitHubClient
// implement the GitHubClient interface correctly for PR operations
func TestMCPPRIntegration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("PR engine accepts GitHubClient interface", func(t *testing.T) {
		// This test verifies that the PR engine can work with any GitHubClient implementation
		
		// Create a mock that implements GitHubClient
		mockClient := &MockGitHubClient{}
		
		// Create PR engine with the mock - this should compile if interface is correct
		prEngine := NewPullRequestEngine(mockClient, logger)
		
		assert.NotNil(t, prEngine)
		assert.Equal(t, mockClient, prEngine.githubClient)
	})

	t.Run("MCPGitHubClient implements GitHubClient", func(t *testing.T) {
		// This test verifies that MCPGitHubClient satisfies the GitHubClient interface
		var _ GitHubClient = (*MCPGitHubClient)(nil)
		// If this compiles, the interface is satisfied
	})

	t.Run("GitHubIntegration implements GitHubClient", func(t *testing.T) {
		// This test verifies that GitHubIntegration satisfies the GitHubClient interface
		var _ GitHubClient = (*GitHubIntegration)(nil)
		// If this compiles, the interface is satisfied
	})
}