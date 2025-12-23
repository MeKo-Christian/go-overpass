package turbo

import "github.com/MeKo-Christian/go-overpass"

// NewClientWithOverride builds a client using Result.EndpointOverride when present.
// If both override and fallback are empty, it returns the default client.
func NewClientWithOverride(fallbackEndpoint string, maxParallel int, httpClient overpass.HTTPClient, res Result) overpass.Client {
	endpoint := ApplyEndpointOverride(fallbackEndpoint, res)
	if endpoint == "" {
		return overpass.New()
	}

	return overpass.NewWithSettings(endpoint, maxParallel, httpClient)
}
