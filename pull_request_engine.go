package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// PullRequestEngine handles automated pull request creation and management
type PullRequestEngine struct {
	githubClient *GitHubIntegration
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

// PRCreationOptions contains options for PR creation
type PRCreationOptions struct {
	BranchName   string   `json:"branch_name"`
	TargetBranch string   `json:"target_branch"`
	Title        string   `json:"title"`
	Body         string   `json:"body"`
	Labels       []string `json:"labels"`
	Reviewers    []string `json:"reviewers"`
	Assignees    []string `json:"assignees"`
	Draft        bool     `json:"draft"`
	AutoMerge    bool     `json:"auto_merge"`
	DeleteBranch bool     `json:"delete_branch"`
}

// NewPullRequestEngine creates a new pull request engine
func NewPullRequestEngine(githubClient *GitHubIntegration, logger *logrus.Logger) *PullRequestEngine {
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

	// Get existing PR
	existingPR, _, err := p.githubClient.client.PullRequests.Get(ctx, p.githubClient.repoOwner, p.githubClient.repoName, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing PR: %w", err)
	}

	// Update PR
	updatedPR := &github.PullRequest{
		Title: &updates.Title,
		Body:  &updates.Body,
	}

	result, _, err := p.githubClient.client.PullRequests.Edit(ctx, p.githubClient.repoOwner, p.githubClient.repoName, prNumber, updatedPR)
	if err != nil {
		return nil, fmt.Errorf("failed to update PR: %w", err)
	}

	// Update labels if provided
	if len(updates.Labels) > 0 {
		if _, _, err := p.githubClient.client.Issues.ReplaceLabelsForIssue(ctx, p.githubClient.repoOwner, p.githubClient.repoName, prNumber, updates.Labels); err != nil {
			p.logger.WithError(err).Warn("Failed to update PR labels")
		}
	}

	return &PullRequest{
		Number:    result.GetNumber(),
		Title:     result.GetTitle(),
		Body:      result.GetBody(),
		URL:       result.GetHTMLURL(),
		Branch:    existingPR.GetHead().GetRef(),
		CommitSHA: result.GetHead().GetSHA(),
		State:     result.GetState(),
		CreatedAt: result.GetCreatedAt(),
		Author:    result.GetUser().GetLogin(),
		Labels:    updates.Labels,
	}, nil
}

// ClosePR closes a pull request
func (p *PullRequestEngine) ClosePR(ctx context.Context, prNumber int, reason string) error {
	p.logger.WithFields(logrus.Fields{
		"pr_number": prNumber,
		"reason":    reason,
	}).Info("Closing pull request")

	// Close PR
	state := "closed"
	updatedPR := &github.PullRequest{
		State: &state,
	}

	_, _, err := p.githubClient.client.PullRequests.Edit(ctx, p.githubClient.repoOwner, p.githubClient.repoName, prNumber, updatedPR)
	if err != nil {
		return fmt.Errorf("failed to close PR: %w", err)
	}

	// Add closing comment
	comment := &github.IssueComment{
		Body: &reason,
	}

	if _, _, err := p.githubClient.client.Issues.CreateComment(ctx, p.githubClient.repoOwner, p.githubClient.repoName, prNumber, comment); err != nil {
		p.logger.WithError(err).Warn("Failed to add closing comment")
	}

	return nil
}

// GetPRStatus gets the status of a pull request
func (p *PullRequestEngine) GetPRStatus(ctx context.Context, prNumber int) (*PullRequest, error) {
	pr, _, err := p.githubClient.client.PullRequests.Get(ctx, p.githubClient.repoOwner, p.githubClient.repoName, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	// Get labels
	labels, _, err := p.githubClient.client.Issues.ListLabelsByIssue(ctx, p.githubClient.repoOwner, p.githubClient.repoName, prNumber, nil)
	if err != nil {
		p.logger.WithError(err).Warn("Failed to get PR labels")
	}

	var labelNames []string
	for _, label := range labels {
		labelNames = append(labelNames, label.GetName())
	}

	return &PullRequest{
		Number:    pr.GetNumber(),
		Title:     pr.GetTitle(),
		Body:      pr.GetBody(),
		URL:       pr.GetHTMLURL(),
		Branch:    pr.GetHead().GetRef(),
		CommitSHA: pr.GetHead().GetSHA(),
		State:     pr.GetState(),
		CreatedAt: pr.GetCreatedAt(),
		Author:    pr.GetUser().GetLogin(),
		Labels:    labelNames,
	}, nil
}

// Private helper methods

func (p *PullRequestEngine) generateBranchName(analysis *FailureAnalysisResult, fix *ProposedFix) string {
	timestamp := time.Now().Format("20060102-150405")
	fixType := strings.ToLower(string(fix.Type))
	return fmt.Sprintf("autofix/%s/%s-%s", fixType, analysis.ID, timestamp)
}

func (p *PullRequestEngine) createBranch(ctx context.Context, branchName string, changes []CodeChange) error {
	p.logger.WithField("branch", branchName).Debug("Creating branch with changes")

	// Get the default branch reference
	mainRef, _, err := p.githubClient.client.Git.GetRef(ctx, p.githubClient.repoOwner, p.githubClient.repoName, "heads/main")
	if err != nil {
		return fmt.Errorf("failed to get main branch ref: %w", err)
	}

	// Create new branch
	newRef := &github.Reference{
		Ref: github.String("refs/heads/" + branchName),
		Object: &github.GitObject{
			SHA: mainRef.Object.SHA,
		},
	}

	_, _, err = p.githubClient.client.Git.CreateRef(ctx, p.githubClient.repoOwner, p.githubClient.repoName, newRef)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// Apply changes to the branch
	for _, change := range changes {
		if err := p.applyChange(ctx, branchName, change); err != nil {
			p.logger.WithError(err).Warnf("Failed to apply change to %s", change.FilePath)
			// Continue with other changes even if one fails
		}
	}

	return nil
}

func (p *PullRequestEngine) applyChange(ctx context.Context, branch string, change CodeChange) error {
	p.logger.WithFields(logrus.Fields{
		"file":      change.FilePath,
		"operation": change.Operation,
		"branch":    branch,
	}).Debug("Applying code change")

	switch change.Operation {
	case "add":
		return p.createFile(ctx, branch, change)
	case "modify":
		return p.updateFile(ctx, branch, change)
	case "delete":
		return p.deleteFile(ctx, branch, change)
	default:
		return fmt.Errorf("unknown operation: %s", change.Operation)
	}
}

func (p *PullRequestEngine) createFile(ctx context.Context, branch string, change CodeChange) error {
	fileContent := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("Add %s", change.FilePath)),
		Content: []byte(change.NewContent),
		Branch:  &branch,
	}

	_, _, err := p.githubClient.client.Repositories.CreateFile(ctx, p.githubClient.repoOwner, p.githubClient.repoName, change.FilePath, fileContent)
	return err
}

