package ai

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func setupRoutes(app *fiber.App, plugin *Plugin) {
	api := app.Group("/api/ai")

	api.Post("/chat", plugin.handleChat)
	api.Post("/chat/stream", plugin.handleChatStream)

	providers := api.Group("/providers")
	providers.Post("/", plugin.handleCreateProvider)
	providers.Get("/:id", plugin.handleGetProvider)
	providers.Get("/", plugin.handleListProviders)
	providers.Put("/:id", plugin.handleUpdateProvider)
	providers.Delete("/:id", plugin.handleDeleteProvider)

	api.Get("/usage", plugin.handleGetUsage)
	api.Get("/usage/quota", plugin.handleGetQuota)

	requests := api.Group("/requests")
	requests.Get("/", plugin.handleListRequests)
	requests.Get("/:id", plugin.handleGetRequest)
}

func (p *Plugin) handleChat(c *fiber.Ctx) error {
	var req ChatRequestDTO
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponseDTO{
			Error:   "invalid_request",
			Message: "Failed to parse request body",
			Details: err.Error(),
		})
	}

	var userID *uuid.UUID
	if p.config.RequireAuth {
	}

	if p.config.EnableQuota && userID != nil {
		allowed, err := p.service.CheckQuota(c.Context(), *userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponseDTO{
				Error:   "quota_check_failed",
				Message: "Failed to check quota",
			})
		}
		if !allowed {
			return c.Status(fiber.StatusTooManyRequests).JSON(ErrorResponseDTO{
				Error:   "quota_exceeded",
				Message: "You have exceeded your usage quota",
			})
		}
	}

	response, err := p.service.Chat(c.Context(), &req, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponseDTO{
			Error:   "chat_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(response)
}

func (p *Plugin) handleChatStream(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(ErrorResponseDTO{
		Error:   "not_implemented",
		Message: "Streaming is not yet implemented",
	})
}

func (p *Plugin) handleCreateProvider(c *fiber.Ctx) error {
	var dto ProviderCreateDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponseDTO{
			Error:   "invalid_request",
			Message: "Failed to parse request body",
		})
	}

	provider := ToProviderModel(&dto)

	query := `INSERT INTO ai_providers (id, name, display_name, api_key, base_url, enabled,
		priority, max_tokens, temperature, rate_limit, cost_per_token, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := p.db.Exec(query, provider.ID, provider.Name, provider.DisplayName,
		provider.APIKey, provider.BaseURL, provider.Enabled, provider.Priority,
		provider.MaxTokens, provider.Temperature, provider.RateLimit,
		provider.CostPerToken, provider.CreatedAt)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponseDTO{
			Error:   "create_failed",
			Message: "Failed to create provider",
		})
	}

	responseDTO := ToProviderResponseDTO(provider)

	return c.Status(fiber.StatusCreated).JSON(responseDTO)
}

func (p *Plugin) handleGetProvider(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponseDTO{
			Error:   "invalid_id",
			Message: "Invalid provider ID",
		})
	}

	var provider AIProvider
	query := `SELECT id, name, display_name, api_key, base_url, enabled, priority,
		max_tokens, temperature, rate_limit, cost_per_token, created_at, updated_at
		FROM ai_providers WHERE id = $1`

	err = p.db.QueryRow(query, id).Scan(
		&provider.ID, &provider.Name, &provider.DisplayName, &provider.APIKey,
		&provider.BaseURL, &provider.Enabled, &provider.Priority, &provider.MaxTokens,
		&provider.Temperature, &provider.RateLimit, &provider.CostPerToken,
		&provider.CreatedAt, &provider.UpdatedAt,
	)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponseDTO{
			Error:   "not_found",
			Message: "Provider not found",
		})
	}

	responseDTO := ToProviderResponseDTO(&provider)
	return c.JSON(responseDTO)
}

func (p *Plugin) handleListProviders(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", p.config.PaginationLimit)
	offset := c.QueryInt("offset", 0)

	if limit > p.config.MaxPaginationLimit {
		limit = p.config.MaxPaginationLimit
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM ai_providers`
	p.db.QueryRow(countQuery).Scan(&total)

	query := `SELECT id, name, display_name, api_key, base_url, enabled, priority,
		max_tokens, temperature, rate_limit, cost_per_token, created_at, updated_at
		FROM ai_providers ORDER BY priority ASC, created_at DESC LIMIT $1 OFFSET $2`

	rows, err := p.db.Query(query, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponseDTO{
			Error:   "query_failed",
			Message: "Failed to query providers",
		})
	}
	defer rows.Close()

	providers := make([]*ProviderResponseDTO, 0)
	for rows.Next() {
		var provider AIProvider
		err := rows.Scan(
			&provider.ID, &provider.Name, &provider.DisplayName, &provider.APIKey,
			&provider.BaseURL, &provider.Enabled, &provider.Priority, &provider.MaxTokens,
			&provider.Temperature, &provider.RateLimit, &provider.CostPerToken,
			&provider.CreatedAt, &provider.UpdatedAt,
		)
		if err != nil {
			continue
		}

		providers = append(providers, ToProviderResponseDTO(&provider))
	}

	return c.JSON(PaginatedResponseDTO{
		Data:    providers,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: offset+limit < total,
	})
}

