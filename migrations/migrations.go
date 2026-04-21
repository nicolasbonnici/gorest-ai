package migrations

import (
	"context"
	"fmt"

	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/migrations"
)

// NewSource creates a new migration source for the AI plugin
func NewSource() migrations.MigrationSource {
	return migrations.NewGoMigrationSource("ai", GetMigrations())
}

// GetMigrations returns all migrations for the AI plugin
func GetMigrations() []migrations.Migration {
	return []migrations.Migration{
		createAIProvidersTable(),
		createAIRequestsTable(),
		createAICacheTable(),
		createAIQuotasTable(),
	}
}

// createAIProvidersTable creates the ai_providers table
func createAIProvidersTable() migrations.Migration {
	return migrations.Migration{
		Version: "20250101000001",
		Name:    "create_ai_providers_table",
		Source:  "ai",
		Executor: &dynamicSQLExecutor{
			upFunc: func(ctx context.Context, db database.Database) error {
				var query string

				switch db.DriverName() {
				case "postgres":
					query = `
					CREATE TABLE IF NOT EXISTS ai_providers (
						id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
						name VARCHAR(50) NOT NULL UNIQUE,
						display_name VARCHAR(100) NOT NULL,
						api_key TEXT NOT NULL,
						base_url VARCHAR(255),
						enabled BOOLEAN NOT NULL DEFAULT true,
						priority INTEGER NOT NULL DEFAULT 0,
						max_tokens INTEGER NOT NULL DEFAULT 4096,
						temperature DOUBLE PRECISION NOT NULL DEFAULT 0.7,
						rate_limit INTEGER NOT NULL DEFAULT 60,
						cost_per_token DOUBLE PRECISION NOT NULL DEFAULT 0.0,
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP
					);
					CREATE INDEX IF NOT EXISTS idx_ai_providers_enabled ON ai_providers(enabled);
					CREATE INDEX IF NOT EXISTS idx_ai_providers_priority ON ai_providers(priority);
				`
				case "mysql":
					query = `
					CREATE TABLE IF NOT EXISTS ai_providers (
						id CHAR(36) PRIMARY KEY,
						name VARCHAR(50) NOT NULL UNIQUE,
						display_name VARCHAR(100) NOT NULL,
						api_key TEXT NOT NULL,
						base_url VARCHAR(255),
						enabled BOOLEAN NOT NULL DEFAULT true,
						priority INT NOT NULL DEFAULT 0,
						max_tokens INT NOT NULL DEFAULT 4096,
						temperature DOUBLE NOT NULL DEFAULT 0.7,
						rate_limit INT NOT NULL DEFAULT 60,
						cost_per_token DOUBLE NOT NULL DEFAULT 0.0,
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP NULL,
						INDEX idx_ai_providers_enabled (enabled),
						INDEX idx_ai_providers_priority (priority)
					);
				`
				case "sqlite":
					query = `
					CREATE TABLE IF NOT EXISTS ai_providers (
						id TEXT PRIMARY KEY,
						name TEXT NOT NULL UNIQUE,
						display_name TEXT NOT NULL,
						api_key TEXT NOT NULL,
						base_url TEXT,
						enabled INTEGER NOT NULL DEFAULT 1,
						priority INTEGER NOT NULL DEFAULT 0,
						max_tokens INTEGER NOT NULL DEFAULT 4096,
						temperature REAL NOT NULL DEFAULT 0.7,
						rate_limit INTEGER NOT NULL DEFAULT 60,
						cost_per_token REAL NOT NULL DEFAULT 0.0,
						created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME
					);
					CREATE INDEX IF NOT EXISTS idx_ai_providers_enabled ON ai_providers(enabled);
					CREATE INDEX IF NOT EXISTS idx_ai_providers_priority ON ai_providers(priority);
				`
				default:
					return fmt.Errorf("unsupported database driver: %s", db.DriverName())
				}

				_, err := db.Exec(ctx, query)
				return err
			},
			downFunc: func(ctx context.Context, db database.Database) error {
				_, err := db.Exec(ctx, "DROP TABLE IF EXISTS ai_providers")
				return err
			},
		},
	}
}

