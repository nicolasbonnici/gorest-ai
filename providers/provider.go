package providers

import (
	"context"
	"fmt"
	"sync"
)

type Message struct {
	Role    string
	Content string
}

type ChatRequest struct {
	Messages    []Message
	Model       string
	Temperature float64
	MaxTokens   int
	Stream      bool
}

type ChatResponse struct {
	ID               string
	Content          string
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	FinishReason     string
}

type TokenMetrics struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type StreamChunk struct {
	Delta   string
	Done    bool
	Metrics *TokenMetrics
}

type Provider interface {
	Name() string
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, <-chan error)
	CountTokens(ctx context.Context, messages []Message) (int, error)
	ValidateConfig() error
	HealthCheck(ctx context.Context) error
}

type ProviderRegistry struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]Provider),
	}
}

func (r *ProviderRegistry) Register(provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := provider.Name()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %s is already registered", name)
	}

	r.providers[name] = provider
	return nil
}

func (r *ProviderRegistry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

func (r *ProviderRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}

	return names
}

func (r *ProviderRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	delete(r.providers, name)
	return nil
}

func (r *ProviderRegistry) GetAll() map[string]Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]Provider, len(r.providers))
	for name, provider := range r.providers {
		result[name] = provider
	}

	return result
}
