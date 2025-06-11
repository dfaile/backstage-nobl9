#Nobl9 Bot

A chat bot for managing Nobl9 projects and user roles through the Nolb9 SDK.

## Features

- **Interactive Project Creation**: Step-by-step guided project creation with confirmation
- **Role Assignment**: Assign roles to users in Nobl9 projects  
- **Natural Language Processing**: Understands phrases like "create project" and "assign role"
- **Configuration Management**: Supports both config files and command-line arguments
- **Official Nobl9 SDK**: Uses the official [Nobl9 Go SDK](https://github.com/nobl9/nobl9-go) for reliable API integration

## Quick Start

### 1. Configure Credentials

The bot uses the official Nobl9 SDK configuration system. You can configure it in several ways:

#### Option A: Configuration File (Recommended)
Create the file ```~/.nobl9/config.toml```

```defaultContext = "default"

[contexts]
  [contexts.default]
    clientId = "N9_CLIENT_ID"
    clientSecret = "N9_CLIENT_SECRET"
    project = "default"
```


#### Option B: Environment Variables

Set the following environment variables:
```bash
export NOBL9_SDK_CLIENT_ID="your-client-id"
export NOBL9_SDK_CLIENT_SECRET="your-client-secret"  
export NOBL9_SDK_ORGANIZATION="your-organization"
export NOBL9_SDK_URL="https://app.nobl9.com"
```

#### Option C: Command Line Arguments

```bash
./bin/nobl9-bot \
  --client-id "your-client-id" \
  --client-secret "your-client-secret" \
  --organization "your-organization" \
  --url "https://app.nobl9.com"
```

### 2. Build and Run

```bash
# Build the bot
make build

# Run with config file (will create default if doesn't exist)
./bin/nobl9-bot

# Or run with command line arguments
./bin/nobl9-bot --client-id YOUR_CLIENT_ID --client-secret YOUR_CLIENT_SECRET --organization YOUR_ORG
```

### 3. Interact with the Bot

Once running, you can interact with the bot using natural language or commands:

```
👋 Hello! I'm your Nobl9 Project Bot. I can help you:

🏗️  **Create new projects** - Just say "create project" or "new project"
👥 **Assign user roles** - Say "assign role" to manage user permissions
📋 **List projects** - Say "list projects" to see available projects

What would you like to do?

> create project
🤖 What's the name of your new project?

> my-awesome-service
🤖 Please provide a description for project 'my-awesome-service':

> This is my awesome microservice
🤖 Create project 'my-awesome-service' with description 'This is my awesome microservice'?

(Y/n) Y
✅ Project 'my-awesome-service' created successfully!
```

## Configuration

The bot follows the [Nobl9 SDK configuration precedence](https://github.com/nobl9/nobl9-go#reading-configuration):

1. **Command line arguments** (highest priority)
2. **Environment variables** 
3. **Configuration file**
4. **Default values** (lowest priority)

### Configuration File Locations

The bot will look for configuration in these locations:
- `~/.nobl9/config.toml` (standard Nobl9 SDK location)
- Path specified with `--config` flag
- Current directory `config.json` (legacy format)

## Development

### Prerequisites

- Go 1.21 or higher
- Access to a Nobl9 organization with API credentials

### Building

```bash
# Install dependencies
go mod tidy

# Build
make build

# Run tests
make test

# Run linters
make lint
```

### Testing

```bash
# Run unit tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/bot
```

## Authentication & Authorization

The bot uses OAuth2 Client Credentials flow through the official Nobl9 SDK. You'll need:

1. **Client ID**: Your Nobl9 API client identifier
2. **Client Secret**: Your Nobl9 API client secret  
3. **Organization**: Your Nobl9 organization name

### Getting Nobl9 Credentials

1. Log into your Nobl9 account
2. Navigate to Settings → Access Keys
3. Create a new access key
4. Use the generated Client ID and Secret

## Commands

The bot supports both natural language and explicit commands:

### Available Commands

- **create-project** `<name>` - Create a new Nobl9 project
- **assign-role** `<project>` `<user-email>` - Assign roles to users  
- **list-projects** - List available projects
- **help** - Show help message

### Natural Language Examples

- "I want to create a new project"
- "Create a project called analytics-service"  
- "Help me assign roles to users"
- "Show me the available projects"

## Error Handling

The bot includes comprehensive error handling:

- **Authentication errors**: Clear messages about credential issues
- **API errors**: Detailed error information from Nobl9 API
- **Input validation**: Helpful guidance for invalid input
- **Network errors**: Retry logic for transient failures

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Projects

- [Nobl9 Go SDK](https://github.com/nobl9/nobl9-go) - Official Nobl9 SDK for Go
- [Backstage](https://backstage.io/) - Platform for building developer portals
- [Nobl9 Documentation](https://docs.nobl9.com/) - Official Nobl9 documentation
