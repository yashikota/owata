package main

import (
	"fmt"
	"os"

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

func handleInit(cm *config.Manager, global bool) error {
	path, created, err := cm.CreateTemplate(global)
	if err != nil {
		return err
	}

	if created {
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
		fmt.Printf("ℹ️ Config file already exists: %s\n", path)
		output, err := cm.DisplayConfig(path)
		if err != nil {
			return err
		}
		fmt.Print(output)
	}

	return nil
}

func handleConfig(cm *config.Manager, args *cli.Args) error {
	// If no parameters were provided, show current configuration
	if args.WebhookURL == "" && args.Username == "" && args.AvatarURL == "" {
		configPath, err := cm.GetPathWithError(args.Global)
		if err != nil {
			return fmt.Errorf("failed to get config path: %v", err)
		}

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
		output, err := cm.DisplayConfig(configPath)
		if err != nil {
			return err
		}
		fmt.Print(output)
		return nil
	}

	// Load existing config or create new one
	configPath, pathErr := cm.GetPathWithError(args.Global)
	if pathErr != nil {
		return fmt.Errorf("failed to get config path: %v", pathErr)
	}
	cfg, err := cm.LoadFromPath(configPath)
	if err != nil {
		// If config doesn't exist, create new one
		cfg = &config.Config{}
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
	output, err := cm.DisplayConfig(path)
	if err != nil {
		return err
	}
	fmt.Print(output)
	return nil
}

func handleNotify(cm *config.Manager, args *cli.Args) error {
	var webhookURL string
	var configToUse *config.Config
	preferGlobal := args.Global

	cfg, _, err := cm.Load(preferGlobal)
	if err == nil {
		configToUse = cfg
		if configToUse.WebhookURL != "" && args.WebhookURL == "" {
			webhookURL = configToUse.WebhookURL
		}
	}

	if args.WebhookURL != "" {
		webhookURL = args.WebhookURL
	}

	if webhookURL == "" {
		configType := "local"
		if args.Global {
			configType = "global"
		}
		return fmt.Errorf("no webhook URL provided in command line or %s config", configType)
	}

	sendErr := discord.SendNotification(webhookURL, args.Message, args.Source, configToUse)
	if sendErr != nil {
		return sendErr
	}

	fmt.Println("✅ Discord notification sent successfully")
	return nil
}
