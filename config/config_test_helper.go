package config

import (
	"sync"
)

var (
	configDirMu   sync.RWMutex
	testConfigDir string
	originalFunc  = userConfigDirFunc
)

// SetTestConfigDir sets a custom config directory for testing
func SetTestConfigDir(dir string) {
	configDirMu.Lock()
	defer configDirMu.Unlock()

	// Save the test directory
	testConfigDir = dir

	// Replace the function with our test version
	userConfigDirFunc = func() (string, error) {
		return testConfigDir, nil
	}
}

// ResetTestConfigDir resets to the original function
func ResetTestConfigDir() {
	configDirMu.Lock()
	defer configDirMu.Unlock()

	// Reset the directory
	testConfigDir = ""

	// Restore the original function
	userConfigDirFunc = originalFunc
}
