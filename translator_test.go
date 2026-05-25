package ai

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mocks ---

type mockLocaleProvider struct {
	defaultLocale string
	targetLocales []string
}

func (m *mockLocaleProvider) DefaultLocale() string   { return m.defaultLocale }
func (m *mockLocaleProvider) TargetLocales() []string { return m.targetLocales }

type mockChatter struct {
	response *ChatResponseDTO
	err      error
	calls    int
}

func (m *mockChatter) Chat(_ context.Context, _ *ChatRequestDTO, _ *uuid.UUID) (*ChatResponseDTO, error) {
	m.calls++
	return m.response, m.err
}

type mockRow struct {
	value string
	err   error
}

func (r *mockRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) > 0 {
		if s, ok := dest[0].(*string); ok {
			*s = r.value
		}
	}
	return nil
}

type mockResult struct{}

func (r *mockResult) LastInsertId() (int64, error) { return 0, nil }
func (r *mockResult) RowsAffected() (int64, error) { return 1, nil }

type mockDB struct {
	queryRowFns []func(query string, args ...interface{}) database.Row
	execFns     []func(query string, args ...interface{}) (database.Result, error)
	qIdx        int
	eIdx        int
}

func (m *mockDB) QueryRow(_ context.Context, query string, args ...interface{}) database.Row {
	if m.qIdx < len(m.queryRowFns) {
		fn := m.queryRowFns[m.qIdx]
		m.qIdx++
		return fn(query, args...)
	}
	return &mockRow{err: sql.ErrNoRows}
}

func (m *mockDB) Exec(_ context.Context, query string, args ...interface{}) (database.Result, error) {
	if m.eIdx < len(m.execFns) {
		fn := m.execFns[m.eIdx]
		m.eIdx++
		return fn(query, args...)
	}
	return &mockResult{}, nil
}

func (m *mockDB) Query(_ context.Context, _ string, _ ...interface{}) (database.Rows, error) {
	panic("mockDB.Query not implemented")
}
func (m *mockDB) Connect(_ context.Context, _ string) error { panic("not implemented") }
func (m *mockDB) Close() error                              { return nil }
func (m *mockDB) Ping(_ context.Context) error              { return nil }
func (m *mockDB) Begin(_ context.Context) (database.Tx, error) {
	panic("not implemented")
}
func (m *mockDB) Dialect() database.Dialect                 { return nil }
func (m *mockDB) DriverName() string                        { return "postgres" }
func (m *mockDB) Introspector() database.SchemaIntrospector { return nil }

// --- helpers ---

func rowReturning(value string) func(string, ...interface{}) database.Row {
	return func(_ string, _ ...interface{}) database.Row {
		return &mockRow{value: value}
	}
}

func rowReturningErr(err error) func(string, ...interface{}) database.Row {
	return func(_ string, _ ...interface{}) database.Row {
		return &mockRow{err: err}
	}
}

func execOK() func(string, ...interface{}) (database.Result, error) {
	return func(_ string, _ ...interface{}) (database.Result, error) {
		return &mockResult{}, nil
	}
}

// --- tests ---

func TestSha256Hash(t *testing.T) {
	h := sha256Hash("hello")
	assert.Len(t, h, 64)
	assert.Equal(t, sha256Hash("hello"), sha256Hash("hello"))
	assert.NotEqual(t, sha256Hash("hello"), sha256Hash("world"))
}

func TestExtractJSON(t *testing.T) {
	plain := `{"a":"b"}`
	assert.Equal(t, plain, extractJSON(plain))

	wrapped := "```json\n{\"a\":\"b\"}\n```"
	assert.Equal(t, `{"a":"b"}`, extractJSON(wrapped))

	wrappedNoLang := "```\n{\"a\":\"b\"}\n```"
	assert.Equal(t, `{"a":"b"}`, extractJSON(wrappedNoLang))
}

func TestAutoTranslator_IsAllowed(t *testing.T) {
	cfg := &Config{}
	at := NewAutoTranslator(nil, nil, nil, cfg)

	assert.True(t, at.isAllowed("post"), "empty allowed list means all allowed")
	assert.True(t, at.isAllowed("comment"))

	cfg.AllowedResourceTypes = []string{"post"}
	assert.True(t, at.isAllowed("post"))
	assert.False(t, at.isAllowed("comment"))
}

func TestAutoTranslator_Translate_TypeNotAllowed(t *testing.T) {
	at := NewAutoTranslator(nil, nil, &mockLocaleProvider{"en", []string{"fr"}}, &Config{
		AllowedResourceTypes: []string{"post"},
	})
	result, err := at.Translate(context.Background(), "comment", uuid.New().String(), nil)
	require.NoError(t, err)
	assert.Empty(t, result.Translated)
	assert.Empty(t, result.Skipped)
	assert.Empty(t, result.Failed)
}

func TestAutoTranslator_Translate_NoTargetLocales(t *testing.T) {
	at := NewAutoTranslator(nil, nil, &mockLocaleProvider{"en", []string{}}, &Config{})
	result, err := at.Translate(context.Background(), "post", uuid.New().String(), nil)
	require.NoError(t, err)
	assert.Empty(t, result.Translated)
}

