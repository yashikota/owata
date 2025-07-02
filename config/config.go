package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ConfigFileName  = "owata-config.json"
	DefaultUsername = "Owata"
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

func (m *Manager) GetPath(global bool) string {
	path, _ := m.GetPathWithError(global)
	return path
}

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

func (m *Manager) Load(preferGlobal bool) (*Config, string, error) {
	localPath := m.GetPath(false)
	globalPath := m.GetPath(true)

	localExists := fileExists(localPath)
	globalExists := fileExists(globalPath)

	var configPath string

	if preferGlobal {
		if !globalExists {
			return nil, "", fmt.Errorf("global config file not found at %s", globalPath)
		}
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

func (m *Manager) Save(config *Config, global bool) (string, error) {
	configPath := m.GetPath(global)

	if global {
		dirPath := filepath.Dir(configPath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return configPath, fmt.Errorf("failed to create config directory: %v", err)
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
	configPath := m.GetPath(global)

	if global {
		dirPath := filepath.Dir(configPath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return configPath, false, fmt.Errorf("failed to create config directory: %v", err)
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
		url := config.WebhookURL
		if len(url) > 10 {
			url = "..." + url[len(url)-10:]
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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		// Any error from Stat (e.g., not found, permission denied) means we can't treat it as an existing file.
		return false
	}
	return !info.IsDir() // Make sure it's not a directory
}
