package wechat

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TokenProvider manages WeChat access token lifecycle.
// Supports direct, stable, and gateway_v2 modes per spec Section 2.
type TokenProvider interface {
	// GetAccessToken returns a valid access token for the given account.
	GetAccessToken(ctx context.Context, appID string) (string, error)
}

// CachedToken represents a cached WeChat access token with expiry.
type CachedToken struct {
	Token     string
	ExpiresAt time.Time
}

// InMemoryTokenCache is a simple in-memory token cache.
// In production, this should be replaced with Redis-backed cache.
type InMemoryTokenCache struct {
	mu     sync.RWMutex
	tokens map[string]*CachedToken
}

// NewInMemoryTokenCache creates a new in-memory token cache.
func NewInMemoryTokenCache() *InMemoryTokenCache {
	return &InMemoryTokenCache{
		tokens: make(map[string]*CachedToken),
	}
}

// Get retrieves a cached token if it exists and hasn't expired.
func (c *InMemoryTokenCache) Get(appID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.tokens[appID]
	if !ok || time.Now().After(cached.ExpiresAt) {
		return "", false
	}
	return cached.Token, true
}

// Set stores a token with the given TTL.
func (c *InMemoryTokenCache) Set(appID, token string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tokens[appID] = &CachedToken{
		Token:     token,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// StubTokenProvider is a placeholder implementation for development.
// Will be replaced with real WeChat API calls.
type StubTokenProvider struct {
	cache *InMemoryTokenCache
}

// NewStubTokenProvider creates a new stub token provider.
func NewStubTokenProvider() *StubTokenProvider {
	return &StubTokenProvider{
		cache: NewInMemoryTokenCache(),
	}
}

// GetAccessToken returns a stub token for development.
func (p *StubTokenProvider) GetAccessToken(_ context.Context, appID string) (string, error) {
	if token, ok := p.cache.Get(appID); ok {
		return token, nil
	}

	// In production, this would call WeChat's getAccessToken or getStableAccessToken API
	token := fmt.Sprintf("stub_token_%s_%d", appID, time.Now().Unix())
	p.cache.Set(appID, token, 2*time.Hour)
	return token, nil
}
