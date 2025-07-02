package discord

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yashikota/owata/config"
)

// Mock HTTP server for testing webhook requests
func setupMockServer(t *testing.T, statusCode int, validatePayload func(payload *Webhook)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", contentType)
		}

		// Read and validate payload
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		// Unmarshal webhook payload
		var payload Webhook
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}

		// Call payload validator if provided
		if validatePayload != nil {
			validatePayload(&payload)
		}

		// Set response status
		w.WriteHeader(statusCode)
	}))
}

func TestSendNotification(t *testing.T) {
	// Test cases
	tests := []struct {
		name        string
		message     string
		source      string
		config      *config.Config
		statusCode  int
		expectError bool
		validator   func(*Webhook)
	}{
		{
			name:       "Successful notification with default config",
			message:    "Test message",
			source:     "Test",
			statusCode: http.StatusNoContent,
			validator: func(payload *Webhook) {
				if payload.Username != config.DefaultUsername {
					t.Errorf("Expected username %s, got %s", config.DefaultUsername, payload.Username)
				}
				if payload.AvatarURL != "" {
					t.Errorf("Expected empty avatar URL, got %s", payload.AvatarURL)
				}
				if len(payload.Embeds) != 1 {
					t.Fatalf("Expected 1 embed, got %d", len(payload.Embeds))
				}
				embed := payload.Embeds[0]
				if embed.Description != "Test message" {
					t.Errorf("Expected description %q, got %q", "Test message", embed.Description)
				}

				// Find source field
				var sourceFound bool
				for _, field := range embed.Fields {
					if field.Name == "Source" && field.Value == "Test" {
						sourceFound = true
						break
					}
				}
				if !sourceFound {
					t.Errorf("Source field with value 'Test' not found in embed fields")
				}
			},
		},
		{
			name:    "Successful notification with custom config",
			message: "Test message",
			source:  "Test",
			config: &config.Config{
				Username:  "CustomUser",
				AvatarURL: "https://example.com/avatar.png",
			},
			statusCode: http.StatusNoContent,
			validator: func(payload *Webhook) {
				if payload.Username != "CustomUser" {
					t.Errorf("Expected username %s, got %s", "CustomUser", payload.Username)
				}
				if payload.AvatarURL != "https://example.com/avatar.png" {
					t.Errorf("Expected avatar URL %s, got %s", "https://example.com/avatar.png", payload.AvatarURL)
				}
			},
		},
		{
			name:        "Failed notification",
			message:     "Test message",
			source:      "Test",
			statusCode:  http.StatusBadRequest,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server for this test case
			server := setupMockServer(t, tt.statusCode, tt.validator)
			defer server.Close()

			// Send notification
			err := SendNotification(server.URL, tt.message, tt.source, tt.config)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Test marshalling and structure of webhook payload
func TestWebhookPayload(t *testing.T) {
	webhook := Webhook{
		Username:  "TestUser",
		AvatarURL: "https://example.com/avatar.png",
		Embeds: []Embed{
			{
				Title:       "Test Title",
				Description: "Test Description",
				Color:       12345,
				Fields: []Field{
					{
						Name:   "Field1",
						Value:  "Value1",
						Inline: true,
					},
					{
						Name:   "Field2",
						Value:  "Value2",
						Inline: false,
					},
				},
				Footer: Footer{
					Text: "Test Footer",
				},
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(webhook)
	if err != nil {
		t.Fatalf("Failed to marshal webhook: %v", err)
	}

	// Unmarshal back to verify structure
	var unmarshalled Webhook
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal webhook: %v", err)
	}

	// Verify fields
	if unmarshalled.Username != webhook.Username {
		t.Errorf("Username mismatch: expected %q, got %q", webhook.Username, unmarshalled.Username)
	}

	if unmarshalled.AvatarURL != webhook.AvatarURL {
		t.Errorf("AvatarURL mismatch: expected %q, got %q", webhook.AvatarURL, unmarshalled.AvatarURL)
	}

	if len(unmarshalled.Embeds) != 1 {
		t.Fatalf("Expected 1 embed, got %d", len(unmarshalled.Embeds))
	}

	embed := unmarshalled.Embeds[0]
	if embed.Title != "Test Title" {
		t.Errorf("Embed title mismatch: expected %q, got %q", "Test Title", embed.Title)
	}

	if embed.Description != "Test Description" {
		t.Errorf("Embed description mismatch: expected %q, got %q", "Test Description", embed.Description)
	}

	if embed.Color != 12345 {
		t.Errorf("Embed color mismatch: expected %d, got %d", 12345, embed.Color)
	}

	if len(embed.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(embed.Fields))
	}

	if embed.Footer.Text != "Test Footer" {
		t.Errorf("Footer text mismatch: expected %q, got %q", "Test Footer", embed.Footer.Text)
	}

	// Check specific field properties
	field1 := embed.Fields[0]
	if field1.Name != "Field1" || field1.Value != "Value1" || !field1.Inline {
		t.Errorf("Field1 mismatch: expected {Name:Field1, Value:Value1, Inline:true}, got %+v", field1)
	}

	field2 := embed.Fields[1]
	if field2.Name != "Field2" || field2.Value != "Value2" || field2.Inline {
		t.Errorf("Field2 mismatch: expected {Name:Field2, Value:Value2, Inline:false}, got %+v", field2)
	}
}
