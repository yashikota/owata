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
	Command    CommandType
	Message    string
	WebhookURL string
	Source     string
	Username   string
	AvatarURL  string
	Global     bool
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

	// Extract global flag if present at the beginning of arguments
	// This allows commands like "owata -g init" or "owata --global config"
	var globalFlag bool
	var processedArgs []string

	// Process args to extract global flag at any position before the command
	for i := range args {
		if args[i] == "-g" || args[i] == "--global" {
			globalFlag = true
		} else {
			processedArgs = append(processedArgs, args[i])
		}
	}

	// If no args left after global flag extraction, treat as notification command
	if len(processedArgs) == 0 {
		return &Args{
			Command: CommandNotify,
			Global:  globalFlag,
		}, fmt.Errorf("missing required message argument")
	}

	// Check for explicit commands
	if processedArgs[0] == "init" {
		return &Args{Command: CommandInit, Global: globalFlag}, nil
	}

	if processedArgs[0] == "config" {
		result, err := parseConfigArgs(processedArgs[1:])
		if err == nil && result != nil {
			// Merge global flag from initial parsing
			result.Global = result.Global || globalFlag
		}
		return result, err
	}

	// Default is notification command
	result, err := parseNotifyArgs(processedArgs)
	if err == nil && result != nil {
		// Merge global flag from initial parsing
		result.Global = result.Global || globalFlag
	}
	return result, err
}

// parseNotifyArgs parses arguments for the notify command
func parseNotifyArgs(args []string) (*Args, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("missing required message argument")
	}

	result := &Args{
		Command: CommandNotify,
		Source:  "Unknown", // Default source
	}

	// First pass: process all flags and find the message
	var messageArgs []string
	var messageFound bool

	for i := range args {
		arg := args[i]

		// Process known flags
		if after, ok := strings.CutPrefix(arg, "--source="); ok {
			result.Source = strings.Trim(after, "'\"")
		} else if after, ok := strings.CutPrefix(arg, "--webhook="); ok {
			result.WebhookURL = strings.Trim(after, "'\"")
		} else if arg == "-g" || arg == "--global" {
			result.Global = true
		} else if strings.HasPrefix(arg, "-") {
			// Unknown flag
			return nil, fmt.Errorf("unknown option for notify command: %s", arg)
		} else {
			// This must be the message
			messageArgs = append(messageArgs, arg)
			messageFound = true
		}
	}

	if !messageFound {
		return nil, fmt.Errorf("missing required message argument")
	}

	// Join all non-flag arguments as the message
	result.Message = strings.Join(messageArgs, " ")

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
		} else if after, ok := strings.CutPrefix(arg, "--webhook="); ok {
			result.WebhookURL = strings.Trim(after, "'\"")
		} else if after, ok := strings.CutPrefix(arg, "--username="); ok {
			result.Username = strings.Trim(after, "'\"")
		} else if after, ok := strings.CutPrefix(arg, "--avatar="); ok {
			result.AvatarURL = strings.Trim(after, "'\"")
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
	fmt.Println("  owata <message> [--webhook=<url>] [--source=<source>] [-g|--global]")
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
	fmt.Println("  config --username=<name>      Set bot username in local config")
	fmt.Println("  config -g --username=<name>   Set bot username in global config")
	fmt.Println("  config --avatar=<url>      Set avatar URL in local config")
	fmt.Println("  config -g --avatar=<url>   Set avatar URL in global config")
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
	fmt.Println("  owata 'Task completed!' -g # Send notification using global config")
}

// PrintVersion prints version information
func PrintVersion() {
	fmt.Printf("Owata v%s\n", Version)
}
