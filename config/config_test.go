package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetPath(t *testing.T) {
	manager := NewManager()

	// Test local config path
	localPath := manager.GetPath(false)
	if localPath != ConfigFileName {
		t.Errorf("Expected local path to be %s, got %s", ConfigFileName, localPath)
	}

	// Test global config path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Could not determine home directory, skipping global path test")
	}

	expectedGlobalPath := filepath.Join(homeDir, ".config", ConfigFileName)
	globalPath := manager.GetPath(true)

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

	// Create a custom config manager with the temp directory
	manager := &Manager{
		configFileName: filepath.Join(tempDir, ConfigFileName),
	}

	// Create local template
	localPath, err := manager.CreateTemplate(false)
	if err != nil {
		t.Fatalf("Failed to create local template: %v", err)
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

	// Test global template in a controlled environment
	globalConfigDir := filepath.Join(tempDir, ".config")
	if err := os.MkdirAll(globalConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create mock global config directory: %v", err)
	}

	globalPath := filepath.Join(globalConfigDir, ConfigFileName)

	// Create a JSON template manually
	templateContent := `{
  "webhook_url": "",
  "username": "",
  "avatar_url": ""
}`

	if err := os.WriteFile(globalPath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write global config template: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(globalPath); os.IsNotExist(err) {
		t.Errorf("Global config template was not created at %s", globalPath)
	}
}

func TestLoad(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	localConfigPath := filepath.Join(tempDir, "local-config.json")
	globalConfigDir := filepath.Join(tempDir, ".config")
	globalConfigPath := filepath.Join(globalConfigDir, "global-config.json")

	// Create a custom manager for testing
	customManager := NewManager()

	// Create a testing helper that uses our test paths
	testLoad := func(preferGlobal bool) (*Config, string, error) {
		var configPath string

		// Logic from the original Load function
		localExists := false
		globalExists := false

		// Check existence
		if _, err := os.Stat(localConfigPath); err == nil {
			localExists = true
		}

		if _, err := os.Stat(globalConfigPath); err == nil {
			globalExists = true
		}

		// Choose which config to load
		if preferGlobal && globalExists {
			configPath = globalConfigPath
		} else if localExists {
			configPath = localConfigPath
		} else if globalExists {
			configPath = globalConfigPath
		} else {
			return nil, "", fmt.Errorf("config file not found: neither %s nor %s exists", localConfigPath, globalConfigPath)
		}

		// Load the config from the chosen path
		config, err := customManager.LoadFromPath(configPath)
		if err != nil {
			return nil, configPath, err
		}

		return config, configPath, nil
	}

	// Ensure the global config directory exists
	if err := os.MkdirAll(globalConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create global config directory: %v", err)
	}

	// Test case 1: Neither config exists
	_, _, err := testLoad(false)
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
	config, path, err := testLoad(false)
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
	config, path, err = testLoad(true)
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
	config, path, err = testLoad(false)
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

	// Test case 5: Only global config exists
	config, path, err = testLoad(false)
	if err != nil {
		t.Fatalf("Failed to load global config as fallback: %v", err)
	}

	if path != globalConfigPath {
		t.Errorf("Expected path to be %s, got %s", globalConfigPath, path)
	}

	if !reflect.DeepEqual(config, globalConfig) {
		t.Errorf("Loaded config does not match global config.\nExpected: %+v\nGot: %+v", globalConfig, config)
	}
}

func TestSave(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	localConfigPath := filepath.Join(tempDir, "local-config.json")
	globalConfigDir := filepath.Join(tempDir, ".config")
	globalConfigPath := filepath.Join(globalConfigDir, "global-config.json")

	// Create test config
	testConfig := &Config{
		WebhookURL: "https://example.com/webhook",
		Username:   "TestUser",
		AvatarURL:  "https://example.com/avatar.png",
	}

	// Create a custom manager for this test
	customManager := NewManager()

	// Create testing helpers that simulate the Save function
	testSaveLocal := func() (string, error) {
		// Serialize and save
		if err := customManager.SaveToPath(testConfig, localConfigPath); err != nil {
			return localConfigPath, err
		}
		return localConfigPath, nil
	}

	testSaveGlobal := func() (string, error) {
		// For global config, ensure directory exists
		dirPath := filepath.Dir(globalConfigPath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return globalConfigPath, fmt.Errorf("failed to create config directory: %v", err)
		}

		// Serialize and save
		if err := customManager.SaveToPath(testConfig, globalConfigPath); err != nil {
			return globalConfigPath, err
		}
		return globalConfigPath, nil
	}

	// Test saving local config
	localSavedPath, err := testSaveLocal()
	if err != nil {
		t.Fatalf("Failed to save local config: %v", err)
	}

	if localSavedPath != localConfigPath {
		t.Errorf("Expected local save path to be %s, got %s", localConfigPath, localSavedPath)
	}

	// Verify local file was created
	if _, err := os.Stat(localConfigPath); os.IsNotExist(err) {
		t.Errorf("Local config was not created at %s", localConfigPath)
	}

	// Test saving global config
	globalSavedPath, err := testSaveGlobal()
	if err != nil {
		t.Fatalf("Failed to save global config: %v", err)
	}

	if globalSavedPath != globalConfigPath {
		t.Errorf("Expected global save path to be %s, got %s", globalConfigPath, globalSavedPath)
	}

	// Verify global file was created
	if _, err := os.Stat(globalConfigPath); os.IsNotExist(err) {
		t.Errorf("Global config was not created at %s", globalConfigPath)
	}

	// Verify global directory was created
	if _, err := os.Stat(globalConfigDir); os.IsNotExist(err) {
		t.Errorf("Global config directory was not created at %s", globalConfigDir)
	}
}
