package mapon

import (
	"context"
	"net/http"
	"runtime/debug"
	"time"

	maponv1connect "github.com/way-platform/mapon-go/proto/gen/go/wayplatform/connect/mapon/v1/maponv1connect"
)

var _ maponv1connect.MaponApiClient = (*Client)(nil)

// BaseURL is the default base URL for the Mapon API.
const BaseURL = "https://mapon.com/api/v1"

// Client to the Mapon management APIs.
type Client struct {
	baseURL string
	config  clientConfig
}

// NewClient creates a new Mapon API client.
func NewClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
	config := newClientConfig()
	for _, opt := range opts {
		opt(&config)
	}
	client := &Client{
		baseURL: BaseURL,
		config:  config,
	}
	return client, nil
}

// clientConfig configures a [Client].
type clientConfig struct {
	apiKey       string
	httpClient   *http.Client
	retryCount   int
	timeout      time.Duration
	interceptors []func(http.RoundTripper) http.RoundTripper
}

func newClientConfig() clientConfig {
	return clientConfig{
		retryCount: 3,
		timeout:    30 * time.Second,
	}
}

// ClientOption is a configuration option for a [Client].
type ClientOption func(*clientConfig)

// WithAPIKey sets the API key for the client.
func WithAPIKey(apiKey string) ClientOption {
	return func(config *clientConfig) {
		config.apiKey = apiKey
	}
}

// WithHTTPClient sets the base HTTP client for the SDK client.
// The client's transport is used as the innermost transport in the chain.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(config *clientConfig) {
		config.httpClient = httpClient
	}
}

// WithRetryCount sets the number of retries for API requests.
func WithRetryCount(retryCount int) ClientOption {
	return func(config *clientConfig) {
		config.retryCount = retryCount
	}
}

// WithTimeout sets the timeout for API requests.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(config *clientConfig) {
		config.timeout = timeout
	}
}

// WithInterceptor adds a request interceptor for the [Client].
func WithInterceptor(interceptor func(http.RoundTripper) http.RoundTripper) ClientOption {
	return func(config *clientConfig) {
		config.interceptors = append(config.interceptors, interceptor)
	}
}

func (c *Client) httpClient(cfg clientConfig) *http.Client {
	transport := http.RoundTripper(http.DefaultTransport)
	timeout := cfg.timeout
	if cfg.httpClient != nil {
		if cfg.httpClient.Transport != nil {
			transport = cfg.httpClient.Transport
		}
		if cfg.httpClient.Timeout > 0 {
			timeout = cfg.httpClient.Timeout
		}
	}
	if cfg.apiKey != "" {
		transport = &apiKeyTransport{
			apiKey: cfg.apiKey,
			next:   transport,
		}
	}
	if len(cfg.interceptors) > 0 {
		transport = &interceptorTransport{
			interceptors: cfg.interceptors,
			next:         transport,
		}
	}
	if cfg.retryCount > 0 {
		transport = &retryTransport{
			maxRetries: cfg.retryCount,
			next:       transport,
		}
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

func getUserAgent() string {
	userAgent := "WayPlatformMaponGo"
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		userAgent += "/" + info.Main.Version
	}
	return userAgent
}
