package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// QuotaConfig represents quota middleware configuration
type QuotaConfig struct {
	// Enabled determines if quota checking is enabled
	Enabled bool

	// QuotaChecker is a function that checks if a user has quota available
	QuotaChecker func(userID uuid.UUID) (bool, error)

	// GetUserID extracts the user ID from the context
	GetUserID func(c *fiber.Ctx) (*uuid.UUID, error)
}

// QuotaMiddleware creates a middleware for enforcing user quotas
func QuotaMiddleware(config QuotaConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if not enabled
		if !config.Enabled {
			return c.Next()
		}

		// Get user ID
		userID, err := config.GetUserID(c)
		if err != nil || userID == nil {
			// If no user ID, allow request (anonymous or auth disabled)
			return c.Next()
		}

		// Check quota
		allowed, err := config.QuotaChecker(*userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "quota_check_failed",
				"message": "Failed to check quota",
			})
		}

		if !allowed {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "quota_exceeded",
				"message": "You have exceeded your usage quota",
			})
		}

		return c.Next()
	}
}