// createAIRequestsTable creates the ai_requests table
func createAIRequestsTable() migrations.Migration {
	return migrations.Migration{
		Version: "20250101000002",
		Name:    "create_ai_requests_table",
		Source:  "ai",
		Executor: &dynamicSQLExecutor{
			upFunc: func(ctx context.Context, db database.Database) error {
				var query string

				switch db.DriverName() {
				case "postgres":
					query = `
					CREATE TABLE IF NOT EXISTS ai_requests (
						id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
						user_id UUID,
						provider_id UUID NOT NULL,
						provider_name VARCHAR(50) NOT NULL,
						model VARCHAR(100) NOT NULL,
						request_type VARCHAR(50) NOT NULL,
						prompt TEXT NOT NULL,
						response_text TEXT,
						prompt_tokens INTEGER NOT NULL DEFAULT 0,
						completion_tokens INTEGER NOT NULL DEFAULT 0,
						total_tokens INTEGER NOT NULL DEFAULT 0,
						cost DOUBLE PRECISION NOT NULL DEFAULT 0.0,
						duration_ms INTEGER NOT NULL DEFAULT 0,
						status VARCHAR(20) NOT NULL,
						error_message TEXT,
						cached BOOLEAN NOT NULL DEFAULT false,
						ip_address VARCHAR(45),
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
					);
					CREATE INDEX IF NOT EXISTS idx_ai_requests_user_id ON ai_requests(user_id, created_at DESC);
					CREATE INDEX IF NOT EXISTS idx_ai_requests_provider_id ON ai_requests(provider_id, created_at DESC);
					CREATE INDEX IF NOT EXISTS idx_ai_requests_status ON ai_requests(status);
					CREATE INDEX IF NOT EXISTS idx_ai_requests_created_at ON ai_requests(created_at DESC);
				`
				case "mysql":
					query = `
					CREATE TABLE IF NOT EXISTS ai_requests (
						id CHAR(36) PRIMARY KEY,
						user_id CHAR(36),
						provider_id CHAR(36) NOT NULL,
						provider_name VARCHAR(50) NOT NULL,
						model VARCHAR(100) NOT NULL,
						request_type VARCHAR(50) NOT NULL,
						prompt TEXT NOT NULL,
						response_text TEXT,
						prompt_tokens INT NOT NULL DEFAULT 0,
						completion_tokens INT NOT NULL DEFAULT 0,
						total_tokens INT NOT NULL DEFAULT 0,
						cost DOUBLE NOT NULL DEFAULT 0.0,
						duration_ms INT NOT NULL DEFAULT 0,
						status VARCHAR(20) NOT NULL,
						error_message TEXT,
						cached BOOLEAN NOT NULL DEFAULT false,
						ip_address VARCHAR(45),
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						INDEX idx_ai_requests_user_id (user_id, created_at),
						INDEX idx_ai_requests_provider_id (provider_id, created_at),
						INDEX idx_ai_requests_status (status),
						INDEX idx_ai_requests_created_at (created_at)
					);
				`
				case "sqlite":
					query = `
					CREATE TABLE IF NOT EXISTS ai_requests (
						id TEXT PRIMARY KEY,
						user_id TEXT,
						provider_id TEXT NOT NULL,
						provider_name TEXT NOT NULL,
						model TEXT NOT NULL,
						request_type TEXT NOT NULL,
						prompt TEXT NOT NULL,
						response_text TEXT,
						prompt_tokens INTEGER NOT NULL DEFAULT 0,
						completion_tokens INTEGER NOT NULL DEFAULT 0,
						total_tokens INTEGER NOT NULL DEFAULT 0,
						cost REAL NOT NULL DEFAULT 0.0,
						duration_ms INTEGER NOT NULL DEFAULT 0,
						status TEXT NOT NULL,
						error_message TEXT,
						cached INTEGER NOT NULL DEFAULT 0,
						ip_address TEXT,
						created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
					);
					CREATE INDEX IF NOT EXISTS idx_ai_requests_user_id ON ai_requests(user_id, created_at);
					CREATE INDEX IF NOT EXISTS idx_ai_requests_provider_id ON ai_requests(provider_id, created_at);
					CREATE INDEX IF NOT EXISTS idx_ai_requests_status ON ai_requests(status);
					CREATE INDEX IF NOT EXISTS idx_ai_requests_created_at ON ai_requests(created_at);
				`
				default:
					return fmt.Errorf("unsupported database driver: %s", db.DriverName())
				}

				_, err := db.Exec(ctx, query)
				return err
			},
			downFunc: func(ctx context.Context, db database.Database) error {
				_, err := db.Exec(ctx, "DROP TABLE IF EXISTS ai_requests")
				return err
			},
		},
	}
}

