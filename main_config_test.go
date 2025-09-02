package main

import (
	"context"
	"testing"

	"dagger.io/dagger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Test configuration setters ensure fields are assigned and fluent API is preserved
func TestDaggerAutofixConfigurationSetters(t *testing.T) {
	fakeDir := &dagger.Directory{}
	token := createTestSecret("token", "ghp_test")
	apiKey := createTestSecret("api", "sk-test")

	tests := []struct {
		name   string
		apply  func(*DaggerAutofix) *DaggerAutofix
		verify func(*testing.T, *DaggerAutofix)
	}{
		{
			name:   "WithSource",
			apply:  func(m *DaggerAutofix) *DaggerAutofix { return m.WithSource(fakeDir) },
			verify: func(t *testing.T, m *DaggerAutofix) { assert.Equal(t, fakeDir, m.Source) },
		},
		{
			name:   "WithGitHubToken",
			apply:  func(m *DaggerAutofix) *DaggerAutofix { return m.WithGitHubToken(token) },
			verify: func(t *testing.T, m *DaggerAutofix) { assert.Equal(t, token, m.GitHubToken) },
		},
		{
			name:  "WithLLMProvider",
			apply: func(m *DaggerAutofix) *DaggerAutofix { return m.WithLLMProvider("anthropic", apiKey) },
			verify: func(t *testing.T, m *DaggerAutofix) {
				assert.Equal(t, LLMProvider("anthropic"), m.LLMProvider)
				assert.Equal(t, apiKey, m.LLMAPIKey)
			},
		},
		{
			name:  "WithRepository",
			apply: func(m *DaggerAutofix) *DaggerAutofix { return m.WithRepository("owner", "repo") },
			verify: func(t *testing.T, m *DaggerAutofix) {
				assert.Equal(t, "owner", m.RepoOwner)
				assert.Equal(t, "repo", m.RepoName)
			},
		},
		{
			name:   "WithTargetBranch",
			apply:  func(m *DaggerAutofix) *DaggerAutofix { return m.WithTargetBranch("develop") },
			verify: func(t *testing.T, m *DaggerAutofix) { assert.Equal(t, "develop", m.TargetBranch) },
		},
		{
			name:   "WithMinCoverage",
			apply:  func(m *DaggerAutofix) *DaggerAutofix { return m.WithMinCoverage(90) },
			verify: func(t *testing.T, m *DaggerAutofix) { assert.Equal(t, 90, m.MinCoverage) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			returned := tt.apply(m)
			assert.Equal(t, m, returned)
			tt.verify(t, m)
		})
	}
}

// TestInitializeAndEnsureInitialized verifies initialization logic with dependency injection
func TestInitializeAndEnsureInitialized(t *testing.T) {
	t.Run("MissingConfiguration", func(t *testing.T) {
		module := New()
		_, err := module.Initialize(context.Background())
		assert.Error(t, err)
	})

	t.Run("SuccessfulInitialization", func(t *testing.T) {
		module := New().
			WithGitHubToken(createTestSecret("token", "ghp_test")).
			WithLLMProvider("openai", createTestSecret("key", "sk-test")).
			WithRepository("owner", "repo")

		// Inject mocks
		oldGH := newGitHubIntegration
		oldLLM := newLLMClient
		oldFailure := newFailureAnalysisEngine
		oldTest := newTestEngine
		oldPR := newPullRequestEngine
		defer func() {
			newGitHubIntegration = oldGH
			newLLMClient = oldLLM
			newFailureAnalysisEngine = oldFailure
			newTestEngine = oldTest
			newPullRequestEngine = oldPR
		}()

		newGitHubIntegration = func(ctx context.Context, token *dagger.Secret, owner, name string) (*GitHubIntegration, error) {
			return &GitHubIntegration{}, nil
		}
		newLLMClient = func(ctx context.Context, provider LLMProvider, apiKey *dagger.Secret) (*LLMClient, error) {
			return &LLMClient{}, nil
		}
		newFailureAnalysisEngine = func(llm *LLMClient, logger *logrus.Logger) *FailureAnalysisEngine {
			return &FailureAnalysisEngine{}
		}
		newTestEngine = func(minCoverage int, logger *logrus.Logger) *TestEngine {
			return &TestEngine{}
		}
		newPullRequestEngine = func(gh *GitHubIntegration, logger *logrus.Logger) *PullRequestEngine {
			return &PullRequestEngine{}
		}

		initialized, err := module.Initialize(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, module, initialized)

		// ensureInitialized should pass now
		assert.NoError(t, module.ensureInitialized())
	})
}
