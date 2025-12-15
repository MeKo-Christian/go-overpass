package overpass

import (
	"context"
	"net/http"
)

const apiEndpoint = "https://overpass-api.de/api/interpreter"

// HTTPClient interface for making HTTP requests with context support.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// A Client manages communication with the Overpass API.
type Client struct {
	apiEndpoint string
	httpClient  HTTPClient
	semaphore   chan struct{}
	retryConfig RetryConfig
	cache       *cache
	cacheCtx    context.Context
	cacheCancel context.CancelFunc
}

// New returns Client instance with default overpass-api.de endpoint.
func New() Client {
	return NewWithSettings(apiEndpoint, 1, http.DefaultClient)
}

// NewWithSettings returns Client with custom settings.
func NewWithSettings(
	apiEndpoint string,
	maxParallel int,
	httpClient HTTPClient,
) Client {
	ctx, cancel := context.WithCancel(context.Background())

	c := Client{
		apiEndpoint: apiEndpoint,
		httpClient:  httpClient,
		semaphore:   make(chan struct{}, maxParallel),
		retryConfig: DefaultRetryConfig(),
		cache:       newCache(DefaultCacheConfig()),
		cacheCtx:    ctx,
		cacheCancel: cancel,
	}
	for i := 0; i < maxParallel; i++ {
		c.semaphore <- struct{}{}
	}

	c.cache.startCleanupRoutine(ctx)
	return c
}

// NewWithRetry returns Client with custom retry configuration.
func NewWithRetry(
	apiEndpoint string,
	maxParallel int,
	httpClient HTTPClient,
	retryConfig RetryConfig,
) Client {
	c := NewWithSettings(apiEndpoint, maxParallel, httpClient)
	c.retryConfig = retryConfig
	return c
}

// SetRetryConfig updates the retry configuration for the client.
func (c *Client) SetRetryConfig(config RetryConfig) {
	c.retryConfig = config
}

// SetCacheConfig updates the cache configuration for the client.
func (c *Client) SetCacheConfig(config CacheConfig) {
	c.cache.mu.Lock()
	c.cache.config = config
	c.cache.mu.Unlock()

	// Restart cleanup routine if enabling cache
	if config.Enabled {
		c.cache.startCleanupRoutine(c.cacheCtx)
	}
}

// ClearCache removes all cached entries.
func (c *Client) ClearCache() {
	c.cache.clear()
}

// CacheSize returns the number of cached entries.
func (c *Client) CacheSize() int {
	return c.cache.size()
}

// Close stops the cache cleanup routine and releases resources.
func (c *Client) Close() {
	if c.cacheCancel != nil {
		c.cacheCancel()
	}
}

// QueryContext sends request to OverpassAPI with provided querystring and context for cancellation/timeout.
func (c *Client) QueryContext(ctx context.Context, query string) (Result, error) {
	// Check cache first
	if result, hit := c.cache.get(c.apiEndpoint, query); hit {
		return result, nil
	}

	var body []byte
	var err error

	// Use retry logic if MaxRetries > 0
	if c.retryConfig.MaxRetries > 0 {
		body, err = c.retryableHTTPPost(ctx, query)
	} else {
		body, err = c.httpPost(ctx, query)
	}

	if err != nil {
		return Result{}, err
	}

	result, err := unmarshal(body)
	if err != nil {
		return Result{}, err
	}

	// Store in cache
	c.cache.set(c.apiEndpoint, query, result)

	return result, nil
}

// Query is deprecated: use QueryContext instead.
// It sends request to OverpassAPI with context.Background().
func (c *Client) Query(query string) (Result, error) {
	return c.QueryContext(context.Background(), query)
}

// QueryWithBuilder executes query from builder (convenience method)
func (c *Client) QueryWithBuilder(ctx context.Context, builder *QueryBuilder) (Result, error) {
	return c.QueryContext(ctx, builder.Build())
}

var DefaultClient = New()

// QueryWithBuilder executes query from builder using DefaultClient
func QueryWithBuilder(ctx context.Context, builder *QueryBuilder) (Result, error) {
	return DefaultClient.QueryWithBuilder(ctx, builder)
}
