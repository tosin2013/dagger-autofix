package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

// Example of using dagger-autofix with MCP GitHub integration
func main() {
	ctx := context.Background()
	
	// Load MCP configuration
	configFile := "examples/mcp_config.json"
	configData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to read MCP config: %v", err)
	}
	
	var config struct {
		GitHubMCP *MCPConfig `json:"github_mcp"`
	}
	
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse MCP config: %v", err)
	}
	
	// Create Dagger autofix instance with MCP
	autofix := New().
		WithRepository("your-org", "your-repo").
		WithMCPGitHub(config.GitHubMCP).
		WithLLMProvider("openai", nil) // You'd provide actual API key
	
	// Initialize the autofix agent
	agent, err := autofix.Initialize(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}
	
	logrus.Info("MCP-enabled GitHub autofix agent initialized successfully!")
	
	// Example: Get a workflow run
	run, err := agent.AnalyzeFailure(ctx, 12345) // Replace with actual run ID
	if err != nil {
		log.Printf("Failed to analyze failure: %v", err)
		return
	}
	
	logrus.Infof("Analyzed failure: %s", run.ID)
}