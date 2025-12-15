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
	c := Client{
		apiEndpoint: apiEndpoint,
		httpClient:  httpClient,
		semaphore:   make(chan struct{}, maxParallel),
	}
	for i := 0; i < maxParallel; i++ {
		c.semaphore <- struct{}{}
	}

	return c
}

// QueryContext sends request to OverpassAPI with provided querystring and context for cancellation/timeout.
func (c *Client) QueryContext(ctx context.Context, query string) (Result, error) {
	body, err := c.httpPost(ctx, query)
	if err != nil {
		return Result{}, err
	}
	return unmarshal(body)
}

// Query is deprecated: use QueryContext instead.
// It sends request to OverpassAPI with context.Background().
func (c *Client) Query(query string) (Result, error) {
	return c.QueryContext(context.Background(), query)
}

var DefaultClient = New()
