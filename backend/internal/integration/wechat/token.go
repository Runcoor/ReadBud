// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package wechat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	"readbud/internal/domain"
	"readbud/internal/pkg/crypto"
	"readbud/internal/repository/postgres"
)

// TokenProvider manages WeChat access token lifecycle.
// Supports direct, stable, and gateway_v2 modes per spec Section 2.
type TokenProvider interface {
	// GetAccessToken returns a valid access token for the given account (looked up by AppID).
	GetAccessToken(ctx context.Context, appID string) (string, error)
}

// ----- Errors -----

var (
	// ErrAccountNotFound is returned when no wechat_accounts row matches the AppID.
	ErrAccountNotFound = errors.New("wechat: account not found for app_id")
	// ErrAccountInactive is returned when the account is disabled.
	ErrAccountInactive = errors.New("wechat: account is inactive")
	// ErrMissingSecret is returned when the account has no encrypted AppSecret stored.
	ErrMissingSecret = errors.New("wechat: app_secret not configured")
	// ErrGatewayNotConfigured is returned when token_mode=gateway_v2 but no gateway client is wired.
	ErrGatewayNotConfigured = errors.New("wechat: gateway_v2 token provider not configured (set wechat.gateway.* in config)")
	// ErrUnknownTokenMode is returned for unrecognised token_mode values.
	ErrUnknownTokenMode = errors.New("wechat: unknown token_mode")
)

// APIError wraps a non-zero errcode response from WeChat.
type APIError struct {
	Code    int
	Message string
	Op      string // e.g. "getAccessToken", "stableToken"
}

func (e *APIError) Error() string {
	return fmt.Sprintf("wechat %s: errcode=%d errmsg=%q", e.Op, e.Code, e.Message)
}

// ----- Cache abstraction -----

// TokenCache is a TTL'd token store. Implementations should be safe for concurrent use.
type TokenCache interface {
	Get(ctx context.Context, appID string) (string, bool)
	Set(ctx context.Context, appID, token string, ttl time.Duration)
}

// InMemoryTokenCache is a process-local TTL cache. Suitable for single-instance deployments
// or as a fallback when Redis is unavailable.
type InMemoryTokenCache struct {
	mu     sync.RWMutex
	tokens map[string]*cachedEntry
}

type cachedEntry struct {
	token     string
	expiresAt time.Time
}

// NewInMemoryTokenCache creates a new in-memory token cache.
func NewInMemoryTokenCache() *InMemoryTokenCache {
	return &InMemoryTokenCache{tokens: make(map[string]*cachedEntry)}
}

// Get retrieves a cached token if it exists and hasn't expired.
func (c *InMemoryTokenCache) Get(_ context.Context, appID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cached, ok := c.tokens[appID]
	if !ok || time.Now().After(cached.expiresAt) {
		return "", false
	}
	return cached.token, true
}

// Set stores a token with the given TTL.
func (c *InMemoryTokenCache) Set(_ context.Context, appID, token string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokens[appID] = &cachedEntry{token: token, expiresAt: time.Now().Add(ttl)}
}

// RedisTokenCache stores tokens in Redis under "wechat:token:<appid>".
type RedisTokenCache struct {
	rdb       *redis.Client
	keyPrefix string
}

// NewRedisTokenCache creates a Redis-backed token cache. keyPrefix defaults to "wechat:token:".
func NewRedisTokenCache(rdb *redis.Client, keyPrefix string) *RedisTokenCache {
	if keyPrefix == "" {
		keyPrefix = "wechat:token:"
	}
	return &RedisTokenCache{rdb: rdb, keyPrefix: keyPrefix}
}

func (c *RedisTokenCache) key(appID string) string { return c.keyPrefix + appID }

// Get returns the cached token, or ("", false) if missing/expired/Redis error.
func (c *RedisTokenCache) Get(ctx context.Context, appID string) (string, bool) {
	v, err := c.rdb.Get(ctx, c.key(appID)).Result()
	if err != nil || v == "" {
		return "", false
	}
	return v, true
}