func (p *PullRequestEngine) updateFile(ctx context.Context, branch string, change CodeChange) error {
	// Get current file to get SHA
	fileContent, _, _, err := p.githubClient.client.Repositories.GetContents(ctx, p.githubClient.repoOwner, p.githubClient.repoName, change.FilePath, &github.RepositoryContentGetOptions{
		Ref: branch,
	})
	if err != nil {
		return fmt.Errorf("failed to get file content: %w", err)
	}

	updateOptions := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("Update %s - %s", change.FilePath, change.Explanation)),
		Content: []byte(change.NewContent),
		SHA:     fileContent.SHA,
		Branch:  &branch,
	}

	_, _, err = p.githubClient.client.Repositories.UpdateFile(ctx, p.githubClient.repoOwner, p.githubClient.repoName, change.FilePath, updateOptions)
	return err
}

func (p *PullRequestEngine) deleteFile(ctx context.Context, branch string, change CodeChange) error {
	// Get current file to get SHA
	fileContent, _, _, err := p.githubClient.client.Repositories.GetContents(ctx, p.githubClient.repoOwner, p.githubClient.repoName, change.FilePath, &github.RepositoryContentGetOptions{
		Ref: branch,
	})
	if err != nil {
		return fmt.Errorf("failed to get file content: %w", err)
	}

	deleteOptions := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("Delete %s - %s", change.FilePath, change.Explanation)),
		SHA:     fileContent.SHA,
		Branch:  &branch,
	}

	_, _, err = p.githubClient.client.Repositories.DeleteFile(ctx, p.githubClient.repoOwner, p.githubClient.repoName, change.FilePath, deleteOptions)
	return err
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

	return fmt.Sprintf("🤖 Auto-fix: %s for %s failure (Run #%d)", fixType, failureType, analysis.Context.WorkflowRun.ID)
}

