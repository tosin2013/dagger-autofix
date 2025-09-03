package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dagger.io/dagger"
	"github.com/sirupsen/logrus"
)

// TestEngine handles automated testing and validation of fixes
type TestEngine struct {
	minCoverage    int
	logger         *logrus.Logger
	testFrameworks map[string]*TestFramework
	coverageTools  map[string]*CoverageTool
}

// TestFramework defines testing capabilities for a specific language/framework
type TestFramework struct {
	Name            string            `json:"name"`
	Language        string            `json:"language"`
	Framework       string            `json:"framework"`
	TestCommand     string            `json:"test_command"`
	CoverageCommand string            `json:"coverage_command"`
	BuildCommand    string            `json:"build_command"`
	LintCommand     string            `json:"lint_command"`
	ConfigFiles     []string          `json:"config_files"`
	Environment     map[string]string `json:"environment"`
}

// CoverageTool defines coverage analysis capabilities
type CoverageTool struct {
	Name          string             `json:"name"`
	Languages     []string           `json:"languages"`
	Command       string             `json:"command"`
	ReportFormats []string           `json:"report_formats"`
	Thresholds    CoverageThresholds `json:"thresholds"`
}

// CoverageThresholds defines minimum coverage requirements
type CoverageThresholds struct {
	Line      float64 `json:"line"`
	Branch    float64 `json:"branch"`
	Function  float64 `json:"function"`
	Statement float64 `json:"statement"`
}

// NewTestEngine creates a new test engine with specified minimum coverage
func NewTestEngine(minCoverage int, logger *logrus.Logger) *TestEngine {
	return &TestEngine{
		minCoverage:    minCoverage,
		logger:         logger,
		testFrameworks: loadTestFrameworks(),
		coverageTools:  loadCoverageTools(),
	}
}

// RunTests executes the test suite for a given repository and branch
func (e *TestEngine) RunTests(ctx context.Context, owner, repo, branch string) (*TestResult, error) {
	start := time.Now()
	e.logger.WithFields(logrus.Fields{
		"owner":  owner,
		"repo":   repo,
		"branch": branch,
	}).Info("Starting test execution")

	// Create test container
	testContainer, err := e.createTestContainer(ctx, owner, repo, branch)
	if err != nil {
		return nil, fmt.Errorf("failed to create test container: %w", err)
	}

	// Detect project type and framework
	framework, err := e.detectFramework(ctx, testContainer)
	if err != nil {
		return nil, fmt.Errorf("failed to detect framework: %w", err)
	}

	e.logger.WithField("framework", framework.Name).Info("Detected test framework")

	// Run linting
	lintResult, err := e.runLinting(ctx, testContainer, framework)
	if err != nil {
		e.logger.WithError(err).Warn("Linting failed")
	}

	// Run build
	buildResult, err := e.runBuild(ctx, testContainer, framework)
	if err != nil {
		return &TestResult{
			Success:  false,
			Duration: time.Since(start),
			Output:   "Build failed",
			Errors:   []string{err.Error()},
			Details: map[string]interface{}{
				"stage":     "build",
				"framework": framework.Name,
				"lint":      lintResult,
			},
		}, nil
	}

	// Run tests
	testOutput, err := e.runTestSuite(ctx, testContainer, framework)
	if err != nil {
		return &TestResult{
			Success:  false,
			Duration: time.Since(start),
			Output:   testOutput,
			Errors:   []string{err.Error()},
			Details: map[string]interface{}{
				"stage":     "test",
				"framework": framework.Name,
				"lint":      lintResult,
				"build":     buildResult,
			},
		}, nil
	}

	// Run coverage analysis
	coverageResult, err := e.runCoverageAnalysis(ctx, testContainer, framework)
	if err != nil {
		e.logger.WithError(err).Warn("Coverage analysis failed")
		coverageResult = &CoverageResult{Coverage: 0.0}
	}

	// Parse test results
	testStats := e.parseTestOutput(testOutput, framework)

	result := &TestResult{
		Success:      testStats.Passed > 0 && coverageResult.Coverage >= float64(e.minCoverage),
		TotalTests:   testStats.Total,
		PassedTests:  testStats.Passed,
		FailedTests:  testStats.Failed,
		SkippedTests: testStats.Skipped,
		Coverage:     coverageResult.Coverage,
		Duration:     time.Since(start),
		Output:       testOutput,
		Details: map[string]interface{}{
			"framework":       framework.Name,
			"lint":            lintResult,
			"build":           buildResult,
			"coverage_detail": coverageResult,
			"min_coverage":    e.minCoverage,
		},
	}

	if testStats.Failed > 0 {
		result.Errors = append(result.Errors, "Some tests failed")
	}

	if coverageResult.Coverage < float64(e.minCoverage) {
		result.Errors = append(result.Errors, fmt.Sprintf("Coverage %.2f%% below minimum %.2f%%", coverageResult.Coverage, float64(e.minCoverage)))
	}

	e.logger.WithFields(logrus.Fields{
		"success":      result.Success,
		"total_tests":  result.TotalTests,
		"passed_tests": result.PassedTests,
		"coverage":     result.Coverage,
		"duration":     result.Duration,
	}).Info("Test execution completed")

	return result, nil
}