func (p *Plugin) handleUpdateProvider(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponseDTO{
			Error:   "invalid_id",
			Message: "Invalid provider ID",
		})
	}

	var dto ProviderUpdateDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponseDTO{
			Error:   "invalid_request",
			Message: "Failed to parse request body",
		})
	}

	var provider AIProvider
	query := `SELECT id, name, display_name, api_key, base_url, enabled, priority,
		max_tokens, temperature, rate_limit, cost_per_token, created_at, updated_at
		FROM ai_providers WHERE id = $1`

	err = p.db.QueryRow(query, id).Scan(
		&provider.ID, &provider.Name, &provider.DisplayName, &provider.APIKey,
		&provider.BaseURL, &provider.Enabled, &provider.Priority, &provider.MaxTokens,
		&provider.Temperature, &provider.RateLimit, &provider.CostPerToken,
		&provider.CreatedAt, &provider.UpdatedAt,
	)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponseDTO{
			Error:   "not_found",
			Message: "Provider not found",
		})
	}

	UpdateProviderModel(&provider, &dto)

	updateQuery := `UPDATE ai_providers SET display_name = $1, api_key = $2, base_url = $3,
		enabled = $4, priority = $5, max_tokens = $6, temperature = $7, rate_limit = $8,
		updated_at = $9 WHERE id = $10`

	_, err = p.db.Exec(updateQuery, provider.DisplayName, provider.APIKey, provider.BaseURL,
		provider.Enabled, provider.Priority, provider.MaxTokens, provider.Temperature,
		provider.RateLimit, provider.UpdatedAt, provider.ID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponseDTO{
			Error:   "update_failed",
			Message: "Failed to update provider",
		})
	}

	responseDTO := ToProviderResponseDTO(&provider)
	return c.JSON(responseDTO)
}

func (p *Plugin) handleDeleteProvider(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponseDTO{
			Error:   "invalid_id",
			Message: "Invalid provider ID",
		})
	}

	query := `DELETE FROM ai_providers WHERE id = $1`
	result, err := p.db.Exec(query, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponseDTO{
			Error:   "delete_failed",
			Message: "Failed to delete provider",
		})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponseDTO{
			Error:   "not_found",
			Message: "Provider not found",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

func (p *Plugin) handleGetUsage(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(ErrorResponseDTO{
		Error:   "not_implemented",
		Message: "Usage statistics not yet implemented",
	})
}

func (p *Plugin) handleGetQuota(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(ErrorResponseDTO{
		Error:   "not_implemented",
		Message: "Quota status not yet implemented",
	})
}

func (p *Plugin) handleListRequests(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(ErrorResponseDTO{
		Error:   "not_implemented",
		Message: "Request history not yet implemented",
	})
}

func (p *Plugin) handleGetRequest(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(ErrorResponseDTO{
		Error:   "not_implemented",
		Message: "Request details not yet implemented",
	})
}
