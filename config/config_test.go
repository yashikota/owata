package config

import (
	"encoding/json"
	"os"
	"path/filepath"
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

	// Test global config path - create a temporary directory
	tempDir := t.TempDir()
	SetTestConfigDir(tempDir)
	defer ResetTestConfigDir()

	expectedGlobalPath := filepath.Join(tempDir, ConfigFileName)
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
	if loadedConfig.WebhookURL != testConfig.WebhookURL ||
		loadedConfig.Username != testConfig.Username ||
		loadedConfig.AvatarURL != testConfig.AvatarURL {
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
	if loadedConfig.WebhookURL != testConfig.WebhookURL ||
		loadedConfig.Username != testConfig.Username ||
		loadedConfig.AvatarURL != testConfig.AvatarURL {
		t.Errorf("Loaded config does not match test config.\nExpected: %+v\nGot: %+v", testConfig, loadedConfig)
	}
}

func TestCreateTemplate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Set test config directory
	SetTestConfigDir(tempDir)
	defer ResetTestConfigDir()

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
	// First, make sure the file doesn't exist already
	expectedGlobalPath := filepath.Join(tempDir, "owata-config.json")
	os.Remove(expectedGlobalPath)

	globalPath, created, err := manager.CreateTemplate(true)
	if err != nil {
		t.Fatalf("Failed to create global template: %v", err)
	}

	// Check that the file was created and the created flag is true
	if !created {
		t.Error("Expected created=true for new global template, got false")
	}

	// Expected global path
	if globalPath != expectedGlobalPath {
		t.Errorf("Expected global path %s, got %s", expectedGlobalPath, globalPath)
	}

	// Make a new test with the same function call for checking existence
	secondGlobalPath, created, err := manager.CreateTemplate(true)
	if err != nil {
		t.Fatalf("Failed on second global template creation: %v", err)
	}

	// Check that the path is the same
	if secondGlobalPath != globalPath {
		t.Errorf("Expected same path on second call, got %s vs %s", secondGlobalPath, globalPath)
	}

	// Since file exists now, created should be false
	if created {
		t.Error("Expected created=false for existing global template, got true")
	}

	// Check template content - empty config should have been created
	globalConfig, err := manager.LoadFromPath(expectedGlobalPath)
	if err != nil {
		t.Fatalf("Failed to load global config template: %v", err)
	}

	if globalConfig.WebhookURL != "" || globalConfig.Username != "" || globalConfig.AvatarURL != "" {
		t.Errorf("Global template config should have empty fields, got %+v", globalConfig)
	}
}

