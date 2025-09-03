package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// FailureAnalysisEngine analyzes CI/CD failures using LLM-powered intelligence
type FailureAnalysisEngine struct {
	llmClient LLMClientInterface
	logger    *logrus.Logger
	patterns  *ErrorPatternDatabase
	prompts   *PromptTemplates
}

// ErrorPatternDatabase contains known error patterns and their solutions
type ErrorPatternDatabase struct {
	Patterns map[string]*ErrorPatternRule `json:"patterns"`
}

// ErrorPatternRule defines a rule for matching and categorizing errors
type ErrorPatternRule struct {
	Pattern     string          `json:"pattern"`
	Type        FailureType     `json:"type"`
	Category    FailureCategory `json:"category"`
	Severity    SeverityLevel   `json:"severity"`
	Description string          `json:"description"`
	Solutions   []string        `json:"solutions"`
	Confidence  float64         `json:"confidence"`
	Tags        []string        `json:"tags"`
}

// PromptTemplates contains templates for different types of analysis
type PromptTemplates struct {
	FailureAnalysis  string `json:"failure_analysis"`
	CodeAnalysis     string `json:"code_analysis"`
	FixGeneration    string `json:"fix_generation"`
	TestGeneration   string `json:"test_generation"`
	SecurityAnalysis string `json:"security_analysis"`
}

// NewFailureAnalysisEngine creates a new failure analysis engine
func NewFailureAnalysisEngine(llmClient LLMClientInterface, logger *logrus.Logger) *FailureAnalysisEngine {
	return &FailureAnalysisEngine{
		llmClient: llmClient,
		logger:    logger,
		patterns:  loadErrorPatterns(),
		prompts:   loadPromptTemplates(),
	}
}

// AnalyzeFailure performs comprehensive failure analysis using LLM
func (e *FailureAnalysisEngine) AnalyzeFailure(ctx context.Context, failureCtx FailureContext) (*FailureAnalysisResult, error) {
	start := time.Now()
	e.logger.WithField("run_id", failureCtx.WorkflowRun.ID).Info("Starting failure analysis")

	// Step 1: Pre-classify using pattern matching
	preClassification := e.preClassifyFailure(failureCtx)

	// Step 2: Prepare comprehensive context for LLM
	analysisPrompt := e.buildAnalysisPrompt(failureCtx, preClassification)

	// Step 3: Analyze with LLM
	req := &LLMRequest{
		SystemMsg: e.prompts.FailureAnalysis,
		Prompt:    analysisPrompt,
		Context: map[string]interface{}{
			"failure_type": preClassification.Type,
			"repository":   failureCtx.Repository,
			"workflow_run": failureCtx.WorkflowRun,
		},
	}

	response, err := e.llmClient.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	// Step 4: Parse and structure the analysis result
	analysis, err := e.parseAnalysisResponse(response.Content, failureCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse analysis response: %w", err)
	}

	// Step 5: Enhance with pattern-based insights
	e.enhanceWithPatterns(analysis, preClassification)

	// Step 6: Finalize result
	analysis.ID = fmt.Sprintf("analysis-%d-%d", failureCtx.WorkflowRun.ID, time.Now().Unix())
	analysis.Context = failureCtx
	analysis.Timestamp = time.Now()

	// Set LLM provider if available
	if realClient, ok := e.llmClient.(*LLMClient); ok {
		analysis.LLMProvider = realClient.provider
	} else {
		analysis.LLMProvider = "mock" // For testing
	}

	analysis.ProcessingTime = time.Since(start)

	e.logger.WithFields(logrus.Fields{
		"analysis_id":     analysis.ID,
		"failure_type":    analysis.Classification.Type,
		"confidence":      analysis.Classification.Confidence,
		"processing_time": analysis.ProcessingTime,
	}).Info("Failure analysis completed")

	return analysis, nil
}

