package providers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockProvider implements Provider for testing
type MockProvider struct {
	name string
}

func NewMockProvider(name string) *MockProvider {
	return &MockProvider{name: name}
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	return &ChatResponse{
		Content: "Mock response",
		Model:   "mock-model",
	}, nil
}

func (m *MockProvider) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, <-chan error) {
	chunkChan := make(chan StreamChunk, 1)
	errChan := make(chan error)

	go func() {
		chunkChan <- StreamChunk{Delta: "Mock", Done: false}
		chunkChan <- StreamChunk{Delta: "", Done: true}
		close(chunkChan)
		close(errChan)
	}()

	return chunkChan, errChan
}

func (m *MockProvider) CountTokens(ctx context.Context, messages []Message) (int, error) {
	total := 0
	for _, msg := range messages {
		total += len(msg.Content) / 4
	}
	return total, nil
}

func (m *MockProvider) ValidateConfig() error {
	return nil
}

func (m *MockProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func TestProviderRegistry(t *testing.T) {
	registry := NewProviderRegistry()

	// Test registration
	provider1 := NewMockProvider("test1")
	err := registry.Register(provider1)
	assert.NoError(t, err)

	// Test duplicate registration
	err = registry.Register(provider1)
	assert.Error(t, err)

	// Test empty name
	providerEmptyName := NewMockProvider("")
	err = registry.Register(providerEmptyName)
	assert.Error(t, err)

	// Test Get
	retrieved, err := registry.Get("test1")
	assert.NoError(t, err)
	assert.Equal(t, "test1", retrieved.Name())

	// Test Get non-existent
	_, err = registry.Get("nonexistent")
	assert.Error(t, err)

	// Test List
	provider2 := NewMockProvider("test2")
	registry.Register(provider2)

	names := registry.List()
	assert.Contains(t, names, "test1")
	assert.Contains(t, names, "test2")
	assert.Len(t, names, 2)

	// Test GetAll
	all := registry.GetAll()
	assert.Len(t, all, 2)
	assert.NotNil(t, all["test1"])
	assert.NotNil(t, all["test2"])

	// Test Unregister
	err = registry.Unregister("test1")
	assert.NoError(t, err)

	_, err = registry.Get("test1")
	assert.Error(t, err)

	// Test Unregister non-existent
	err = registry.Unregister("nonexistent")
	assert.Error(t, err)
}

func TestCacheEntryIsExpired(t *testing.T) {
	// Not implemented in this test file as CacheEntry is defined in cache package
	// This is just a placeholder to show the test structure
}