// Set stores the token with the given TTL. Errors are intentionally swallowed so that
// a Redis hiccup doesn't break the request — singleflight + WeChat-side cache still
// keep things correct, you just lose the local cache benefit for one cycle.
func (c *RedisTokenCache) Set(ctx context.Context, appID, token string, ttl time.Duration) {
	_ = c.rdb.Set(ctx, c.key(appID), token, ttl).Err()
}

// ----- Stub provider (kept for tests/local dev without DB) -----

// StubTokenProvider is a placeholder implementation for development without a real DB.
type StubTokenProvider struct {
	cache *InMemoryTokenCache
}

// NewStubTokenProvider creates a new stub token provider.
func NewStubTokenProvider() *StubTokenProvider {
	return &StubTokenProvider{cache: NewInMemoryTokenCache()}
}

// GetAccessToken returns a deterministic stub token. Do NOT use in production.
func (p *StubTokenProvider) GetAccessToken(ctx context.Context, appID string) (string, error) {
	if t, ok := p.cache.Get(ctx, appID); ok {
		return t, nil
	}
	token := fmt.Sprintf("stub_token_%s_%d", appID, time.Now().Unix())
	p.cache.Set(ctx, appID, token, 2*time.Hour)
	return token, nil
}

// ----- Real provider -----

// WeChat API endpoints.
const (
	endpointDirectToken = "https://api.weixin.qq.com/cgi-bin/token"
	endpointStableToken = "https://api.weixin.qq.com/cgi-bin/stable_token"

	// safetyMargin shaves time off the WeChat-reported expires_in so we refresh
	// before WeChat itself rejects the token.
	safetyMargin = 5 * time.Minute
	// minCacheTTL guards against pathological 0/negative expires_in.
	minCacheTTL = 30 * time.Second
)

// AccountResolver fetches and decrypts a WeChat account by AppID.
// Extracted as an interface so tests can stub the DB out.
type AccountResolver interface {
	ResolveByAppID(ctx context.Context, appID string) (*domain.WechatAccount, string, error)
}

// dbAccountResolver looks up WechatAccount via repo and decrypts AppSecretEnc with encKey.
type dbAccountResolver struct {
	repo   postgres.WechatAccountRepository
	encKey []byte
}

// NewDBAccountResolver creates an AccountResolver backed by Postgres.
func NewDBAccountResolver(repo postgres.WechatAccountRepository, encKey []byte) AccountResolver {
	return &dbAccountResolver{repo: repo, encKey: encKey}
}

func (r *dbAccountResolver) ResolveByAppID(ctx context.Context, appID string) (*domain.WechatAccount, string, error) {
	acct, err := r.repo.FindByAppID(ctx, appID)
	if err != nil {
		return nil, "", fmt.Errorf("resolve account: %w", err)
	}
	if acct == nil {
		return nil, "", ErrAccountNotFound
	}
	if acct.Status != domain.StatusActive {
		return nil, "", ErrAccountInactive
	}
	if strings.TrimSpace(acct.AppSecretEnc) == "" {
		// Gateway mode legitimately has no secret, callers will handle that.
		return acct, "", nil
	}
	secret, err := crypto.Decrypt(r.encKey, acct.AppSecretEnc)
	if err != nil {
		return nil, "", fmt.Errorf("decrypt app_secret: %w", err)
	}
	return acct, secret, nil
}

// AccountPersister persists rotated stable tokens back to the wechat_accounts row.
// This is best-effort — failures are logged, not propagated.
type AccountPersister interface {
	UpdateStableToken(ctx context.Context, acctID int64, token string, expireAt time.Time) error
}

type repoAccountPersister struct {
	repo postgres.WechatAccountRepository
}

// NewRepoAccountPersister wires a Postgres-backed persister.
func NewRepoAccountPersister(repo postgres.WechatAccountRepository) AccountPersister {
	return &repoAccountPersister{repo: repo}
}

