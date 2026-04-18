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
		logFields := map[string]interface{}{
			"method":  method,
			"path":    path,
			"ip":      ip,
			"user_id": userID,
		}

		if config.IncludeRequestBody {
			logFields["body"] = string(c.Body())
		}

		logger.Log.Info("AI API request", logFields)

		// Execute request
		err := c.Next()

		// Record duration
		duration := time.Since(start)

		// Log response
		responseFields := map[string]interface{}{
			"method":   method,
			"path":     path,
			"ip":       ip,
			"user_id":  userID,
			"status":   c.Response().StatusCode(),
			"duration": duration.Milliseconds(),
		}

		if config.IncludeResponseBody {
			responseFields["response"] = string(c.Response().Body())
		}

		if err != nil {
			responseFields["error"] = err.Error()
			logger.Log.Error("AI API request failed", responseFields)
		} else {
			logger.Log.Info("AI API request completed", responseFields)
		}

		return err
	}
}
