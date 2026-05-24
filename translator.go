package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/database"
)

// Chatter allows Chat calls to be mocked in tests.
type Chatter interface {
	Chat(ctx context.Context, req *ChatRequestDTO, userID *uuid.UUID) (*ChatResponseDTO, error)
}

type AutoTranslator struct {
	service        Chatter
	db             database.Database
	localeProvider LocaleProvider
	config         *Config
}

type TranslationResult struct {
	Translated []string `json:"translated"`
	Skipped    []string `json:"skipped"`
	Failed     []string `json:"failed"`
}

func NewAutoTranslator(service Chatter, db database.Database, localeProvider LocaleProvider, config *Config) *AutoTranslator {
	return &AutoTranslator{
		service:        service,
		db:             db,
		localeProvider: localeProvider,
		config:         config,
	}
}

func (t *AutoTranslator) Translate(ctx context.Context, resourceType, resourceID string, userID *uuid.UUID) (*TranslationResult, error) {
	if !t.isAllowed(resourceType) {
		return &TranslationResult{}, nil
	}

	targetLocales := t.localeProvider.TargetLocales()
	if len(targetLocales) == 0 {
		return &TranslationResult{}, nil
	}

	sourceLocale := t.localeProvider.DefaultLocale()
	rawContent, err := t.readSourceContent(ctx, resourceType, resourceID, sourceLocale)
	if err != nil {
		return nil, fmt.Errorf("source translation not found for %s/%s locale=%s: %w", resourceType, resourceID, sourceLocale, err)
	}

	contentHash := sha256Hash(rawContent)
	result := &TranslationResult{}

	needed := make([]string, 0, len(targetLocales))
	for _, locale := range targetLocales {
		upToDate, _ := t.isUpToDate(ctx, resourceType, resourceID, sourceLocale, locale, contentHash)
		if upToDate {
			result.Skipped = append(result.Skipped, locale)
			continue
		}
		needed = append(needed, locale)
	}

	if len(needed) == 0 {
		return result, nil
	}

	translations, batchErr := t.batchTranslate(ctx, rawContent, sourceLocale, needed, userID)
	if batchErr != nil {
		for _, locale := range needed {
			translated, err := t.translateOne(ctx, rawContent, sourceLocale, locale, userID)
			if err != nil {
				result.Failed = append(result.Failed, locale)
				t.logTranslation(ctx, resourceType, resourceID, sourceLocale, locale, contentHash, "error", err.Error())
				continue
			}
			if err := t.upsertTranslation(ctx, resourceType, resourceID, locale, translated, userID); err != nil {
				result.Failed = append(result.Failed, locale)
				t.logTranslation(ctx, resourceType, resourceID, sourceLocale, locale, contentHash, "error", err.Error())
				continue
			}
			t.logTranslation(ctx, resourceType, resourceID, sourceLocale, locale, contentHash, "success", "")
			result.Translated = append(result.Translated, locale)
		}
		return result, nil
	}

	for _, locale := range needed {
		translated, ok := translations[locale]
		if !ok {
			result.Failed = append(result.Failed, locale)
			t.logTranslation(ctx, resourceType, resourceID, sourceLocale, locale, contentHash, "error", "locale missing from batch response")
			continue
		}
		if err := t.upsertTranslation(ctx, resourceType, resourceID, locale, translated, userID); err != nil {
			result.Failed = append(result.Failed, locale)
			t.logTranslation(ctx, resourceType, resourceID, sourceLocale, locale, contentHash, "error", err.Error())
			continue
		}
		t.logTranslation(ctx, resourceType, resourceID, sourceLocale, locale, contentHash, "success", "")
		result.Translated = append(result.Translated, locale)
	}

	return result, nil
}

func (t *AutoTranslator) TranslateAsync(ctx context.Context, resourceType, resourceID string, userID *uuid.UUID) {
	go func() {
		_, _ = t.Translate(context.Background(), resourceType, resourceID, userID)
	}()
}

func (t *AutoTranslator) isAllowed(resourceType string) bool {
	if len(t.config.AllowedResourceTypes) == 0 {
		return true
	}
	for _, rt := range t.config.AllowedResourceTypes {
		if rt == resourceType {
			return true
		}
	}
	return false
}

func (t *AutoTranslator) readSourceContent(ctx context.Context, resourceType, resourceID, locale string) (string, error) {
	row := t.db.QueryRow(ctx,
		`SELECT content FROM translations WHERE translatable = $1 AND translatable_id = $2 AND locale = $3 LIMIT 1`,
		resourceType, resourceID, locale)
	var content string
	if err := row.Scan(&content); err != nil {
		return "", err
	}
	return content, nil
}

