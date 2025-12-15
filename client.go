package mapon

import (
	"context"
	"net/http"
	"runtime/debug"
	"time"
)

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
	debug        bool
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

func (cc clientConfig) with(opts ...ClientOption) clientConfig {
	for _, opt := range opts {
		opt(&cc)
	}
	return cc
}

// ClientOption is a configuration option for a [Client].
type ClientOption func(*clientConfig)

// WithAPIKey sets the API key for the client.
func WithAPIKey(apiKey string) ClientOption {
	return func(config *clientConfig) {
		config.apiKey = apiKey
	}
}

// WithDebug toggles debug mode (request/response dumps to stderr).
func WithDebug(debug bool) ClientOption {
	return func(config *clientConfig) {
		config.debug = debug
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
	transport := http.DefaultTransport
	if cfg.debug {
		transport = &debugTransport{next: transport}
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
		Timeout:   cfg.timeout,
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