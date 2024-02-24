package webexapi

import (
	"context"
	"net/http"
)

// Client provides methods to interact with the Webex API
type Client interface {
	// CreateMessage creates a new message with the specified content
	CreateMessage(ctx context.Context, request CreateMessageRequest) (*Message, *WebexError, error)

	// GetMessage gets a message with the specified identifier
	GetMessage(ctx context.Context, messageId string) (*Message, *WebexError, error)

	// GetAttachmentAction gets an attachment action with the specified identifier
	GetAttachmentAction(ctx context.Context, attachmentActionId string) (*AttachmentAction, *WebexError, error)
}

// client is an implementation of the Client interface
type client struct {
	webexURL   string
	authToken  string
	httpClient http.Client
}

// DefaultWebexURL is used as the default URL to the Webex API
const DefaultWebexURL = "https://webexapis.com/v1"

// DefaultHTTPClient is used as the default HTTP client for making requests to the Webex API
var DefaultHTTPClient = http.Client{}

// NewClient creates a new instance of the Client with the specified authentication token. It uses the DefaultWebexURL
// as a URL to the Webex API and the DefaultHTTPClient as an HTTP client for making requests to the API. If it is
// needed, these parameters could be redefined by options.
func NewClient(token string, options ...Option) Client {
	client := &client{
		webexURL:   DefaultWebexURL,
		authToken:  token,
		httpClient: DefaultHTTPClient,
	}

	for _, option := range options {
		option(client)
	}
	return client
}

// Option configures the Client behavior
type Option func(*client)

// WithHTTPClient creates an option which defines an HTTP client used by the Client
func WithHTTPClient(httpClient http.Client) Option {
	return func(c *client) {
		c.httpClient = httpClient
	}
}

// WithWebexURL creates an option which defines a Webex API URL used by the Client
func WithWebexURL(webexURL string) Option {
	return func(c *client) {
		c.webexURL = webexURL
	}
}
