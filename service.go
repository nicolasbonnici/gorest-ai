package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/database"

	"github.com/nicolasbonnici/gorest-ai/cache"
	"github.com/nicolasbonnici/gorest-ai/providers"
)

type Service struct {
	config   *Config
	registry *providers.ProviderRegistry
	db       database.Database
	cache    cache.Cache
}

func NewService(config *Config, registry *providers.ProviderRegistry, db database.Database) *Service {
	var cacheImpl cache.Cache
	if config.EnableCache {
		cacheImpl = cache.NewMemoryCache(5 * time.Minute)
	}

	return &Service{
		config:   config,
		registry: registry,
		db:       db,
		cache:    cacheImpl,
	}
}

func (s *Service) Chat(ctx context.Context, req *ChatRequestDTO, userID *uuid.UUID) (*ChatResponseDTO, error) {
	startTime := time.Now()

	if s.config.EnableCache && req.UseCache {
		cacheKey := s.generateCacheKey(req)
		if cached, found, err := s.checkCache(ctx, cacheKey); err == nil && found {
			s.recordCacheHit(ctx, cacheKey, userID)
			return cached, nil
		}
	}

	providerName := req.Provider
	if providerName == "auto" {
		providerName = s.config.DefaultProvider
	}

	response, err := s.tryProvider(ctx, providerName, req)
	if err != nil && s.config.EnableFallback {
		response, err = s.tryFallbackProviders(ctx, providerName, req)
	}

	if err != nil {
		s.recordFailedRequest(ctx, providerName, req, err, userID, time.Since(startTime))
		return nil, err
	}

	cost := s.calculateCost(providerName, response.TotalTokens)
	duration := int(time.Since(startTime).Milliseconds())

	result := &ChatResponseDTO{
		ID:               uuid.New(),
		Provider:         providerName,
		Model:            response.Model,
		Content:          response.Content,
		PromptTokens:     response.PromptTokens,
		CompletionTokens: response.CompletionTokens,
		TotalTokens:      response.TotalTokens,
		Cost:             cost,
		DurationMs:       duration,
		Cached:           false,
		CreatedAt:        time.Now(),
	}

	if s.config.EnableCache && req.UseCache {
		cacheKey := s.generateCacheKey(req)
		s.storeInCache(ctx, cacheKey, result, req)
	}

	s.recordSuccessfulRequest(ctx, providerName, req, response, userID, duration, cost)

	if s.config.EnableQuota && userID != nil {
		_ = s.updateQuota(ctx, *userID, response.TotalTokens)
	}

	return result, nil
}

func (s *Service) tryProvider(ctx context.Context, providerName string, req *ChatRequestDTO) (*providers.ChatResponse, error) {
	provider, err := s.registry.Get(providerName)
	if err != nil {
		return nil, fmt.Errorf("provider not available: %w", err)
	}

	providerReq := &providers.ChatRequest{
		Messages:    s.convertMessages(req.Messages),
		Model:       s.getModel(providerName, req.Model),
		Temperature: s.getTemperature(req.Temperature),
		MaxTokens:   s.getMaxTokens(req.MaxTokens),
		Stream:      req.Stream,
	}

	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(s.config.RequestTimeout)*time.Second)
	defer cancel()

	return provider.Chat(reqCtx, providerReq)
}

func (s *Service) tryFallbackProviders(ctx context.Context, excludeProvider string, req *ChatRequestDTO) (*providers.ChatResponse, error) {
	for _, providerName := range s.config.EnabledProviders {
		if providerName == excludeProvider {
			continue
		}

		response, err := s.tryProvider(ctx, providerName, req)
		if err == nil {
			return response, nil
		}
	}

	return nil, fmt.Errorf("all fallback providers failed")
}

