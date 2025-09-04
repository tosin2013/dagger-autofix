package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCreateTestContainerWithMocks(t *testing.T) {
	tests := []struct {
		name          string
		owner         string
		repo          string
		branch        string
		expectedError bool
	}{
		{
			name:          "successful container creation",
			owner:         "test-owner",
			repo:          "test-repo",
			branch:        "main",
			expectedError: false,
		},
		{
			name:          "empty parameters",
			owner:         "",
			repo:          "",
			branch:        "",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup with mocks
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)

			mockProvider := NewMockContainerProvider()
			engine := NewTestEngine(85, logger)
			engine.SetContainerProvider(mockProvider)

			// Execute
			container, err := engine.createTestContainer(context.Background(), tt.owner, tt.repo, tt.branch)

			// Validate
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			assert.NotNil(t, container)

			// Validate mock behavior
			mock := mockProvider.MockContainer
			assert.Equal(t, "ubuntu:22.04", mock.BaseImage)
			assert.Equal(t, "/workspace", mock.WorkingDir)
			assert.True(t, len(mock.ExecHistory) > 0, "Expected exec commands")
		})
	}
}

func TestDetectFrameworkWithMocks(t *testing.T) {
	tests := []struct {
		name              string
		setupFiles        func(*MockDaggerContainer)
		expectedFramework string
		expectedError     bool
	}{
		{
			name: "detect Go framework",
			setupFiles: func(mock *MockDaggerContainer) {
				// Clear all default files first
				mock.FileSystem = make(map[string]string)
				mock.SetFileContent("go.mod", "module test\ngo 1.19")
			},
			expectedFramework: "golang",
			expectedError:     false,
		},
		{
			name: "detect Node.js framework",
			setupFiles: func(mock *MockDaggerContainer) {
				mock.SetFileContent("package.json", `{"name":"test","scripts":{"test":"jest"}}`)
				// Remove go.mod so package.json is found first
				delete(mock.FileSystem, "go.mod")
			},
			expectedFramework: "nodejs",  // Fixed to match actual framework name
			expectedError:     false,
		},
		{
			name: "no framework files - use default",
			setupFiles: func(mock *MockDaggerContainer) {
				// Clear all files
				mock.FileSystem = make(map[string]string)
			},
			expectedFramework: "generic",
			expectedError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)

			mockProvider := NewMockContainerProvider()
			tt.setupFiles(mockProvider.MockContainer)

			engine := NewTestEngine(85, logger)
			engine.SetContainerProvider(mockProvider)

			container := &MockContainerWrapper{mockProvider.MockContainer}

			// Execute
			framework, err := engine.detectFramework(context.Background(), container)

			// Validate
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if framework != nil {
				if tt.expectedFramework == "generic" {
					assert.Equal(t, "generic", framework.Name)
				} else {
					assert.Equal(t, tt.expectedFramework, framework.Name)
				}
			}
		})
	}
}

func TestRunTestSuiteWithMocks(t *testing.T) {
	tests := []struct {
		name           string
		framework      *TestFramework
		setupMock      func(*MockDaggerContainer)
		expectedError  bool
		validateOutput func(*testing.T, string)
	}{
		{
			name: "successful Go test execution",
			framework: &TestFramework{
				Name:        "go",
				Language:    "go",
				TestCommand: "go test ./...",
				Environment: map[string]string{"GO_ENV": "test"},
			},
			setupMock: func(mock *MockDaggerContainer) {
				mock.SetCommandOutput("go test ./...",
					"PASS\ncoverage: 87.5% of statements\nok\ttest\t0.005s",
					"", 0, nil)
			},
			expectedError: false,
			validateOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "PASS")
				assert.Contains(t, output, "87.5%")
			},
		},
		{
			name: "test execution with failure",
			framework: &TestFramework{
				Name:        "go",
				Language:    "go",
				TestCommand: "go test ./...",
				Environment: map[string]string{},
			},
			setupMock: func(mock *MockDaggerContainer) {
				mock.SetCommandOutput("go test ./...",
					"FAIL", "", 1,
					fmt.Errorf("tests failed"))
			},
			expectedError: true,
			validateOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "FAIL")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)

			mockProvider := NewMockContainerProvider()
			tt.setupMock(mockProvider.MockContainer)

			engine := NewTestEngine(85, logger)
			engine.SetContainerProvider(mockProvider)

			container := &MockContainerWrapper{mockProvider.MockContainer}

			// Execute
			output, err := engine.runTestSuite(context.Background(), container, tt.framework)

			// Validate
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			tt.validateOutput(t, output)

			// Validate mock interactions
			mock := mockProvider.MockContainer
			assert.Equal(t, tt.framework.Environment["GO_ENV"], mock.EnvVars["GO_ENV"])

			// Verify command was executed
			found := false
			for _, execCmd := range mock.ExecHistory {
				cmdStr := strings.Join(execCmd, " ")
				if strings.Contains(cmdStr, "go test") {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected test command to be executed")
		})
	}
}

