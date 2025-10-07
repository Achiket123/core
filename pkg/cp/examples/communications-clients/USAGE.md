# Usage Guide

## Quick Start

The example includes a built-in mock server, so it works out of the box:

```bash
go run main.go clients.go mockserver.go
```

## Using Real Services

### Slack

To use a real Slack workspace:

1. Create a Slack App at https://api.slack.com/apps
2. Add the following Bot Token Scopes:
   - `chat:write`
   - `channels:read`
   - `im:write`
3. Install the app to your workspace
4. Set environment variable:

```bash
export SLACK_TOKEN="xoxb-your-bot-token"
```

5. Update tenant config in `main.go`:

```go
{
    ID:                  "tenant-1",
    PreferredPlatform:   "slack",
    SlackToken:          os.Getenv("SLACK_TOKEN"),
    SlackTeamID:         "your-team-id",
    SlackBaseURL:        "", // Empty uses real Slack API
    NotificationChannel: "general",
}
```

### Microsoft Teams

To use Microsoft Teams:

1. Create an Incoming Webhook in your Teams channel:
   - Go to your Teams channel
   - Click "..." → Connectors → Incoming Webhook
   - Copy the webhook URL

2. Set environment variable:

```bash
export TEAMS_WEBHOOK_URL="https://outlook.office.com/webhook/..."
```

3. Update tenant config:

```go
{
    ID:                "tenant-2",
    PreferredPlatform: "teams",
    TeamsWebhookURL:   os.Getenv("TEAMS_WEBHOOK_URL"),
    NotificationChannel: "General",
}
```

### Discord

To use Discord:

1. Create a webhook in your Discord server:
   - Go to Server Settings → Integrations → Webhooks
   - Create webhook and copy URL

2. Set environment variable:

```bash
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
```

3. Update tenant config:

```go
{
    ID:                "tenant-3",
    PreferredPlatform: "discord",
    DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
    NotificationChannel: "general",
}
```

## Running with Real Services

```bash
# Set all webhooks
export SLACK_TOKEN="xoxb-..."
export TEAMS_WEBHOOK_URL="https://..."
export DISCORD_WEBHOOK_URL="https://..."

# Run the example
go run main.go clients.go mockserver.go
```

The example will automatically use real services if environment variables are set, otherwise it falls back to the mock server.

## Testing Different Scenarios

### Test Priority Routing

Create different message contexts with priority levels to see routing in action.

### Test Cache Efficiency

The example demonstrates cache hits by sending multiple messages for the same tenant.

### Test Concurrent Operations

The client pool is thread-safe and supports concurrent message sending across multiple tenants.

## Mock Server Details

The mock server runs on `localhost:8080` and provides:

- **Slack API**: `/slack/chat.postMessage`, `/slack/conversations.open`, `/slack/conversations.list`
- **Teams Webhook**: `/teams/webhook`
- **Discord Webhook**: `/discord/webhook`
- **Health Check**: `/health`

All endpoints return appropriate success responses and log received messages to console.

## Building

```bash
# Build executable
go build -o communications-example main.go clients.go mockserver.go

# Run executable
./communications-example
```

## Expected Output

```
=== Communications Client Pooling Example ===

Mock communication server starting on :8080
Demonstrating multi-tenant communications routing:

--- Tenant: tenant-1 (Platform: slack) ---
[Mock Slack] Received message to channel general: Hello from multi-tenant system!
[Slack] ✓ Sent message to general: Hello from multi-tenant system!
...

=== Example completed successfully ===
```

## Troubleshooting

**Port 8080 already in use**:
```bash
# Kill process on port 8080
lsof -ti:8080 | xargs kill -9

# Or change the port in main.go:
mockServer := NewMockServer("8081") // Use port 8081 instead
```

**Connection refused**:
- Check that mock server started successfully
- Wait a moment after starting for server to initialize

**Real API errors**:
- Verify your tokens/webhooks are correct
- Check that your Slack app has the required scopes
- Ensure webhooks are enabled in Teams/Discord channels
