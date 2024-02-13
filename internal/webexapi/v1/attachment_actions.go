package webexapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type AttachmentAction struct {
	Id        string               `json:"id,omitempty"`
	PersonId  string               `json:"personId,omitempty"`
	RoomId    string               `json:"roomId,omitempty"`
	Type      AttachmentActionType `json:"type,omitempty"`
	MessageId string               `json:"messageId,omitempty"`
	Inputs    json.RawMessage      `json:"inputs,omitempty"`
	Created   time.Time            `json:"created"`
}

type AttachmentActionType string

const AttachmentActionTypeSubmit AttachmentActionType = "submit"

func (c *client) GetAttachmentAction(
	ctx context.Context,
	attachmentActionId string,
) (*AttachmentAction, *WebexError, error) {
	httpRequest, err := c.newHTTPRequestToGetAttachmentAction(ctx, attachmentActionId)
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

	response := new(AttachmentAction)
	err = json.NewDecoder(httpResponse.Body).Decode(response)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to decode the response body : %w", err)
	}
	return response, nil, nil
}

func (c *client) newHTTPRequestToGetAttachmentAction(
	ctx context.Context,
	attachmentActionId string,
) (*http.Request, error) {
	if len(attachmentActionId) == 0 {
		return nil, errors.New("the attachment action identifier is not specified")
	}

	fullURL, err := url.JoinPath(c.webexURL, "/attachment/actions", attachmentActionId)
	if err != nil {
		return nil, fmt.Errorf("unable to constract the request URL : %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create a new HTTP request : %w", err)
	}

	authorizationValue := fmt.Sprintf("Bearer %s", c.authToken)
	httpRequest.Header.Add("Authorization", authorizationValue)
	httpRequest.Header.Add("Content-Type", "application/json")

	return httpRequest, nil
}