// GenerateFixes generates multiple fix proposals for the analyzed failure
func (e *FailureAnalysisEngine) GenerateFixes(ctx context.Context, analysis *FailureAnalysisResult) ([]*ProposedFix, error) {
	e.logger.WithField("analysis_id", analysis.ID).Info("Generating fixes")

	// Build fix generation prompt
	fixPrompt := e.buildFixGenerationPrompt(analysis)

	req := &LLMRequest{
		SystemMsg: e.prompts.FixGeneration,
		Prompt:    fixPrompt,
		Context: map[string]interface{}{
			"analysis":     analysis,
			"failure_type": analysis.Classification.Type,
		},
	}

	response, err := e.llmClient.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fix generation failed: %w", err)
	}

	// Parse fixes from response
	fixes, err := e.parseFixesResponse(response.Content, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fixes response: %w", err)
	}

	// Enhance fixes with validation steps
	for _, fix := range fixes {
		e.addValidationSteps(fix, analysis)
	}

	e.logger.WithFields(logrus.Fields{
		"analysis_id": analysis.ID,
		"fixes_count": len(fixes),
	}).Info("Fix generation completed")

	return fixes, nil
}

// preClassifyFailure performs initial classification using pattern matching
func (e *FailureAnalysisEngine) preClassifyFailure(ctx FailureContext) *FailureClassification {
	var errorLines, allLogs string
	if ctx.Logs != nil {
		errorLines = strings.Join(ctx.Logs.ErrorLines, "\n")
		allLogs = ctx.Logs.RawLogs
	}

	// Check against known patterns, sorted by pattern length (longest first) for specificity
	type patternEntry struct {
		name string
		rule *ErrorPatternRule
	}
	var patterns []patternEntry
	for name, rule := range e.patterns.Patterns {
		patterns = append(patterns, patternEntry{name, rule})
	}
	// Sort by pattern length descending to match more specific patterns first
	for i := 0; i < len(patterns)-1; i++ {
		for j := i + 1; j < len(patterns); j++ {
			if len(patterns[i].rule.Pattern) < len(patterns[j].rule.Pattern) {
				patterns[i], patterns[j] = patterns[j], patterns[i]
			}
		}
	}

	for _, entry := range patterns {
		if strings.Contains(errorLines, entry.rule.Pattern) || strings.Contains(allLogs, entry.rule.Pattern) {
			e.logger.WithField("pattern", entry.name).Debug("Matched error pattern")
			return &FailureClassification{
				Type:       entry.rule.Type,
				Severity:   entry.rule.Severity,
				Category:   entry.rule.Category,
				Confidence: entry.rule.Confidence,
				Tags:       entry.rule.Tags,
			}
		}
	}

	// Default classification if no pattern matches
	return &FailureClassification{
		Type:       InfrastructureFailure,
		Severity:   Medium,
		Category:   Systematic,
		Confidence: 0.3,
		Tags:       []string{"unclassified"},
	}
}

