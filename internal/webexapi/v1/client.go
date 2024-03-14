package webexapi

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"
)

// Client provides methods to interact with the Webex API
type Client interface {
	// GetMyOwnDetails gets information about the currently authenticated user
	GetMyOwnDetails(ctx context.Context) (*Person, *WebexError, error)

	// CreateMessage creates a new message with the specified content
	CreateMessage(ctx context.Context, request CreateMessageRequest) (*Message, *WebexError, error)

	// GetMessage gets a message with the specified identifier
	GetMessage(ctx context.Context, messageId string) (*Message, *WebexError, error)

	// GetAttachmentAction gets an attachment action with the specified identifier
	GetAttachmentAction(ctx context.Context, attachmentActionId string) (*AttachmentAction, *WebexError, error)

	// CreateWebhook creates a new webhook with the specified parameters
	CreateWebhook(ctx context.Context, request CreateWebhookRequest) (*Webhook, *WebexError, error)
}

// client is an implementation of the Client interface
type client struct {
	webexURL   string
	authToken  string
	httpClient http.Client
}

// DefaultWebexURL is used as the default URL to the Webex API
const DefaultWebexURL = "https://webexapis.com/v1"

// defaultHTTPClient is used as the default HTTP client for making requests to the Webex API
var defaultHTTPClient = http.Client{
	Timeout: 15 * time.Second,
}

// NewClient creates a new instance of the Client with the specified authentication token. It uses the DefaultWebexURL
// as a URL to the Webex API and the DefaultHTTPClient as an HTTP client for making requests to the API. If it is
// needed, these parameters could be redefined by options.
func NewClient(token string, options ...Option) Client {
	client := &client{
		webexURL:   DefaultWebexURL,
		authToken:  token,
		httpClient: defaultHTTPClient,
	}

	for _, option := range options {
		option(client)
	}
	return client
}

// Option configures the Client behavior
type Option func(*client)

// WithTLSConfig creates an option which defines a TLS config used by the Client
func WithTLSConfig(config *tls.Config) Option {
	return func(c *client) {
		c.httpClient.Transport = &http.Transport{
			TLSClientConfig: config,
		}
	}
}

// WithWebexURL creates an option which defines a Webex API URL used by the Client
func WithWebexURL(webexURL string) Option {
	return func(c *client) {
		c.webexURL = webexURL
	}
}
