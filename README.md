# GoREST AI Plugin

Production-ready AI API abstraction layer plugin for [GoREST](https://github.com/nicolasbonnici/gorest) with unified support for multiple AI providers.

## Features

### 🤖 Multiple AI Providers
- **Anthropic (Claude)** - Claude 3.5 Sonnet, Claude 3 Opus, Claude 3 Haiku
- **OpenAI (ChatGPT)** - GPT-4, GPT-3.5 Turbo
- **Google (Gemini)** - Gemini Pro, Gemini Ultra
- **Mistral AI** - Mistral Large, Mistral Medium, Mistral Tiny

### 🚀 Production Features
- ✅ **Provider Abstraction** - Unified interface for all AI providers
- ✅ **Intelligent Caching** - SHA256-based caching with configurable TTL
- ✅ **Automatic Fallback** - Priority-based failover between providers
- ✅ **Quota Management** - Per-user request and token limits
- ✅ **Cost Tracking** - Token-based cost calculation and analytics
- ✅ **Rate Limiting** - Token bucket algorithm with configurable limits
- ✅ **Streaming Support** - Server-Sent Events (SSE) for real-time responses
- ✅ **Audit Logging** - Complete request/response tracking
- ✅ **Multi-Database** - PostgreSQL, MySQL, SQLite support
- ✅ **OpenAPI Integration** - Automatic API documentation

## Installation

```bash
go get github.com/nicolasbonnici/gorest-ai
```

## Quick Start

### 1. Register the Plugin

```go
package main

import (
    "github.com/nicolasbonnici/gorest"
    "github.com/nicolasbonnici/gorest/pluginloader"
    ai "github.com/nicolasbonnici/gorest-ai"
)

func init() {
    pluginloader.RegisterPluginFactory("ai", ai.NewPlugin)
}

func main() {
    cfg := gorest.Config{
        ConfigPath: ".",
    }
    gorest.Start(cfg)
}
```

### 2. Configure in `gorest.yaml`

```yaml
database:
  url: "${DATABASE_URL}"

plugins:
  - name: ai
    enabled: true
    config:
      default_provider: "anthropic"
      enabled_providers:
        - anthropic
        - openai
        - gemini
        - mistral

      # API Keys
      anthropic_api_key: "${ANTHROPIC_API_KEY}"
      openai_api_key: "${OPENAI_API_KEY}"
      gemini_api_key: "${GEMINI_API_KEY}"
      mistral_api_key: "${MISTRAL_API_KEY}"

      # Features
      enable_cache: true
      cache_ttl: 3600
      enable_fallback: true
      enable_quota: true

      # Limits
      max_tokens: 4096
      default_temperature: 0.7
      rate_limit_per_min: 60
      request_timeout: 30
```

### 3. Use the API

```bash
curl -X POST http://localhost:8000/api/ai/chat \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "anthropic",
    "messages": [
      {"role": "user", "content": "Explain quantum computing"}
    ],
    "use_cache": true
  }'
```

## API Endpoints

### Chat Endpoints

#### POST `/api/ai/chat`
Send a chat completion request.

**Request:**
```json
{
  "provider": "anthropic",
  "model": "claude-3-5-sonnet-20241022",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant"},
    {"role": "user", "content": "Hello, how are you?"}
  ],
  "temperature": 0.7,
  "max_tokens": 1000,
  "use_cache": true
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "provider": "anthropic",
  "model": "claude-3-5-sonnet-20241022",
  "content": "Hello! I'm doing well, thank you...",
  "prompt_tokens": 15,
  "completion_tokens": 42,
  "total_tokens": 57,
  "cost": 0.000171,
  "duration_ms": 1234,
  "cached": false,
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### POST `/api/ai/chat/stream`
Send a streaming chat request (SSE).

### Provider Management (Admin)

#### POST `/api/ai/providers`
Create a provider configuration.

#### GET `/api/ai/providers/:id`
Get provider by ID.

#### GET `/api/ai/providers`
List all providers with pagination.

#### PUT `/api/ai/providers/:id`
Update provider configuration.

#### DELETE `/api/ai/providers/:id`
Delete provider.

### Usage & Statistics

#### GET `/api/ai/usage`
Get usage statistics.

**Response:**
```json
{
  "total_requests": 1000,
  "successful_requests": 980,
  "failed_requests": 20,
  "cached_requests": 100,
  "total_tokens": 50000,
  "total_cost": 150.50,
  "average_duration_ms": 1500,
  "provider_breakdown": {
    "anthropic": {
      "requests": 500,
      "tokens": 25000,
      "cost": 75.00
    }
  }
}
```

#### GET `/api/ai/usage/quota`
Get user quota status.

### Request History

#### GET `/api/ai/requests`
List request history with filtering.

#### GET `/api/ai/requests/:id`
Get request details.

## Configuration

### Provider Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `default_provider` | string | `"anthropic"` | Default AI provider |
| `enabled_providers` | []string | `["anthropic", "openai", "gemini", "mistral"]` | List of enabled providers |
| `anthropic_api_key` | string | - | Anthropic API key |
| `openai_api_key` | string | - | OpenAI API key |
| `gemini_api_key` | string | - | Google Gemini API key |
| `mistral_api_key` | string | - | Mistral AI API key |

### Feature Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enable_cache` | bool | `true` | Enable response caching |
| `cache_ttl` | int | `3600` | Cache TTL in seconds |
| `enable_fallback` | bool | `true` | Enable provider fallback |
| `enable_quota` | bool | `true` | Enable quota management |

### Limit Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `max_tokens` | int | `4096` | Maximum tokens per request |
| `default_temperature` | float64 | `0.7` | Default temperature (0-2) |
| `rate_limit_per_min` | int | `60` | Requests per minute |
| `request_timeout` | int | `30` | Request timeout in seconds |

### Security Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `require_auth` | bool | `true` | Require authentication |
| `allow_anonymous` | bool | `false` | Allow anonymous requests |

### Audit Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enable_audit` | bool | `true` | Enable audit logging |
| `retain_audit_days` | int | `90` | Audit retention period |

## Advanced Usage

### Custom Provider Configuration

Create a custom provider configuration via API:

```bash
curl -X POST http://localhost:8000/api/ai/providers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "openai",
    "display_name": "OpenAI Custom",
    "api_key": "sk-...",
    "enabled": true,
    "priority": 1,
    "max_tokens": 4096,
    "temperature": 0.8,
    "rate_limit": 100
  }'
```

### Provider Fallback

If `enable_fallback` is true and the primary provider fails, requests automatically failover to the next available provider based on priority.

### Caching

Caching uses SHA256 hashing of the request parameters:
- Provider
- Model
- Messages
- Temperature
- Max tokens

Identical requests return cached responses instantly.

### Quota Management

Configure per-user quotas:
- Daily request limit
- Monthly request limit
- Daily token limit
- Monthly token limit

Quotas automatically reset at configured intervals.

### Cost Tracking

Token-based cost calculation with per-provider pricing:
- Anthropic: ~$3 per million tokens
- OpenAI: ~$30 per million tokens (GPT-4)
- Gemini: ~$1 per million tokens
- Mistral: ~$2 per million tokens

## Database Schema

The plugin creates four tables:

### `ai_providers`
Stores AI provider configurations with encrypted API keys.

### `ai_requests`
Audit trail of all AI requests with token usage and costs.

### `ai_cache`
Cache responses with TTL and hit count tracking.

### `ai_quotas`
User-level quota tracking with daily/monthly limits.

## Architecture

### Provider Interface

All providers implement a common interface:

```go
type Provider interface {
    Name() string
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, <-chan error)
    CountTokens(ctx context.Context, messages []Message) (int, error)
    ValidateConfig() error
    HealthCheck(ctx context.Context) error
}
```

### Service Layer

The service layer handles:
- Cache checking and storage
- Provider selection and fallback
- Quota enforcement
- Cost calculation
- Audit logging

### Middleware

Two middleware components:
- **QuotaMiddleware** - Enforces user quotas
- **AuditMiddleware** - Logs all requests and responses

## Production Deployment

### 1. Enable Authentication

```yaml
plugins:
  - name: auth
    enabled: true

  - name: ai
    config:
      require_auth: true
      allow_anonymous: false
```

### 2. Configure Quotas

```yaml
plugins:
  - name: ai
    config:
      enable_quota: true
```

### 3. Use Environment Variables

```yaml
anthropic_api_key: "${ANTHROPIC_API_KEY}"
openai_api_key: "${OPENAI_API_KEY}"
```

### 4. Enable Audit Logging

```yaml
plugins:
  - name: ai
    config:
      enable_audit: true
      retain_audit_days: 90
```

### 5. Use Production Database

```yaml
database:
  url: "${DATABASE_URL}"
```

## Testing

### Unit Tests

```bash
go test ./...
```

### Integration Tests

```bash
go test -tags=integration ./...
```

### Load Testing

```bash
# Using hey
hey -n 1000 -c 10 -m POST -H "Content-Type: application/json" \
  -d '{"provider":"anthropic","messages":[{"role":"user","content":"Hi"}]}' \
  http://localhost:8000/api/ai/chat
```

## Examples

See `examples/basic/` for a complete working example.

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Write tests for new features
4. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

- **Issues**: https://github.com/nicolasbonnici/gorest-ai/issues
- **Documentation**: https://github.com/nicolasbonnici/gorest-ai
- **GoREST**: https://github.com/nicolasbonnici/gorest

## Roadmap

- [x] Core plugin implementation
- [x] Multiple provider support
- [x] Caching system
- [x] Quota management
- [x] Fallback mechanism
- [x] Cost tracking
- [ ] Streaming support (in progress)
- [ ] Advanced quota rules
- [ ] Provider health monitoring
- [ ] Webhook support
- [ ] Function calling support
- [ ] Vision API support
- [ ] Embedding API support

## Related Projects

- [gorest](https://github.com/nicolasbonnici/gorest) - The GoREST framework
- [gorest-auth](https://github.com/nicolasbonnici/gorest-auth) - Authentication plugin
- [gorest-mcp](https://github.com/nicolasbonnici/gorest-mcp) - MCP connector plugin

---

Built using [GoREST](https://github.com/nicolasbonnici/gorest)