// buildAnalysisPrompt creates a comprehensive prompt for failure analysis
func (e *FailureAnalysisEngine) buildAnalysisPrompt(ctx FailureContext, preClass *FailureClassification) string {
	var prompt strings.Builder

	prompt.WriteString("## GitHub Actions Workflow Failure Analysis\n\n")

	// Workflow context
	prompt.WriteString(fmt.Sprintf("**Workflow**: %s\n", ctx.WorkflowRun.Name))
	prompt.WriteString(fmt.Sprintf("**Branch**: %s\n", ctx.WorkflowRun.Branch))
	prompt.WriteString(fmt.Sprintf("**Commit**: %s\n", ctx.WorkflowRun.CommitSHA))
	prompt.WriteString(fmt.Sprintf("**Status**: %s/%s\n", ctx.WorkflowRun.Status, ctx.WorkflowRun.Conclusion))
	prompt.WriteString(fmt.Sprintf("**Repository**: %s/%s\n\n", ctx.Repository.Owner, ctx.Repository.Name))

	// Repository context
	if ctx.Repository.Language != "" {
		prompt.WriteString(fmt.Sprintf("**Language**: %s\n", ctx.Repository.Language))
	}
	if ctx.Repository.Framework != "" {
		prompt.WriteString(fmt.Sprintf("**Framework**: %s\n", ctx.Repository.Framework))
	}

	// Pre-classification
	prompt.WriteString(fmt.Sprintf("**Initial Classification**: %s (confidence: %.2f)\n\n", preClass.Type.DisplayName(), preClass.Confidence))

	// Error logs
	prompt.WriteString("## Error Information\n\n")
	if len(ctx.Logs.ErrorLines) > 0 {
		prompt.WriteString("**Error Lines**:\n```\n")
		for _, line := range ctx.Logs.ErrorLines {
			prompt.WriteString(line + "\n")
		}
		prompt.WriteString("```\n\n")
	}

	// Full logs (truncated if too long)
	if len(ctx.Logs.RawLogs) > 0 {
		prompt.WriteString("**Full Logs**:\n```\n")
		logContent := ctx.Logs.RawLogs
		if len(logContent) > 8000 {
			logContent = logContent[:4000] + "\n...\n[TRUNCATED]\n...\n" + logContent[len(logContent)-4000:]
		}
		prompt.WriteString(logContent)
		prompt.WriteString("\n```\n\n")
	}

	// Recent commits context
	if len(ctx.RecentCommits) > 0 {
		prompt.WriteString("## Recent Changes\n\n")
		for i, commit := range ctx.RecentCommits {
			if i >= 3 {
				break // Limit to 3 recent commits
			}
			shaShort := commit.SHA
			if len(shaShort) > 8 {
				shaShort = shaShort[:8]
			}
			prompt.WriteString(fmt.Sprintf("**Commit %s**: %s (by %s)\n", shaShort, commit.Message, commit.Author))
			for _, change := range commit.Changes {
				prompt.WriteString(fmt.Sprintf("  - %s: %s (+%d/-%d)\n", change.Status, change.Filename, change.Additions, change.Deletions))
			}
			prompt.WriteString("\n")
		}
	}

	prompt.WriteString("\n## Analysis Instructions\n\n")
	prompt.WriteString("Please provide a comprehensive analysis including:\n")
	prompt.WriteString("1. **Root Cause**: What exactly caused this failure?\n")
	prompt.WriteString("2. **Classification**: Confirm or adjust the failure type, severity, and category\n")
	prompt.WriteString("3. **Affected Components**: Which files, dependencies, or configurations are involved?\n")
	prompt.WriteString("4. **Error Patterns**: Identify specific error patterns and their meanings\n")
	prompt.WriteString("5. **Context Analysis**: How do recent changes relate to this failure?\n")
	prompt.WriteString("\nFormat your response as structured JSON with the required fields.\n")

	return prompt.String()
}

// buildFixGenerationPrompt creates a prompt for generating fixes
func (e *FailureAnalysisEngine) buildFixGenerationPrompt(analysis *FailureAnalysisResult) string {
	var prompt strings.Builder

	prompt.WriteString("## Fix Generation for CI/CD Failure\n\n")

	// Analysis summary
	prompt.WriteString(fmt.Sprintf("**Failure Type**: %s\n", analysis.Classification.Type.DisplayName()))
	prompt.WriteString(fmt.Sprintf("**Root Cause**: %s\n", analysis.RootCause))
	prompt.WriteString(fmt.Sprintf("**Description**: %s\n\n", analysis.Description))

	// Affected files
	if len(analysis.AffectedFiles) > 0 {
		prompt.WriteString("**Affected Files**:\n")
		for _, file := range analysis.AffectedFiles {
			prompt.WriteString(fmt.Sprintf("- %s\n", file))
		}
		prompt.WriteString("\n")
	}

	// Error patterns
	if len(analysis.ErrorPatterns) > 0 {
		prompt.WriteString("**Error Patterns**:\n")
		for _, pattern := range analysis.ErrorPatterns {
			prompt.WriteString(fmt.Sprintf("- %s: %s (confidence: %.2f)\n", pattern.Pattern, pattern.Description, pattern.Confidence))
		}
		prompt.WriteString("\n")
	}

	// Repository context
	prompt.WriteString(fmt.Sprintf("**Repository**: %s/%s\n", analysis.Context.Repository.Owner, analysis.Context.Repository.Name))
	if analysis.Context.Repository.Language != "" {
		prompt.WriteString(fmt.Sprintf("**Language**: %s\n", analysis.Context.Repository.Language))
	}
	if analysis.Context.Repository.Framework != "" {
		prompt.WriteString(fmt.Sprintf("**Framework**: %s\n\n", analysis.Context.Repository.Framework))
	}

	prompt.WriteString("## Fix Generation Instructions\n\n")
	prompt.WriteString("Generate 2-3 different fix proposals, each with:\n")
	prompt.WriteString("1. **Type**: The type of fix (code, configuration, dependency, etc.)\n")
	prompt.WriteString("2. **Description**: Clear description of what the fix does\n")
	prompt.WriteString("3. **Rationale**: Why this fix addresses the root cause\n")
	prompt.WriteString("4. **Changes**: Specific code/configuration changes needed\n")
	prompt.WriteString("5. **Confidence**: Your confidence level (0.0-1.0)\n")
	prompt.WriteString("6. **Risks**: Potential risks or side effects\n")
	prompt.WriteString("7. **Benefits**: Expected benefits\n")
	prompt.WriteString("\nOrder fixes by confidence level (highest first).\n")
	prompt.WriteString("Format response as JSON array of fix objects.\n")

	return prompt.String()
}

