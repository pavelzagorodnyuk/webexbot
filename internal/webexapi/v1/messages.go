package webexapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
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
		return nil, fmt.Errorf("unable to construct the request URL : %w", err)
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

type CreateMessageRequest struct {
	RoomId        string      `json:"roomId,omitempty"`
	ParentId      string      `json:"parentId,omitempty"`
	ToPersonId    string      `json:"toPersonId,omitempty"`
	ToPersonEmail string      `json:"toPersonEmail,omitempty"`
	Text          string      `json:"text,omitempty"`
	Markdown      string      `json:"markdown,omitempty"`
	Html          string      `json:"html,omitempty"`
	File          *File       `json:"-"`
	Attachment    *Attachment `json:"attachments,omitempty"`
}

// TODO: implement the use of public URLs as a file source
type File struct {
	Name    string
	Content io.Reader
}

func (c *client) CreateMessage(ctx context.Context, request CreateMessageRequest) (*Message, *WebexError, error) {
	httpRequest, err := c.newHTTPRequestToCreateMessage(ctx, request)
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

func (c *client) newHTTPRequestToCreateMessage(
	ctx context.Context,
	request CreateMessageRequest,
) (*http.Request, error) {
	bodyReader, contentType, err := c.newHTTPRequestBodyToCreateMessage(request)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare a body for the HTTP request : %w", err)
	}

	fullURL, err := url.JoinPath(c.webexURL, "/messages")
	if err != nil {
		return nil, fmt.Errorf("unable to construct the request URL : %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("unable to create a new HTTP request : %w", err)
	}

	authorizationValue := fmt.Sprintf("Bearer %s", c.authToken)
	httpRequest.Header.Add("Authorization", authorizationValue)
	httpRequest.Header.Add("Content-Type", contentType)

	return httpRequest, nil
}

func (c *client) newHTTPRequestBodyToCreateMessage(
	request CreateMessageRequest,
) (
	body io.Reader,
	contentType string,
	err error,
) {
	switch {
	case request.File != nil && request.Attachment != nil:
		return nil, "", errors.New("a message cannot contain both a file and an attachment")

	case request.File != nil:
		return c.newHTTPRequestMultipartBodyToCreateMessage(request)

	default:
		encodedBody, err := json.Marshal(request)
		if err != nil {
			return nil, "", fmt.Errorf("unable to encode the request body in JSON format : %w", err)
		}
		return bytes.NewReader(encodedBody), "application/json", nil
	}
}

func (c *client) newHTTPRequestMultipartBodyToCreateMessage(
	request CreateMessageRequest,
) (
	body io.Reader,
	contentType string,
	err error,
) {
	var bodyBuffer = new(bytes.Buffer)
	var multipartWriter = multipart.NewWriter(bodyBuffer)

	const unableToWriteField = "unable to write the '%s' field into the multipart body : %w"

	if request.RoomId != "" {
		const fieldName = "roomId"
		err = multipartWriter.WriteField(fieldName, request.RoomId)
		if err != nil {
			return nil, "", fmt.Errorf(unableToWriteField, fieldName, err)
		}
	}

	if request.ParentId != "" {
		const fieldName = "parentId"
		err = multipartWriter.WriteField(fieldName, request.ParentId)
		if err != nil {
			return nil, "", fmt.Errorf(unableToWriteField, fieldName, err)
		}
	}

	if request.ToPersonId != "" {
		const fieldName = "toPersonId"
		err = multipartWriter.WriteField(fieldName, request.ToPersonId)
		if err != nil {
			return nil, "", fmt.Errorf(unableToWriteField, fieldName, err)
		}
	}

	if request.ToPersonEmail != "" {
		const fieldName = "toPersonEmail"
		err = multipartWriter.WriteField(fieldName, request.ToPersonEmail)
		if err != nil {
			return nil, "", fmt.Errorf(unableToWriteField, fieldName, err)
		}
	}

	if request.Text != "" {
		const fieldName = "text"
		err = multipartWriter.WriteField(fieldName, request.Text)
		if err != nil {
			return nil, "", fmt.Errorf(unableToWriteField, fieldName, err)
		}
	}

	if request.Markdown != "" {
		const fieldName = "markdown"
		err = multipartWriter.WriteField(fieldName, request.Markdown)
		if err != nil {
			return nil, "", fmt.Errorf(unableToWriteField, fieldName, err)
		}
	}

	if request.Html != "" {
		const fieldName = "html"
		err = multipartWriter.WriteField(fieldName, request.Html)
		if err != nil {
			return nil, "", fmt.Errorf(unableToWriteField, fieldName, err)
		}
	}

	if request.File != nil {
		const fieldName = "files"
		fileWriter, err := multipartWriter.CreateFormFile(fieldName, request.File.Name)
		if err != nil {
			return nil, "", fmt.Errorf("unable to write the file header into the multipart body : %w", err)
		}

		_, err = io.Copy(fileWriter, request.File.Content)
		if err != nil {
			return nil, "", fmt.Errorf("unable to write the file content into the multipart body : %w", err)
		}
	}

	err = multipartWriter.Close()
	if err != nil {
		return nil, "", fmt.Errorf("unable to close the multipart writer : %w", err)
	}

	return bodyBuffer, multipartWriter.FormDataContentType(), nil
}
