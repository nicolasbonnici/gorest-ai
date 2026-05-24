package ai

import (
	"github.com/gofiber/fiber/v2"
)

func setupRoutes(app *fiber.App, p *Plugin) {
	ai := app.Group("/ai")

	ai.Post("/translate/:resource/:resource_id", func(c *fiber.Ctx) error {
		if p.autoTranslator == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "auto translation not configured: set locale provider and enable auto_translate",
			})
		}

		resourceType := c.Params("resource")
		resourceID := c.Params("resource_id")

		result, err := p.autoTranslator.Translate(c.UserContext(), resourceType, resourceID, nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(result)
	})
}