// parseAnalysisResponse parses the LLM response into a structured analysis result
func (e *FailureAnalysisEngine) parseAnalysisResponse(content string, ctx FailureContext) (*FailureAnalysisResult, error) {
	// Return error for empty content
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("empty response content")
	}

	// Try to extract JSON from the response
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")

	if jsonStart == -1 || jsonEnd == -1 {
		return e.parseUnstructuredAnalysis(content, ctx)
	}

	jsonContent := content[jsonStart : jsonEnd+1]
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &parsed); err != nil {
		return e.parseUnstructuredAnalysis(content, ctx)
	}

	// Extract structured data
	analysis := &FailureAnalysisResult{}

	if rootCause, ok := parsed["root_cause"].(string); ok {
		analysis.RootCause = rootCause
	}
	if description, ok := parsed["description"].(string); ok {
		analysis.Description = description
	}

	// Parse classification
	if classData, ok := parsed["classification"].(map[string]interface{}); ok {
		analysis.Classification = FailureClassification{
			Type:       FailureType(getStringField(classData, "type", string(CodeFailure))),
			Severity:   SeverityLevel(getStringField(classData, "severity", string(Medium))),
			Category:   FailureCategory(getStringField(classData, "category", string(Systematic))),
			Confidence: getFloatField(classData, "confidence", 0.7),
			Tags:       getStringArrayField(classData, "tags"),
		}
	}

	// Parse affected files
	if files, ok := parsed["affected_files"].([]interface{}); ok {
		for _, file := range files {
			if fileStr, ok := file.(string); ok {
				analysis.AffectedFiles = append(analysis.AffectedFiles, fileStr)
			}
		}
	}

	// Parse error patterns
	if patterns, ok := parsed["error_patterns"].([]interface{}); ok {
		for _, pattern := range patterns {
			if patternMap, ok := pattern.(map[string]interface{}); ok {
				analysis.ErrorPatterns = append(analysis.ErrorPatterns, ErrorPattern{
					Pattern:     getStringField(patternMap, "pattern", ""),
					Description: getStringField(patternMap, "description", ""),
					Confidence:  getFloatField(patternMap, "confidence", 0.5),
					Location:    getStringField(patternMap, "location", ""),
				})
			}
		}
	}

	return analysis, nil
}

// parseUnstructuredAnalysis parses unstructured LLM response
func (e *FailureAnalysisEngine) parseUnstructuredAnalysis(content string, ctx FailureContext) (*FailureAnalysisResult, error) {
	// Try to detect failure type from content
	contentLower := strings.ToLower(content)
	var failureType FailureType = CodeFailure

	// Simple pattern matching for unstructured content
	if strings.Contains(contentLower, "dependency") || strings.Contains(contentLower, "dependencies") ||
		strings.Contains(contentLower, "package") {
		failureType = DependencyFailure
	} else if strings.Contains(contentLower, "build") || strings.Contains(contentLower, "compilation") {
		failureType = BuildFailure
	} else if strings.Contains(contentLower, "test") {
		failureType = TestFailure
	} else if strings.Contains(contentLower, "infrastructure") || strings.Contains(contentLower, "network") ||
		strings.Contains(contentLower, "timeout") {
		failureType = InfrastructureFailure
	} else if strings.Contains(contentLower, "security") || strings.Contains(contentLower, "vulnerability") {
		failureType = SecurityFailure
	} else if strings.Contains(contentLower, "configuration") || strings.Contains(contentLower, "config") {
		failureType = ConfigurationFailure
	}

	// Fallback parsing for unstructured responses
	analysis := &FailureAnalysisResult{
		Description: content,
		RootCause:   "Analysis provided in description field",
		Classification: FailureClassification{
			Type:       failureType,
			Severity:   Medium,
			Category:   Systematic,
			Confidence: 0.5,
			Tags:       []string{"unstructured"},
		},
	}

	return analysis, nil
}

