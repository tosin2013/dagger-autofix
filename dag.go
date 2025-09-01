package main

import "dagger.io/dagger"

// dag is the Dagger client instance
// In a real Dagger module, this is provided by the Dagger runtime
// For compilation and testing purposes, we create a mock
var dag daggerClient

type daggerClient interface {
	Container() *dagger.Container
	Host() hostClient
	SetSecret(name, value string) *dagger.Secret
}

type hostClient interface {
	Directory(path string) *dagger.Directory
}

// mockDaggerClient implements daggerClient for testing/compilation
type mockDaggerClient struct{}

func (m *mockDaggerClient) Container() *dagger.Container {
	// Return a mock container for compilation
	return &dagger.Container{}
}

func (m *mockDaggerClient) Host() hostClient {
	return &mockHostClient{}
}

func (m *mockDaggerClient) SetSecret(name, value string) *dagger.Secret {
	// Return a mock secret for compilation
	return &dagger.Secret{}
}

type mockHostClient struct{}

func (m *mockHostClient) Directory(path string) *dagger.Directory {
	// Return a mock directory for compilation
	return &dagger.Directory{}
}

func init() {
	// Initialize with mock client for compilation
	// In a real Dagger environment, this would be provided by the runtime
	dag = &mockDaggerClient{}
}