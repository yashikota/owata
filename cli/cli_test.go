package cli

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedCmd    CommandType
		expectedErr    bool
		expectedGlobal bool
	}{
		{
			name:        "No arguments",
			args:        []string{},
			expectedErr: true,
		},
		{
			name:        "Help flag short",
			args:        []string{"-h"},
			expectedCmd: CommandShowHelp,
		},
		{
			name:        "Help flag long",
			args:        []string{"--help"},
			expectedCmd: CommandShowHelp,
		},
		{
			name:        "Version flag short",
			args:        []string{"-v"},
			expectedCmd: CommandShowVersion,
		},
		{
			name:        "Version flag long",
			args:        []string{"--version"},
			expectedCmd: CommandShowVersion,
		},
		{
			name:           "Init command",
			args:           []string{"init"},
			expectedCmd:    CommandInit,
			expectedGlobal: false,
		},
		{
			name:           "Init command with global flag short",
			args:           []string{"init", "-g"},
			expectedCmd:    CommandInit,
			expectedGlobal: true,
		},
		{
			name:           "Init command with global flag long",
			args:           []string{"init", "--global"},
			expectedCmd:    CommandInit,
			expectedGlobal: true,
		},
		{
			name:           "Config command",
			args:           []string{"config"},
			expectedCmd:    CommandConfig,
			expectedGlobal: false,
		},
		{
			name:           "Config command with global flag",
			args:           []string{"config", "-g"},
			expectedCmd:    CommandConfig,
			expectedGlobal: true,
		},
		{
			name:        "Notify command",
			args:        []string{"Hello world"},
			expectedCmd: CommandNotify,
		},
		{
			name:        "Notify command with webhook",
			args:        []string{"Hello world", "--webhook=https://example.com"},
			expectedCmd: CommandNotify,
		},
		{
			name:        "Notify command with source",
			args:        []string{"Hello world", "--source=Test"},
			expectedCmd: CommandNotify,
		},
		{
			name:        "Notify command with invalid option",
			args:        []string{"Hello world", "--invalid=option"},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := Parse(tt.args)
			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if args.Command != tt.expectedCmd {
				t.Errorf("Expected command %v, got %v", tt.expectedCmd, args.Command)
			}

			if tt.expectedCmd == CommandInit || tt.expectedCmd == CommandConfig {
				if args.Global != tt.expectedGlobal {
					t.Errorf("Expected Global=%v, got %v", tt.expectedGlobal, args.Global)
				}
			}
		})
	}
}

func TestParseConfigArgs(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedErr     bool
		expectedGlobal  bool
		expectedWebhook string
		expectedUser    string
		expectedAvatar  string
	}{
		{
			name: "Empty args",
			args: []string{},
		},
		{
			name:           "Global flag short",
			args:           []string{"-g"},
			expectedGlobal: true,
		},
		{
			name:           "Global flag long",
			args:           []string{"--global"},
			expectedGlobal: true,
		},
		{
			name:            "Webhook URL",
			args:            []string{"--webhook=https://example.com"},
			expectedWebhook: "https://example.com",
		},
		{
			name:         "Username",
			args:         []string{"--username=TestUser"},
			expectedUser: "TestUser",
		},
		{
			name:           "Avatar URL",
			args:           []string{"--avatar=https://example.com/avatar.png"},
			expectedAvatar: "https://example.com/avatar.png",
		},
		{
			name:            "Multiple arguments",
			args:            []string{"-g", "--webhook=https://example.com", "--username=TestUser"},
			expectedGlobal:  true,
			expectedWebhook: "https://example.com",
			expectedUser:    "TestUser",
		},
		{
			name:        "Unknown parameter",
			args:        []string{"--unknown=value"},
			expectedErr: true,
		},
		{
			name:            "Quoted values",
			args:            []string{"--webhook='https://example.com'", "--username=\"TestUser\""},
			expectedWebhook: "https://example.com",
			expectedUser:    "TestUser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := parseConfigArgs(tt.args)
			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if args.Command != CommandConfig {
				t.Errorf("Expected command type CommandConfig, got %v", args.Command)
			}

			if args.Global != tt.expectedGlobal {
				t.Errorf("Expected Global=%v, got %v", tt.expectedGlobal, args.Global)
			}

			if args.WebhookURL != tt.expectedWebhook {
				t.Errorf("Expected WebhookURL=%q, got %q", tt.expectedWebhook, args.WebhookURL)
			}

			if args.Username != tt.expectedUser {
				t.Errorf("Expected Username=%q, got %q", tt.expectedUser, args.Username)
			}

			if args.AvatarURL != tt.expectedAvatar {
				t.Errorf("Expected AvatarURL=%q, got %q", tt.expectedAvatar, args.AvatarURL)
			}
		})
	}
}

func TestParseNotifyArgs(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedErr     bool
		expectedMessage string
		expectedSource  string
		expectedWebhook string
	}{
		{
			name:        "Empty args",
			args:        []string{},
			expectedErr: true,
		},
		{
			name:            "Message only",
			args:            []string{"Hello world"},
			expectedMessage: "Hello world",
			expectedSource:  "Unknown", // Default source
		},
		{
			name:            "Message with webhook",
			args:            []string{"Hello world", "--webhook=https://example.com"},
			expectedMessage: "Hello world",
			expectedSource:  "Unknown", // Default source
			expectedWebhook: "https://example.com",
		},
		{
			name:            "Message with source",
			args:            []string{"Hello world", "--source=Test"},
			expectedMessage: "Hello world",
			expectedSource:  "Test",
		},
		{
			name:            "Message with webhook and source",
			args:            []string{"Hello world", "--webhook=https://example.com", "--source=Test"},
			expectedMessage: "Hello world",
			expectedSource:  "Test",
			expectedWebhook: "https://example.com",
		},
		{
			name:            "Message with quoted source",
			args:            []string{"Hello world", "--source='Test Source'"},
			expectedMessage: "Hello world",
			expectedSource:  "Test Source",
		},
		{
			name:        "Message with unknown flag",
			args:        []string{"Hello world", "--unknown=value"},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := parseNotifyArgs(tt.args)
			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if args.Command != CommandNotify {
				t.Errorf("Expected command type CommandNotify, got %v", args.Command)
			}

			if args.Message != tt.expectedMessage {
				t.Errorf("Expected Message=%q, got %q", tt.expectedMessage, args.Message)
			}

			if args.Source != tt.expectedSource {
				t.Errorf("Expected Source=%q, got %q", tt.expectedSource, args.Source)
			}

			if args.WebhookURL != tt.expectedWebhook {
				t.Errorf("Expected WebhookURL=%q, got %q", tt.expectedWebhook, args.WebhookURL)
			}
		})
	}
}

func TestPrintUsage(t *testing.T) {
	// Redirect stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintUsage()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Check that important parts are in the usage output
	expectedParts := []string{
		fmt.Sprintf("Owata v%s", Version),
		"Discord Webhook Notifier",
		"Usage:",
		"owata <message>",
		"owata init",
		"owata config",
		"-g, --global",
		"Commands:",
		"Options:",
		"Examples:",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Expected usage output to contain %q", part)
		}
	}
}

func TestPrintVersion(t *testing.T) {
	// Redirect stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintVersion()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	expected := fmt.Sprintf("Owata v%s\n", Version)
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}
