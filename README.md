# alexa-cli

A command-line interface for controlling Amazon Alexa devices using the unofficial Alexa API.

Control your Echo devices, send announcements, execute voice commands, and more - all from the terminal or scripts.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install buddyh/tap/alexa
```

### Go

```bash
go install github.com/buddyh/alexa-cli/cmd/alexa@latest
```

### Build locally

```bash
git clone https://github.com/buddyh/alexa-cli
cd alexa-cli
make build
./bin/alexa --help
```

## Authentication

This CLI uses Amazon's unofficial API (the same one the Alexa app uses). You'll need to obtain a refresh token using the [alexa-cookie-cli](https://github.com/adn77/alexa-cookie-cli) tool.

### Step 1: Get your refresh token

Download alexa-cookie-cli from the [releases page](https://github.com/adn77/alexa-cookie-cli/releases), then:

```bash
# For US Amazon accounts
./alexa-cookie-cli --amazonPage amazon.com --baseAmazonPage amazon.com --amazonPageProxyLanguage en_US --acceptLanguage en-US
```

This opens a local proxy at http://127.0.0.1:8080. Log into your Amazon account there, and the refresh token will be displayed.

### Step 2: Configure alexa-cli

```bash
# Interactive
alexa auth

# Direct
alexa auth <your-refresh-token>

# Or set environment variable
export ALEXA_REFRESH_TOKEN=<your-token>
```

Configuration is stored in `~/.alexa-cli/config.json`.

## Usage

### List Devices

```bash
# List all Echo devices
alexa devices

# JSON output for scripting
alexa devices --json
```

### Text-to-Speech

```bash
# Speak on a specific device
alexa speak "Hello world" -d "Kitchen Echo"

# Announce to ALL devices
alexa speak "Dinner is ready!" --announce

# Device name matching is flexible
alexa speak "Build complete" -d Kitchen
alexa speak "Build complete" -d "Living Room"
```

### Voice Commands

Send any command as if you spoke it to Alexa:

```bash
# Control smart home devices
alexa command "turn off the living room lights" -d Kitchen
alexa command "set thermostat to 72 degrees" -d Bedroom

# Play music
alexa command "play jazz music" -d "Living Room"

# Ask questions
alexa command "what's the weather" -d Kitchen

# Set timers and reminders
alexa command "set a timer for 10 minutes" -d Kitchen
```

### Routines (Coming Soon)

```bash
# List available routines
alexa routine list

# Execute a routine
alexa routine run "Good Night"
```

### Smart Home (Coming Soon)

```bash
# List smart home devices
alexa sh list

# Control devices
alexa sh on "Kitchen Light"
alexa sh off "All Lights"
alexa sh brightness "Bedroom Lamp" 50
```

## JSON Output

All commands support `--json` for machine-readable output:

```bash
alexa devices --json | jq '.[].name'
alexa speak "test" -d Kitchen --json
```

## Command Reference

| Command | Description | Status |
|---------|-------------|--------|
| `alexa devices` | List all Echo devices | Working |
| `alexa speak <text> -d <device>` | Text-to-speech on device | Working |
| `alexa speak <text> --announce` | Announce to all devices | Working |
| `alexa command <text> -d <device>` | Send voice command | Working |
| `alexa auth` | Configure authentication | Working |
| `alexa routine list` | List routines | WIP |
| `alexa routine run <name>` | Execute routine | WIP |
| `alexa sh list` | List smart home devices | WIP |
| `alexa sh on/off <device>` | Control device | WIP |

> **Note:** Routines and Smart Home commands are work-in-progress. For now, use `alexa command` to control smart home devices via natural language (e.g., `alexa command "turn off the lights" -d Kitchen`).

## Use Cases

### Claude Code / AI Agent Integration

This CLI was built specifically to allow AI assistants to control smart home devices:

```bash
# In your AI assistant's prompt or tools
alexa speak "The build finished successfully" -d Office
alexa command "turn off all lights" -d Kitchen
```

### Scripting and Automation

```bash
#!/bin/bash
# Announce when a long-running job finishes
make build && alexa speak "Build complete!" --announce
```

### Home Automation

```bash
# Morning routine script
alexa command "good morning" -d Bedroom
alexa command "turn on kitchen lights" -d Kitchen
alexa command "what's on my calendar today" -d Kitchen
```

## Token Refresh

The refresh token is valid for approximately 14 days. If you get authentication errors, run `alexa auth` again with a fresh token from alexa-cookie-cli.

## Troubleshooting

### "not configured" error

Run `alexa auth <token>` to configure your refresh token.

### Device not found

Use `alexa devices` to see exact device names, then match them in your commands. Partial matching is supported.

### Command not working

Try running the same command with `alexa command` instead - this sends it as a voice command which has broader support.

## How It Works

This CLI uses the same unofficial API that the Alexa mobile app uses. It:

1. Exchanges your refresh token for session cookies
2. Obtains a CSRF token from alexa.amazon.com
3. Sends commands to pitangui.amazon.com (US) or layla.amazon.com (EU)

This approach is used by many popular projects including [alexa-remote-control](https://github.com/thorsten-gehrig/alexa-remote-control) and [Home Assistant's Alexa integration](https://github.com/alandtse/alexa_media_player).

## Disclaimer

This is an unofficial tool that uses Amazon's private APIs. It may break at any time if Amazon changes their API. Use at your own risk.

## Acknowledgments

This project builds on the work of several excellent open source projects:

- **[alexa-cookie-cli](https://github.com/adn77/alexa-cookie-cli)** by adn77 - Token authentication tool (required for setup)
- **[alexa-cookie2](https://github.com/Apollon77/alexa-cookie2)** by Apollon77 - The underlying authentication library
- **[alexa-remote-control](https://github.com/thorsten-gehrig/alexa-remote-control)** by thorsten-gehrig - Bash implementation that documented the API
- **[alexa_media_player](https://github.com/alandtse/alexa_media_player)** by alandtse - Home Assistant integration that proved the approach at scale

Without these projects' reverse-engineering efforts and documentation, this CLI wouldn't exist.

## License

MIT
