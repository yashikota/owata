# Owata - Discord Notification Tool

🔔 A cross-platform Go tool for sending Discord notifications. Perfect for LLMs like Claude Code or Gemini CLI to send completion notifications.  

## Features

- 🖥️ **Cross-platform**: Works on Windows, macOS, and Linux
- 📨 **Discord webhooks**: Sends rich embed notifications
- ⚙️ **Configurable**: JSON config file or command-line arguments
- 🚀 **Zero dependencies**: Uses only Go standard library
- 🤖 **LLM-friendly**: Simple command-line interface for automated notifications

## Installation

### Build from source

```bash
git clone https://github.com/yashikota/owata
cd owata
go build -o owata
```

### Download binary

Download the latest release from the [releases page](https://github.com/yashikota/owata/releases).

## Usage

### Basic usage

```bash
# Send notification with webhook URL
owata "Claude Code session completed" https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN

# Send notification about task completion
owata "Task finished successfully" https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN

# Send notification with source specification
owata "Task completed successfully" --source="Claude Code"
```

### Usage in AI/LLM tools

AI tools and LLM agents can use the CLI directly by executing the command:

```bash
# From any programming language
exec("owata 'AI task completed' --source='Claude Code'");
```

### Using config file

1. Copy the example config file:
   ```bash
   cp owata-config.json.example owata-config.json
   ```

2. Edit `owata-config.json` with your settings:
   ```json
   {
     "webhook_url": "https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN",
     "username": "Owata",
     "avatar_url": "https://example.com/avatar.png"
   }
   ```

3. Send notifications:
   ```bash
   owata "Your message here"
   ```

## Configuration

### Config file options

- `webhook_url`: Discord webhook URL (required)
- `username`: Custom username for the bot (optional, default: "Owata Monitor")
- `avatar_url`: Custom avatar URL for the bot (optional)

### Command line

```bash
owata <message> [webhook-url] [--source=<source>]
```

- `message`: The message to send (required)
- `webhook-url`: Discord webhook URL (optional if using config file)
- `--source`: Specify the source of the notification (e.g., "Claude Code", "Gemini", etc.)

## Discord Webhook Setup

1. Go to your Discord server settings
2. Navigate to Integrations → Webhooks
3. Click "New Webhook"
4. Choose a channel and copy the webhook URL
5. Use this URL in your config file or command line

## Examples

```bash
owata "Tasks Done" --source="Claude Code"
```

```bash
owata "Tasks Done" --source="Gemini CLI"
```

```bash
owata "Custom Message" --source="任意のソース"
```

## Notification Format

Owata sends a Discord embed with:
- Message text
- Working directory
- Source (if specified)
- Timestamp