func (t *AutoTranslator) isUpToDate(ctx context.Context, resourceType, resourceID, sourceLocale, targetLocale, contentHash string) (bool, error) {
	row := t.db.QueryRow(ctx,
		`SELECT source_hash FROM ai_translation_log WHERE translatable = $1 AND translatable_id = $2 AND source_locale = $3 AND target_locale = $4 AND status = 'success' LIMIT 1`,
		resourceType, resourceID, sourceLocale, targetLocale)
	var existingHash string
	if err := row.Scan(&existingHash); err != nil {
		return false, err
	}
	return existingHash == contentHash, nil
}

func (t *AutoTranslator) batchTranslate(ctx context.Context, rawContent, sourceLocale string, targetLocales []string, userID *uuid.UUID) (map[string]string, error) {
	systemPrompt := fmt.Sprintf(
		`You are a translator. Translate every string value in the given JSON from %s to the target locales listed. `+
			`Return only valid JSON where top-level keys are locale codes (e.g. "fr", "es") and values are objects with the same keys as the input. `+
			`Do not translate URLs, slugs, or content inside code blocks.`,
		sourceLocale,
	)
	userPrompt := fmt.Sprintf("Translate to locales %s:\n%s", strings.Join(targetLocales, ", "), rawContent)

	resp, err := t.service.Chat(ctx, &ChatRequestDTO{
		Provider: "auto",
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		UseCache: false,
	}, userID)
	if err != nil {
		return nil, err
	}

	content := extractJSON(resp.Content)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return nil, fmt.Errorf("batch response is not valid JSON: %w", err)
	}

	translations := make(map[string]string, len(targetLocales))
	for _, locale := range targetLocales {
		if v, ok := raw[locale]; ok {
			translations[locale] = string(v)
		}
	}
	if len(translations) == 0 {
		return nil, fmt.Errorf("batch response contained no matching locale keys")
	}
	return translations, nil
}

func (t *AutoTranslator) translateOne(ctx context.Context, rawContent, sourceLocale, targetLocale string, userID *uuid.UUID) (string, error) {
	systemPrompt := fmt.Sprintf(
		`You are a translator. Translate every string value in the given JSON from %s to %s. `+
			`Return only valid JSON with the same keys as the input. `+
			`Do not translate URLs, slugs, or content inside code blocks.`,
		sourceLocale, targetLocale,
	)

	resp, err := t.service.Chat(ctx, &ChatRequestDTO{
		Provider: "auto",
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: rawContent},
		},
		UseCache: false,
	}, userID)
	if err != nil {
		return "", err
	}

	content := extractJSON(resp.Content)
	var check json.RawMessage
	if err := json.Unmarshal([]byte(content), &check); err != nil {
		return "", fmt.Errorf("per-locale response is not valid JSON: %w", err)
	}
	return content, nil
}

func (t *AutoTranslator) upsertTranslation(ctx context.Context, resourceType, resourceID, locale, content string, userID *uuid.UUID) error {
	now := time.Now()
	var userIDStr interface{}
	if userID != nil {
		userIDStr = userID.String()
	}
	_, err := t.db.Exec(ctx,
		`INSERT INTO translations (id, user_id, translatable, translatable_id, locale, content, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (translatable_id, translatable, locale) DO UPDATE SET content = $6, updated_at = $8`,
		uuid.New().String(), userIDStr, resourceType, resourceID, locale, content, now, now)
	return err
}

func (t *AutoTranslator) logTranslation(ctx context.Context, resourceType, resourceID, sourceLocale, targetLocale, contentHash, status, errMsg string) {
	now := time.Now()
	var errMsgPtr interface{}
	if errMsg != "" {
		errMsgPtr = errMsg
	}
	_, _ = t.db.Exec(ctx,
		`INSERT INTO ai_translation_log (id, translatable, translatable_id, source_locale, target_locale, source_hash, status, error_message, translated_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (translatable, translatable_id, source_locale, target_locale) DO UPDATE SET source_hash = $6, status = $7, error_message = $8, translated_at = $9`,
		uuid.New().String(), resourceType, resourceID, sourceLocale, targetLocale, contentHash, status, errMsgPtr, now, now)
}

func sha256Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// extractJSON strips markdown code fences if present before JSON parsing.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		end := strings.LastIndex(s, "```")
		if end > 3 {
			s = strings.TrimSpace(s[strings.Index(s, "\n")+1 : end])
		}
	}
	return s
}
