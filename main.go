package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Constants for the application
const (
	defaultUsername = "Owata"
	configPath      = "owata-config.json"
	defaultColor    = 3447003 // Blue color
	version         = "2.0.0"
)

// Config holds the configuration from owata-config.json
type Config struct {
	WebhookURL string `json:"webhook_url"`
	Username   string `json:"username"`
	AvatarURL  string `json:"avatar_url"`
}

// DiscordWebhook represents the Discord webhook payload
type DiscordWebhook struct {
	Username  string  `json:"username,omitempty"`
	AvatarURL string  `json:"avatar_url,omitempty"`
	Embeds    []Embed `json:"embeds"`
}

// Embed represents a Discord embed message
type Embed struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Color       int       `json:"color"`
	Timestamp   time.Time `json:"timestamp"`
	Fields      []Field   `json:"fields"`
	Footer      Footer    `json:"footer"`
}

// Field represents a field in a Discord embed
type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

// Footer represents the footer of a Discord embed
type Footer struct {
	Text string `json:"text"`
}

// CLIArgs holds the parsed command line arguments
type CLIArgs struct {
	Message      string
	WebhookURL   string
	Source       string
	ConfigMode   bool
	Username     string
	AvatarURL    string
	CreateConfig bool
}

func main() {
	args, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		printUsage()
		os.Exit(1)
	}

	// Handle create config flag
	if args.CreateConfig {
		if err := createConfigTemplate(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle config mode
	if args.ConfigMode {
		if err := handleConfigCommand(args); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Get webhook URL from config if not provided in args
	if args.WebhookURL == "" {
		config, err := loadConfig()
		if err != nil || config.WebhookURL == "" {
			fmt.Println("Error: No webhook URL provided. Use command line argument or config file.")
			os.Exit(1)
		}
		args.WebhookURL = config.WebhookURL
	}

	if err := sendDiscordNotification(args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// parseArgs parses command line arguments
func parseArgs(args []string) (*CLIArgs, error) {
	// Check for help or version flags first (even if no other args)
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			printUsage()
			os.Exit(0)
		}
		if arg == "--version" || arg == "-v" {
			printVersion()
			os.Exit(0)
		}
	}

	// Check for init command to create config template
	if len(args) > 0 && args[0] == "init" {
		return &CLIArgs{CreateConfig: true}, nil
	}

	// Check for config command
	if len(args) > 0 && args[0] == "config" {
		return parseConfigCommand(args[1:])
	}

	if len(args) < 1 {
		return nil, errors.New("missing required message argument")
	}

	result := &CLIArgs{
		Message: args[0],
		Source:  "Unknown", // Default source
	}

	// Parse arguments after the message
	for i := 1; i < len(args); i++ {
		arg := args[i]

		// Parse source flag
		if after, ok := strings.CutPrefix(arg, "--source="); ok {
			result.Source = after
			// Remove quotes if present
			result.Source = strings.Trim(result.Source, "'\"")
		} else if after, ok := strings.CutPrefix(arg, "--webhook="); ok {
			result.WebhookURL = strings.Trim(after, "'\"")
		} else {
			return nil, fmt.Errorf("unknown option: %s", arg)
		}
	}

	return result, nil
}

// printUsage prints command line usage information
func printUsage() {
	fmt.Printf("Owata v%s - Discord Webhook Notifier\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  owata <message> [--webhook=<url>] [--source=<source>]")
	fmt.Println("  owata init")
	fmt.Println("  owata config [--webhook=<url>] [--username=<name>] [--avatar=<url>]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  init                       Create configuration template file")
	fmt.Println("  config                     Show current configuration")
	fmt.Println("  config --webhook=<url>     Set Discord webhook URL")
	fmt.Println("  config --username=<name>   Set bot username")
	fmt.Println("  config --avatar=<url>      Set bot avatar URL")
	fmt.Println("")
	fmt.Println("Arguments:")
	fmt.Println("  message                    The notification message to send")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --webhook=<url>            Discord webhook URL (overrides config)")
	fmt.Println("  --source=<source>          Set the source of the notification")
	fmt.Println("  --help, -h                 Show this help message")
	fmt.Println("  --version, -v              Show version information")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  owata init                 # Create config template")
	fmt.Println("  owata config               # Show current settings")
	fmt.Println("  owata config --webhook='https://discord.com/api/webhooks/...'")
	fmt.Println("  owata 'Task completed!'    # Send notification (using config)")
	fmt.Println("  owata 'Build finished' --webhook='https://...' --source='CI'")
}

// printVersion prints version information
func printVersion() {
	fmt.Printf("Owata v%s\n", version)
}

// loadConfig loads configuration from the config file
func loadConfig() (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
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

// sendDiscordNotification sends a notification to a Discord webhook
func sendDiscordNotification(args *CLIArgs) error {
	config, _ := loadConfig() // Ignore error as we'll use defaults

	// Set default values
	username := defaultUsername
	var avatarURL string

	// Override with config values if available
	if config != nil {
		if config.Username != "" {
			username = config.Username
		}
		if config.AvatarURL != "" {
			avatarURL = config.AvatarURL
		}
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "Unknown"
	}

	// Create the Discord embed
	embed := Embed{
		Title:       "🔔 Notification",
		Description: args.Message,
		Color:       defaultColor,
		Timestamp:   time.Now(),
		Fields: []Field{
			{
				Name:   "Working Directory",
				Value:  cwd,
				Inline: false,
			},
			{
				Name:   "Source",
				Value:  args.Source,
				Inline: true,
			},
		},
		Footer: Footer{
			Text: "Owata",
		},
	}

	webhook := DiscordWebhook{
		Username:  username,
		AvatarURL: avatarURL,
		Embeds:    []Embed{embed},
	}

	// Marshal the webhook payload
	jsonData, err := json.Marshal(webhook)
	if err != nil {
		return fmt.Errorf("error marshaling webhook data: %v", err)
	}

	// Send the webhook request
	resp, err := http.Post(args.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error sending webhook: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode == http.StatusNoContent {
		fmt.Println("✅ Discord notification sent successfully")
		return nil
	}

	return fmt.Errorf("discord webhook returned status: %d", resp.StatusCode)
}

// parseConfigCommand parses the config subcommand arguments
func parseConfigCommand(args []string) (*CLIArgs, error) {
	result := &CLIArgs{
		ConfigMode: true,
	}

	// No parameters means show current config
	if len(args) == 0 {
		return result, nil
	}

	// Parse config arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "--webhook=") {
			if after, ok := strings.CutPrefix(arg, "--webhook="); ok {
				result.WebhookURL = strings.Trim(after, "'\"")
			}
		} else if strings.HasPrefix(arg, "--username=") {
			if after, ok := strings.CutPrefix(arg, "--username="); ok {
				result.Username = strings.Trim(after, "'\"")
			}
		} else if strings.HasPrefix(arg, "--avatar=") {
			if after, ok := strings.CutPrefix(arg, "--avatar="); ok {
				result.AvatarURL = strings.Trim(after, "'\"")
			}
		} else {
			return nil, fmt.Errorf("unknown config parameter: %s", arg)
		}
	}

	return result, nil
}

// handleConfigCommand handles the config subcommand
func handleConfigCommand(args *CLIArgs) error {
	// If no parameters were provided, show current configuration
	if args.WebhookURL == "" && args.Username == "" && args.AvatarURL == "" {
		return showCurrentConfig()
	}

	// Load existing config or create new one
	config, err := loadConfig()
	if err != nil {
		// If config doesn't exist, create new one
		config = &Config{}
	}

	// Update config with provided values
	if args.WebhookURL != "" {
		config.WebhookURL = args.WebhookURL
	}
	if args.Username != "" {
		config.Username = args.Username
	}
	if args.AvatarURL != "" {
		config.AvatarURL = args.AvatarURL
	}

	// Save config
	if err := saveConfig(config); err != nil {
		return err
	}

	fmt.Printf("✅ Configuration updated\n")

	// Display current config
	return showCurrentConfig()
}

// saveConfig saves configuration to the config file
func saveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// createConfigTemplate creates a configuration template file
func createConfigTemplate() error {
	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("⚠️  Configuration file %s already exists.\n", configPath)
		fmt.Println("Current configuration:")

		// Load and display existing config
		config, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to read existing config: %v", err)
		}

		if config.WebhookURL != "" {
			fmt.Printf("  Webhook URL: %s\n", config.WebhookURL)
		} else {
			fmt.Println("  Webhook URL: (not set)")
		}
		if config.Username != "" {
			fmt.Printf("  Username: %s\n", config.Username)
		} else {
			fmt.Println("  Username: (not set)")
		}
		if config.AvatarURL != "" {
			fmt.Printf("  Avatar URL: %s\n", config.AvatarURL)
		} else {
			fmt.Println("  Avatar URL: (not set)")
		}

		return nil
	}

	// Create JSON template
	templateContent := `{
  "webhook_url": "",
  "username": "",
  "avatar_url": ""
}`

	// Write template to file
	if err := os.WriteFile(configPath, []byte(templateContent), 0644); err != nil {
		return fmt.Errorf("failed to create config template: %v", err)
	}

	fmt.Printf("✅ Configuration template created: %s\n", configPath)
	fmt.Println("\nPlease edit the configuration file and set the following values:")
	fmt.Println("  webhook_url: Your Discord webhook URL")
	fmt.Println("  username:    Bot display name (optional)")
	fmt.Println("  avatar_url:  Bot avatar image URL (optional)")
	fmt.Println("\nOr use the config command with parameters:")
	fmt.Println("  owata config --webhook='https://discord.com/api/webhooks/...'")
	fmt.Println("  owata config --username='MyBot' --avatar='https://example.com/avatar.png'")

	return nil
}

// showCurrentConfig displays the current configuration
func showCurrentConfig() error {
	config, err := loadConfig()
	if err != nil {
		fmt.Println("❌ No configuration found. Run 'owata init' to create a config file.")
		return nil
	}

	fmt.Println("\n📋 Current configuration:")
	if config.WebhookURL != "" {
		// Hide webhook URL for security (show only last 10 characters)
		url := config.WebhookURL
		if len(url) > 10 {
			url = "..." + url[len(url)-10:]
		}
		fmt.Printf("  🔗 Webhook URL: %s\n", url)
	} else {
		fmt.Println("  🔗 Webhook URL: (not set)")
	}

	if config.Username != "" {
		fmt.Printf("  👤 Username: %s\n", config.Username)
	} else {
		fmt.Println("  👤 Username: (not set)")
	}

	if config.AvatarURL != "" {
		fmt.Printf("  🖼️  Avatar URL: %s\n", config.AvatarURL)
	} else {
		fmt.Println("  🖼️  Avatar URL: (not set)")
	}

	return nil
}