// createAICacheTable creates the ai_cache table
func createAICacheTable() migrations.Migration {
	return migrations.Migration{
		Version: "20250101000003",
		Name:    "create_ai_cache_table",
		Source:  "ai",
		Executor: &dynamicSQLExecutor{
			upFunc: func(ctx context.Context, db database.Database) error {
				var query string

				switch db.DriverName() {
				case "postgres":
					query = `
					CREATE TABLE IF NOT EXISTS ai_cache (
						id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
						cache_key VARCHAR(64) NOT NULL UNIQUE,
						provider_name VARCHAR(50) NOT NULL,
						model VARCHAR(100) NOT NULL,
						request_type VARCHAR(50) NOT NULL,
						response_text TEXT NOT NULL,
						prompt_tokens INTEGER NOT NULL DEFAULT 0,
						completion_tokens INTEGER NOT NULL DEFAULT 0,
						total_tokens INTEGER NOT NULL DEFAULT 0,
						hit_count INTEGER NOT NULL DEFAULT 0,
						expires_at TIMESTAMP NOT NULL,
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP
					);
					CREATE INDEX IF NOT EXISTS idx_ai_cache_cache_key ON ai_cache(cache_key);
					CREATE INDEX IF NOT EXISTS idx_ai_cache_expires_at ON ai_cache(expires_at);
				`
				case "mysql":
					query = `
					CREATE TABLE IF NOT EXISTS ai_cache (
						id CHAR(36) PRIMARY KEY,
						cache_key VARCHAR(64) NOT NULL UNIQUE,
						provider_name VARCHAR(50) NOT NULL,
						model VARCHAR(100) NOT NULL,
						request_type VARCHAR(50) NOT NULL,
						response_text TEXT NOT NULL,
						prompt_tokens INT NOT NULL DEFAULT 0,
						completion_tokens INT NOT NULL DEFAULT 0,
						total_tokens INT NOT NULL DEFAULT 0,
						hit_count INT NOT NULL DEFAULT 0,
						expires_at TIMESTAMP NOT NULL,
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP NULL,
						INDEX idx_ai_cache_cache_key (cache_key),
						INDEX idx_ai_cache_expires_at (expires_at)
					);
				`
				case "sqlite":
					query = `
					CREATE TABLE IF NOT EXISTS ai_cache (
						id TEXT PRIMARY KEY,
						cache_key TEXT NOT NULL UNIQUE,
						provider_name TEXT NOT NULL,
						model TEXT NOT NULL,
						request_type TEXT NOT NULL,
						response_text TEXT NOT NULL,
						prompt_tokens INTEGER NOT NULL DEFAULT 0,
						completion_tokens INTEGER NOT NULL DEFAULT 0,
						total_tokens INTEGER NOT NULL DEFAULT 0,
						hit_count INTEGER NOT NULL DEFAULT 0,
						expires_at DATETIME NOT NULL,
						created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME
					);
					CREATE INDEX IF NOT EXISTS idx_ai_cache_cache_key ON ai_cache(cache_key);
					CREATE INDEX IF NOT EXISTS idx_ai_cache_expires_at ON ai_cache(expires_at);
				`
				default:
					return fmt.Errorf("unsupported database driver: %s", db.DriverName())
				}

				_, err := db.Exec(ctx, query)
				return err
			},
			downFunc: func(ctx context.Context, db database.Database) error {
				_, err := db.Exec(ctx, "DROP TABLE IF EXISTS ai_cache")
				return err
			},
		},
	}
}