func (p *repoAccountPersister) UpdateStableToken(ctx context.Context, acctID int64, token string, expireAt time.Time) error {
	acct, err := p.repo.FindByID(ctx, acctID)
	if err != nil {
		return err
	}
	if acct == nil {
		return ErrAccountNotFound
	}
	acct.StableAccessToken = &token
	t := expireAt
	acct.TokenExpireAt = &t
	return p.repo.Update(ctx, acct)
}

// GatewayClient resolves access tokens via an external gateway (e.g. WeChat 开放平台
// component access_token, or an in-house proxy). Implementations are out of scope here.
type GatewayClient interface {
	GetAccessToken(ctx context.Context, gatewayUserAPIID string) (token string, ttl time.Duration, err error)
}

// RealTokenProvider is the production implementation of TokenProvider.
type RealTokenProvider struct {
	resolver  AccountResolver
	persister AccountPersister // optional; may be nil
	cache     TokenCache
	gateway   GatewayClient // optional; nil unless gateway_v2 is configured
	http      *http.Client
	sf        singleflight.Group
	logger    *zap.Logger
}

// RealTokenProviderConfig configures RealTokenProvider.
type RealTokenProviderConfig struct {
	Resolver   AccountResolver
	Persister  AccountPersister
	Cache      TokenCache
	Gateway    GatewayClient
	HTTPClient *http.Client
	Logger     *zap.Logger
}

// NewRealTokenProvider builds a production-ready token provider.
// Resolver and Cache are required; Persister, Gateway, HTTPClient, Logger are optional.
func NewRealTokenProvider(cfg RealTokenProviderConfig) (*RealTokenProvider, error) {
	if cfg.Resolver == nil {
		return nil, errors.New("wechat: RealTokenProvider requires Resolver")
	}
	if cfg.Cache == nil {
		return nil, errors.New("wechat: RealTokenProvider requires Cache")
	}
	httpc := cfg.HTTPClient
	if httpc == nil {
		httpc = &http.Client{Timeout: 10 * time.Second}
	}
	lg := cfg.Logger
	if lg == nil {
		lg = zap.NewNop()
	}
	return &RealTokenProvider{
		resolver:  cfg.Resolver,
		persister: cfg.Persister,
		cache:     cfg.Cache,
		gateway:   cfg.Gateway,
		http:      httpc,
		logger:    lg,
	}, nil
}