// parseFixesResponse parses fix generation response
func (e *FailureAnalysisEngine) parseFixesResponse(content string, analysis *FailureAnalysisResult) ([]*ProposedFix, error) {
	// Try to extract JSON array
	jsonStart := strings.Index(content, "[")
	jsonEnd := strings.LastIndex(content, "]")

	if jsonStart == -1 || jsonEnd == -1 {
		return e.parseUnstructuredFixes(content, analysis)
	}

	jsonContent := content[jsonStart : jsonEnd+1]
	var parsed []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &parsed); err != nil {
		return e.parseUnstructuredFixes(content, analysis)
	}

	var fixes []*ProposedFix
	for i, fixData := range parsed {
		fix := &ProposedFix{
			ID:          fmt.Sprintf("%s-fix-%d", analysis.ID, i+1),
			Type:        FixType(getStringField(fixData, "type", string(CodeFix))),
			Description: getStringField(fixData, "description", ""),
			Rationale:   getStringField(fixData, "rationale", ""),
			Confidence:  getFloatField(fixData, "confidence", 0.5),
			Risks:       getStringArrayField(fixData, "risks"),
			Benefits:    getStringArrayField(fixData, "benefits"),
			Timestamp:   time.Now(),
		}

		// Parse changes if present
		if changes, ok := fixData["changes"].([]interface{}); ok {
			for _, change := range changes {
				if changeMap, ok := change.(map[string]interface{}); ok {
					fix.Changes = append(fix.Changes, CodeChange{
						FilePath:    getStringField(changeMap, "file_path", ""),
						OldContent:  getStringField(changeMap, "old_content", ""),
						NewContent:  getStringField(changeMap, "new_content", ""),
						Operation:   getStringField(changeMap, "operation", "modify"),
						Explanation: getStringField(changeMap, "explanation", ""),
					})
				}
			}
		}

		fixes = append(fixes, fix)
	}

	return fixes, nil
}

// parseUnstructuredFixes creates a basic fix from unstructured content
func (e *FailureAnalysisEngine) parseUnstructuredFixes(content string, analysis *FailureAnalysisResult) ([]*ProposedFix, error) {
	fix := &ProposedFix{
		ID:          fmt.Sprintf("%s-fix-1", analysis.ID),
		Type:        CodeFix,
		Description: "Fix based on analysis",
		Rationale:   content,
		Confidence:  0.3,
		Timestamp:   time.Now(),
	}

	return []*ProposedFix{fix}, nil
}

// enhanceWithPatterns adds pattern-based insights to the analysis
func (e *FailureAnalysisEngine) enhanceWithPatterns(analysis *FailureAnalysisResult, preClass *FailureClassification) {
	// Use pre-classification if analysis confidence is low
	if analysis.Classification.Confidence < 0.6 && preClass.Confidence > 0.5 {
		analysis.Classification = *preClass
		analysis.Classification.Tags = append(analysis.Classification.Tags, "pattern-enhanced")
	}
}

// addValidationSteps adds appropriate validation steps to fixes
func (e *FailureAnalysisEngine) addValidationSteps(fix *ProposedFix, analysis *FailureAnalysisResult) {
	validationSteps := []ValidationStep{
		{
			Name:     "Syntax Check",
			Command:  "make lint",
			Expected: "No syntax errors",
			Timeout:  30 * time.Second,
		},
		{
			Name:     "Unit Tests",
			Command:  "make test",
			Expected: "All tests pass",
			Timeout:  5 * time.Minute,
		},
		{
			Name:     "Build Check",
			Command:  "make build",
			Expected: "Build succeeds",
			Timeout:  10 * time.Minute,
		},
	}

	// Add type-specific validation
	switch fix.Type {
	case DependencyFix:
		validationSteps = append(validationSteps, ValidationStep{
			Name:     "Dependency Check",
			Command:  "make deps-check",
			Expected: "Dependencies resolved",
			Timeout:  2 * time.Minute,
		})
	case SecurityFix:
		validationSteps = append(validationSteps, ValidationStep{
			Name:     "Security Scan",
			Command:  "make security-scan",
			Expected: "No security issues",
			Timeout:  3 * time.Minute,
		})
	}

	fix.Validation = validationSteps
}

