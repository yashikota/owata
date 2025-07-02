package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/yashikota/owata/config"
)

const DefaultColor = 3447003 // Blue color

// Webhook represents the Discord webhook payload
type Webhook struct {
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

// SendNotification sends a notification to a Discord webhook
func SendNotification(webhookURL, message, source string, cfg *config.Config) error {
	// Set default values
	username := config.DefaultUsername
	var avatarURL string

	// Override with config values if available
	if cfg != nil {
		if cfg.Username != "" {
			username = cfg.Username
		}
		if cfg.AvatarURL != "" {
			avatarURL = cfg.AvatarURL
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
		Description: message,
		Color:       DefaultColor,
		Timestamp:   time.Now(),
		Fields: []Field{
			{
				Name:   "Working Directory",
				Value:  cwd,
				Inline: false,
			},
			{
				Name:   "Source",
				Value:  source,
				Inline: true,
			},
		},
		Footer: Footer{
			Text: "Owata",
		},
	}

	webhook := Webhook{
		Username:  username,
		AvatarURL: avatarURL,
		Embeds:    []Embed{embed},
	}

	// Marshal the webhook payload
	jsonData, err := json.Marshal(webhook)
	if err != nil {
		return fmt.Errorf("error marshaling webhook data: %v", err)
	}

	// Create HTTP client with timeout to prevent hanging requests
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Create request
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Send the webhook request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending webhook: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// Read response body for better error messages
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("discord webhook returned status %d, but failed to read response body: %v", resp.StatusCode, readErr)
	}
	return fmt.Errorf("discord webhook returned status: %d, body: %s", resp.StatusCode, string(body))
}
