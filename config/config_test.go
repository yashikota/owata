package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetPathWithError(t *testing.T) {
	manager := NewManager()

	// Test local config path
	localPath, err := manager.GetPathWithError(false)
	if err != nil {
		t.Fatalf("Expected no error for local path, got: %v", err)
	}
	if localPath != ConfigFileName {
		t.Errorf("Expected local path to be %s, got %s", ConfigFileName, localPath)
	}

	// Test global config path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Could not determine home directory, skipping global path test")
	}

	expectedGlobalPath := filepath.Join(homeDir, ".config", ConfigFileName)
	globalPath, err := manager.GetPathWithError(true)
	if err != nil {
		t.Fatalf("Expected no error for global path, got: %v", err)
	}

	if globalPath != expectedGlobalPath {
		t.Errorf("Expected global path to be %s, got %s", expectedGlobalPath, globalPath)
	}
}

func TestLoadFromPath(t *testing.T) {
	// Create a temporary config file for testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-config.json")

	// Create test config
	testConfig := &Config{
		WebhookURL: "https://example.com/webhook",
		Username:   "TestUser",
		AvatarURL:  "https://example.com/avatar.png",
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	// Write to temp file
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Create config manager
	manager := NewManager()

	// Test loading config from path
	loadedConfig, err := manager.LoadFromPath(tempFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded config matches test config
	if !reflect.DeepEqual(loadedConfig, testConfig) {
		t.Errorf("Loaded config does not match test config.\nExpected: %+v\nGot: %+v", testConfig, loadedConfig)
	}

	// Test loading non-existent file
	_, err = manager.LoadFromPath(filepath.Join(tempDir, "nonexistent.json"))
	if err == nil {
		t.Error("Expected error when loading non-existent file, got nil")
	}

	// Test loading invalid JSON
	invalidFile := filepath.Join(tempDir, "invalid.json")
	if err := os.WriteFile(invalidFile, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid test file: %v", err)
	}

	_, err = manager.LoadFromPath(invalidFile)
	if err == nil {
		t.Error("Expected error when loading invalid JSON, got nil")
	}
}

func TestSaveToPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-config.json")

	// Create test config
	testConfig := &Config{
		WebhookURL: "https://example.com/webhook",
		Username:   "TestUser",
		AvatarURL:  "https://example.com/avatar.png",
	}

	// Create config manager
	manager := NewManager()

	// Save config to path
	if err := manager.SaveToPath(testConfig, tempFile); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Errorf("Config file was not created at %s", tempFile)
	}

	// Load saved config
	loadedConfig, err := manager.LoadFromPath(tempFile)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	// Verify loaded config matches test config
	if !reflect.DeepEqual(loadedConfig, testConfig) {
		t.Errorf("Loaded config does not match test config.\nExpected: %+v\nGot: %+v", testConfig, loadedConfig)
	}
}

func TestCreateTemplate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Save original home directory and set mock HOME
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Save original working directory and change to tempDir
	currentDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(currentDir)

	// Create a config manager with standard filename
	manager := NewManager()

	// Test case 1: Create local template when it doesn't exist
	localPath, created, err := manager.CreateTemplate(false)
	if err != nil {
		t.Fatalf("Failed to create local template: %v", err)
	}

	// Check that the file was created and the created flag is true
	if !created {
		t.Error("Expected created=true for new local template, got false")
	}

	// Verify file was created
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		t.Errorf("Local config template was not created at %s", localPath)
	}

	// Check template content
	localConfig, err := manager.LoadFromPath(localPath)
	if err != nil {
		t.Fatalf("Failed to load local config template: %v", err)
	}

	if localConfig.WebhookURL != "" || localConfig.Username != "" || localConfig.AvatarURL != "" {
		t.Errorf("Template config should have empty fields, got %+v", localConfig)
	}

	// Test case 2: Create local template when it already exists
	secondLocalPath, created, err := manager.CreateTemplate(false)
	if err != nil {
		t.Fatalf("Failed on second local template creation: %v", err)
	}

	// Check that no new file was created and created flag is false
	if created {
		t.Error("Expected created=false for existing template, got true")
	}
	if secondLocalPath != localPath {
		t.Errorf("Expected same path on second call, got %s vs %s", secondLocalPath, localPath)
	}

	// Test case 3: Create global template
	globalPath, created, err := manager.CreateTemplate(true)
	if err != nil {
		t.Fatalf("Failed to create global template: %v", err)
	}

	// Check that the file was created and the created flag is true
	if !created {
		t.Error("Expected created=true for new global template, got false")
	}

	// Expected global path
	expectedGlobalPath := filepath.Join(tempDir, ".config", "owata-config.json")
	if globalPath != expectedGlobalPath {
		t.Errorf("Expected global path %s, got %s", expectedGlobalPath, globalPath)
	}

	// Verify file was created
	if _, err := os.Stat(globalPath); os.IsNotExist(err) {
		t.Errorf("Global config template was not created at %s", globalPath)
	}

	// Check global template content
	globalConfig, err := manager.LoadFromPath(globalPath)
	if err != nil {
		t.Fatalf("Failed to load global config template: %v", err)
	}

	if globalConfig.WebhookURL != "" || globalConfig.Username != "" || globalConfig.AvatarURL != "" {
		t.Errorf("Global template config should have empty fields, got %+v", globalConfig)
	}
}