// Helper functions for parsing

func getStringField(data map[string]interface{}, key, defaultValue string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return defaultValue
}

func getFloatField(data map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}
	return defaultValue
}

func getStringArrayField(data map[string]interface{}, key string) []string {
	if val, ok := data[key].([]interface{}); ok {
		var result []string
		for _, item := range val {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return []string{}
}

// loadErrorPatterns loads predefined error patterns
func loadErrorPatterns() *ErrorPatternDatabase {
	return &ErrorPatternDatabase{
		Patterns: map[string]*ErrorPatternRule{
			"connection_timeout": {
				Pattern:     "connection timeout",
				Type:        InfrastructureFailure,
				Category:    Environmental,
				Severity:    High,
				Description: "Network connection timeout",
				Solutions:   []string{"Check network connectivity", "Increase timeout", "Retry connection"},
				Confidence:  0.8,
				Tags:        []string{"network", "timeout", "infrastructure"},
			},
			"build_failure": {
				Pattern:     "build failed",
				Type:        BuildFailure,
				Category:    Systematic,
				Severity:    High,
				Description: "Build compilation failure",
				Solutions:   []string{"Check syntax errors", "Update dependencies", "Fix import paths"},
				Confidence:  0.8,
				Tags:        []string{"build", "compilation"},
			},
			"go_build_failure": {
				Pattern:     "go build",
				Type:        BuildFailure,
				Category:    Systematic,
				Severity:    High,
				Description: "Go compilation failure",
				Solutions:   []string{"Check syntax errors", "Update dependencies", "Fix import paths"},
				Confidence:  0.8,
				Tags:        []string{"go", "build", "compilation"},
			},
			"test_failure": {
				Pattern:     "test failed",
				Type:        TestFailure,
				Category:    Systematic,
				Severity:    Medium,
				Description: "Test execution failure",
				Solutions:   []string{"Fix test logic", "Update test expectations", "Check test dependencies"},
				Confidence:  0.8,
				Tags:        []string{"test", "failure"},
			},
			"npm_install_failure": {
				Pattern:     "npm install",
				Type:        DependencyFailure,
				Category:    Systematic,
				Severity:    High,
				Description: "NPM package installation failure",
				Solutions:   []string{"Check package.json", "Clear node_modules", "Update npm version"},
				Confidence:  0.8,
				Tags:        []string{"npm", "dependency", "nodejs"},
			},
			"test_timeout": {
				Pattern:     "timeout",
				Type:        TestFailure,
				Category:    Transient,
				Severity:    Medium,
				Description: "Test execution timeout",
				Solutions:   []string{"Increase timeout", "Optimize test performance", "Check for infinite loops"},
				Confidence:  0.7,
				Tags:        []string{"timeout", "test", "performance"},
			},
			"service_unavailable": {
				Pattern:     "service unavailable",
				Type:        InfrastructureFailure,
				Category:    Environmental,
				Severity:    High,
				Description: "Service availability issue",
				Solutions:   []string{"Check service status", "Verify endpoints", "Check dependencies"},
				Confidence:  0.8,
				Tags:        []string{"service", "availability", "infrastructure"},
			},
			"docker_build_failure": {
				Pattern:     "docker build",
				Type:        InfrastructureFailure,
				Category:    Environmental,
				Severity:    High,
				Description: "Docker image build failure",
				Solutions:   []string{"Check Dockerfile syntax", "Verify base image", "Check build context"},
				Confidence:  0.8,
				Tags:        []string{"docker", "containerization", "build"},
			},
			"memory_error": {
				Pattern:     "out of memory",
				Type:        InfrastructureFailure,
				Category:    Environmental,
				Severity:    Critical,
				Description: "Memory exhaustion error",
				Solutions:   []string{"Increase memory limits", "Optimize memory usage", "Check for memory leaks"},
				Confidence:  0.9,
				Tags:        []string{"memory", "resource", "performance"},
			},
			"security_vulnerability": {
				Pattern:     "security vulnerability",
				Type:        SecurityFailure,
				Category:    Systematic,
				Severity:    Critical,
				Description: "Security vulnerability detected",
				Solutions:   []string{"Update vulnerable dependencies", "Apply security patches", "Review security policies"},
				Confidence:  0.9,
				Tags:        []string{"security", "vulnerability"},
			},
			"insecure_dependency": {
				Pattern:     "insecure dependency",
				Type:        SecurityFailure,
				Category:    Systematic,
				Severity:    High,
				Description: "Insecure dependency detected",
				Solutions:   []string{"Update dependency", "Replace with secure alternative", "Add security controls"},
				Confidence:  0.8,
				Tags:        []string{"security", "dependency"},
			},
			"invalid_configuration": {
				Pattern:     "invalid configuration",
				Type:        ConfigurationFailure,
				Category:    Systematic,
				Severity:    Medium,
				Description: "Configuration validation failure",
				Solutions:   []string{"Fix configuration syntax", "Update config values", "Check config schema"},
				Confidence:  0.8,
				Tags:        []string{"configuration", "validation"},
			},
			"config_file_not_found": {
				Pattern:     "config file not found",
				Type:        ConfigurationFailure,
				Category:    Systematic,
				Severity:    Medium,
				Description: "Configuration file missing",
				Solutions:   []string{"Create missing config file", "Fix file path", "Check file permissions"},
				Confidence:  0.8,
				Tags:        []string{"configuration", "file", "missing"},
			},
		},
	}
}

// loadPromptTemplates loads prompt templates for different analysis types
func loadPromptTemplates() *PromptTemplates {
	return &PromptTemplates{
		FailureAnalysis: `You are an expert DevOps engineer and CI/CD specialist with deep knowledge of software development workflows, testing frameworks, and deployment pipelines. Your task is to analyze GitHub Actions workflow failures and provide comprehensive, actionable insights.

Analyze the provided failure information using a systematic approach:

1. **Root Cause Analysis**: Identify the primary cause of the failure by examining error messages, stack traces, and context
2. **Classification**: Categorize the failure type (infrastructure, code, test, dependency, build, deployment, configuration, security)
3. **Severity Assessment**: Determine the impact level (critical, high, medium, low)
4. **Category Determination**: Classify as transient, systematic, environmental, or flaky
5. **Pattern Recognition**: Identify recurring error patterns and their implications
6. **Context Correlation**: Connect the failure to recent code changes, dependencies, or environmental factors

Provide your analysis in a structured format that enables automated processing and decision-making.`,

		CodeAnalysis: `You are a senior software engineer specializing in code quality and static analysis. Analyze the provided code context to understand how it relates to the reported CI/CD failure.

Focus on:
- Code structure and patterns
- Potential bugs or logic errors
- Dependency issues
- Configuration problems
- Security vulnerabilities
- Performance concerns

Provide specific, actionable recommendations for code improvements.`,

		FixGeneration: `You are an expert software developer and DevOps engineer. Generate specific, implementable fixes for the analyzed CI/CD failure.

For each fix proposal:
1. Provide clear, specific code changes with exact file paths and line numbers when possible
2. Explain the rationale behind each change
3. Assess the confidence level of the fix
4. Identify potential risks and mitigation strategies
5. Suggest validation steps to verify the fix
6. Consider backwards compatibility and deployment implications

Generate multiple fix alternatives when possible, ordered by confidence and risk level.`,

		TestGeneration: `You are a test automation expert. Generate comprehensive test cases to validate the proposed fixes and prevent regression of the identified issues.

Include:
- Unit tests for specific code changes
- Integration tests for system interactions
- End-to-end tests for user workflows
- Performance tests if relevant
- Security tests if applicable

Ensure tests are maintainable, reliable, and follow best practices for the target language and framework.`,

		SecurityAnalysis: `You are a cybersecurity expert specializing in application security and DevSecOps. Analyze the failure context for security implications and vulnerabilities.

Focus on:
- Security misconfigurations
- Dependency vulnerabilities
- Code injection risks
- Authentication/authorization issues
- Data exposure risks
- Infrastructure security

Provide specific security recommendations and remediation steps.`,
	}
}
