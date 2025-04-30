// Package testutil provides shared utilities for testing
package testutil

import (
	"path/filepath"
	"runtime"
)

// GetTestDictionaryPath returns the path to the test dictionary file
func GetTestDictionaryPath() string {
	// Get the current file path
	_, filename, _, _ := runtime.Caller(0)

	// Navigate up one directory from testutil to the tests directory
	testsDir := filepath.Dir(filepath.Dir(filename))

	// Join with the data directory and file name
	return filepath.Join(testsDir, "data", "test_dictionary.json")
}