func TestLoad(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Save original home directory and set mock HOME
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Save original working directory and change to tempDir
	currentDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(currentDir)

	// Set up a test config manager with standard filename
	testManager := NewManager()

	// Define paths for local and global configs
	localConfigPath := ConfigFileName // In current dir (tempDir)
	globalConfigDir := filepath.Join(tempDir, ".config")
	globalConfigPath := filepath.Join(globalConfigDir, ConfigFileName)

	// Ensure the global config directory exists
	if err := os.MkdirAll(globalConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create global config directory: %v", err)
	}

	// Test case 1: Neither config exists
	_, _, err := testManager.Load(false)
	if err == nil {
		t.Error("Expected error when no config exists, got nil")
	}

	// Create local config
	localConfig := &Config{
		WebhookURL: "https://example.com/local-webhook",
		Username:   "LocalUser",
		AvatarURL:  "https://example.com/local-avatar.png",
	}

	localData, _ := json.MarshalIndent(localConfig, "", "  ")
	if err := os.WriteFile(localConfigPath, localData, 0644); err != nil {
		t.Fatalf("Failed to write local config: %v", err)
	}

	// Test case 2: Local config exists, prefer local
	config, path, err := testManager.Load(false)
	if err != nil {
		t.Fatalf("Failed to load local config: %v", err)
	}

	if path != localConfigPath {
		t.Errorf("Expected path to be %s, got %s", localConfigPath, path)
	}

	if !reflect.DeepEqual(config, localConfig) {
		t.Errorf("Loaded config does not match local config.\nExpected: %+v\nGot: %+v", localConfig, config)
	}

	// Create global config
	globalConfig := &Config{
		WebhookURL: "https://example.com/global-webhook",
		Username:   "GlobalUser",
		AvatarURL:  "https://example.com/global-avatar.png",
	}

	globalData, _ := json.MarshalIndent(globalConfig, "", "  ")
	if err := os.WriteFile(globalConfigPath, globalData, 0644); err != nil {
		t.Fatalf("Failed to write global config: %v", err)
	}

	// Test case 3: Both configs exist, prefer global
	config, path, err = testManager.Load(true)
	if err != nil {
		t.Fatalf("Failed to load global config: %v", err)
	}

	if path != globalConfigPath {
		t.Errorf("Expected path to be %s, got %s", globalConfigPath, path)
	}

	if !reflect.DeepEqual(config, globalConfig) {
		t.Errorf("Loaded config does not match global config.\nExpected: %+v\nGot: %+v", globalConfig, config)
	}

	// Test case 4: Both configs exist, prefer local
	config, path, err = testManager.Load(false)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if path != localConfigPath {
		t.Errorf("Expected path to be %s, got %s", localConfigPath, path)
	}

	if !reflect.DeepEqual(config, localConfig) {
		t.Errorf("Loaded config does not match local config.\nExpected: %+v\nGot: %+v", localConfig, config)
	}

	// Remove local config
	if err := os.Remove(localConfigPath); err != nil {
		t.Fatalf("Failed to remove local config: %v", err)
	}

	// Test case 5: Only global config exists, no preference (should fallback to global)
	config, path, err = testManager.Load(false)
	if err != nil {
		t.Fatalf("Failed to load global config as fallback: %v", err)
	}

	if path != globalConfigPath {
		t.Errorf("Expected path to be %s, got %s", globalConfigPath, path)
	}

	if !reflect.DeepEqual(config, globalConfig) {
		t.Errorf("Loaded config does not match global config.\nExpected: %+v\nGot: %+v", globalConfig, config)
	}

	// Test case 6: Only global config exists, but explicitly prefer global
	config, path, err = testManager.Load(true)
	if err != nil {
		t.Fatalf("Failed to load global config when preferred: %v", err)
	}

	if path != globalConfigPath {
		t.Errorf("Expected path to be %s, got %s", globalConfigPath, path)
	}

	if !reflect.DeepEqual(config, globalConfig) {
		t.Errorf("Loaded config does not match global config.\nExpected: %+v\nGot: %+v", globalConfig, config)
	}

	// Remove global config
	if err := os.Remove(globalConfigPath); err != nil {
		t.Fatalf("Failed to remove global config: %v", err)
	}

	// Test case 7: No configs exist with preference for global
	_, _, err = testManager.Load(true)
	if err == nil {
		t.Error("Expected error when no config exists with global preference, got nil")
	}
}