func TestAutoTranslator_Translate_SourceNotFound(t *testing.T) {
	db := &mockDB{
		queryRowFns: []func(string, ...interface{}) database.Row{
			rowReturningErr(sql.ErrNoRows),
		},
	}
	at := NewAutoTranslator(nil, db, &mockLocaleProvider{"en", []string{"fr"}}, &Config{})
	_, err := at.Translate(context.Background(), "post", uuid.New().String(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source translation not found")
}

func TestAutoTranslator_Translate_AllSkipped(t *testing.T) {
	sourceContent := `{"title":"Hello"}`
	hash := sha256Hash(sourceContent)

	db := &mockDB{
		queryRowFns: []func(string, ...interface{}) database.Row{
			rowReturning(sourceContent),
			rowReturning(hash),
			rowReturning(hash),
		},
	}
	at := NewAutoTranslator(nil, db, &mockLocaleProvider{"en", []string{"fr", "es"}}, &Config{})
	result, err := at.Translate(context.Background(), "post", uuid.New().String(), nil)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"fr", "es"}, result.Skipped)
	assert.Empty(t, result.Translated)
	assert.Empty(t, result.Failed)
}

func TestAutoTranslator_Translate_BatchSuccess(t *testing.T) {
	sourceContent := `{"title":"Hello"}`

	db := &mockDB{
		queryRowFns: []func(string, ...interface{}) database.Row{
			rowReturning(sourceContent),
			rowReturningErr(sql.ErrNoRows),
			rowReturningErr(sql.ErrNoRows),
		},
		execFns: []func(string, ...interface{}) (database.Result, error){
			execOK(), execOK(), execOK(), execOK(),
		},
	}

	chatter := &mockChatter{
		response: &ChatResponseDTO{
			Content:   `{"fr":{"title":"Bonjour"},"es":{"title":"Hola"}}`,
			CreatedAt: time.Now(),
		},
	}

	at := NewAutoTranslator(chatter, db, &mockLocaleProvider{"en", []string{"fr", "es"}}, &Config{})
	result, err := at.Translate(context.Background(), "post", uuid.New().String(), nil)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"fr", "es"}, result.Translated)
	assert.Empty(t, result.Failed)
	assert.Equal(t, 1, chatter.calls)
}

func TestAutoTranslator_Translate_BatchFallbackToPerLocale(t *testing.T) {
	sourceContent := `{"title":"Hello"}`

	db := &mockDB{
		queryRowFns: []func(string, ...interface{}) database.Row{
			rowReturning(sourceContent),
			rowReturningErr(sql.ErrNoRows),
			rowReturningErr(sql.ErrNoRows),
		},
		execFns: []func(string, ...interface{}) (database.Result, error){
			execOK(), execOK(), execOK(), execOK(),
		},
	}

	callCount := 0
	chatter := &mockChatterFn{
		fn: func(req *ChatRequestDTO) (*ChatResponseDTO, error) {
			callCount++
			if callCount == 1 {
				return nil, errors.New("batch failed")
			}
			return &ChatResponseDTO{Content: `{"title":"translated"}`, CreatedAt: time.Now()}, nil
		},
	}

	at := NewAutoTranslator(chatter, db, &mockLocaleProvider{"en", []string{"fr", "es"}}, &Config{})
	result, err := at.Translate(context.Background(), "post", uuid.New().String(), nil)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"fr", "es"}, result.Translated)
	assert.Empty(t, result.Failed)
	assert.Equal(t, 3, callCount)
}

func TestAutoTranslator_Translate_PerLocaleFails(t *testing.T) {
	sourceContent := `{"title":"Hello"}`

	db := &mockDB{
		queryRowFns: []func(string, ...interface{}) database.Row{
			rowReturning(sourceContent),
			rowReturningErr(sql.ErrNoRows),
		},
		execFns: []func(string, ...interface{}) (database.Result, error){
			execOK(),
		},
	}

	chatter := &mockChatterFn{
		fn: func(_ *ChatRequestDTO) (*ChatResponseDTO, error) {
			return nil, errors.New("AI unavailable")
		},
	}

	at := NewAutoTranslator(chatter, db, &mockLocaleProvider{"en", []string{"fr"}}, &Config{})
	result, err := at.Translate(context.Background(), "post", uuid.New().String(), nil)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"fr"}, result.Failed)
	assert.Empty(t, result.Translated)
}

func TestAutoTranslator_Translate_BatchResponseMissingLocale(t *testing.T) {
	sourceContent := `{"title":"Hello"}`

	db := &mockDB{
		queryRowFns: []func(string, ...interface{}) database.Row{
			rowReturning(sourceContent),
			rowReturningErr(sql.ErrNoRows),
		},
		execFns: []func(string, ...interface{}) (database.Result, error){
			execOK(),
		},
	}

	// batch returns a response but missing the requested locale; per-locale fallback also fails
	callCount := 0
	chatter := &mockChatterFn{
		fn: func(_ *ChatRequestDTO) (*ChatResponseDTO, error) {
			callCount++
			if callCount == 1 {
				return &ChatResponseDTO{Content: `{"de":{"title":"Hallo"}}`, CreatedAt: time.Now()}, nil
			}
			return nil, errors.New("per-locale also failed")
		},
	}

	at := NewAutoTranslator(chatter, db, &mockLocaleProvider{"en", []string{"fr"}}, &Config{})
	result, err := at.Translate(context.Background(), "post", uuid.New().String(), nil)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"fr"}, result.Failed)
}

func TestAutoTranslator_TranslateAsync_DoesNotBlock(t *testing.T) {
	db := &mockDB{
		queryRowFns: []func(string, ...interface{}) database.Row{
			rowReturningErr(sql.ErrNoRows),
		},
	}
	at := NewAutoTranslator(nil, db, &mockLocaleProvider{"en", []string{"fr"}}, &Config{})
	at.TranslateAsync(context.Background(), "post", uuid.New().String(), nil)
}

// mockChatterFn allows per-call control.
type mockChatterFn struct {
	fn func(*ChatRequestDTO) (*ChatResponseDTO, error)
}

func (m *mockChatterFn) Chat(_ context.Context, req *ChatRequestDTO, _ *uuid.UUID) (*ChatResponseDTO, error) {
	return m.fn(req)
}