func (p *PullRequestEngine) generatePRBody(analysis *FailureAnalysisResult, fix *FixValidationResult) string {
	var body strings.Builder

	body.WriteString("## 🤖 Automated Fix\n\n")
	body.WriteString("This pull request was automatically generated to fix a CI/CD pipeline failure.\n\n")

	// Failure summary
	body.WriteString("## 📊 Failure Analysis\n\n")
	body.WriteString(fmt.Sprintf("**Workflow Run**: [#%d](%s)\n", analysis.Context.WorkflowRun.ID, analysis.Context.WorkflowRun.URL))
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
	body.WriteString(fmt.Sprintf("**Tests Passed**: %s\n", boolToEmoji(fix.TestResult.Success)))
	body.WriteString(fmt.Sprintf("**Test Coverage**: %.1f%% (Required: 85%%)\n", fix.TestResult.Coverage))
	body.WriteString(fmt.Sprintf("**Tests Run**: %d passed, %d failed, %d skipped\n\n", fix.TestResult.PassedTests, fix.TestResult.FailedTests, fix.TestResult.SkippedTests))

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
	newPR := &github.NewPullRequest{
		Title: &options.Title,
		Head:  &options.BranchName,
		Base:  &options.TargetBranch,
		Body:  &options.Body,
		Draft: &options.Draft,
	}

	pr, _, err := p.githubClient.client.PullRequests.Create(ctx, p.githubClient.repoOwner, p.githubClient.repoName, newPR)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	// Add labels
	if len(options.Labels) > 0 {
		if _, _, err := p.githubClient.client.Issues.AddLabelsToIssue(ctx, p.githubClient.repoOwner, p.githubClient.repoName, pr.GetNumber(), options.Labels); err != nil {
			p.logger.WithError(err).Warn("Failed to add labels to PR")
		}
	}

	// Request reviewers
	if len(options.Reviewers) > 0 {
		reviewersRequest := github.ReviewersRequest{
			Reviewers: options.Reviewers,
		}
		if _, _, err := p.githubClient.client.PullRequests.RequestReviewers(ctx, p.githubClient.repoOwner, p.githubClient.repoName, pr.GetNumber(), reviewersRequest); err != nil {
			p.logger.WithError(err).Warn("Failed to request reviewers")
		}
	}

	// Assign assignees
	if len(options.Assignees) > 0 {
		if _, _, err := p.githubClient.client.Issues.AddAssignees(ctx, p.githubClient.repoOwner, p.githubClient.repoName, pr.GetNumber(), options.Assignees); err != nil {
			p.logger.WithError(err).Warn("Failed to add assignees")
		}
	}

	return &PullRequest{
		Number:    pr.GetNumber(),
		Title:     pr.GetTitle(),
		Body:      pr.GetBody(),
		URL:       pr.GetHTMLURL(),
		Branch:    options.BranchName,
		CommitSHA: pr.GetHead().GetSHA(),
		State:     pr.GetState(),
		CreatedAt: pr.GetCreatedAt(),
		Author:    pr.GetUser().GetLogin(),
		Labels:    options.Labels,
	}, nil
}

func (p *PullRequestEngine) addPRMetadata(ctx context.Context, pr *PullRequest, analysis *FailureAnalysisResult, fix *FixValidationResult) error {
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
		fix.TestResult.Duration,
		truncateString(fix.TestResult.Output, 500),
		analysis.LLMProvider,
		"N/A", // Model info would need to be added to the response
	)

	comment := &github.IssueComment{
		Body: &metadataComment,
	}

	_, _, err := p.githubClient.client.Issues.CreateComment(ctx, p.githubClient.repoOwner, p.githubClient.repoName, pr.Number, comment)
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
