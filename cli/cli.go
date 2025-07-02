package cli

import (
	"fmt"
	"strings"
)

const Version = "2.0.0"

type CommandType int

const (
	CommandNotify CommandType = iota
	CommandInit
	CommandConfig
	CommandShowHelp
	CommandShowVersion
)

type Args struct {
	Command    CommandType
	Message    string
	WebhookURL string
	Source     string
	Username   string
	AvatarURL  string
	Global     bool
}

func Parse(args []string) (*Args, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("missing arguments")
	}

	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return &Args{Command: CommandShowHelp}, nil
		}
		if arg == "--version" || arg == "-v" {
			return &Args{Command: CommandShowVersion}, nil
		}
	}

	var globalFlag bool
	var processedArgs []string

	for i := range args {
		if args[i] == "-g" || args[i] == "--global" {
			globalFlag = true
		} else {
			processedArgs = append(processedArgs, args[i])
		}
	}

	if len(processedArgs) == 0 {
		return nil, fmt.Errorf("missing command; please specify 'init', 'config', or a notification message")
	}

	if processedArgs[0] == "init" {
		return &Args{Command: CommandInit, Global: globalFlag}, nil
	}

	if processedArgs[0] == "config" {
		result, err := parseConfigArgs(processedArgs[1:])
		if err == nil && result != nil {
			// Merge global flag from initial parsing
			result.Global = globalFlag
		}
		return result, err
	}

	result, err := parseNotifyArgs(processedArgs)
	if err == nil && result != nil {
		// Merge global flag from initial parsing
		result.Global = globalFlag
	}
	return result, err
}

func parseNotifyArgs(args []string) (*Args, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("missing required message argument")
	}

	result := &Args{
		Command: CommandNotify,
		Source:  "Unknown", // Default source
	}

	var messageArgs []string
	var messageFound bool

	for i := range args {
		arg := args[i]

		if after, ok := strings.CutPrefix(arg, "--source="); ok {
			result.Source = strings.Trim(after, "'\"")
		} else if after, ok := strings.CutPrefix(arg, "--webhook="); ok {
			result.WebhookURL = strings.Trim(after, "'\"")
		} else if strings.HasPrefix(arg, "-") {
			// Unknown flag
			return nil, fmt.Errorf("unknown option for notify command: %s", arg)
		} else {
			messageArgs = append(messageArgs, arg)
			messageFound = true
		}
	}

	if !messageFound {
		return nil, fmt.Errorf("missing required message argument")
	}

	result.Message = strings.Join(messageArgs, " ")

	return result, nil
}

func parseConfigArgs(args []string) (*Args, error) {
	result := &Args{
		Command: CommandConfig,
	}

	if len(args) == 0 {
		return result, nil
	}

	for i := range args {
		arg := args[i]

		if after, ok := strings.CutPrefix(arg, "--webhook="); ok {
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

func PrintUsage() {
	fmt.Printf("Owata v%s - Discord Webhook Notifier\n\n", Version)
	fmt.Println("Usage:")
	fmt.Println("  owata <message> [--webhook=<url>] [--source=<source>] [-g|--global]")
	fmt.Println("  owata init [-g|--global]")
	fmt.Println("  owata config [-g|--global] [--webhook=<url>] [--username=<name>] [--avatar=<url>]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Printf("  %-30s Create local configuration template file\n", "init")
	fmt.Printf("  %-30s Create global configuration template file\n", "init -g, --global")
	fmt.Printf("  %-30s Show current local configuration\n", "config")
	fmt.Printf("  %-30s Show current global configuration\n", "config -g, --global")
	fmt.Printf("  %-30s Set Discord webhook URL in local config\n", "config --webhook=<url>")
	fmt.Printf("  %-30s Set Discord webhook URL in global config\n", "config -g --webhook=<url>")
	fmt.Printf("  %-30s Set bot username in local config\n", "config --username=<name>")
	fmt.Printf("  %-30s Set bot username in global config\n", "config -g --username=<name>")
	fmt.Printf("  %-30s Set avatar URL in local config\n", "config --avatar=<url>")
	fmt.Printf("  %-30s Set avatar URL in global config\n", "config -g --avatar=<url>")
	fmt.Println("")
	fmt.Println("Arguments:")
	fmt.Println("  message                    The notification message to send")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --webhook=<url>            Discord webhook URL (overrides config)")
	fmt.Println("  --source=<source>          Set the source of the notification")
	fmt.Println("  -g, --global               Use global configuration (in system config directory)")
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

func PrintVersion() {
	fmt.Printf("Owata v%s\n", Version)
}
