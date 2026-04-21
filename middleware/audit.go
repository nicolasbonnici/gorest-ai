package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/logger"
)

// AuditConfig represents audit middleware configuration
type AuditConfig struct {
	// Enabled determines if audit logging is enabled
	Enabled bool

	// IncludeRequestBody determines if request body should be logged
	IncludeRequestBody bool

	// IncludeResponseBody determines if response body should be logged
	IncludeResponseBody bool
}

// AuditMiddleware creates a middleware for audit logging
func AuditMiddleware(config AuditConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if not enabled
		if !config.Enabled {
			return c.Next()
		}

		// Record start time
		start := time.Now()

		// Get request details
		method := c.Method()
		path := c.Path()
		ip := c.IP()

		// Get user ID if available
		userID := c.Locals("userID")

		// Log request
		if config.IncludeRequestBody {
			logger.Log.Info("AI API request",
				"method", method,
				"path", path,
				"ip", ip,
				"user_id", userID,
				"body", string(c.Body()),
			)
		} else {
			logger.Log.Info("AI API request",
				"method", method,
				"path", path,
				"ip", ip,
				"user_id", userID,
			)
		}

		// Execute request
		err := c.Next()

		// Record duration
		duration := time.Since(start)

		// Log response
		statusCode := c.Response().StatusCode()
		durationMs := duration.Milliseconds()

		if err != nil {
			if config.IncludeResponseBody {
				logger.Log.Error("AI API request failed",
					"method", method,
					"path", path,
					"ip", ip,
					"user_id", userID,
					"status", statusCode,
					"duration", durationMs,
					"response", string(c.Response().Body()),
					"error", err.Error(),
				)
			} else {
				logger.Log.Error("AI API request failed",
					"method", method,
					"path", path,
					"ip", ip,
					"user_id", userID,
					"status", statusCode,
					"duration", durationMs,
					"error", err.Error(),
				)
			}
		} else {
			if config.IncludeResponseBody {
				logger.Log.Info("AI API request completed",
					"method", method,
					"path", path,
					"ip", ip,
					"user_id", userID,
					"status", statusCode,
					"duration", durationMs,
					"response", string(c.Response().Body()),
				)
			} else {
				logger.Log.Info("AI API request completed",
					"method", method,
					"path", path,
					"ip", ip,
					"user_id", userID,
					"status", statusCode,
					"duration", durationMs,
				)
			}
		}

		return err
	}
}
