package main

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
)

// ContainerProvider interface for dependency injection into TestEngine
type ContainerProvider interface {
	CreateContainer() ContainerInterface
}

// ContainerInterface abstracts Dagger container operations
type ContainerInterface interface {
	From(image string) ContainerInterface
	WithExec(args []string) ContainerInterface
	WithEnvVariable(key, value string) ContainerInterface
	WithWorkdir(path string) ContainerInterface
	File(path string) FileInterface
	Stdout(ctx context.Context) (string, error)
	Stderr(ctx context.Context) (string, error)
}

// FileInterface abstracts Dagger file operations
type FileInterface interface {
	Contents(ctx context.Context) (string, error)
}

// RealContainerProvider uses actual Dagger (production)
type RealContainerProvider struct{}

func (r *RealContainerProvider) CreateContainer() ContainerInterface {
	if dag == nil {
		panic("Dagger client not initialized - use mock provider for testing")
	}
	return &RealContainerWrapper{dag.Container()}
}

// RealContainerWrapper wraps real Dagger container
type RealContainerWrapper struct {
	container *dagger.Container
}

func (r *RealContainerWrapper) From(image string) ContainerInterface {
	return &RealContainerWrapper{r.container.From(image)}
}

func (r *RealContainerWrapper) WithExec(args []string) ContainerInterface {
	return &RealContainerWrapper{r.container.WithExec(args)}
}

func (r *RealContainerWrapper) WithEnvVariable(key, value string) ContainerInterface {
	return &RealContainerWrapper{r.container.WithEnvVariable(key, value)}
}

func (r *RealContainerWrapper) WithWorkdir(path string) ContainerInterface {
	return &RealContainerWrapper{r.container.WithWorkdir(path)}
}

func (r *RealContainerWrapper) File(path string) FileInterface {
	return &RealFileWrapper{r.container.File(path)}
}

func (r *RealContainerWrapper) Stdout(ctx context.Context) (string, error) {
	return r.container.Stdout(ctx)
}

func (r *RealContainerWrapper) Stderr(ctx context.Context) (string, error) {
	return r.container.Stderr(ctx)
}

// RealFileWrapper wraps real Dagger file
type RealFileWrapper struct {
	file *dagger.File
}

func (r *RealFileWrapper) Contents(ctx context.Context) (string, error) {
	return r.file.Contents(ctx)
}

// MockContainerProvider for testing
type MockContainerProvider struct {
	MockContainer *MockDaggerContainer
}

func NewMockContainerProvider() *MockContainerProvider {
	return &MockContainerProvider{
		MockContainer: NewMockDaggerContainer(),
	}
}

func (m *MockContainerProvider) CreateContainer() ContainerInterface {
	return &MockContainerWrapper{m.MockContainer}
}

// MockDaggerContainer simulates all Dagger operations
type MockDaggerContainer struct {
	BaseImage      string
	WorkingDir     string
	EnvVars        map[string]string
	ExecHistory    [][]string
	FileSystem     map[string]string
	CommandOutputs map[string]MockCommandResult
	ShouldFail     bool
	FailureMessage string
}

type MockCommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

func NewMockDaggerContainer() *MockDaggerContainer {
	mock := &MockDaggerContainer{
		BaseImage:      "",
		WorkingDir:     "/",
		EnvVars:        make(map[string]string),
		ExecHistory:    make([][]string, 0),
		FileSystem:     make(map[string]string),
		CommandOutputs: make(map[string]MockCommandResult),
		ShouldFail:     false,
	}

	// Setup default framework detection files
	mock.setupDefaultFiles()

	return mock
}

func (m *MockDaggerContainer) setupDefaultFiles() {
	// Default files for framework detection
	m.FileSystem["package.json"] = `{"name":"test-project","scripts":{"test":"jest","lint":"eslint"}}`
	m.FileSystem["go.mod"] = "module test\n\ngo 1.19"
	m.FileSystem["pom.xml"] = `<?xml version="1.0"?><project><modelVersion>4.0.0</modelVersion></project>`

	// Default command outputs
	m.CommandOutputs["go test ./..."] = MockCommandResult{
		Stdout:   "PASS\ncoverage: 87.5% of statements\nok\ttest\t0.005s",
		ExitCode: 0,
	}
	m.CommandOutputs["npm test"] = MockCommandResult{
		Stdout:   "Test Suites: 1 passed, 1 total\nTests: 5 passed, 5 total\nCoverage: 89.2%",
		ExitCode: 0,
	}
	m.CommandOutputs["go build ."] = MockCommandResult{
		Stdout:   "Build successful",
		ExitCode: 0,
	}
	m.CommandOutputs["eslint ."] = MockCommandResult{
		Stdout:   "No linting issues found",
		ExitCode: 0,
	}
}

// SetCommandOutput allows configuring specific command outputs
func (m *MockDaggerContainer) SetCommandOutput(command string, stdout, stderr string, exitCode int, err error) {
	m.CommandOutputs[command] = MockCommandResult{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
		Error:    err,
	}
}

// SetFileContent allows setting specific file contents
func (m *MockDaggerContainer) SetFileContent(path, content string) {
	m.FileSystem[path] = content
}

// MockContainerWrapper implements ContainerInterface
type MockContainerWrapper struct {
	mock *MockDaggerContainer
}

func (m *MockContainerWrapper) From(image string) ContainerInterface {
	m.mock.BaseImage = image
	return m
}

func (m *MockContainerWrapper) WithExec(args []string) ContainerInterface {
	m.mock.ExecHistory = append(m.mock.ExecHistory, args)
	return m
}

func (m *MockContainerWrapper) WithEnvVariable(key, value string) ContainerInterface {
	m.mock.EnvVars[key] = value
	return m
}

func (m *MockContainerWrapper) WithWorkdir(path string) ContainerInterface {
	m.mock.WorkingDir = path
	return m
}

func (m *MockContainerWrapper) File(path string) FileInterface {
	return &MockFileWrapper{path: path, container: m.mock}
}

func (m *MockContainerWrapper) Stdout(ctx context.Context) (string, error) {
	if m.mock.ShouldFail {
		return "", fmt.Errorf("mock container failed: %s", m.mock.FailureMessage)
	}

	// Find the last executed command and return its output
	if len(m.mock.ExecHistory) > 0 {
		lastCmd := strings.Join(m.mock.ExecHistory[len(m.mock.ExecHistory)-1], " ")
		if result, exists := m.mock.CommandOutputs[lastCmd]; exists {
			if result.Error != nil {
				return result.Stdout, result.Error
			}
			return result.Stdout, nil
		}
	}

	return "mock output", nil
}

func (m *MockContainerWrapper) Stderr(ctx context.Context) (string, error) {
	if m.mock.ShouldFail {
		return m.mock.FailureMessage, fmt.Errorf("mock container failed")
	}
	return "", nil
}

// MockFileWrapper implements FileInterface
type MockFileWrapper struct {
	path      string
	container *MockDaggerContainer
}

func (f *MockFileWrapper) Contents(ctx context.Context) (string, error) {
	if f.container.ShouldFail {
		return "", fmt.Errorf("file not found: %s", f.path)
	}

	if content, exists := f.container.FileSystem[f.path]; exists {
		return content, nil
	}

	return "", fmt.Errorf("file not found: %s", f.path)
}