func TestSave(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Save original home directory and set mock HOME
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Save original working directory and change to tempDir
	currentDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(currentDir)

	// Use standard manager
	testManager := NewManager()

	// Create test config
	testConfig := &Config{
		WebhookURL: "https://example.com/webhook",
		Username:   "TestUser",
		AvatarURL:  "https://example.com/avatar.png",
	}

	// Test saving local config
	localSavedPath, err := testManager.Save(testConfig, false)
	if err != nil {
		t.Fatalf("Failed to save local config: %v", err)
	}

	// Get expected local path should be in current directory
	expectedLocalPath := ConfigFileName
	if localSavedPath != expectedLocalPath {
		t.Errorf("Expected local save path to be %s, got %s", expectedLocalPath, localSavedPath)
	}

	// Verify local file was created
	if _, err := os.Stat(localSavedPath); os.IsNotExist(err) {
		t.Errorf("Local config was not created at %s", localSavedPath)
	}

	// Test saving global config
	globalSavedPath, err := testManager.Save(testConfig, true)
	if err != nil {
		t.Fatalf("Failed to save global config: %v", err)
	}

	// Get expected global path
	expectedGlobalPath := filepath.Join(tempDir, ".config", ConfigFileName)
	if globalSavedPath != expectedGlobalPath {
		t.Errorf("Expected global save path to be %s, got %s", expectedGlobalPath, globalSavedPath)
	}

	// Verify global file was created
	if _, err := os.Stat(globalSavedPath); os.IsNotExist(err) {
		t.Errorf("Global config was not created at %s", globalSavedPath)
	}

	// Verify global directory was created
	globalConfigDir := filepath.Dir(globalSavedPath)
	if _, err := os.Stat(globalConfigDir); os.IsNotExist(err) {
		t.Errorf("Global config directory was not created at %s", globalConfigDir)
	}

	// Verify config was written correctly
	loadedConfig, err := testManager.LoadFromPath(localSavedPath)
	if err != nil {
		t.Fatalf("Failed to load saved local config: %v", err)
	}
	if !reflect.DeepEqual(loadedConfig, testConfig) {
		t.Errorf("Loaded local config doesn't match original.\nExpected: %+v\nGot: %+v", testConfig, loadedConfig)
	}

	loadedConfig, err = testManager.LoadFromPath(globalSavedPath)
	if err != nil {
		t.Fatalf("Failed to load saved global config: %v", err)
	}
	if !reflect.DeepEqual(loadedConfig, testConfig) {
		t.Errorf("Loaded global config doesn't match original.\nExpected: %+v\nGot: %+v", testConfig, loadedConfig)
	}
}
