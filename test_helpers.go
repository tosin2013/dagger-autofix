package main

import "dagger.io/dagger"

// createTestSecret creates a secret for testing purposes
// When dag is nil (unit tests), it returns a dummy secret
// When dag is available (Dagger context), it uses dag.SetSecret
func createTestSecret(name, value string) *dagger.Secret {
	if dag != nil {
		return dag.SetSecret(name, value)
	}
	// Return a dummy secret for unit tests
	// This won't work in actual Dagger operations but allows tests to compile and run
	return &dagger.Secret{}
}

// createTestContainer creates a container for testing purposes
// Returns nil when dag is not available
func createTestContainer() *dagger.Container {
	if dag != nil {
		return dag.Container()
	}
	return nil
}