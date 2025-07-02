package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yashikota/owata/cli"
	"github.com/yashikota/owata/config"
	"github.com/yashikota/owata/discord"
)

// TestInitCommand tests the init command functionality
func TestInitCommand(t *testing.T) {
	// Create a temp directory for test
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a config manager
	manager := config.NewManager()

	// Run init command directly
	path, _, err := manager.CreateTemplate(false)
	if err != nil {
		t.Fatalf("Failed to create config template: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Config file was not created at %s", path)
	}

	// Check content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Failed to parse config file: %v", err)
	}

	if cfg.WebhookURL != "" || cfg.Username != "" || cfg.AvatarURL != "" {
		t.Errorf("Expected empty config values, got %+v", cfg)
	}
}

// TestConfigCommand tests the config command functionality
func TestConfigCommand(t *testing.T) {
	// Create a temp directory for test
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a config manager
	manager := config.NewManager()

	// Create initial config
	testConfig := &config.Config{
		WebhookURL: "",
		Username:   "",
		AvatarURL:  "",
	}

	// Save initial config
	path, err := manager.Save(testConfig, false)
	if err != nil {
		t.Fatalf("Failed to save initial config: %v", err)
	}

	// Update config with new values
	testConfig.Username = "TestUser"
	testConfig.AvatarURL = "https://example.com/avatar.png"

	// Save updated config
	_, err = manager.Save(testConfig, false)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Check that config file was updated
	loadedConfig, err := manager.LoadFromPath(path)
	if err != nil {
		t.Fatalf("Failed to load updated config: %v", err)
	}

	if loadedConfig.Username != "TestUser" {
		t.Errorf("Expected username to be 'TestUser', got %q", loadedConfig.Username)
	}

	if loadedConfig.AvatarURL != "https://example.com/avatar.png" {
		t.Errorf("Expected avatar URL to be 'https://example.com/avatar.png', got %q", loadedConfig.AvatarURL)
	}
}

// TestGlobalConfig tests the global config functionality
func TestGlobalConfig(t *testing.T) {
	// Create a temp directory for test
	tempDir := t.TempDir()

	// Setup mock home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create a config manager
	manager := config.NewManager()

	// Create global config
	path, _, err := manager.CreateTemplate(true)
	if err != nil {
		t.Fatalf("Failed to create global config: %v", err)
	}

	// Check global path
	expectedPath := filepath.Join(tempDir, ".config", config.ConfigFileName)
	if path != expectedPath {
		t.Errorf("Expected global path to be %q, got %q", expectedPath, path)
	}

	// Check file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Global config file was not created")
	}

	// Update global config
	testConfig := &config.Config{
		WebhookURL: "https://example.com/webhook",
		Username:   "GlobalUser",
		AvatarURL:  "https://example.com/avatar.png",
	}

	// Save updated config
	path, err = manager.Save(testConfig, true)
	if err != nil {
		t.Fatalf("Failed to update global config: %v", err)
	}

	// Check that config file was updated
	loadedConfig, err := manager.LoadFromPath(path)
	if err != nil {
		t.Fatalf("Failed to load updated global config: %v", err)
	}

	if loadedConfig.Username != "GlobalUser" {
		t.Errorf("Expected username to be 'GlobalUser', got %q", loadedConfig.Username)
	}
}

// TestNotification tests the notification sending functionality directly
func TestNotification(t *testing.T) {
	// Create test server
	var requestReceived bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		// Check if it's a webhook request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		// Return success
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create test config
	testConfig := &config.Config{
		Username:  "TestUser",
		AvatarURL: "https://example.com/avatar.png",
	}

	// Send notification
	err := discord.SendNotification(server.URL, "Test message", "TestSource", testConfig)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}

	// Check request was received
	if !requestReceived {
		t.Error("No request was received by test server")
	}
}

