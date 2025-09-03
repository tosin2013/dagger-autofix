//go:build !dagger
// +build !dagger

package main

import "dagger.io/dagger"

// dag is a global variable provided by the Dagger runtime
// When running in Dagger, this is automatically injected
// For local compilation and testing, we define it as nil
var dag *dagger.Client
