package main

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestDaggerAutofix tests the main module functionality
func TestDaggerAutofix(t *testing.T) {
	t.Run("NewDaggerAutofix", func(t *testing.T) {
		module := New()
		assert.NotNil(t, module)
		assert.Equal(t, LLMProvider("openai"), module.LLMProvider)
		assert.Equal(t, "main", module.TargetBranch)
		assert.Equal(t, 85, module.MinCoverage)
	})

	t.Run("WithConfiguration", func(t *testing.T) {
		_ = context.Background()
		module := New().
			WithLLMProvider("anthropic", createTestSecret("test-key", "test-value")).
			WithRepository("owner", "repo").
			WithTargetBranch("develop").
			WithMinCoverage(90)

		assert.Equal(t, LLMProvider("anthropic"), module.LLMProvider)
		assert.Equal(t, "owner", module.RepoOwner)
		assert.Equal(t, "repo", module.RepoName)
		assert.Equal(t, "develop", module.TargetBranch)
		assert.Equal(t, 90, module.MinCoverage)
	})

	t.Run("ValidateConfiguration", func(t *testing.T) {
		module := New()

		// Missing required fields
		err := module.validateConfiguration()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GitHub token is required")

		// Valid configuration
		module.GitHubToken = createTestSecret("github-token", "test-token")
		module.LLMAPIKey = createTestSecret("llm-key", "test-key")
		module.RepoOwner = "owner"
		module.RepoName = "repo"

		err = module.validateConfiguration()
		assert.NoError(t, err)
	})
}

// TestFailureClassification tests failure type classification
func TestFailureClassification(t *testing.T) {
	testCases := []struct {
		name     string
		context  FailureContext
		expected FailureType
	}{
		{
			name: "NPM Install Failure",
			context: FailureContext{
				Logs: &WorkflowLogs{
					ErrorLines: []string{"npm install failed"},
				},
			},
			expected: DependencyFailure,
		},
		{
			name: "Go Build Failure",
			context: FailureContext{
				Logs: &WorkflowLogs{
					ErrorLines: []string{"go build failed with errors"},
				},
			},
			expected: BuildFailure,
		},
		{
			name: "Test Timeout",
			context: FailureContext{
				Logs: &WorkflowLogs{
					ErrorLines: []string{"test execution timeout"},
				},
			},
			expected: TestFailure,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine := NewFailureAnalysisEngine(nil, logrus.New())
			classification := engine.preClassifyFailure(tc.context)
			assert.Equal(t, tc.expected, classification.Type)
		})
	}
}

// TestLLMClient tests LLM provider integration
func TestLLMClient(t *testing.T) {
	// Mock test - in real tests, you'd use actual API keys for integration testing
	t.Run("NewLLMClient", func(t *testing.T) {
		ctx := context.Background()
		apiKey := createTestSecret("test-key", "test-value")

		// This would fail in real testing without valid API key
		_, err := NewLLMClient(ctx, OpenAI, apiKey)
		if err != nil {
			t.Skip("Skipping LLM client test - requires valid API key")
		}
	})

	t.Run("SupportedProviders", func(t *testing.T) {
		supportedProviders := []LLMProvider{
			OpenAI, Anthropic, Gemini, DeepSeek, LiteLLM,
		}

		for _, provider := range supportedProviders {
			config := getDefaultConfig(provider)
			assert.NotNil(t, config)
			assert.NotEmpty(t, config.Model)
			assert.Greater(t, config.MaxTokens, 0)
		}
	})
}

// TestTestEngine tests the testing and validation functionality
func TestTestEngine(t *testing.T) {
	t.Run("NewTestEngine", func(t *testing.T) {
		engine := NewTestEngine(85, logrus.New())
		assert.NotNil(t, engine)
		assert.Equal(t, 85, engine.minCoverage)
		assert.NotNil(t, engine.testFrameworks)
		assert.NotNil(t, engine.coverageTools)
	})

	t.Run("TestFrameworkDetection", func(t *testing.T) {
		testCases := []struct {
			filename string
			expected string
		}{
			{"package.json", "nodejs"},
			{"go.mod", "golang"},
			{"pom.xml", "maven"},
			{"requirements.txt", "python"},
			{"Cargo.toml", "rust"},
			{"unknown.txt", "generic"},
		}

		engine := NewTestEngine(85, logrus.New())
		for _, tc := range testCases {
			framework := engine.getFrameworkByFile(tc.filename)
			assert.Equal(t, tc.expected, framework.Name)
		}
	})

	t.Run("CoverageValidation", func(t *testing.T) {
		engine := NewTestEngine(80, logrus.New())

		// Coverage above threshold
		coverageResult := &CoverageResult{Coverage: 85.0}
		err := engine.ValidateTestCoverage(context.Background(), coverageResult)
		assert.NoError(t, err)

		// Coverage below threshold
		coverageResult = &CoverageResult{Coverage: 75.0}
		err = engine.ValidateTestCoverage(context.Background(), coverageResult)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "coverage")
	})
}