// ValidateTestCoverage validates that test coverage meets minimum requirements
func (e *TestEngine) ValidateTestCoverage(ctx context.Context, coverage *CoverageResult) error {
	if coverage.Coverage < float64(e.minCoverage) {
		return fmt.Errorf("coverage %.2f%% is below minimum required %.2f%%", coverage.Coverage, float64(e.minCoverage))
	}
	return nil
}

// GenerateTestsForFix generates additional tests to validate a specific fix
func (e *TestEngine) GenerateTestsForFix(ctx context.Context, fix *ProposedFix, analysis *FailureAnalysisResult) ([]string, error) {
	e.logger.WithField("fix_id", fix.ID).Info("Generating tests for fix")

	var tests []string

	// Generate tests based on fix type
	switch fix.Type {
	case CodeFix:
		tests = append(tests, e.generateCodeFixTests(fix, analysis)...)
	case DependencyFix:
		tests = append(tests, e.generateDependencyTests(fix, analysis)...)
	case ConfigurationFix:
		tests = append(tests, e.generateConfigurationTests(fix, analysis)...)
	case TestFix:
		tests = append(tests, e.generateTestFixTests(fix, analysis)...)
	}

	// Add regression tests
	tests = append(tests, e.generateRegressionTests(fix, analysis)...)

	e.logger.WithFields(logrus.Fields{
		"fix_id":     fix.ID,
		"test_count": len(tests),
	}).Info("Test generation completed")

	return tests, nil
}

// Private helper methods

func (e *TestEngine) createTestContainer(ctx context.Context, owner, repo, branch string) (*dagger.Container, error) {
	// Create a container for testing
	repoURL := fmt.Sprintf("https://github.com/%s/%s", owner, repo)

	container := dag.Container().
		From("ubuntu:22.04").
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "git", "curl", "wget", "build-essential"}).
		WithExec([]string{"git", "clone", "-b", branch, repoURL, "/workspace"}).
		WithWorkdir("/workspace")

	return container, nil
}

func (e *TestEngine) detectFramework(ctx context.Context, container *dagger.Container) (*TestFramework, error) {
	// Check for various framework indicators
	files := []string{"package.json", "go.mod", "pom.xml", "requirements.txt", "Cargo.toml", "composer.json"}

	for _, file := range files {
		_, err := container.File(file).Contents(ctx)
		if err == nil {
			return e.getFrameworkByFile(file), nil
		}
	}

	// Default to generic framework
	return e.testFrameworks["generic"], nil
}

