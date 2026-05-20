// Package ibapi provides a Go client library for the Interactive Brokers Web API.
//
// The IB Web API delivers real-time access to Interactive Brokers' trading functionality,
// including live market data, market scanners, and intra-day portfolio updates.
//
// Usage:
//
//	client, err := ibapi.NewClient(ibapi.WithBaseURL("https://localhost:5000"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Check authentication status
//	status, err := client.Session().AuthStatus(context.Background())
//
// For more information about the IB Web API, visit:
// https://www.interactivebrokers.com/campus/ibkr-api-page/cpapi-v1/
package ibapi

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client represents an Interactive Brokers Web API client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	httpClientConfig
	auth *AuthState
}

// ClientOption is a functional option for configuring the client.
type ClientOption func(*Client)

// WithBaseURL sets the base URL for the API.
// Default: "https://localhost:5000"
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithInsecureSkipVerify controls whether TLS certificate verification is skipped.
// This is useful when connecting to localhost with a self-signed certificate.
func WithInsecureSkipVerify(skip bool) ClientOption {
	return func(c *Client) {
		c.InsecureSkipVerify = skip
	}
}

// WithTimeout sets the request timeout for the HTTP client.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.Timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// httpClientConfig holds HTTP client configuration.
type httpClientConfig struct {
	InsecureSkipVerify bool
	Timeout            time.Duration
}

// AuthState holds authentication state for the client.
type AuthState struct {
	SessionID  string
	SSOExpires int64
	UserID     int64
}

// NewClient creates a new Interactive Brokers Web API client.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		baseURL: "https://localhost:5000",
		httpClientConfig: httpClientConfig{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	// Create HTTP client with TLS configuration
	if c.httpClient == nil {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.InsecureSkipVerify,
			},
		}
		c.httpClient = &http.Client{
			Transport: transport,
			Timeout:   c.Timeout,
		}
	}

	return c, nil
}

// BaseURL returns the base URL for API requests.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// BuildURL constructs a full URL from a path.
// If path contains a query string (e.g., "/path?key=val"),
// it will be properly parsed into path and query components.
func (c *Client) BuildURL(path string) string {
	u, _ := url.Parse(c.baseURL)
	// Parse the provided path to extract query string properly
	if parsed, err := url.Parse(path); err == nil {
		u.Path = parsed.Path
		u.RawQuery = parsed.RawQuery
	} else {
		u.Path = path
	}
	return u.String()
}

// doRequest performs an HTTP request and returns the response.
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

// Close closes the HTTP client and releases resources.
func (c *Client) Close() error {
	if c.httpClient != nil {
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
	return nil
}

// Session returns the Session service for authentication and session management.
func (c *Client) Session() *SessionService {
	return &SessionService{client: c}
}

// Contract returns the Contract service for security definitions.
func (c *Client) Contract() *ContractService {
	return &ContractService{client: c}
}

// MarketData returns the MarketData service for market data requests.
func (c *Client) MarketData() *MarketDataService {
	return &MarketDataService{client: c}
}

// Scanner returns the Scanner service for market scanner requests.
func (c *Client) Scanner() *ScannerService {
	return &ScannerService{client: c}
}

// APIError represents an API error response.
type APIError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"error,omitempty"`
	Status  string `json:"status,omitempty"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Status != "" {
		return e.Status
	}
	return fmt.Sprintf("API error: code=%d", e.Code)
}

// HTTPError is returned when the API returns an error response.
type HTTPError struct {
	HTTPStatusCode int
	APIError       APIError
}

func (e *HTTPError) Error() string {
	return e.APIError.Error()
}