// GetAccessToken returns a valid access token, fetching from WeChat if needed.
// Caches in Redis (or in-memory fallback) and dedupes concurrent fetches per AppID.
func (p *RealTokenProvider) GetAccessToken(ctx context.Context, appID string) (string, error) {
	if strings.TrimSpace(appID) == "" {
		return "", errors.New("wechat: app_id is required")
	}
	if t, ok := p.cache.Get(ctx, appID); ok {
		return t, nil
	}
	v, err, _ := p.sf.Do(appID, func() (any, error) {
		// Double-check after acquiring the singleflight slot; another goroutine may
		// have populated the cache while we were waiting.
		if t, ok := p.cache.Get(ctx, appID); ok {
			return t, nil
		}
		acct, secret, err := p.resolver.ResolveByAppID(ctx, appID)
		if err != nil {
			return "", err
		}
		token, ttl, err := p.fetchToken(ctx, acct, secret)
		if err != nil {
			return "", err
		}
		if ttl < minCacheTTL {
			ttl = minCacheTTL
		}
		p.cache.Set(ctx, appID, token, ttl)
		if acct.TokenMode == domain.TokenModeStable && p.persister != nil {
			if err := p.persister.UpdateStableToken(ctx, acct.ID, token, time.Now().Add(ttl)); err != nil {
				p.logger.Warn("persist stable token failed", zap.String("app_id", appID), zap.Error(err))
			}
		}
		return token, nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

// fetchToken dispatches to the right backend based on TokenMode and returns
// (token, ttlAfterSafetyMargin, err).
func (p *RealTokenProvider) fetchToken(ctx context.Context, acct *domain.WechatAccount, secret string) (string, time.Duration, error) {
	switch acct.TokenMode {
	case domain.TokenModeDirect:
		if secret == "" {
			return "", 0, ErrMissingSecret
		}
		return p.fetchDirect(ctx, acct.AppID, secret)
	case domain.TokenModeStable:
		if secret == "" {
			return "", 0, ErrMissingSecret
		}
		return p.fetchStable(ctx, acct.AppID, secret, false)
	case domain.TokenModeGateway:
		if p.gateway == nil {
			return "", 0, ErrGatewayNotConfigured
		}
		if acct.GatewayUserAPIID == nil || strings.TrimSpace(*acct.GatewayUserAPIID) == "" {
			return "", 0, fmt.Errorf("%w: gateway_user_api_id is empty", ErrGatewayNotConfigured)
		}
		token, ttl, err := p.gateway.GetAccessToken(ctx, *acct.GatewayUserAPIID)
		if err != nil {
			return "", 0, fmt.Errorf("gateway token fetch: %w", err)
		}
		return token, ttl - safetyMargin, nil
	default:
		return "", 0, fmt.Errorf("%w: %q", ErrUnknownTokenMode, acct.TokenMode)
	}
}

// directTokenResp models /cgi-bin/token success and error responses.
type directTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
}

// fetchDirect calls /cgi-bin/token (classic mode).
func (p *RealTokenProvider) fetchDirect(ctx context.Context, appID, secret string) (string, time.Duration, error) {
	url := fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s", endpointDirectToken, appID, secret)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", 0, fmt.Errorf("build direct token request: %w", err)
	}
	var body directTokenResp
	if err := p.doJSON(req, &body); err != nil {
		return "", 0, fmt.Errorf("direct token: %w", err)
	}
	if body.Errcode != 0 {
		return "", 0, &APIError{Code: body.Errcode, Message: body.Errmsg, Op: "getAccessToken"}
	}
	if body.AccessToken == "" {
		return "", 0, errors.New("direct token: empty access_token in response")
	}
	return body.AccessToken, time.Duration(body.ExpiresIn)*time.Second - safetyMargin, nil
}

// stableTokenReq models /cgi-bin/stable_token request body.
type stableTokenReq struct {
	GrantType    string `json:"grant_type"`
	AppID        string `json:"appid"`
	Secret       string `json:"secret"`
	ForceRefresh bool   `json:"force_refresh"`
}

// fetchStable calls /cgi-bin/stable_token (recommended mode).
func (p *RealTokenProvider) fetchStable(ctx context.Context, appID, secret string, forceRefresh bool) (string, time.Duration, error) {
	body, err := json.Marshal(stableTokenReq{
		GrantType:    "client_credential",
		AppID:        appID,
		Secret:       secret,
		ForceRefresh: forceRefresh,
	})
	if err != nil {
		return "", 0, fmt.Errorf("marshal stable token request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointStableToken, strings.NewReader(string(body)))
	if err != nil {
		return "", 0, fmt.Errorf("build stable token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	var resp directTokenResp
	if err := p.doJSON(req, &resp); err != nil {
		return "", 0, fmt.Errorf("stable token: %w", err)
	}
	if resp.Errcode != 0 {
		return "", 0, &APIError{Code: resp.Errcode, Message: resp.Errmsg, Op: "getStableAccessToken"}
	}
	if resp.AccessToken == "" {
		return "", 0, errors.New("stable token: empty access_token in response")
	}
	return resp.AccessToken, time.Duration(resp.ExpiresIn)*time.Second - safetyMargin, nil
}

// doJSON executes the request and decodes the JSON body into v.
func (p *RealTokenProvider) doJSON(req *http.Request, v any) error {
	res, err := p.http.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("http %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("decode json: %w (body=%s)", err, strings.TrimSpace(string(raw)))
	}
	return nil
}

// Compile-time interface checks.
var (
	_ TokenProvider = (*RealTokenProvider)(nil)
	_ TokenProvider = (*StubTokenProvider)(nil)
	_ TokenCache    = (*InMemoryTokenCache)(nil)
	_ TokenCache    = (*RedisTokenCache)(nil)
)