func (e *TestEngine) getFrameworkByFile(filename string) *TestFramework {
	if filename == "" {
		return nil
	}

	switch filename {
	case "package.json":
		return e.testFrameworks["nodejs"]
	case "go.mod":
		return e.testFrameworks["golang"]
	case "pom.xml":
		return e.testFrameworks["maven"]
	case "requirements.txt":
		return e.testFrameworks["python"]
	case "Cargo.toml":
		return e.testFrameworks["rust"]
	case "composer.json":
		return e.testFrameworks["php"]
	case "Makefile", "makefile":
		return e.testFrameworks["generic"]
	default:
		// For truly unknown files, check if it's a generic build file pattern
		if strings.HasSuffix(filename, ".file") || strings.Contains(filename, "Makefile") {
			return e.testFrameworks["generic"]
		}
		// Return nil for unknown file extensions
		return nil
	}
}

func (e *TestEngine) runLinting(ctx context.Context, container *dagger.Container, framework *TestFramework) (string, error) {
	if framework.LintCommand == "" {
		return "No linting configured", nil
	}

	e.logger.WithField("command", framework.LintCommand).Debug("Running linting")

	// Setup environment
	for key, value := range framework.Environment {
		container = container.WithEnvVariable(key, value)
	}

	output, err := container.WithExec(strings.Split(framework.LintCommand, " ")).Stdout(ctx)
	if err != nil {
		return output, fmt.Errorf("linting failed: %w", err)
	}

	return output, nil
}

func (e *TestEngine) runBuild(ctx context.Context, container *dagger.Container, framework *TestFramework) (string, error) {
	if framework.BuildCommand == "" {
		return "No build configured", nil
	}

	e.logger.WithField("command", framework.BuildCommand).Debug("Running build")

	// Setup environment
	for key, value := range framework.Environment {
		container = container.WithEnvVariable(key, value)
	}

	output, err := container.WithExec(strings.Split(framework.BuildCommand, " ")).Stdout(ctx)
	if err != nil {
		return output, fmt.Errorf("build failed: %w", err)
	}

	return output, nil
}

func (e *TestEngine) runTestSuite(ctx context.Context, container *dagger.Container, framework *TestFramework) (string, error) {
	e.logger.WithField("command", framework.TestCommand).Debug("Running test suite")

	// Setup environment
	for key, value := range framework.Environment {
		container = container.WithEnvVariable(key, value)
	}

	output, err := container.WithExec(strings.Split(framework.TestCommand, " ")).Stdout(ctx)
	if err != nil {
		return output, fmt.Errorf("tests failed: %w", err)
	}

	return output, nil
}

type CoverageResult struct {
	Coverage     float64                `json:"coverage"`
	Details      map[string]interface{} `json:"details"`
	ReportFormat string                 `json:"report_format"`
}

func (e *TestEngine) runCoverageAnalysis(ctx context.Context, container *dagger.Container, framework *TestFramework) (*CoverageResult, error) {
	if framework.CoverageCommand == "" {
		return &CoverageResult{Coverage: 0.0}, nil
	}

	e.logger.WithField("command", framework.CoverageCommand).Debug("Running coverage analysis")

	// Setup environment
	for key, value := range framework.Environment {
		container = container.WithEnvVariable(key, value)
	}

	output, err := container.WithExec(strings.Split(framework.CoverageCommand, " ")).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("coverage analysis failed: %w", err)
	}

	// Parse coverage from output (simplified)
	coverage := e.parseCoverageOutput(output, framework)

	return &CoverageResult{
		Coverage:     coverage,
		ReportFormat: "text",
		Details: map[string]interface{}{
			"raw_output": output,
			"framework":  framework.Name,
		},
	}, nil
}

type TestStats struct {
	Total   int
	Passed  int
	Failed  int
	Skipped int
}

