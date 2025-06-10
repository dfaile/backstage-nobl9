# Nobl9 Project Bot Technical Documentation

## Architecture

### Components

1. Bot Core (`internal/bot`)
   - Handles conversation state
   - Processes user messages
   - Manages command execution

2. Nobl9 Client (`internal/nobl9`)
   - Implements Nobl9 API integration
   - Handles rate limiting
   - Manages project and role operations

3. Command System (`internal/command`)
   - Parses and validates commands
   - Executes command handlers
   - Formats responses

4. Formatting (`internal/format`)
   - Formats messages and responses
   - Handles error formatting
   - Manages interactive prompts

5. Error Handling (`internal/errors`)
   - Defines error types
   - Implements error wrapping
   - Provides error utilities

6. Logging (`internal/logging`)
   - Structured logging
   - Context propagation
   - Log level management

### Data Flow

1. Message Reception
   ```
   User -> Slack -> Bot -> Message Handler
   ```

2. Command Processing
   ```
   Message Handler -> Command Parser -> Command Executor -> Response Formatter
   ```

3. API Interaction
   ```
   Command Executor -> Nobl9 Client -> Nobl9 API
   ```

## Implementation Details

### State Management

```go
type ConversationState struct {
    LastUpdated      time.Time
    CurrentStep      string
    PendingPrompt    *interactive.Prompt
    ProjectName      string
    ProjectDescription string
    RoleUser         string
    RoleType         string
}
```

### Rate Limiting

```go
type RateLimiter interface {
    Wait(ctx context.Context) error
    Success()
    Failure()
}
```

### Error Types

```go
type ErrorType string

const (
    ErrorTypeNotFound    ErrorType = "NOT_FOUND"
    ErrorTypeConflict    ErrorType = "CONFLICT"
    ErrorTypeValidation  ErrorType = "VALIDATION"
    ErrorTypeRateLimit   ErrorType = "RATE_LIMIT"
    ErrorTypeInternal    ErrorType = "INTERNAL"
)
```

## API Integration

### Nobl9 API Client

```go
type Client interface {
    GetProject(ctx context.Context, name string) (*Project, error)
    CreateProject(ctx context.Context, name, description string) (*Project, error)
    ValidateProjectName(ctx context.Context, name string) (bool, error)
    ValidateUser(ctx context.Context, email string) (bool, error)
    AssignRoles(ctx context.Context, project string, assignments map[string][]string) error
}
```

### Rate Limiting

- Exponential backoff
- Maximum retry attempts
- Rate limit detection

## Testing

### Unit Tests

- Mock-based testing
- Error scenario coverage
- State management verification

### Integration Tests

- Real API integration
- End-to-end flows
- Error recovery testing

## Configuration

### Environment Variables

```bash
NOBL9_API_KEY=your-api-key
NOBL9_ORG=your-org
NOBL9_BASE_URL=https://api.nobl9.com
```

### Rate Limiting Configuration

```go
type RateLimitConfig struct {
    MaxRetries     int
    InitialBackoff time.Duration
    MaxBackoff     time.Duration
}
```

## Deployment

### Requirements

- Go 1.21 or later
- Slack workspace
- Nobl9 API access

### Build

```bash
go build -o nobl9-bot cmd/bot/main.go
```

### Run

```bash
./nobl9-bot
```

## Monitoring

### Logging

- Structured JSON logs
- Context propagation
- Log levels (debug, info, warn, error)

### Metrics

- Command execution time
- API call latency
- Error rates
- Rate limit hits

## Security

### API Key Management

- Environment variable storage
- Key rotation support
- Access logging

### Access Control

- Role-based access
- Project-level permissions
- Audit logging

## Development

### Code Style

- Go standard formatting
- Linting with golangci-lint
- Documentation requirements

### Testing

- Unit test coverage > 80%
- Integration test coverage
- Error scenario testing

### Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests
4. Submit a pull request

## Troubleshooting

### Common Issues

1. Rate Limiting
   - Check rate limit configuration
   - Verify retry mechanism
   - Monitor API usage

2. State Management
   - Verify state transitions
   - Check cleanup procedures
   - Monitor memory usage

3. API Integration
   - Verify API credentials
   - Check API version
   - Monitor API responses

### Debugging

1. Enable debug logging
2. Check state transitions
3. Monitor API calls
4. Verify error handling

## Future Improvements

1. Enhanced Monitoring
   - Prometheus metrics
   - Grafana dashboards
   - Alerting

2. Additional Features
   - Bulk operations
   - Project templates
   - Role inheritance

3. Performance
   - Caching
   - Batch processing
   - Connection pooling 