func TestRunBuildWithMocks(t *testing.T) {
	tests := []struct {
		name           string
		framework      *TestFramework
		setupMock      func(*MockDaggerContainer)
		expectedError  bool
		validateOutput func(*testing.T, string)
	}{
		{
			name: "successful build",
			framework: &TestFramework{
				Name:         "go",
				Language:     "go",
				BuildCommand: "go build .",
				Environment:  map[string]string{"GOOS": "linux"},
			},
			setupMock: func(mock *MockDaggerContainer) {
				mock.SetCommandOutput("go build .",
					"Build successful", "", 0, nil)
			},
			expectedError: false,
			validateOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "Build successful")
			},
		},
		{
			name: "no build command configured",
			framework: &TestFramework{
				Name:         "no-build",
				BuildCommand: "", // Empty build command
			},
			setupMock: func(mock *MockDaggerContainer) {
				// No setup needed
			},
			expectedError: false,
			validateOutput: func(t *testing.T, output string) {
				assert.Equal(t, "No build configured", output)
			},
		},
		{
			name: "build failure",
			framework: &TestFramework{
				Name:         "go",
				Language:     "go", 
				BuildCommand: "go build .",
				Environment:  map[string]string{},
			},
			setupMock: func(mock *MockDaggerContainer) {
				mock.SetCommandOutput("go build .",
					"Build failed: compilation error", "", 1,
					fmt.Errorf("build failed"))
			},
			expectedError: true,
			validateOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "Build failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)

			mockProvider := NewMockContainerProvider()
			tt.setupMock(mockProvider.MockContainer)

			engine := NewTestEngine(85, logger)
			engine.SetContainerProvider(mockProvider)

			container := &MockContainerWrapper{mockProvider.MockContainer}

			// Execute
			output, err := engine.runBuild(context.Background(), container, tt.framework)

			// Validate
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			tt.validateOutput(t, output)

			// Validate mock interactions
			mock := mockProvider.MockContainer
			if tt.framework.Environment["GOOS"] != "" {
				assert.Equal(t, tt.framework.Environment["GOOS"], mock.EnvVars["GOOS"])
			}
		})
	}
}

func TestRunLintingWithMocks(t *testing.T) {
	tests := []struct {
		name           string
		framework      *TestFramework
		setupMock      func(*MockDaggerContainer)
		expectedError  bool
		validateOutput func(*testing.T, string)
	}{
		{
			name: "successful linting",
			framework: &TestFramework{
				Name:        "javascript",
				Language:    "javascript",
				LintCommand: "eslint .",
				Environment: map[string]string{"NODE_ENV": "test"},
			},
			setupMock: func(mock *MockDaggerContainer) {
				mock.SetCommandOutput("eslint .",
					"No linting issues found", "", 0, nil)
			},
			expectedError: false,
			validateOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "No linting issues found")
			},
		},
		{
			name: "no lint command configured",
			framework: &TestFramework{
				Name:        "no-lint",
				LintCommand: "", // Empty lint command
			},
			setupMock: func(mock *MockDaggerContainer) {
				// No setup needed
			},
			expectedError: false,
			validateOutput: func(t *testing.T, output string) {
				assert.Equal(t, "No linting configured", output)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)

			mockProvider := NewMockContainerProvider()
			tt.setupMock(mockProvider.MockContainer)

			engine := NewTestEngine(85, logger)
			engine.SetContainerProvider(mockProvider)

			container := &MockContainerWrapper{mockProvider.MockContainer}

			// Execute
			output, err := engine.runLinting(context.Background(), container, tt.framework)

			// Validate
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			tt.validateOutput(t, output)
		})
	}
}