func (e *TestEngine) parseTestOutput(output string, framework *TestFramework) TestStats {
	// Framework-specific test parsing
	stats := TestStats{}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Go test output parsing
		if strings.Contains(line, "PASS:") || strings.Contains(line, "FAIL:") {
			if strings.Contains(line, "PASS:") {
				stats.Passed++
			} else if strings.Contains(line, "FAIL:") {
				stats.Failed++
			}
			stats.Total++
		}

		// Jest output parsing
		if strings.Contains(line, "Tests:") {
			// Example: "Tests:       5 passed, 2 failed, 7 total"
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "passed," && i > 0 {
					if val, err := strconv.Atoi(parts[i-1]); err == nil {
						stats.Passed = val
					}
				}
				if part == "failed," && i > 0 {
					if val, err := strconv.Atoi(parts[i-1]); err == nil {
						stats.Failed = val
					}
				}
				if part == "total" && i > 0 {
					if val, err := strconv.Atoi(parts[i-1]); err == nil {
						stats.Total = val
					}
				}
			}
		}

		// Individual test result lines for Go
		if (strings.Contains(line, "PASS") || strings.Contains(line, "FAIL")) &&
			(strings.Contains(line, "Test") || strings.Contains(line, "Example")) {
			// Only count if not already counted by summary lines
			continue
		}
	}

	// If no summary found, calculate total
	if stats.Total == 0 && (stats.Passed > 0 || stats.Failed > 0) {
		stats.Total = stats.Passed + stats.Failed + stats.Skipped
	}

	return stats
}

func (e *TestEngine) parseCoverageOutput(output string, framework *TestFramework) float64 {
	// Framework-specific coverage parsing
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Go coverage output: "coverage: 75.5% of statements"
		if strings.Contains(line, "coverage:") && strings.Contains(line, "% of statements") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasSuffix(part, "%") {
					percentStr := strings.TrimSuffix(part, "%")
					if val, err := strconv.ParseFloat(percentStr, 64); err == nil {
						return val
					}
				}
			}
		}

		// Jest coverage output: "All files      |   85.25 |"
		if strings.Contains(line, "All files") && strings.Contains(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 2 {
				percentStr := strings.TrimSpace(parts[1])
				if val, err := strconv.ParseFloat(percentStr, 64); err == nil {
					return val
				}
			}
		}

		// Python coverage output: "TOTAL          92%"
		if strings.Contains(line, "TOTAL") && strings.Contains(line, "%") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasSuffix(part, "%") {
					percentStr := strings.TrimSuffix(part, "%")
					if val, err := strconv.ParseFloat(percentStr, 64); err == nil {
						return val
					}
				}
			}
		}
	}
	return 0.0
}

func (e *TestEngine) generateCodeFixTests(fix *ProposedFix, analysis *FailureAnalysisResult) []string {
	var tests []string

	// Generate unit tests for changed functions
	for _, change := range fix.Changes {
		if change.Operation == "modify" || change.Operation == "add" {
			tests = append(tests, fmt.Sprintf(`
// Test for changes in %s
func Test_%s_Fix(t *testing.T) {
	// Test the fixed functionality
	// TODO: Implement specific test logic
}
`, change.FilePath, strings.ReplaceAll(change.FilePath, "/", "_")))
		}
	}

	return tests
}

func (e *TestEngine) generateDependencyTests(fix *ProposedFix, analysis *FailureAnalysisResult) []string {
	return []string{
		`// Test dependency installation
func TestDependencyInstallation(t *testing.T) {
	// Verify dependencies are correctly installed
	// TODO: Implement dependency validation
}`,
	}
}

func (e *TestEngine) generateConfigurationTests(fix *ProposedFix, analysis *FailureAnalysisResult) []string {
	return []string{
		`// Test configuration validity
func TestConfigurationValidity(t *testing.T) {
	// Verify configuration is valid
	// TODO: Implement configuration validation
}`,
	}
}

func (e *TestEngine) generateTestFixTests(fix *ProposedFix, analysis *FailureAnalysisResult) []string {
	return []string{
		`// Test for test fixes
func TestFixedTests(t *testing.T) {
	// Verify previously failing tests now pass
	// TODO: Implement test validation
}`,
	}
}

func (e *TestEngine) generateRegressionTests(fix *ProposedFix, analysis *FailureAnalysisResult) []string {
	return []string{
		`// Regression test
func TestRegression_` + analysis.ID + `(t *testing.T) {
	// Ensure the original failure doesn't reoccur
	// TODO: Implement regression test logic
}`,
	}
}

// Load predefined test frameworks and coverage tools

