# Backstage Nobl9 Bot

A Slack bot for managing Nobl9 projects and user roles through Backstage.

## Prerequisites

- Go 1.21 or later
- Nobl9 account with OAuth2 credentials (Client ID and Client Secret)
- Slack workspace with bot token

## Configuration

The bot can be configured using a config file or command line flags. The config file is located at `~/.nobl9/config.json` by default.

### Config File Format

```json
{
  "client_id": "your-client-id",
  "client_secret": "your-client-secret",
  "organization": "your-org",
  "base_url": "https://app.nobl9.com"
}
```

### Command Line Flags

```bash
--client-id string     Nobl9 client ID
--client-secret string Nobl9 client secret
--org string          Nobl9 organization
--base-url string     Nobl9 base URL (default: https://app.nobl9.com)
--config string       Path to config file (default: ~/.nobl9/config.json)
```

## Building

```bash
make build
```

## Running

```bash
./bin/nobl9-bot --client-id YOUR_CLIENT_ID --client-secret YOUR_CLIENT_SECRET --org YOUR_ORG
```

## Docker

Build the Docker image:

```bash
make docker-build
```

Run the container:

```bash
docker run -v ~/.nobl9:/root/.nobl9 backstage-nobl9-bot
```

## Development

1. Clone the repository
2. Install dependencies: `go mod download`
3. Run tests: `make test`
4. Build: `make build`

## License

MIT