func TestRunCoverageAnalysisWithMocks(t *testing.T) {
	tests := []struct {
		name              string
		framework         *TestFramework
		setupMock         func(*MockDaggerContainer)
		expectedError     bool
		expectedCoverage  float64
		validateResult    func(*testing.T, *CoverageResult)
	}{
		{
			name: "successful coverage analysis - Go",
			framework: &TestFramework{
				Name:            "go",
				Language:        "go",
				CoverageCommand: "go test -cover ./...",
				Environment:     map[string]string{"GO_ENV": "test"},
			},
			setupMock: func(mock *MockDaggerContainer) {
				mock.SetCommandOutput("go test -cover ./...",
					"PASS\ncoverage: 87.5% of statements\nok\ttest\t0.005s",
					"", 0, nil)
			},
			expectedError:    false,
			expectedCoverage: 87.5,
			validateResult: func(t *testing.T, result *CoverageResult) {
				assert.Equal(t, 87.5, result.Coverage)
				assert.Equal(t, "text", result.ReportFormat)
				assert.Contains(t, result.Details["raw_output"], "87.5%")
			},
		},
		{
			name: "no coverage command configured",
			framework: &TestFramework{
				Name:            "no-coverage",
				CoverageCommand: "", // Empty coverage command
			},
			setupMock: func(mock *MockDaggerContainer) {
				// No setup needed
			},
			expectedError:    false,
			expectedCoverage: 0.0,
			validateResult: func(t *testing.T, result *CoverageResult) {
				assert.Equal(t, 0.0, result.Coverage)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)

			mockProvider := NewMockContainerProvider()
			tt.setupMock(mockProvider.MockContainer)

			engine := NewTestEngine(85, logger)
			engine.SetContainerProvider(mockProvider)

			container := &MockContainerWrapper{mockProvider.MockContainer}

			// Execute
			result, err := engine.runCoverageAnalysis(context.Background(), container, tt.framework)

			// Validate
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != nil {
				assert.Equal(t, tt.expectedCoverage, result.Coverage)
				tt.validateResult(t, result)
			}
		})
	}
}

func TestRunTestsIntegrationWithMocks(t *testing.T) {
	tests := []struct {
		name            string
		owner           string
		repo            string
		branch          string
		setupMock       func(*MockDaggerContainer)
		expectedSuccess bool
		minCoverage     int
	}{
		{
			name:   "successful integration test with good coverage",
			owner:  "test-owner",
			repo:   "test-repo",
			branch: "main",
			setupMock: func(mock *MockDaggerContainer) {
				// Clear default files and setup Go project
				mock.FileSystem = make(map[string]string)
				mock.SetFileContent("go.mod", "module test\ngo 1.19")
				
				// Setup successful commands - make sure to match actual framework commands
				mock.SetCommandOutput("go test ./...",
					"PASS: 5 passed, 0 failed\ncoverage: 90.0% of statements\nok\ttest\t0.005s",
					"", 0, nil)
				mock.SetCommandOutput("go build ./...",
					"Build successful", "", 0, nil)
				mock.SetCommandOutput("go test -coverprofile=coverage.out ./...",
					"PASS\ncoverage: 90.0% of statements\nok\ttest\t0.005s",
					"", 0, nil)
				mock.SetCommandOutput("golangci-lint run",
					"No linting issues found", "", 0, nil)
			},
			expectedSuccess: true,
			minCoverage:     85,
		},
		{
			name:   "integration test with insufficient coverage",
			owner:  "test-owner",
			repo:   "test-repo", 
			branch: "main",
			setupMock: func(mock *MockDaggerContainer) {
				// Clear default files and setup Go project
				mock.FileSystem = make(map[string]string)
				mock.SetFileContent("go.mod", "module test\ngo 1.19")
				
				// Setup commands with low coverage
				mock.SetCommandOutput("go test ./...",
					"PASS: 3 passed, 0 failed\ncoverage: 60.0% of statements\nok\ttest\t0.005s",
					"", 0, nil)
				mock.SetCommandOutput("go build ./...",
					"Build successful", "", 0, nil)
				mock.SetCommandOutput("go test -coverprofile=coverage.out ./...",
					"PASS\ncoverage: 60.0% of statements\nok\ttest\t0.005s",
					"", 0, nil)
				mock.SetCommandOutput("golangci-lint run",
					"No linting issues found", "", 0, nil)
			},
			expectedSuccess: false,
			minCoverage:     85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)

			mockProvider := NewMockContainerProvider()
			tt.setupMock(mockProvider.MockContainer)

			engine := NewTestEngine(tt.minCoverage, logger)
			engine.SetContainerProvider(mockProvider)

			// Execute
			result, err := engine.RunTests(context.Background(), tt.owner, tt.repo, tt.branch)

			// Validate
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedSuccess, result.Success)

			// Validate mock behavior
			mock := mockProvider.MockContainer
			assert.Equal(t, "ubuntu:22.04", mock.BaseImage)
			assert.Equal(t, "/workspace", mock.WorkingDir)
			assert.True(t, len(mock.ExecHistory) > 0, "Expected exec commands")
		})
	}
}