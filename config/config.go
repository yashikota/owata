package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ConfigFileName  = "owata-config.json"
	DefaultUsername = "Owata"
)

// Sentinel errors
var (
	ErrConfigFileNotFound = errors.New("config file not found")
)

type Config struct {
	WebhookURL string `json:"webhook_url"`
	Username   string `json:"username"`
	AvatarURL  string `json:"avatar_url"`
}

type Manager struct {
	configFileName string
}

func NewManager() *Manager {
	return &Manager{
		configFileName: ConfigFileName,
	}
}

// For testing purposes
var userConfigDirFunc = os.UserConfigDir

func (m *Manager) GetPathWithError(global bool) (string, error) {
	if global {
		configDir, err := userConfigDirFunc()
		if err != nil {
			return "", fmt.Errorf("could not determine config directory: %w", err)
		}
		return filepath.Join(configDir, m.configFileName), nil
	}
	return m.configFileName, nil
}

func (m *Manager) Load(preferGlobal bool) (*Config, string, error) {
	localPath, _ := m.GetPathWithError(false)
	globalPath, globalPathErr := m.GetPathWithError(true)

	// If we can't get global path but it was requested, return the error
	if preferGlobal && globalPathErr != nil {
		return nil, "", fmt.Errorf("failed to get global config path: %w", globalPathErr)
	}

	localExists := fileExists(localPath)
	globalExists := fileExists(globalPath)

	var configPath string

	if preferGlobal {
		if !globalExists {
			return nil, "", fmt.Errorf("global config file not found at %s", globalPath)
		}
		// Only assign configPath if globalExists is true (we wouldn't reach here otherwise)
		configPath = globalPath
	} else if localExists {
		configPath = localPath
	} else if globalExists {
		configPath = globalPath
	} else {
		return nil, "", fmt.Errorf("config file not found: neither %s nor %s exists", localPath, globalPath)
	}

	config, err := m.LoadFromPath(configPath)
	if err != nil {
		return nil, configPath, err
	}

	return config, configPath, nil
}

func (m *Manager) LoadFromPath(configPath string) (*Config, error) {
	if !fileExists(configPath) {
		return nil, fmt.Errorf("%w: %s", ErrConfigFileNotFound, configPath)
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

func (m *Manager) Save(config *Config, global bool) (string, error) {
	configPath, pathErr := m.GetPathWithError(global)
	if pathErr != nil {
		return "", fmt.Errorf("failed to get config path: %w", pathErr)
	}

	// Ensure directory exists - only needed for non-current directories
	dirPath := filepath.Dir(configPath)
	if dirPath != "." {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return "", fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	if err := m.SaveToPath(config, configPath); err != nil {
		return configPath, err
	}

	return configPath, nil
}

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

func (m *Manager) CreateTemplate(global bool) (string, bool, error) {
	configPath, pathErr := m.GetPathWithError(global)
	if pathErr != nil {
		return "", false, fmt.Errorf("failed to get config path: %w", pathErr)
	}

	// Ensure directory exists - only needed for non-current directories
	dirPath := filepath.Dir(configPath)
	if dirPath != "." {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return "", false, fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	if fileExists(configPath) {
		return configPath, false, nil // File already exists, not created
	}

	templateContent := `{
  "webhook_url": "",
  "username": "",
  "avatar_url": ""
}`

	if err := os.WriteFile(configPath, []byte(templateContent), 0644); err != nil {
		return configPath, false, fmt.Errorf("failed to create config template: %v", err)
	}

	return configPath, true, nil // New file was created
}

func (m *Manager) DisplayConfig(path string) (string, error) {
	config, err := m.LoadFromPath(path)
	if err != nil {
		return "", err
	}

	var output string
	output += fmt.Sprintf("\n📋 Current configuration (%s):\n", path)

	if config.WebhookURL != "" {
		// Safely obfuscate the webhook URL - show only last few characters
		url := config.WebhookURL
		if len(url) > 10 {
			// Take last 10 characters only
			lastTen := url[len(url)-10:]
			url = "..." + lastTen
		}
		output += fmt.Sprintf("  🔗 Webhook URL: %s\n", url)
	} else {
		output += "  🔗 Webhook URL: (not set)\n"
	}

	if config.Username != "" {
		output += fmt.Sprintf("  👤 Username: %s\n", config.Username)
	} else {
		output += "  👤 Username: (not set)\n"
	}

	if config.AvatarURL != "" {
		output += fmt.Sprintf("  🖼️  Avatar URL: %s\n", config.AvatarURL)
	} else {
		output += "  🖼️  Avatar URL: (not set)\n"
	}

	return output, nil
}

func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // File does not exist
		}
		return false, err // Propagate other errors
	}
	return !info.IsDir(), nil // Make sure it's not a directory
}
