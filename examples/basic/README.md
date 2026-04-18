# GoREST AI Plugin - Basic Example

This example demonstrates how to use the GoREST AI plugin to add AI capabilities to your GoREST application.

## Prerequisites

- Go 1.22 or later
- PostgreSQL, MySQL, or SQLite database
- API keys for at least one AI provider:
  - [Anthropic (Claude)](https://console.anthropic.com/)
  - [OpenAI (ChatGPT)](https://platform.openai.com/)
  - [Google (Gemini)](https://makersuite.google.com/app/apikey)
  - [Mistral AI](https://console.mistral.ai/)

## Setup

### 1. Install Dependencies

```bash
go mod download
```

### 2. Configure Environment Variables

Create a `.env` file:

```bash
# Database
export DATABASE_URL="postgres://localhost/gorest_ai?sslmode=disable"

# AI Provider API Keys (at least one required)
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="..."
export MISTRAL_API_KEY="..."
```

Load the environment:

```bash
source .env
```

### 3. Run the Application

```bash
go run main.go
```

The server will start on `http://localhost:8000`

## Usage Examples

### Send a Chat Request

```bash
curl -X POST http://localhost:8000/api/ai/chat \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "anthropic",
    "messages": [
      {"role": "user", "content": "Explain quantum computing in simple terms"}
    ],
    "temperature": 0.7,
    "max_tokens": 1000,
    "use_cache": true
  }'
```

### Send a Chat Request with Auto Provider Selection

```bash
curl -X POST http://localhost:8000/api/ai/chat \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "auto",
    "messages": [
      {"role": "user", "content": "Write a haiku about programming"}
    ]
  }'
```

### List All Providers

```bash
curl http://localhost:8000/api/ai/providers
```

### Create a Custom Provider Configuration

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

### Get Usage Statistics

```bash
curl http://localhost:8000/api/ai/usage
```

### Get User Quota Status

```bash
curl http://localhost:8000/api/ai/usage/quota
```

## Features Demonstrated

### 1. Multiple AI Providers
- Anthropic (Claude)
- OpenAI (ChatGPT)
- Google (Gemini)
- Mistral AI

### 2. Intelligent Caching
- SHA256-based cache keys
- Configurable TTL
- Automatic cache hit tracking

### 3. Fallback Mechanism
- Automatic failover on provider errors
- Priority-based provider ordering

### 4. Quota Management
- Per-user request limits
- Token-based quota tracking
- Automatic quota reset

### 5. Cost Tracking
- Token-based cost calculation
- Per-provider cost configuration
- Usage analytics

### 6. Audit Logging
- Complete request/response tracking
- Duration and token metrics
- Error tracking

## Configuration

Edit `gorest.yaml` to customize:

- **Provider Selection**: Choose default provider and enable/disable providers
- **Caching**: Enable/disable cache and set TTL
- **Quota**: Configure daily/monthly limits
- **Rate Limiting**: Set requests per minute
- **Security**: Enable authentication and authorization
- **Audit**: Configure audit logging and retention

## Production Deployment

For production:

1. **Enable Authentication**:
   ```yaml
   require_auth: true
   allow_anonymous: false
   ```

2. **Use Environment Variables** for API keys:
   ```yaml
   anthropic_api_key: "${ANTHROPIC_API_KEY}"
   ```

3. **Configure Quotas**:
   ```yaml
   enable_quota: true
   ```

4. **Enable Audit Logging**:
   ```yaml
   enable_audit: true
   retain_audit_days: 90
   ```

5. **Use a Production Database**:
   ```yaml
   database:
     url: "${DATABASE_URL}"
   ```

## Next Steps

- Add authentication with `gorest-auth` plugin
- Implement RBAC for admin endpoints
- Set up monitoring and alerting
- Configure provider-specific models
- Customize quota limits per user
- Add streaming support for real-time responses

## Support

For issues or questions:
- GitHub: https://github.com/nicolasbonnici/gorest-ai
- Documentation: See main README.md
