package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test the Execute function safely
func TestExecuteFunction(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test version command
	os.Args = []string{"github-autofix", "--version"}
	
	// Capture output
	var buf bytes.Buffer
	cli := NewCLI()
	cli.rootCmd.SetOut(&buf)
	cli.rootCmd.SetErr(&buf)
	
	err := cli.Execute()
	assert.NoError(t, err)
}

// Test some CLI functions that can work without full setup
func TestSimpleCLIFunctions(t *testing.T) {
	// Test that just check basic functionality
	cli := NewCLI()
	
	t.Run("Execute_help", func(t *testing.T) {
		// Save original args
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		// Test help command
		os.Args = []string{"github-autofix", "--help"}
		
		// Capture output
		var buf bytes.Buffer
		cli.rootCmd.SetOut(&buf)
		cli.rootCmd.SetErr(&buf)
		
		err := cli.Execute()
		assert.NoError(t, err)
	})
}