// createAIQuotasTable creates the ai_quotas table
func createAIQuotasTable() migrations.Migration {
	return migrations.Migration{
		Version: "20250101000004",
		Name:    "create_ai_quotas_table",
		Source:  "ai",
		Executor: &dynamicSQLExecutor{
			upFunc: func(ctx context.Context, db database.Database) error {
				var query string

				switch db.DriverName() {
				case "postgres":
					query = `
					CREATE TABLE IF NOT EXISTS ai_quotas (
						id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
						user_id UUID NOT NULL UNIQUE,
						daily_limit INTEGER NOT NULL DEFAULT 100,
						monthly_limit INTEGER NOT NULL DEFAULT 1000,
						daily_token_limit INTEGER NOT NULL DEFAULT 100000,
						monthly_token_limit INTEGER NOT NULL DEFAULT 1000000,
						daily_used INTEGER NOT NULL DEFAULT 0,
						monthly_used INTEGER NOT NULL DEFAULT 0,
						daily_tokens_used INTEGER NOT NULL DEFAULT 0,
						monthly_tokens_used INTEGER NOT NULL DEFAULT 0,
						reset_daily TIMESTAMP NOT NULL,
						reset_monthly TIMESTAMP NOT NULL,
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP
					);
					CREATE INDEX IF NOT EXISTS idx_ai_quotas_user_id ON ai_quotas(user_id);
				`
				case "mysql":
					query = `
					CREATE TABLE IF NOT EXISTS ai_quotas (
						id CHAR(36) PRIMARY KEY,
						user_id CHAR(36) NOT NULL UNIQUE,
						daily_limit INT NOT NULL DEFAULT 100,
						monthly_limit INT NOT NULL DEFAULT 1000,
						daily_token_limit INT NOT NULL DEFAULT 100000,
						monthly_token_limit INT NOT NULL DEFAULT 1000000,
						daily_used INT NOT NULL DEFAULT 0,
						monthly_used INT NOT NULL DEFAULT 0,
						daily_tokens_used INT NOT NULL DEFAULT 0,
						monthly_tokens_used INT NOT NULL DEFAULT 0,
						reset_daily TIMESTAMP NOT NULL,
						reset_monthly TIMESTAMP NOT NULL,
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP NULL,
						INDEX idx_ai_quotas_user_id (user_id)
					);
				`
				case "sqlite":
					query = `
					CREATE TABLE IF NOT EXISTS ai_quotas (
						id TEXT PRIMARY KEY,
						user_id TEXT NOT NULL UNIQUE,
						daily_limit INTEGER NOT NULL DEFAULT 100,
						monthly_limit INTEGER NOT NULL DEFAULT 1000,
						daily_token_limit INTEGER NOT NULL DEFAULT 100000,
						monthly_token_limit INTEGER NOT NULL DEFAULT 1000000,
						daily_used INTEGER NOT NULL DEFAULT 0,
						monthly_used INTEGER NOT NULL DEFAULT 0,
						daily_tokens_used INTEGER NOT NULL DEFAULT 0,
						monthly_tokens_used INTEGER NOT NULL DEFAULT 0,
						reset_daily DATETIME NOT NULL,
						reset_monthly DATETIME NOT NULL,
						created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME
					);
					CREATE INDEX IF NOT EXISTS idx_ai_quotas_user_id ON ai_quotas(user_id);
				`
				default:
					return fmt.Errorf("unsupported database driver: %s", db.DriverName())
				}

				_, err := db.Exec(ctx, query)
				return err
			},
			downFunc: func(ctx context.Context, db database.Database) error {
				_, err := db.Exec(ctx, "DROP TABLE IF EXISTS ai_quotas")
				return err
			},
		},
	}
}

// dynamicSQLExecutor implements MigrationExecutor using function callbacks
type dynamicSQLExecutor struct {
	upFunc   func(ctx context.Context, db database.Database) error
	downFunc func(ctx context.Context, db database.Database) error
}

func (e *dynamicSQLExecutor) Up(ctx context.Context, db database.Database) error {
	return e.upFunc(ctx, db)
}

func (e *dynamicSQLExecutor) Down(ctx context.Context, db database.Database) error {
	return e.downFunc(ctx, db)
}

func (e *dynamicSQLExecutor) Checksum() string {
	// For dynamic executors, we'll return empty checksum
	// The migrations package will handle checksum calculation
	return ""
}
