package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yashikota/owata/cli"
	"github.com/yashikota/owata/config"
	"github.com/yashikota/owata/discord"
)

func main() {
	// Parse command-line arguments
	args, err := cli.Parse(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		cli.PrintUsage()
		os.Exit(1)
	}

	// Create a new config manager
	configManager := config.NewManager()

	// Handle the appropriate command
	switch args.Command {
	case cli.CommandShowHelp:
		cli.PrintUsage()

	case cli.CommandShowVersion:
		cli.PrintVersion()

	case cli.CommandInit:
		if err := handleInit(configManager, args.Global); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case cli.CommandConfig:
		if err := handleConfig(configManager, args); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case cli.CommandNotify:
		if err := handleNotify(configManager, args); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}
}

// handleInit handles the init command
func handleInit(cm *config.Manager, global bool) error {
	path, err := cm.CreateTemplate(global)
	if err != nil {
		return err
	}

	if path != "" {
		fmt.Printf("✅ Configuration template created: %s\n", path)
		fmt.Println("\nPlease edit the configuration file and set the following values:")
		fmt.Println("  webhook_url: Your Discord webhook URL")
		fmt.Println("  username:    Bot display name (optional)")
		fmt.Println("  avatar_url:  Bot avatar image URL (optional)")
		fmt.Println("\nOr use the config command with parameters:")
		fmt.Println("  owata config --webhook='https://discord.com/api/webhooks/...'")
		fmt.Println("  owata config --username='MyBot' --avatar='https://example.com/avatar.png'")
	} else {
		// Config file already exists, display it
		if err := cm.DisplayConfig(cm.GetPath(global)); err != nil {
			return err
		}
	}

	return nil
}

// handleConfig handles the config command
func handleConfig(cm *config.Manager, args *cli.Args) error {
	// If no parameters were provided, show current configuration
	if args.WebhookURL == "" && args.Username == "" && args.AvatarURL == "" {
		configPath := cm.GetPath(args.Global)

		// Check if the config file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			globalFlag := ""
			if args.Global {
				globalFlag = " -g"
			}
			fmt.Printf("❌ No configuration found at %s. Run 'owata init%s' to create a config file.\n",
				configPath, globalFlag)
			return nil
		}

		// Display config if it exists
		return cm.DisplayConfig(configPath)
	}

	// Load existing config or create new one
	configPath := cm.GetPath(args.Global)
	cfg, err := cm.LoadFromPath(configPath)
	if err != nil {
		// If config doesn't exist, create new one
		cfg = &config.Config{}

		// For global config, ensure directory exists
		if args.Global {
			dirPath := filepath.Dir(configPath)
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				return fmt.Errorf("failed to create config directory: %v", err)
			}
		}
	}

	// Update config with provided values
	if args.WebhookURL != "" {
		cfg.WebhookURL = args.WebhookURL
	}
	if args.Username != "" {
		cfg.Username = args.Username
	}
	if args.AvatarURL != "" {
		cfg.AvatarURL = args.AvatarURL
	}

	// Save config
	path, err := cm.Save(cfg, args.Global)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Configuration updated in %s\n", path)

	// Display updated config
	return cm.DisplayConfig(path)
}

// handleNotify handles sending a notification
func handleNotify(cm *config.Manager, args *cli.Args) error {
	webhookURL := args.WebhookURL

	// If webhook URL is not provided in args, try to load from config
	if webhookURL == "" {
		cfg, _, err := cm.Load(false) // Prefer local config
		if err != nil || cfg.WebhookURL == "" {
			return fmt.Errorf("no webhook URL provided. Use command line argument or config file")
		}
		webhookURL = cfg.WebhookURL
	}

	// Load config for other settings (username, avatar)
	cfg, _, _ := cm.Load(false) // Ignore error as we'll use defaults if needed

	return discord.SendNotification(webhookURL, args.Message, args.Source, cfg)
}
