package ai

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/logger"
	"github.com/nicolasbonnici/gorest/plugin"

	"github.com/nicolasbonnici/gorest-ai/migrations"
	"github.com/nicolasbonnici/gorest-ai/providers"
)

type Plugin struct {
	config   Config
	db       database.Database
	app      *fiber.App
	registry *providers.ProviderRegistry
	service  *Service
}

func NewPlugin() plugin.Plugin {
	return &Plugin{
		registry: providers.NewProviderRegistry(),
	}
}

func (p *Plugin) Name() string {
	return "ai"
}

func (p *Plugin) Initialize(config map[string]interface{}) error {
	p.config = DefaultConfig()

	if db, ok := config["database"].(database.Database); ok {
		p.db = db
		p.config.Database = db
	} else {
		return fmt.Errorf("database connection is required")
	}

	if err := p.parseConfig(config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	if err := p.config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	if err := p.initializeProviders(); err != nil {
		return fmt.Errorf("failed to initialize providers: %w", err)
	}

	p.service = NewService(&p.config, p.registry, p.db)

	logger.Log.Info("AI plugin initialized",
		"default_provider", p.config.DefaultProvider,
		"enabled_providers", p.config.EnabledProviders,
		"cache_enabled", p.config.EnableCache,
		"quota_enabled", p.config.EnableQuota,
	)

	return nil
}

func (p *Plugin) parseConfig(config map[string]interface{}) error {
	if v, ok := config["default_provider"].(string); ok {
		p.config.DefaultProvider = v
	}
	if v, ok := config["enabled_providers"].([]interface{}); ok {
		providers := make([]string, 0, len(v))
		for _, provider := range v {
			if str, ok := provider.(string); ok {
				providers = append(providers, str)
			}
		}
		p.config.EnabledProviders = providers
	}

	if v, ok := config["anthropic_api_key"].(string); ok {
		p.config.AnthropicAPIKey = v
	}
	if v, ok := config["anthropic_base_url"].(string); ok {
		p.config.AnthropicBaseURL = v
	}
	if v, ok := config["openai_api_key"].(string); ok {
		p.config.OpenAIAPIKey = v
	}
	if v, ok := config["openai_base_url"].(string); ok {
		p.config.OpenAIBaseURL = v
	}
	if v, ok := config["gemini_api_key"].(string); ok {
		p.config.GeminiAPIKey = v
	}
	if v, ok := config["gemini_base_url"].(string); ok {
		p.config.GeminiBaseURL = v
	}
	if v, ok := config["mistral_api_key"].(string); ok {
		p.config.MistralAPIKey = v
	}
	if v, ok := config["mistral_base_url"].(string); ok {
		p.config.MistralBaseURL = v
	}

	if v, ok := config["enable_cache"].(bool); ok {
		p.config.EnableCache = v
	}
	if v, ok := config["cache_ttl"].(int); ok {
		p.config.CacheTTL = v
	}
	if v, ok := config["enable_fallback"].(bool); ok {
		p.config.EnableFallback = v
	}
	if v, ok := config["enable_quota"].(bool); ok {
		p.config.EnableQuota = v
	}

	if v, ok := config["max_tokens"].(int); ok {
		p.config.MaxTokens = v
	}
	if v, ok := config["default_temperature"].(float64); ok {
		p.config.DefaultTemperature = v
	}
	if v, ok := config["rate_limit_per_min"].(int); ok {
		p.config.RateLimitPerMin = v
	}
	if v, ok := config["request_timeout"].(int); ok {
		p.config.RequestTimeout = v
	}

	if v, ok := config["pagination_limit"].(int); ok {
		p.config.PaginationLimit = v
	}
	if v, ok := config["max_pagination_limit"].(int); ok {
		p.config.MaxPaginationLimit = v
	}

	if v, ok := config["require_auth"].(bool); ok {
		p.config.RequireAuth = v
	}
	if v, ok := config["allow_anonymous"].(bool); ok {
		p.config.AllowAnonymous = v
	}

	if v, ok := config["enable_audit"].(bool); ok {
		p.config.EnableAudit = v
	}
	if v, ok := config["retain_audit_days"].(int); ok {
		p.config.RetainAuditDays = v
	}

	return nil
}

func (p *Plugin) initializeProviders() error {
	logger.Log.Info("Provider initialization pending implementation",
		"enabled_providers", p.config.EnabledProviders,
	)

	return nil
}

func (p *Plugin) MigrationSource() any {
	return migrations.NewSource()
}

func (p *Plugin) MigrationDependencies() []string {
	return nil
}

func (p *Plugin) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

func (p *Plugin) SetupEndpoints(app *fiber.App) error {
	p.app = app

	setupRoutes(app, p)

	logger.Log.Info("AI plugin endpoints registered")
	return nil
}

func (p *Plugin) GetService() *Service {
	return p.service
}

func (p *Plugin) GetRegistry() *providers.ProviderRegistry {
	return p.registry
}
