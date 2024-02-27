package webexapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type ResourceKind string

const (
	AttachmentActions   ResourceKind = "attachmentActions"
	Memberships         ResourceKind = "memberships"
	Messages            ResourceKind = "messages"
	Rooms               ResourceKind = "rooms"
	Meetings            ResourceKind = "meetings"
	Recordings          ResourceKind = "recordings"
	MeetingParticipants ResourceKind = "meetingParticipants"
	MeetingTranscripts  ResourceKind = "meetingTranscripts"
)

type ResourceEvent string

const (
	ResourceCreated  ResourceEvent = "created"
	ResourceUpdated  ResourceEvent = "updated"
	ResourceDeleted  ResourceEvent = "deleted"
	ResourceStarted  ResourceEvent = "started"
	ResourceEnded    ResourceEvent = "ended"
	ResourceJoined   ResourceEvent = "joined"
	ResourceLeft     ResourceEvent = "left"
	ResourceMigrated ResourceEvent = "migrated"
	AllEvents        ResourceEvent = "all"
)

type WebhookStatus string

const (
	WebhookStatusActive   WebhookStatus = "active"
	WebhookStatusInactive WebhookStatus = "inactive"
)

type WebhookCallback struct {
	Id        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Resource  ResourceKind    `json:"resource,omitempty"`
	Event     ResourceEvent   `json:"event,omitempty"`
	Filter    string          `json:"filter,omitempty"`
	OrgId     string          `json:"orgId,omitempty"`
	CreatedBy string          `json:"createdBy,omitempty"`
	AppId     string          `json:"appId,omitempty"`
	OwnedBy   string          `json:"ownedBy,omitempty"`
	Status    WebhookStatus   `json:"status,omitempty"`
	ActorId   string          `json:"actorId,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

type CreateWebhookRequest struct {
	Name      string        `json:"name,omitempty"`
	TargetUrl string        `json:"targetUrl,omitempty"`
	Resource  ResourceKind  `json:"resource,omitempty"`
	Event     ResourceEvent `json:"event,omitempty"`
	Filter    string        `json:"filter,omitempty"`
	Secret    string        `json:"secret,omitempty"`
	OwnedBy   string        `json:"ownedBy,omitempty"`
}

type Webhook struct {
	Id        string        `json:"id,omitempty"`
	Name      string        `json:"name,omitempty"`
	TargetUrl string        `json:"targetUrl,omitempty"`
	Resource  ResourceKind  `json:"resource,omitempty"`
	Event     ResourceEvent `json:"event,omitempty"`
	Filter    string        `json:"filter,omitempty"`
	Secret    string        `json:"secret,omitempty"`
	Status    WebhookStatus `json:"status,omitempty"`
	Created   time.Time     `json:"created"`
	OwnedBy   string        `json:"ownedBy,omitempty"`
}

func (c *client) CreateWebhook(
	ctx context.Context,
	request CreateWebhookRequest,
) (*Webhook, *WebexError, error) {
	httpRequest, err := c.newHTTPRequestToCreateWebhook(ctx, request)
	if err != nil {
		return nil, nil, err
	}

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return nil, nil, fmt.Errorf("the following error occurred during the request execution : %w", err)
	}
	defer httpResponse.Body.Close()

	if !isSuccessful(httpResponse) {
		webexError, err := fetchWebexErrorFrom(httpResponse)
		return nil, webexError, err
	}

	response := new(Webhook)
	err = json.NewDecoder(httpResponse.Body).Decode(response)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to decode the response body : %w", err)
	}
	return response, nil, nil
}

func (c *client) newHTTPRequestToCreateWebhook(
	ctx context.Context,
	request CreateWebhookRequest,
) (*http.Request, error) {
	encodedBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("unable to encode the request body in JSON format : %w", err)
	}
	bodyReader := bytes.NewReader(encodedBody)

	fullURL, err := url.JoinPath(c.webexURL, "/webhooks")
	if err != nil {
		return nil, fmt.Errorf("unable to construct the request URL : %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("unable to create a new HTTP request : %w", err)
	}

	authorizationValue := fmt.Sprintf("Bearer %s", c.authToken)
	httpRequest.Header.Add("Authorization", authorizationValue)
	httpRequest.Header.Add("Content-Type", "application/json")

	return httpRequest, nil
}