func loadTestFrameworks() map[string]*TestFramework {
	return map[string]*TestFramework{
		"nodejs": {
			Name:            "nodejs",
			Language:        "javascript",
			Framework:       "npm",
			TestCommand:     "npm test",
			CoverageCommand: "npm run coverage",
			BuildCommand:    "npm run build",
			LintCommand:     "npm run lint",
			ConfigFiles:     []string{"package.json", "jest.config.js", ".eslintrc.js"},
			Environment: map[string]string{
				"NODE_ENV": "test",
			},
		},
		"golang": {
			Name:            "golang",
			Language:        "go",
			Framework:       "go",
			TestCommand:     "go test ./...",
			CoverageCommand: "go test -coverprofile=coverage.out ./...",
			BuildCommand:    "go build ./...",
			LintCommand:     "golangci-lint run",
			ConfigFiles:     []string{"go.mod", "go.sum"},
			Environment: map[string]string{
				"GO111MODULE": "on",
				"CGO_ENABLED": "0",
			},
		},
		"python": {
			Name:            "python",
			Language:        "python",
			Framework:       "pytest",
			TestCommand:     "pytest",
			CoverageCommand: "pytest --cov=.",
			BuildCommand:    "pip install -e .",
			LintCommand:     "flake8",
			ConfigFiles:     []string{"requirements.txt", "setup.py", "pyproject.toml", "pytest.ini"},
			Environment: map[string]string{
				"PYTHONPATH": ".",
			},
		},
		"maven": {
			Name:            "maven",
			Language:        "java",
			Framework:       "maven",
			TestCommand:     "mvn test",
			CoverageCommand: "mvn jacoco:report",
			BuildCommand:    "mvn compile",
			LintCommand:     "mvn checkstyle:check",
			ConfigFiles:     []string{"pom.xml"},
			Environment:     map[string]string{},
		},
		"rust": {
			Name:            "rust",
			Language:        "rust",
			Framework:       "cargo",
			TestCommand:     "cargo test",
			CoverageCommand: "cargo tarpaulin --out xml",
			BuildCommand:    "cargo build",
			LintCommand:     "cargo clippy",
			ConfigFiles:     []string{"Cargo.toml", "Cargo.lock"},
			Environment:     map[string]string{},
		},
		"php": {
			Name:            "php",
			Language:        "php",
			Framework:       "phpunit",
			TestCommand:     "./vendor/bin/phpunit",
			CoverageCommand: "./vendor/bin/phpunit --coverage-xml coverage",
			BuildCommand:    "composer dump-autoload --optimize",
			LintCommand:     "./vendor/bin/phpcs",
			ConfigFiles:     []string{"composer.json"},
			Environment:     map[string]string{},
		},
		"generic": {
			Name:            "generic",
			Language:        "unknown",
			Framework:       "make",
			TestCommand:     "make test",
			CoverageCommand: "make coverage",
			BuildCommand:    "make build",
			LintCommand:     "make lint",
			ConfigFiles:     []string{"Makefile", "makefile"},
			Environment:     map[string]string{},
		},
	}
}

func loadCoverageTools() map[string]*CoverageTool {
	return map[string]*CoverageTool{
		"jest": {
			Name:          "Jest",
			Languages:     []string{"javascript", "typescript"},
			Command:       "jest --coverage",
			ReportFormats: []string{"lcov", "json", "text"},
			Thresholds: CoverageThresholds{
				Line:      80.0,
				Branch:    75.0,
				Function:  80.0,
				Statement: 80.0,
			},
		},
		"go-cover": {
			Name:          "Go Coverage",
			Languages:     []string{"go"},
			Command:       "go test -coverprofile=coverage.out",
			ReportFormats: []string{"text", "html"},
			Thresholds: CoverageThresholds{
				Line:      85.0,
				Branch:    80.0,
				Function:  85.0,
				Statement: 85.0,
			},
		},
		"pytest-cov": {
			Name:          "pytest-cov",
			Languages:     []string{"python"},
			Command:       "pytest --cov=.",
			ReportFormats: []string{"term", "html", "xml"},
			Thresholds: CoverageThresholds{
				Line:      85.0,
				Branch:    80.0,
				Function:  85.0,
				Statement: 85.0,
			},
		},
	}
}