// TestPullRequestEngine tests PR creation functionality
func TestPullRequestEngine(t *testing.T) {
	t.Run("GeneratePRTitle", func(t *testing.T) {
		engine := NewPullRequestEngine(nil, logrus.New())
		analysis := &FailureAnalysisResult{
			Classification: FailureClassification{
				Type: BuildFailure,
			},
			Context: FailureContext{
				WorkflowRun: &WorkflowRun{ID: 12345},
			},
		}
		fix := &ProposedFix{
			Type: CodeFix,
		}

		title := engine.generatePRTitle(analysis, fix)
		assert.Contains(t, title, "Auto-fix")
		assert.Contains(t, title, "Code")
		assert.Contains(t, title, "Build")
		assert.Contains(t, title, "12345")
	})

	t.Run("GeneratePRLabels", func(t *testing.T) {
		engine := NewPullRequestEngine(nil, logrus.New())
		analysis := &FailureAnalysisResult{
			Classification: FailureClassification{
				Type:     TestFailure,
				Severity: High,
			},
		}
		fix := &ProposedFix{
			Type:       TestFix,
			Confidence: 0.9,
		}

		labels := engine.generatePRLabels(analysis, fix)
		assert.Contains(t, labels, "autofix")
		assert.Contains(t, labels, "test-fix")
		assert.Contains(t, labels, "test-failure")
		assert.Contains(t, labels, "priority-high")
		assert.Contains(t, labels, "high-confidence")
	})

	t.Run("GenerateBranchName", func(t *testing.T) {
		engine := NewPullRequestEngine(nil, logrus.New())
		analysis := &FailureAnalysisResult{ID: "analysis-123"}
		fix := &ProposedFix{Type: DependencyFix}

		branchName := engine.generateBranchName(analysis, fix)
		assert.Contains(t, branchName, "autofix")
		assert.Contains(t, branchName, "dependency")
		assert.Contains(t, branchName, "analysis-123")
	})
}

// TestErrorPatterns tests error pattern matching
func TestErrorPatterns(t *testing.T) {
	t.Run("LoadErrorPatterns", func(t *testing.T) {
		patterns := loadErrorPatterns()
		assert.NotNil(t, patterns)
		assert.NotEmpty(t, patterns.Patterns)

		// Check specific patterns
		npmPattern, exists := patterns.Patterns["npm_install_failure"]
		assert.True(t, exists)
		assert.Equal(t, DependencyFailure, npmPattern.Type)
		assert.Equal(t, "npm install", npmPattern.Pattern)

		goPattern, exists := patterns.Patterns["go_build_failure"]
		assert.True(t, exists)
		assert.Equal(t, BuildFailure, goPattern.Type)
		assert.Equal(t, "go build", goPattern.Pattern)
	})
}

// TestPromptTemplates tests prompt template loading
func TestPromptTemplates(t *testing.T) {
	t.Run("LoadPromptTemplates", func(t *testing.T) {
		templates := loadPromptTemplates()
		assert.NotNil(t, templates)
		assert.NotEmpty(t, templates.FailureAnalysis)
		assert.NotEmpty(t, templates.FixGeneration)
		assert.NotEmpty(t, templates.CodeAnalysis)
		assert.NotEmpty(t, templates.TestGeneration)
		assert.NotEmpty(t, templates.SecurityAnalysis)
	})
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	t.Run("ValidConfiguration", func(t *testing.T) {
		module := &DaggerAutofix{
			GitHubToken: createTestSecret("github-token", "valid-token"),
			LLMAPIKey:   createTestSecret("llm-key", "valid-key"),
			RepoOwner:   "test-owner",
			RepoName:    "test-repo",
		}

		err := module.validateConfiguration()
		assert.NoError(t, err)
	})

	t.Run("MissingGitHubToken", func(t *testing.T) {
		module := &DaggerAutofix{
			LLMAPIKey: createTestSecret("llm-key", "valid-key"),
			RepoOwner: "test-owner",
			RepoName:  "test-repo",
		}

		err := module.validateConfiguration()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GitHub token is required")
	})

	t.Run("MissingLLMKey", func(t *testing.T) {
		module := &DaggerAutofix{
			GitHubToken: createTestSecret("github-token", "valid-token"),
			RepoOwner:   "test-owner",
			RepoName:    "test-repo",
		}

		err := module.validateConfiguration()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LLM API key is required")
	})

	t.Run("MissingRepository", func(t *testing.T) {
		module := &DaggerAutofix{
			GitHubToken: createTestSecret("github-token", "valid-token"),
			LLMAPIKey:   createTestSecret("llm-key", "valid-key"),
		}

		err := module.validateConfiguration()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository owner and name are required")
	})
}

