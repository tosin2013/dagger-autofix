package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// PullRequestEngine handles automated pull request creation and management
type PullRequestEngine struct {
	githubClient GitHubClient
	logger       *logrus.Logger
	templates    *PRTemplates
}

// PRTemplates contains templates for pull request content
type PRTemplates struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	CommitMsg string   `json:"commit_message"`
	Labels    []string `json:"labels"`
	Reviewers []string `json:"reviewers"`
}

// NewPullRequestEngine creates a new pull request engine
func NewPullRequestEngine(githubClient GitHubClient, logger *logrus.Logger) *PullRequestEngine {
	return &PullRequestEngine{
		githubClient: githubClient,
		logger:       logger,
		templates:    loadPRTemplates(),
	}
}

// CreateFixPR creates a pull request for an automated fix
func (p *PullRequestEngine) CreateFixPR(ctx context.Context, analysis *FailureAnalysisResult, fix *FixValidationResult) (*PullRequest, error) {
	if !fix.Valid {
		return nil, fmt.Errorf("cannot create PR for invalid fix")
	}

	p.logger.WithFields(logrus.Fields{
		"analysis_id": analysis.ID,
		"fix_id":      fix.Fix.ID,
	}).Info("Creating automated fix pull request")

	// Generate branch name
	branchName := p.generateBranchName(analysis, fix.Fix)

	// Create branch with changes
	if err := p.createBranch(ctx, branchName, fix.Fix.Changes); err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	// Generate PR content
	prOptions := p.generatePRContent(analysis, fix)
	prOptions.BranchName = branchName

	// Create pull request
	pr, err := p.createPullRequest(ctx, prOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	// Add additional metadata
	if err := p.addPRMetadata(ctx, pr, analysis, fix); err != nil {
		p.logger.WithError(err).Warn("Failed to add PR metadata")
	}

	p.logger.WithFields(logrus.Fields{
		"pr_number": pr.Number,
		"pr_url":    pr.URL,
		"branch":    branchName,
	}).Info("Pull request created successfully")

	return pr, nil
}

// CreateManualPR creates a pull request for manual review
func (p *PullRequestEngine) CreateManualPR(ctx context.Context, analysis *FailureAnalysisResult, options *PRCreationOptions) (*PullRequest, error) {
	p.logger.WithField("analysis_id", analysis.ID).Info("Creating manual review pull request")

	// Create pull request
	pr, err := p.createPullRequest(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"pr_number": pr.Number,
		"pr_url":    pr.URL,
	}).Info("Manual pull request created successfully")

	return pr, nil
}

// UpdatePR updates an existing pull request
func (p *PullRequestEngine) UpdatePR(ctx context.Context, prNumber int, updates *PRCreationOptions) (*PullRequest, error) {
	p.logger.WithField("pr_number", prNumber).Info("Updating pull request")

	// Convert PRCreationOptions to PRUpdateOptions
	updateOptions := &PRUpdateOptions{
		Title:  &updates.Title,
		Body:   &updates.Body,
		Labels: updates.Labels,
	}

	// Update PR using the interface
	result, err := p.githubClient.UpdatePullRequest(ctx, prNumber, updateOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to update PR: %w", err)
	}

	return result, nil
}

// ClosePR closes a pull request
func (p *PullRequestEngine) ClosePR(ctx context.Context, prNumber int, reason string) error {
	p.logger.WithFields(logrus.Fields{
		"pr_number": prNumber,
		"reason":    reason,
	}).Info("Closing pull request")

	// Close PR using the interface
	err := p.githubClient.ClosePullRequest(ctx, prNumber)
	if err != nil {
		return fmt.Errorf("failed to close PR: %w", err)
	}

	// Add closing comment
	if err := p.githubClient.AddPullRequestComment(ctx, prNumber, reason); err != nil {
		p.logger.WithError(err).Warn("Failed to add closing comment")
	}

	return nil
}

