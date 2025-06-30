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
	Message    string
	WebhookURL string
	Source     string
}

func main() {
	args, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		printUsage()
		os.Exit(1)
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
		if strings.HasPrefix(arg, "--source=") {
			result.Source = strings.TrimPrefix(arg, "--source=")
			// Remove quotes if present
			result.Source = strings.Trim(result.Source, "'\"")
		} else if !strings.HasPrefix(arg, "--") && result.WebhookURL == "" {
			// If not a flag and webhook not set, assume it's the webhook URL
			result.WebhookURL = arg
		}
	}

	return result, nil
}

// printUsage prints command line usage information
func printUsage() {
	fmt.Println("Usage: owata <message> [webhook-url] [--source=<source>]")
	fmt.Println("Example: owata 'Claude Code session ended'")
	fmt.Println("Example: owata 'Task completed' https://discord.com/api/webhooks/...")
	fmt.Println("Example: owata 'Task completed' --source='Claude Code'")
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
			Text: "Owata Notifier",
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