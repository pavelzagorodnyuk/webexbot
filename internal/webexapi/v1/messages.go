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

type Message struct {
	Id              string       `json:"id,omitempty"`
	ParentId        string       `json:"parentId,omitempty"`
	RoomId          string       `json:"roomId,omitempty"`
	RoomType        RoomType     `json:"roomType,omitempty"`
	Text            string       `json:"text,omitempty"`
	Markdown        string       `json:"markdown,omitempty"`
	Html            string       `json:"html,omitempty"`
	Files           []string     `json:"files,omitempty"`
	PersonId        string       `json:"personId,omitempty"`
	PersonEmail     string       `json:"personEmail,omitempty"`
	MentionedPeople []string     `json:"mentionedPeople,omitempty"`
	MentionedGroups []string     `json:"mentionedGroups,omitempty"`
	Attachments     []Attachment `json:"attachments,omitempty"`
	Created         time.Time    `json:"created"`
	Updated         time.Time    `json:"updated"`
	IsVoiceClip     bool         `json:"isVoiceClip,omitempty"`
}

type RoomType string

const (
	RoomTypeDirect RoomType = "direct"
	RoomTypeGroup  RoomType = "group"
)

type Attachment struct {
	ContentType string          `json:"contentType,omitempty"`
	Content     json.RawMessage `json:"content,omitempty"`
}

func (c *client) GetMessage(ctx context.Context, messageId string) (*Message, *WebexError, error) {
	httpRequest, err := c.newHTTPRequestToGetMessage(ctx, messageId)
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

	response := new(Message)
	err = json.NewDecoder(httpResponse.Body).Decode(response)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to decode the response body : %w", err)
	}
	return response, nil, nil
}

func (c *client) newHTTPRequestToGetMessage(ctx context.Context, messageId string) (*http.Request, error) {
	if len(messageId) == 0 {
		return nil, errors.New("the message identifier is not specified")
	}

	fullURL, err := url.JoinPath(c.webexURL, "/messages", messageId)
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
