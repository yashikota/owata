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
			fmt.Printf("Warning: Could not determine home directory: %v\n", err)
			return m.configFileName
		}
		return filepath.Join(homeDir, ".config", m.configFileName)
	}
	return m.configFileName
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
func (m *Manager) CreateTemplate(global bool) (string, error) {
	configPath := m.GetPath(global)

	// For global config, ensure directory exists
	if global {
		dirPath := filepath.Dir(configPath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return configPath, fmt.Errorf("failed to create config directory: %v", err)
		}
	}

	// Check if config file already exists
	if fileExists(configPath) {
		return configPath, nil
	}

	// Create JSON template
	templateContent := `{
  "webhook_url": "",
  "username": "",
  "avatar_url": ""
}`

	// Write template to file
	if err := os.WriteFile(configPath, []byte(templateContent), 0644); err != nil {
		return configPath, fmt.Errorf("failed to create config template: %v", err)
	}

	return configPath, nil
}

// DisplayConfig shows the current config at the specified path
func (m *Manager) DisplayConfig(path string) error {
	config, err := m.LoadFromPath(path)
	if err != nil {
		return err
	}

	fmt.Printf("\n📋 Current configuration (%s):\n", path)

	// Display webhook URL (with security masking)
	if config.WebhookURL != "" {
		url := config.WebhookURL
		if len(url) > 10 {
			url = "..." + url[len(url)-10:]
		}
		fmt.Printf("  🔗 Webhook URL: %s\n", url)
	} else {
		fmt.Println("  🔗 Webhook URL: (not set)")
	}

	// Display username
	if config.Username != "" {
		fmt.Printf("  👤 Username: %s\n", config.Username)
	} else {
		fmt.Println("  👤 Username: (not set)")
	}

	// Display avatar URL
	if config.AvatarURL != "" {
		fmt.Printf("  🖼️  Avatar URL: %s\n", config.AvatarURL)
	} else {
		fmt.Println("  🖼️  Avatar URL: (not set)")
	}

	return nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