// GetPRStatus gets the status of a pull request
func (p *PullRequestEngine) GetPRStatus(ctx context.Context, prNumber int) (*PullRequest, error) {
	pr, err := p.githubClient.GetPullRequest(ctx, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	return pr, nil
}

// Private helper methods

func (p *PullRequestEngine) generateBranchName(analysis *FailureAnalysisResult, fix *ProposedFix) string {
	timestamp := time.Now().Format("20060102-150405")
	fixType := strings.ToLower(string(fix.Type))
	return fmt.Sprintf("autofix/%s/%s-%s", fixType, analysis.ID, timestamp)
}

func (p *PullRequestEngine) createBranch(ctx context.Context, branchName string, changes []CodeChange) error {
	p.logger.WithField("branch", branchName).Debug("Creating branch with changes")

	// Use the GitHubClient interface to create test branch with changes
	cleanup, err := p.githubClient.CreateTestBranch(ctx, branchName, changes)
	if err != nil {
		return fmt.Errorf("failed to create branch with changes: %w", err)
	}

	// Store cleanup function for later use (could be stored in a map if needed)
	_ = cleanup // For now, we don't call cleanup as the branch should persist for the PR

	return nil
}

// applyChange is no longer needed as we use GitHubClient.CreateTestBranch
// which handles all change operations internally
func (p *PullRequestEngine) applyChange(ctx context.Context, branch string, change CodeChange) error {
	// This method is deprecated - changes are now applied via GitHubClient.CreateTestBranch
	return fmt.Errorf("applyChange is deprecated, use GitHubClient.CreateTestBranch instead")
}

// createFile is no longer needed - handled by GitHubClient.CreateTestBranch
func (p *PullRequestEngine) createFile(ctx context.Context, branch string, change CodeChange) error {
	return fmt.Errorf("createFile is deprecated, use GitHubClient.CreateTestBranch instead")
}

// updateFile is no longer needed - handled by GitHubClient.CreateTestBranch
func (p *PullRequestEngine) updateFile(ctx context.Context, branch string, change CodeChange) error {
	return fmt.Errorf("updateFile is deprecated, use GitHubClient.CreateTestBranch instead")
}

// deleteFile is no longer needed - handled by GitHubClient.CreateTestBranch
func (p *PullRequestEngine) deleteFile(ctx context.Context, branch string, change CodeChange) error {
	return fmt.Errorf("deleteFile is deprecated, use GitHubClient.CreateTestBranch instead")
}

func (p *PullRequestEngine) generatePRContent(analysis *FailureAnalysisResult, fix *FixValidationResult) *PRCreationOptions {
	// Generate title
	title := p.generatePRTitle(analysis, fix.Fix)

	// Generate body
	body := p.generatePRBody(analysis, fix)

	// Generate labels
	labels := p.generatePRLabels(analysis, fix.Fix)

	return &PRCreationOptions{
		Title:        title,
		Body:         body,
		Labels:       labels,
		TargetBranch: "main",
		Draft:        false,
		AutoMerge:    false,
		DeleteBranch: true,
	}
}

func (p *PullRequestEngine) generatePRTitle(analysis *FailureAnalysisResult, fix *ProposedFix) string {
	caser := cases.Title(language.English)
	fixType := caser.String(string(fix.Type))
	failureType := caser.String(string(analysis.Classification.Type))

	runID := int64(0)
	if analysis.Context.WorkflowRun != nil {
		runID = analysis.Context.WorkflowRun.ID
	}

	return fmt.Sprintf("🤖 Auto-fix: %s for %s failure (Run #%d)", fixType, failureType, runID)
}

func (p *PullRequestEngine) generatePRBody(analysis *FailureAnalysisResult, fix *FixValidationResult) string {
	var body strings.Builder

	body.WriteString("## 🤖 Automated Fix\n\n")
	body.WriteString("This pull request was automatically generated to fix a CI/CD pipeline failure.\n\n")

	// Failure summary
	body.WriteString("## 📊 Failure Analysis\n\n")
	if analysis.Context.WorkflowRun != nil {
		body.WriteString(fmt.Sprintf("**Workflow Run**: [#%d](%s)\n", analysis.Context.WorkflowRun.ID, analysis.Context.WorkflowRun.URL))
	} else {
		body.WriteString("**Workflow Run**: Not available\n")
	}
	body.WriteString(fmt.Sprintf("**Failure Type**: %s\n", analysis.Classification.Type))
	body.WriteString(fmt.Sprintf("**Severity**: %s\n", analysis.Classification.Severity))
	body.WriteString(fmt.Sprintf("**Confidence**: %.1f%%\n", analysis.Classification.Confidence*100))
	body.WriteString(fmt.Sprintf("**Root Cause**: %s\n\n", analysis.RootCause))

	if analysis.Description != "" {
		body.WriteString(fmt.Sprintf("**Description**: %s\n\n", analysis.Description))
	}

	// Fix details
	body.WriteString("## 🔧 Fix Details\n\n")
	body.WriteString(fmt.Sprintf("**Fix Type**: %s\n", fix.Fix.Type))
	body.WriteString(fmt.Sprintf("**Fix Confidence**: %.1f%%\n", fix.Fix.Confidence*100))
	body.WriteString(fmt.Sprintf("**Description**: %s\n\n", fix.Fix.Description))

	if fix.Fix.Rationale != "" {
		body.WriteString(fmt.Sprintf("**Rationale**: %s\n\n", fix.Fix.Rationale))
	}

	// Changes summary
	if len(fix.Fix.Changes) > 0 {
		body.WriteString("## 📝 Changes Made\n\n")
		for _, change := range fix.Fix.Changes {
			caser := cases.Title(language.English)
			body.WriteString(fmt.Sprintf("- **%s**: %s `%s`\n", caser.String(change.Operation), change.FilePath, change.Explanation))
		}
		body.WriteString("\n")
	}

	// Test results
	body.WriteString("## 🧪 Validation Results\n\n")
	if fix.TestResult != nil {
		body.WriteString(fmt.Sprintf("**Tests Passed**: %s\n", boolToEmoji(fix.TestResult.Success)))
		body.WriteString(fmt.Sprintf("**Test Coverage**: %.1f%% (Required: 85%%)\n", fix.TestResult.Coverage))
		body.WriteString(fmt.Sprintf("**Tests Run**: %d passed, %d failed, %d skipped\n\n", fix.TestResult.PassedTests, fix.TestResult.FailedTests, fix.TestResult.SkippedTests))
	} else {
		body.WriteString("**Test Results**: Not available\n\n")
	}

	// Risks and benefits
	if len(fix.Fix.Risks) > 0 {
		body.WriteString("## ⚠️ Potential Risks\n\n")
		for _, risk := range fix.Fix.Risks {
			body.WriteString(fmt.Sprintf("- %s\n", risk))
		}
		body.WriteString("\n")
	}

	if len(fix.Fix.Benefits) > 0 {
		body.WriteString("## ✅ Benefits\n\n")
		for _, benefit := range fix.Fix.Benefits {
			body.WriteString(fmt.Sprintf("- %s\n", benefit))
		}
		body.WriteString("\n")
	}

	// Metadata
	body.WriteString("## 🔍 Metadata\n\n")
	body.WriteString(fmt.Sprintf("**Analysis ID**: `%s`\n", analysis.ID))
	body.WriteString(fmt.Sprintf("**Fix ID**: `%s`\n", fix.Fix.ID))
	body.WriteString(fmt.Sprintf("**LLM Provider**: %s\n", analysis.LLMProvider))
	body.WriteString(fmt.Sprintf("**Generated**: %s\n\n", fix.Fix.Timestamp.Format(time.RFC3339)))

	body.WriteString("---\n")
	body.WriteString("*This PR was automatically generated by the GitHub Actions Auto-Fix Agent*\n")

	return body.String()
}

func (p *PullRequestEngine) generatePRLabels(analysis *FailureAnalysisResult, fix *ProposedFix) []string {
	labels := []string{
		"autofix",
		"automated",
		string(fix.Type) + "-fix",
		string(analysis.Classification.Type) + "-failure",
	}

	// Add severity label
	switch analysis.Classification.Severity {
	case Critical:
		labels = append(labels, "priority-critical")
	case High:
		labels = append(labels, "priority-high")
	case Medium:
		labels = append(labels, "priority-medium")
	case Low:
		labels = append(labels, "priority-low")
	}

	// Add confidence label
	if fix.Confidence > 0.8 {
		labels = append(labels, "high-confidence")
	} else if fix.Confidence > 0.5 {
		labels = append(labels, "medium-confidence")
	} else {
		labels = append(labels, "low-confidence")
	}

	return labels
}

func (p *PullRequestEngine) createPullRequest(ctx context.Context, options *PRCreationOptions) (*PullRequest, error) {
	// Use the GitHubClient interface to create PR
	pr, err := p.githubClient.CreatePullRequest(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	return pr, nil
}

func (p *PullRequestEngine) addPRMetadata(ctx context.Context, pr *PullRequest, analysis *FailureAnalysisResult, fix *FixValidationResult) error {
		validationDuration := "N/A"
		testOutput := "N/A"
		if fix.TestResult != nil {
			validationDuration = fix.TestResult.Duration.String()
			testOutput = truncateString(fix.TestResult.Output, 500)
		}
	// Add a comment with additional metadata
	metadataComment := fmt.Sprintf(`## 🔍 Additional Metadata

**Analysis Details:**
- Processing Time: %v
- Error Patterns: %d detected
- Affected Files: %d
- Classification Tags: %s

**Fix Validation:**
- Validation Duration: %v
- Test Output: %s

**Provider Information:**
- LLM Provider: %s
- Model: %s
`,
		analysis.ProcessingTime,
		len(analysis.ErrorPatterns),
		len(analysis.AffectedFiles),
		strings.Join(analysis.Classification.Tags, ", "),
		validationDuration,
		testOutput,
		analysis.LLMProvider,
		"N/A", // Model info would need to be added to the response
	)

	// Use the GitHubClient interface to add comment
	err := p.githubClient.AddPullRequestComment(ctx, pr.Number, metadataComment)
	return err
}

// Helper functions

func boolToEmoji(b bool) string {
	if b {
		return "✅"
	}
	return "❌"
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func loadPRTemplates() *PRTemplates {
	return &PRTemplates{
		Title: "🤖 Auto-fix: {{.FixType}} for {{.FailureType}} failure",
		Body: `## 🤖 Automated Fix

This pull request was automatically generated to fix a CI/CD pipeline failure.

{{.AnalysisSummary}}

{{.FixDetails}}

{{.ValidationResults}}

---
*This PR was automatically generated by the GitHub Actions Auto-Fix Agent*`,
		CommitMsg: "fix: {{.Description}}\n\nGenerated by GitHub Actions Auto-Fix Agent\nAnalysis ID: {{.AnalysisID}}\nFix ID: {{.FixID}}",
		Labels:    []string{"autofix", "automated", "ci-fix"},
		Reviewers: []string{}, // Can be configured per project
	}
}