// TestFailureAnalysisResult tests analysis result structures
func TestFailureAnalysisResult(t *testing.T) {
	t.Run("CreateAnalysisResult", func(t *testing.T) {
		result := &FailureAnalysisResult{
			ID:          "test-analysis-1",
			RootCause:   "Test root cause",
			Description: "Test description",
			Classification: FailureClassification{
				Type:       CodeFailure,
				Severity:   High,
				Category:   Systematic,
				Confidence: 0.85,
				Tags:       []string{"test", "code"},
			},
			AffectedFiles: []string{"main.go", "test.go"},
			Timestamp:     time.Now(),
		}

		assert.Equal(t, "test-analysis-1", result.ID)
		assert.Equal(t, CodeFailure, result.Classification.Type)
		assert.Equal(t, High, result.Classification.Severity)
		assert.Equal(t, 0.85, result.Classification.Confidence)
		assert.Len(t, result.AffectedFiles, 2)
		assert.Contains(t, result.AffectedFiles, "main.go")
	})
}

// TestProposedFix tests fix proposal structures
func TestProposedFix(t *testing.T) {
	t.Run("CreateProposedFix", func(t *testing.T) {
		fix := &ProposedFix{
			ID:          "test-fix-1",
			Type:        CodeFix,
			Description: "Fix the issue",
			Rationale:   "This fixes the root cause",
			Confidence:  0.9,
			Risks:       []string{"Might break something"},
			Benefits:    []string{"Fixes the issue"},
			Changes: []CodeChange{
				{
					FilePath:    "main.go",
					Operation:   "modify",
					OldContent:  "old code",
					NewContent:  "new code",
					Explanation: "Fixed the bug",
				},
			},
			Timestamp: time.Now(),
		}

		assert.Equal(t, "test-fix-1", fix.ID)
		assert.Equal(t, CodeFix, fix.Type)
		assert.Equal(t, 0.9, fix.Confidence)
		assert.Len(t, fix.Changes, 1)
		assert.Equal(t, "main.go", fix.Changes[0].FilePath)
		assert.Equal(t, "modify", fix.Changes[0].Operation)
	})
}

// TestOperationalMetrics tests metrics collection
func TestOperationalMetrics(t *testing.T) {
	t.Run("CreateMetrics", func(t *testing.T) {
		metrics := &OperationalMetrics{
			TotalFailuresDetected: 10,
			SuccessfulFixes:       8,
			FailedFixes:           2,
			AverageFixTime:        5 * time.Minute,
			TestCoverage:          85.5,
			LLMProviderStats: map[string]int{
				"openai":    5,
				"anthropic": 3,
			},
			LastUpdated: time.Now(),
		}

		assert.Equal(t, 10, metrics.TotalFailuresDetected)
		assert.Equal(t, 8, metrics.SuccessfulFixes)
		assert.Equal(t, 2, metrics.FailedFixes)
		assert.Equal(t, 85.5, metrics.TestCoverage)
		assert.Equal(t, 5, metrics.LLMProviderStats["openai"])
	})
}

// Integration tests (these would require actual API credentials)
func TestIntegration(t *testing.T) {
	t.Skip("Integration tests require actual API credentials")

	t.Run("EndToEndAutoFix", func(t *testing.T) {
		// This would test the complete workflow:
		// 1. Create agent
		// 2. Analyze a real failure
		// 3. Generate fixes
		// 4. Validate fixes
		// 5. Create PR
	})
}

// Benchmark tests
func BenchmarkFailureAnalysis(b *testing.B) {
	engine := NewFailureAnalysisEngine(nil, logrus.New())
	context := FailureContext{
		Logs: &WorkflowLogs{
			ErrorLines: []string{"npm install failed", "build error"},
			RawLogs:    "lots of log content here...",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.preClassifyFailure(context)
	}
}

func BenchmarkPromptGeneration(b *testing.B) {
	engine := NewFailureAnalysisEngine(nil, logrus.New())
	context := FailureContext{
		WorkflowRun: &WorkflowRun{
			ID:     12345,
			Name:   "CI",
			Branch: "main",
		},
		Logs: &WorkflowLogs{
			ErrorLines: []string{"error occurred"},
			RawLogs:    "full log content",
		},
		Repository: RepositoryContext{
			Owner:    "test",
			Name:     "repo",
			Language: "go",
		},
	}
	preClass := &FailureClassification{
		Type:       CodeFailure,
		Confidence: 0.8,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.buildAnalysisPrompt(context, preClass)
	}
}