func TestLoad(t *testing.T) {
	// Create a fresh test environment for each test case
	t.Run("Case1_NeitherConfigExists", func(t *testing.T) {
		tempDir := t.TempDir()
		SetTestConfigDir(tempDir)
		defer ResetTestConfigDir()

		currentDir, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(currentDir)

		manager := NewManager()

		// Test with neither config existing
		_, _, err := manager.Load(false)
		if err == nil {
			t.Error("Expected error when no configs exist, got nil")
		}
	})

	t.Run("Case2_OnlyLocalConfigExists", func(t *testing.T) {
		tempDir := t.TempDir()
		SetTestConfigDir(tempDir)
		defer ResetTestConfigDir()

		currentDir, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(currentDir)

		manager := NewManager()

		// Create local config
		localConfig := &Config{
			WebhookURL: "https://example.com/local-webhook",
			Username:   "LocalUser",
			AvatarURL:  "https://example.com/local-avatar.png",
		}

		localPath := ConfigFileName
		localData, _ := json.MarshalIndent(localConfig, "", "  ")
		if err := os.WriteFile(localPath, localData, 0644); err != nil {
			t.Fatalf("Failed to write local config: %v", err)
		}

		// Test loading with local preference
		config, path, err := manager.Load(false)
		if err != nil {
			t.Fatalf("Failed to load local config: %v", err)
		}

		if path != localPath {
			t.Errorf("Expected path to be %s, got %s", localPath, path)
		}

		if config.WebhookURL != localConfig.WebhookURL ||
			config.Username != localConfig.Username ||
			config.AvatarURL != localConfig.AvatarURL {
			t.Errorf("Loaded config doesn't match local config.\nExpected: %+v\nGot: %+v", localConfig, config)
		}
	})

	t.Run("Case3_OnlyGlobalConfigExists", func(t *testing.T) {
		tempDir := t.TempDir()
		SetTestConfigDir(tempDir)
		defer ResetTestConfigDir()

		// Save original working directory
		currentDir, _ := os.Getwd()
		// Use the temp directory as working directory
		os.Chdir(tempDir)
		defer os.Chdir(currentDir)

		manager := NewManager()

		// Remove any local config if it exists
		os.Remove(ConfigFileName)

		// Create global config
		globalConfig := &Config{
			WebhookURL: "https://example.com/global-webhook",
			Username:   "GlobalUser",
			AvatarURL:  "https://example.com/global-avatar.png",
		}

		globalPath := filepath.Join(tempDir, ConfigFileName)
		globalData, _ := json.MarshalIndent(globalConfig, "", "  ")
		if err := os.WriteFile(globalPath, globalData, 0644); err != nil {
			t.Fatalf("Failed to write global config: %v", err)
		}

		// Test loading with local preference (should fallback to global)
		config, path, err := manager.Load(false)
		if err != nil {
			t.Fatalf("Failed to load global config as fallback: %v", err)
		}

		// On macOS, /var/folders might be symlinked to /private/var/folders
		// So we check if the file exists at the returned path instead of exact string comparison
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("The file at path %s does not exist", path)
		}

		if config.WebhookURL != globalConfig.WebhookURL ||
			config.Username != globalConfig.Username ||
			config.AvatarURL != globalConfig.AvatarURL {
			t.Errorf("Loaded config doesn't match global config.\nExpected: %+v\nGot: %+v", globalConfig, config)
		}

		// Test loading with global preference
		config, path, err = manager.Load(true)
		if err != nil {
			t.Fatalf("Failed to load global config when preferred: %v", err)
		}

		if path != globalPath {
			t.Errorf("Expected path to be %s, got %s", globalPath, path)
		}

		if config.WebhookURL != globalConfig.WebhookURL ||
			config.Username != globalConfig.Username ||
			config.AvatarURL != globalConfig.AvatarURL {
			t.Errorf("Loaded config doesn't match global config.\nExpected: %+v\nGot: %+v", globalConfig, config)
		}
	})

	t.Run("Case4_BothConfigsExist", func(t *testing.T) {
		// This test is more complex and requires careful isolation
		// Create an isolated directory structure
		tempDir := t.TempDir()

		// Override our function to return this directory
		SetTestConfigDir(tempDir)
		defer ResetTestConfigDir()

		// Keep track of original directory
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Move to temp dir
		err := os.Chdir(tempDir)
		if err != nil {
			t.Fatalf("Failed to chdir to temp dir: %v", err)
		}

		// Start with absolutely clean files
		os.Remove(ConfigFileName)
		os.Remove(filepath.Join(tempDir, ConfigFileName))

		// Create our test manager
		manager := NewManager()

		// First, create a local config
		localConfig := &Config{
			WebhookURL: "https://example.com/local-webhook",
			Username:   "LocalUser",
			AvatarURL:  "https://example.com/local-avatar.png",
		}

		// Write it directly to the current directory (which is tempDir)
		localPath := ConfigFileName // Just the filename, in current dir
		localJSON, _ := json.MarshalIndent(localConfig, "", "  ")
		err = os.WriteFile(localPath, localJSON, 0644)
		if err != nil {
			t.Fatalf("Failed to write local config: %v", err)
		}

		// Verify that local config exists and has expected content
		readLocalConfig, err := manager.LoadFromPath(localPath)
		if err != nil {
			t.Fatalf("Failed to read back local config: %v", err)
		}
		if readLocalConfig.WebhookURL != localConfig.WebhookURL ||
			readLocalConfig.Username != localConfig.Username {
			t.Fatalf("Local config content doesn't match what we wrote")
		}

		// Now test that the Load method correctly prefers the local config
		loadedConfig, _, err := manager.Load(false)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Verify the config contents match what we expect
		if loadedConfig.WebhookURL != localConfig.WebhookURL ||
			loadedConfig.Username != localConfig.Username ||
			loadedConfig.AvatarURL != localConfig.AvatarURL {
			t.Errorf("Loaded config doesn't match local config.\nExpected: %+v\nGot: %+v",
				localConfig, loadedConfig)
		}

		// Add test for global config preference separately
		// First create the global config
		globalConfig := &Config{
			WebhookURL: "https://example.com/global-webhook",
			Username:   "GlobalUser",
			AvatarURL:  "https://example.com/global-avatar.png",
		}

		// Write it to the expected global path
		globalPath := filepath.Join(tempDir, ConfigFileName)
		globalJSON, _ := json.MarshalIndent(globalConfig, "", "  ")
		err = os.WriteFile(globalPath, globalJSON, 0644)
		if err != nil {
			t.Fatalf("Failed to write global config: %v", err)
		}

		// Test loading with global preference
		globalLoadedConfig, _, err := manager.Load(true)
		if err != nil {
			t.Fatalf("Failed to load config with global preference: %v", err)
		}

		// Verify the config contents match what we expect
		if globalLoadedConfig.WebhookURL != globalConfig.WebhookURL ||
			globalLoadedConfig.Username != globalConfig.Username ||
			globalLoadedConfig.AvatarURL != globalConfig.AvatarURL {
			t.Errorf("Loaded config doesn't match global config.\nExpected: %+v\nGot: %+v",
				globalConfig, globalLoadedConfig)
		}
	})

	t.Run("Case5_GlobalRequestedButNotFound", func(t *testing.T) {
		tempDir := t.TempDir()
		SetTestConfigDir(tempDir)
		defer ResetTestConfigDir()

		// Save original working directory
		currentDir, _ := os.Getwd()
		// Use the temp directory as working directory
		os.Chdir(tempDir)
		defer os.Chdir(currentDir)

		manager := NewManager()

		// Make sure global config doesn't exist
		globalPath := filepath.Join(tempDir, ConfigFileName)
		os.Remove(globalPath)

		// Create only local config
		localConfig := &Config{
			WebhookURL: "https://example.com/local-webhook",
			Username:   "LocalUser",
			AvatarURL:  "https://example.com/local-avatar.png",
		}

		localPath := ConfigFileName
		localData, _ := json.MarshalIndent(localConfig, "", "  ")
		if err := os.WriteFile(localPath, localData, 0644); err != nil {
			t.Fatalf("Failed to write local config: %v", err)
		}

		// Delete any existing global config to make sure it doesn't exist
		// Important to handle the case where macOS has symlinked /var/folders to /private/var/folders
		os.Remove(filepath.Join(tempDir, ConfigFileName))
		os.Remove(filepath.Join("/private"+tempDir, ConfigFileName))

		// Test loading with global preference - should fail since global doesn't exist
		_, _, err := manager.Load(true)
		if err == nil {
			t.Error("Expected error when global config requested but not found, got nil")
		}
	})
}

func TestSave(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Set test config directory
	SetTestConfigDir(tempDir)
	defer ResetTestConfigDir()

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
	expectedGlobalPath := filepath.Join(tempDir, ConfigFileName)
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
	if loadedConfig.WebhookURL != testConfig.WebhookURL ||
		loadedConfig.Username != testConfig.Username ||
		loadedConfig.AvatarURL != testConfig.AvatarURL {
		t.Errorf("Loaded local config doesn't match original.\nExpected: %+v\nGot: %+v", testConfig, loadedConfig)
	}

	loadedConfig, err = testManager.LoadFromPath(globalSavedPath)
	if err != nil {
		t.Fatalf("Failed to load saved global config: %v", err)
	}
	if loadedConfig.WebhookURL != testConfig.WebhookURL ||
		loadedConfig.Username != testConfig.Username ||
		loadedConfig.AvatarURL != testConfig.AvatarURL {
		t.Errorf("Loaded global config doesn't match original.\nExpected: %+v\nGot: %+v", testConfig, loadedConfig)
	}
}