// TestHandleNotify tests the handleNotify function specifically (integration test)
func TestHandleNotify(t *testing.T) {
	// Create test server
	var requestReceived bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		// Check content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		// Return success
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Setup test environment
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a config manager
	manager := config.NewManager()

	// Create test cases
	tests := []struct {
		name         string
		args         *cli.Args
		setupLocal   bool
		setupGlobal  bool
		expectError  bool
		expectGlobal bool
	}{
		{
			name: "Command line webhook only",
			args: &cli.Args{
				Command:    cli.CommandNotify,
				Message:    "Test message",
				WebhookURL: server.URL,
				Source:     "Test",
				Global:     false,
			},
			setupLocal:  false,
			setupGlobal: false,
			expectError: false,
		},
		{
			name: "Local config exists, no global flag",
			args: &cli.Args{
				Command: cli.CommandNotify,
				Message: "Test message",
				Source:  "Test",
				Global:  false,
			},
			setupLocal:  true,
			setupGlobal: false,
			expectError: false,
		},
		{
			name: "Global config exists, with global flag",
			args: &cli.Args{
				Command: cli.CommandNotify,
				Message: "Test message",
				Source:  "Test",
				Global:  true,
			},
			setupLocal:   false,
			setupGlobal:  true,
			expectError:  false,
			expectGlobal: true,
		},
		{
			name: "Both configs exist, with global flag",
			args: &cli.Args{
				Command: cli.CommandNotify,
				Message: "Test message",
				Source:  "Test",
				Global:  true,
			},
			setupLocal:   true,
			setupGlobal:  true,
			expectError:  false,
			expectGlobal: true,
		},
		{
			name: "Both configs exist, no global flag (prefer local)",
			args: &cli.Args{
				Command: cli.CommandNotify,
				Message: "Test message",
				Source:  "Test",
				Global:  false,
			},
			setupLocal:  true,
			setupGlobal: true,
			expectError: false,
		},
		{
			name: "No configs exist, no webhook URL",
			args: &cli.Args{
				Command: cli.CommandNotify,
				Message: "Test message",
				Source:  "Test",
				Global:  false,
			},
			setupLocal:  false,
			setupGlobal: false,
			expectError: true,
		},
		{
			name: "Global flag but no global config exists",
			args: &cli.Args{
				Command: cli.CommandNotify,
				Message: "Test message",
				Source:  "Test",
				Global:  true,
			},
			setupLocal:  true,
			setupGlobal: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset request flag
			requestReceived = false

			// Clean any existing config files
			os.Remove(config.ConfigFileName) // local
			homeDir, _ := os.UserHomeDir()
			os.Remove(filepath.Join(homeDir, ".config", config.ConfigFileName)) // global

			// Setup local config if needed
			if tt.setupLocal {
				localConfig := &config.Config{
					WebhookURL: server.URL,
					Username:   "LocalUser",
					AvatarURL:  "https://example.com/local-avatar.png",
				}
				_, err := manager.Save(localConfig, false)
				if err != nil {
					t.Fatalf("Failed to setup local config: %v", err)
				}
			}

			// Setup global config if needed
			if tt.setupGlobal {
				globalConfig := &config.Config{
					WebhookURL: server.URL,
					Username:   "GlobalUser",
					AvatarURL:  "https://example.com/global-avatar.png",
				}
				_, err := manager.Save(globalConfig, true)
				if err != nil {
					t.Fatalf("Failed to setup global config: %v", err)
				}
			}

			// Redirect stdout to capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run the test
			err := handleNotify(manager, tt.args)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout
			var output bytes.Buffer
			output.ReadFrom(r)
			outputStr := output.String()

			// Check results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}

				// Check for success message
				if !strings.Contains(outputStr, "Discord notification sent successfully") {
					t.Error("Expected success message in output")
				}

				// Check that request was sent
				if !requestReceived {
					t.Error("No request was received by test server")
				}
			}
		})
	}
}

// TestPrintUsage tests the help output using the CLI package's PrintUsage function
func TestPrintUsage(t *testing.T) {
	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Import and use the PrintUsage function from CLI package
	cli.PrintUsage()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Check for expected content
	if !strings.Contains(output, "Owata") ||
		!strings.Contains(output, "Discord Webhook Notifier") ||
		!strings.Contains(output, "Usage:") {
		t.Errorf("Help output missing expected content")
	}
}
