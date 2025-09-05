package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// MCPConfig represents configuration for an MCP server
type MCPConfig struct {
	ServerCommand []string          `json:"server_command"`
	ServerArgs    []string          `json:"server_args"`
	ServerEnv     map[string]string `json:"server_env"`
	Timeout       int               `json:"timeout"`
}

// MCPClient provides a wrapper around MCP SDK functionality
type MCPClient struct {
	session *mcp.ClientSession
	client  *mcp.Client
	logger  *logrus.Logger
	config  *MCPConfig
}

// NewMCPClient creates a new MCP client instance
func NewMCPClient(config *MCPConfig, logger *logrus.Logger) (*MCPClient, error) {
	if logger == nil {
		logger = logrus.New()
	}

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "dagger-autofix",
		Version: "v1.0.0",
	}, nil)

	return &MCPClient{
		client: client,
		logger: logger,
		config: config,
	}, nil
}

// Connect establishes connection to the MCP server
func (m *MCPClient) Connect(ctx context.Context) error {
	if m.config == nil {
		return fmt.Errorf("MCP config is required")
	}

	if len(m.config.ServerCommand) == 0 {
		return fmt.Errorf("server command is required in MCP config")
	}

	// Create command with args
	cmd := exec.Command(m.config.ServerCommand[0], m.config.ServerCommand[1:]...)
	if len(m.config.ServerArgs) > 0 {
		cmd.Args = append(cmd.Args, m.config.ServerArgs...)
	}

	// Set environment variables
	if len(m.config.ServerEnv) > 0 {
		for key, value := range m.config.ServerEnv {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Create transport and connect
	transport := &mcp.CommandTransport{Command: cmd}
	session, err := m.client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to MCP server: %w", err)
	}

	m.session = session
	m.logger.Info("Successfully connected to MCP server")
	return nil
}

// CallTool calls a tool on the MCP server
func (m *MCPClient) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	if m.session == nil {
		return nil, fmt.Errorf("MCP client not connected")
	}

	params := &mcp.CallToolParams{
		Name:      toolName,
		Arguments: arguments,
	}

	m.logger.WithFields(logrus.Fields{
		"tool": toolName,
		"args": arguments,
	}).Debug("Calling MCP tool")

	result, err := m.session.CallTool(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("MCP tool call failed: %w", err)
	}

	m.logger.WithField("tool", toolName).Debug("MCP tool call successful")
	return result, nil
}

// Close closes the MCP connection
func (m *MCPClient) Close() error {
	if m.session != nil {
		return m.session.Close()
	}
	return nil
}

// MCPGitHubClient implements GitHubClient interface using MCP
type MCPGitHubClient struct {
	*MCPClient
}

// NewMCPGitHubClient creates a new GitHub client using MCP
func NewMCPGitHubClient(config *MCPConfig, logger *logrus.Logger) (*MCPGitHubClient, error) {
	mcpClient, err := NewMCPClient(config, logger)
	if err != nil {
		return nil, err
	}

	return &MCPGitHubClient{
		MCPClient: mcpClient,
	}, nil
}

// GetWorkflowRun retrieves a workflow run via MCP
func (m *MCPGitHubClient) GetWorkflowRun(ctx context.Context, runID int64) (*WorkflowRun, error) {
	result, err := m.CallTool(ctx, "get_workflow_run", map[string]interface{}{
		"run_id": runID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow run: %w", err)
	}

	// Parse the result into WorkflowRun
	var workflowRun WorkflowRun
	if err := parseToolResult(result, &workflowRun); err != nil {
		return nil, fmt.Errorf("failed to parse workflow run result: %w", err)
	}

	return &workflowRun, nil
}

// GetWorkflowLogs retrieves workflow logs via MCP
func (m *MCPGitHubClient) GetWorkflowLogs(ctx context.Context, runID int64) (*WorkflowLogs, error) {
	result, err := m.CallTool(ctx, "get_workflow_logs", map[string]interface{}{
		"run_id": runID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow logs: %w", err)
	}

	var logs WorkflowLogs
	if err := parseToolResult(result, &logs); err != nil {
		return nil, fmt.Errorf("failed to parse workflow logs result: %w", err)
	}

	return &logs, nil
}

// GetFailedWorkflowRuns retrieves failed workflow runs via MCP
func (m *MCPGitHubClient) GetFailedWorkflowRuns(ctx context.Context) ([]*WorkflowRun, error) {
	result, err := m.CallTool(ctx, "list_workflow_runs", map[string]interface{}{
		"status":     "completed",
		"conclusion": "failure",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get failed workflow runs: %w", err)
	}

	var runs []WorkflowRun
	if err := parseToolResult(result, &runs); err != nil {
		return nil, fmt.Errorf("failed to parse workflow runs result: %w", err)
	}

	// Convert to slice of pointers
	var ptrRuns []*WorkflowRun
	for i := range runs {
		ptrRuns = append(ptrRuns, &runs[i])
	}

	return ptrRuns, nil
}

// CreateTestBranch creates a test branch with changes via MCP
func (m *MCPGitHubClient) CreateTestBranch(ctx context.Context, branchName string, changes []CodeChange) (func(), error) {
	// First create the branch
	_, err := m.CallTool(ctx, "create_branch", map[string]interface{}{
		"branch": branchName,
		"from":   "main", // TODO: make configurable
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create test branch: %w", err)
	}

	// Apply changes
	for _, change := range changes {
		var toolName string
		args := map[string]interface{}{
			"path":    change.FilePath,
			"content": change.NewContent,
			"branch":  branchName,
		}

		switch change.Operation {
		case "add":
			toolName = "create_file"
		case "modify":
			toolName = "update_file"
			// Note: SHA would be retrieved by the MCP server if needed
		case "delete":
			toolName = "delete_file"
			// Note: SHA would be retrieved by the MCP server if needed
		default:
			return nil, fmt.Errorf("unsupported operation: %s", change.Operation)
		}

		_, err := m.CallTool(ctx, toolName, args)
		if err != nil {
			return nil, fmt.Errorf("failed to apply change %s: %w", change.FilePath, err)
		}
	}

	// Return cleanup function
	cleanup := func() {
		cleanupCtx := context.Background()
		_, cleanupErr := m.CallTool(cleanupCtx, "delete_branch", map[string]interface{}{
			"branch": branchName,
		})
		if cleanupErr != nil {
			m.logger.WithError(cleanupErr).Error("Failed to cleanup test branch")
		}
	}

	return cleanup, nil
}

// parseToolResult parses MCP tool result into target struct
func parseToolResult(result *mcp.CallToolResult, target interface{}) error {
	if result == nil {
		return fmt.Errorf("nil result")
	}

	// Handle both direct content and structured responses
	var data interface{}
	if len(result.Content) > 0 {
		// Try to parse first content item
		content := result.Content[0]
		
		// Type assert to TextContent
		if textContent, ok := content.(*mcp.TextContent); ok {
			// Try to unmarshal as JSON
			if err := json.Unmarshal([]byte(textContent.Text), &data); err != nil {
				// If not JSON, use text as-is for simple types
				data = textContent.Text
			}
		} else {
			// For other content types, marshal to JSON and use that
			jsonBytes, err := content.MarshalJSON()
			if err != nil {
				return fmt.Errorf("failed to marshal content: %w", err)
			}
			if err := json.Unmarshal(jsonBytes, &data); err != nil {
				return fmt.Errorf("failed to unmarshal content: %w", err)
			}
		}
	}

	// Convert data to target struct
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal into target: %w", err)
	}

	return nil
}