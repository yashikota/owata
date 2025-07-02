package cli

import (
	"fmt"
	"strings"
)

const Version = "2.0.0"

// Command represents the CLI command to execute
type CommandType int

const (
	CommandNotify CommandType = iota
	CommandInit
	CommandConfig
	CommandShowHelp
	CommandShowVersion
)

// Args holds the parsed command line arguments
type Args struct {
	Command     CommandType
	Message     string
	WebhookURL  string
	Source      string
	Username    string
	AvatarURL   string
	Global      bool
}

// Parse parses command line arguments
func Parse(args []string) (*Args, error) {
	// Check for missing arguments
	if len(args) < 1 {
		return nil, fmt.Errorf("missing arguments")
	}

	// Check for help or version flags first (even if no other args)
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return &Args{Command: CommandShowHelp}, nil
		}
		if arg == "--version" || arg == "-v" {
			return &Args{Command: CommandShowVersion}, nil
		}
	}

	// Check for init command
	if args[0] == "init" {
		result := &Args{Command: CommandInit}

		// Check for global flag
		if len(args) > 1 && (args[1] == "-g" || args[1] == "--global") {
			result.Global = true
		}

		return result, nil
	}

	// Check for config command
	if args[0] == "config" {
		return parseConfigArgs(args[1:])
	}

	// Default is notification command
	return parseNotifyArgs(args)
}

// parseNotifyArgs parses arguments for the notify command
func parseNotifyArgs(args []string) (*Args, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("missing required message argument")
	}

	result := &Args{
		Command: CommandNotify,
		Message: args[0],
		Source:  "Unknown", // Default source
	}

	// Parse arguments after the message
	for i := 1; i < len(args); i++ {
		arg := args[i]

		// Parse source flag
		if after, ok := strings.CutPrefix(arg, "--source="); ok {
			result.Source = strings.Trim(after, "'\"")
		} else if after, ok := strings.CutPrefix(arg, "--webhook="); ok {
			result.WebhookURL = strings.Trim(after, "'\"")
		} else {
			return nil, fmt.Errorf("unknown option: %s", arg)
		}
	}

	return result, nil
}

// parseConfigArgs parses arguments for the config command
func parseConfigArgs(args []string) (*Args, error) {
	result := &Args{
		Command: CommandConfig,
	}

	// No parameters means show current config
	if len(args) == 0 {
		return result, nil
	}

	// Parse config arguments
	for i := range args {
		arg := args[i]

		if arg == "-g" || arg == "--global" {
			result.Global = true
		} else if strings.HasPrefix(arg, "--webhook=") {
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

// PrintUsage prints command line usage information
func PrintUsage() {
	fmt.Printf("Owata v%s - Discord Webhook Notifier\n\n", Version)
	fmt.Println("Usage:")
	fmt.Println("  owata <message> [--webhook=<url>] [--source=<source>]")
	fmt.Println("  owata init [-g|--global]")
	fmt.Println("  owata config [-g|--global] [--webhook=<url>] [--username=<name>] [--avatar=<url>]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  init                       Create local configuration template file")
	fmt.Println("  init -g, --global          Create global configuration template file")
	fmt.Println("  config                     Show current local configuration")
	fmt.Println("  config -g, --global        Show current global configuration")
	fmt.Println("  config --webhook=<url>     Set Discord webhook URL in local config")
	fmt.Println("  config -g --webhook=<url>  Set Discord webhook URL in global config")
	fmt.Println("  config --username=<name>   Set bot username")
	fmt.Println("  config --avatar=<url>      Set bot avatar URL")
	fmt.Println("")
	fmt.Println("Arguments:")
	fmt.Println("  message                    The notification message to send")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --webhook=<url>            Discord webhook URL (overrides config)")
	fmt.Println("  --source=<source>          Set the source of the notification")
	fmt.Println("  -g, --global               Use global configuration (~/.config/owata-config.json)")
	fmt.Println("  --help, -h                 Show this help message")
	fmt.Println("  --version, -v              Show version information")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  owata init                 # Create local config template")
	fmt.Println("  owata init -g              # Create global config template")
	fmt.Println("  owata config               # Show current local settings")
	fmt.Println("  owata config -g            # Show current global settings")
	fmt.Println("  owata config --webhook='https://discord.com/api/webhooks/...'")
	fmt.Println("  owata config -g --username='GlobalBot'")
	fmt.Println("  owata 'Task completed!'    # Send notification (using config)")
	fmt.Println("  owata 'Build finished' --webhook='https://...' --source='CI'")
}

// PrintVersion prints version information
func PrintVersion() {
	fmt.Printf("Owata v%s\n", Version)
}