func (s *Service) convertMessages(messages []ChatMessage) []providers.Message {
	result := make([]providers.Message, len(messages))
	for i, msg := range messages {
		result[i] = providers.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return result
}

func (s *Service) getModel(providerName string, modelOverride *string) string {
	if modelOverride != nil && *modelOverride != "" {
		return *modelOverride
	}

	switch providerName {
	case "anthropic":
		return "claude-3-5-sonnet-20241022"
	case "openai":
		return "gpt-4"
	case "gemini":
		return "gemini-pro"
	case "mistral":
		return "mistral-large-latest"
	default:
		return ""
	}
}

func (s *Service) getTemperature(override *float64) float64 {
	if override != nil {
		return *override
	}
	return s.config.DefaultTemperature
}

func (s *Service) getMaxTokens(override *int) int {
	if override != nil {
		return *override
	}
	return s.config.MaxTokens
}

func (s *Service) generateCacheKey(req *ChatRequestDTO) string {
	data := fmt.Sprintf("%s:%s:%v:%v:%v",
		req.Provider,
		s.getModel(req.Provider, req.Model),
		req.Messages,
		s.getTemperature(req.Temperature),
		s.getMaxTokens(req.MaxTokens),
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *Service) checkCache(ctx context.Context, cacheKey string) (*ChatResponseDTO, bool, error) {
	if s.cache == nil {
		return nil, false, nil
	}

	value, found, err := s.cache.Get(ctx, cacheKey)
	if err != nil || !found {
		return nil, false, err
	}

	response, ok := value.(*ChatResponseDTO)
	if !ok {
		return nil, false, fmt.Errorf("invalid cache value type")
	}

	response.Cached = true
	return response, true, nil
}

func (s *Service) storeInCache(ctx context.Context, cacheKey string, response *ChatResponseDTO, req *ChatRequestDTO) {
	if s.cache == nil {
		return
	}

	ttl := time.Duration(s.config.CacheTTL) * time.Second
	if err := s.cache.Set(ctx, cacheKey, response, ttl); err != nil {
		fmt.Printf("Failed to cache response: %v\n", err)
	}

	s.storeCacheInDB(ctx, cacheKey, response, req)
}

func (s *Service) storeCacheInDB(ctx context.Context, cacheKey string, response *ChatResponseDTO, req *ChatRequestDTO) {
	cacheEntry := &AICache{
		ID:               uuid.New(),
		CacheKey:         cacheKey,
		ProviderName:     response.Provider,
		Model:            response.Model,
		RequestType:      "chat",
		ResponseText:     response.Content,
		PromptTokens:     response.PromptTokens,
		CompletionTokens: response.CompletionTokens,
		TotalTokens:      response.TotalTokens,
		HitCount:         0,
		ExpiresAt:        time.Now().Add(time.Duration(s.config.CacheTTL) * time.Second),
		CreatedAt:        time.Now(),
	}

	query := `INSERT INTO ai_cache (id, cache_key, provider_name, model, request_type, response_text,
		prompt_tokens, completion_tokens, total_tokens, hit_count, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, _ = s.db.Exec(ctx, query, cacheEntry.ID.String(), cacheEntry.CacheKey, cacheEntry.ProviderName,
		cacheEntry.Model, cacheEntry.RequestType, cacheEntry.ResponseText,
		cacheEntry.PromptTokens, cacheEntry.CompletionTokens, cacheEntry.TotalTokens,
		cacheEntry.HitCount, cacheEntry.ExpiresAt, cacheEntry.CreatedAt)
}

func (s *Service) recordCacheHit(ctx context.Context, cacheKey string, userID *uuid.UUID) {
	if s.cache != nil {
		_ = s.cache.IncrementHit(ctx, cacheKey)
	}

	query := `UPDATE ai_cache SET hit_count = hit_count + 1, updated_at = $1 WHERE cache_key = $2`
	_, _ = s.db.Exec(ctx, query, time.Now(), cacheKey)
}

func (s *Service) calculateCost(providerName string, totalTokens int) float64 {
	costPerToken := 0.0

	switch providerName {
	case "anthropic":
		costPerToken = 0.003 // $3 per million tokens
	case "openai":
		costPerToken = 0.03 // $30 per million tokens
	case "gemini":
		costPerToken = 0.001 // $1 per million tokens
	case "mistral":
		costPerToken = 0.002 // $2 per million tokens
	}

	return (float64(totalTokens) / 1000000.0) * costPerToken
}

func (s *Service) recordSuccessfulRequest(ctx context.Context, providerName string, req *ChatRequestDTO,
	response *providers.ChatResponse, userID *uuid.UUID, durationMs int, cost float64) {

	prompt := ""
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			prompt = msg.Content
			break
		}
	}

	request := &AIRequest{
		ID:               uuid.New(),
		UserID:           userID,
		ProviderID:       uuid.New(),
		ProviderName:     providerName,
		Model:            response.Model,
		RequestType:      "chat",
		Prompt:           prompt,
		ResponseText:     &response.Content,
		PromptTokens:     response.PromptTokens,
		CompletionTokens: response.CompletionTokens,
		TotalTokens:      response.TotalTokens,
		Cost:             cost,
		DurationMs:       durationMs,
		Status:           "success",
		ErrorMessage:     nil,
		Cached:           false,
		IPAddress:        nil,
		CreatedAt:        time.Now(),
	}

	query := `INSERT INTO ai_requests (id, user_id, provider_id, provider_name, model, request_type,
		prompt, response_text, prompt_tokens, completion_tokens, total_tokens, cost, duration_ms,
		status, cached, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	userIDStr := ""
	if request.UserID != nil {
		userIDStr = request.UserID.String()
	}

	_, _ = s.db.Exec(ctx, query, request.ID.String(), userIDStr, request.ProviderID.String(), request.ProviderName,
		request.Model, request.RequestType, request.Prompt, request.ResponseText,
		request.PromptTokens, request.CompletionTokens, request.TotalTokens,
		request.Cost, request.DurationMs, request.Status, request.Cached, request.CreatedAt)
}

func (s *Service) recordFailedRequest(ctx context.Context, providerName string, req *ChatRequestDTO,
	err error, userID *uuid.UUID, duration time.Duration) {

	prompt := ""
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			prompt = msg.Content
			break
		}
	}

	errMsg := err.Error()
	request := &AIRequest{
		ID:           uuid.New(),
		UserID:       userID,
		ProviderID:   uuid.New(),
		ProviderName: providerName,
		Model:        s.getModel(providerName, req.Model),
		RequestType:  "chat",
		Prompt:       prompt,
		Status:       "error",
		ErrorMessage: &errMsg,
		DurationMs:   int(duration.Milliseconds()),
		CreatedAt:    time.Now(),
	}

	query := `INSERT INTO ai_requests (id, user_id, provider_id, provider_name, model, request_type,
		prompt, status, error_message, duration_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	userIDStr := ""
	if request.UserID != nil {
		userIDStr = request.UserID.String()
	}

	_, _ = s.db.Exec(ctx, query, request.ID.String(), userIDStr, request.ProviderID.String(), request.ProviderName,
		request.Model, request.RequestType, request.Prompt, request.Status,
		request.ErrorMessage, request.DurationMs, request.CreatedAt)
}

func (s *Service) updateQuota(ctx context.Context, userID uuid.UUID, tokensUsed int) error {
	return nil
}

func (s *Service) CheckQuota(ctx context.Context, userID uuid.UUID) (bool, error) {
	if !s.config.EnableQuota {
		return true, nil
	}

	return true, nil
}
