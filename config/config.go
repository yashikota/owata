package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Constants
const (
	ConfigFileName  = "owata-config.json"
	DefaultUsername = "Owata"
)

// Config holds the configuration from owata-config.json
type Config struct {
	WebhookURL string `json:"webhook_url"`
	Username   string `json:"username"`
	AvatarURL  string `json:"avatar_url"`
}

// Manager handles configuration operations
type Manager struct {
	configFileName string
}

// NewManager creates a new config manager with the default configuration file name
func NewManager() *Manager {
	return &Manager{
		configFileName: ConfigFileName,
	}
}

// GetPath returns the path to the config file based on whether global config is requested
func (m *Manager) GetPath(global bool) string {
	if global {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fall back to local config if we can't determine home directory
			return m.configFileName
		}
		return filepath.Join(homeDir, ".config", m.configFileName)
	}
	return m.configFileName
}

// GetPathWithError returns the path to the config file and any error that occurred
func (m *Manager) GetPathWithError(global bool) (string, error) {
	if global {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return m.configFileName, fmt.Errorf("could not determine home directory: %w", err)
		}
		return filepath.Join(homeDir, ".config", m.configFileName), nil
	}
	return m.configFileName, nil
}

// Load loads configuration from local or global file based on preference
func (m *Manager) Load(preferGlobal bool) (*Config, string, error) {
	// Determine which paths to check
	localPath := m.GetPath(false)
	globalPath := m.GetPath(true)

	// Check existence
	localExists := fileExists(localPath)
	globalExists := fileExists(globalPath)

	// Choose which config to load
	var configPath string

	if preferGlobal && globalExists {
		configPath = globalPath
	} else if localExists {
		configPath = localPath
	} else if globalExists {
		configPath = globalPath
	} else {
		return nil, "", fmt.Errorf("config file not found: neither %s nor %s exists", localPath, globalPath)
	}

	// Load the config from the chosen path
	config, err := m.LoadFromPath(configPath)
	if err != nil {
		return nil, configPath, err
	}

	return config, configPath, nil
}

// LoadFromPath loads configuration from the specified config file path
func (m *Manager) LoadFromPath(configPath string) (*Config, error) {
	if !fileExists(configPath) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// Save saves configuration to the specified path (local or global)
func (m *Manager) Save(config *Config, global bool) (string, error) {
	configPath := m.GetPath(global)

	// For global config, ensure directory exists
	if global {
		dirPath := filepath.Dir(configPath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return configPath, fmt.Errorf("failed to create config directory: %v", err)
		}
	}

	// Serialize and save
	if err := m.SaveToPath(config, configPath); err != nil {
		return configPath, err
	}

	return configPath, nil
}

// SaveToPath saves configuration to the specified path
func (m *Manager) SaveToPath(config *Config, configPath string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// CreateTemplate creates a configuration template file
// Returns the config path, a boolean indicating if a new file was created, and any error
func (m *Manager) CreateTemplate(global bool) (string, bool, error) {
	configPath := m.GetPath(global)

	// For global config, ensure directory exists
	if global {
		dirPath := filepath.Dir(configPath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return configPath, false, fmt.Errorf("failed to create config directory: %v", err)
		}
	}

	// Check if config file already exists
	if fileExists(configPath) {
		return configPath, false, nil // File already exists, not created
	}

	// Create JSON template
	templateContent := `{
  "webhook_url": "",
  "username": "",
  "avatar_url": ""
}`

	// Write template to file
	if err := os.WriteFile(configPath, []byte(templateContent), 0644); err != nil {
		return configPath, false, fmt.Errorf("failed to create config template: %v", err)
	}

	return configPath, true, nil // New file was created
}

// DisplayConfig shows the current config at the specified path
// Returns the formatted config information and an error if any
func (m *Manager) DisplayConfig(path string) (string, error) {
	config, err := m.LoadFromPath(path)
	if err != nil {
		return "", err
	}

	var output string
	output += fmt.Sprintf("\n📋 Current configuration (%s):\n", path)

	// Format webhook URL (with security masking)
	if config.WebhookURL != "" {
		url := config.WebhookURL
		if len(url) > 10 {
			url = "..." + url[len(url)-10:]
		}
		output += fmt.Sprintf("  🔗 Webhook URL: %s\n", url)
	} else {
		output += "  🔗 Webhook URL: (not set)\n"
	}

	// Format username
	if config.Username != "" {
		output += fmt.Sprintf("  👤 Username: %s\n", config.Username)
	} else {
		output += "  👤 Username: (not set)\n"
	}

	// Format avatar URL
	if config.AvatarURL != "" {
		output += fmt.Sprintf("  🖼️  Avatar URL: %s\n", config.AvatarURL)
	} else {
		output += "  🖼️  Avatar URL: (not set)\n"
	}

	return output, nil
}

// fileExists checks if a file exists and is accessible
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return !os.IsNotExist(err) // Return true for permission errors, etc.
	}
	return !info.IsDir() // Make sure it's not a directory
}
