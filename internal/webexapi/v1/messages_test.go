package webexapi

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pavelzagorodnyuk/webexbot/internal/testing_tools"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestClient_GetMessage(t *testing.T) {
	const authToken = "auth_token"

	testCases := []struct {
		name                   string
		messageId              string
		getMessageHTTPRequest  testing_tools.Request
		getMessageHTTPResponse testing_tools.Response
		message                *Message
		webexError             *WebexError
		errorMessage           string
	}{
		{
			name:      "OK — message is received",
			messageId: "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
			getMessageHTTPRequest: testing_tools.Request{
				Method: http.MethodGet,
				Path:   "/messages/Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
				Header: testing_tools.Header{
					"Authorization": fmt.Sprintf("Bearer %s", authToken),
					"Content-Type":  "application/json",
				},
			},
			getMessageHTTPResponse: testing_tools.Response{
				StatusCode: http.StatusOK,
				Body: []byte(`{
					"id": "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
					"parentId": "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
					"roomId": "Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv",
					"roomType": "direct",
					"text": "This message contains important information",
					"markdown": "This message contains **important** information",
					"html": "This message contains <p><strong>important</strong> information",
					"files": [
						"https://example.org/assets/logo.svg"
					],
					"personId": "Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
					"personEmail": "john.smith@example.org",
					"mentionedPeople": [
						"Y2lzY29zcGFyazovL3VzL1BFT1BMRS4eCPzf6CprcM4yzM0nBNHWynUhLTbxGaCaQ8gMxx4zc7YCIQS",
						"Y2lzY29zcGFyazovL3VzL1BFT1BMRSWy2fcisBE6wK1PdivUcsPRuMmdhxHwMvVNBXYWruTQ5Sv17TU"
					],
					"mentionedGroups": [
						"all"
					],
					"attachments": [
						{
							"contentType": "application/vnd.microsoft.card.adaptive",
							"content": {
								"type": "AdaptiveCard",
								"version": "1.0",
								"body": [
									{
										"type": "TextBlock",
										"text": "Adaptive Cards",
										"size": "large"
									}
								],
								"actions": [
									{
										"type": "Action.OpenUrl",
										"url": "https://example.org/",
										"title": "Example page"
									}
								]
							}
						}
					],
					"created": "2024-01-03T12:03:42.000Z",
					"updated": "2024-01-03T13:03:42.000Z",
					"isVoiceClip": false
				}`),
			},
			message: &Message{
				Id:       "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
				ParentId: "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
				RoomId:   "Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv",
				RoomType: RoomTypeDirect,
				Text:     "This message contains important information",
				Markdown: "This message contains **important** information",
				Html:     "This message contains <p><strong>important</strong> information",
				Files: []string{
					"https://example.org/assets/logo.svg",
				},
				PersonId:    "Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
				PersonEmail: "john.smith@example.org",
				MentionedPeople: []string{
					"Y2lzY29zcGFyazovL3VzL1BFT1BMRS4eCPzf6CprcM4yzM0nBNHWynUhLTbxGaCaQ8gMxx4zc7YCIQS",
					"Y2lzY29zcGFyazovL3VzL1BFT1BMRSWy2fcisBE6wK1PdivUcsPRuMmdhxHwMvVNBXYWruTQ5Sv17TU",
				},
				MentionedGroups: []string{
					"all",
				},
				Attachments: []Attachment{
					{
						ContentType: "application/vnd.microsoft.card.adaptive",
						Content: testing_tools.MustMinifyJSON(t, `{
							"type": "AdaptiveCard",
							"version": "1.0",
							"body": [
								{
									"type": "TextBlock",
									"text": "Adaptive Cards",
									"size": "large"
								}
							],
							"actions": [
								{
									"type": "Action.OpenUrl",
									"url": "https://example.org/",
									"title": "Example page"
								}
							]
						}`),
					},
				},
				Created:     testing_tools.Inception,
				Updated:     testing_tools.HourAfter,
				IsVoiceClip: false,
			},
		},
		{
			name:      "Error — not found",
			messageId: "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
			getMessageHTTPRequest: testing_tools.Request{
				Method: http.MethodGet,
				Path:   "/messages/Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
				Header: testing_tools.Header{
					"Authorization": fmt.Sprintf("Bearer %s", authToken),
					"Content-Type":  "application/json",
				},
			},
			getMessageHTTPResponse: testing_tools.Response{
				StatusCode: http.StatusBadRequest,
				Body: []byte(`{
					"message": "Message not found",
					"errors": [
						{
							"description": "Message not found"
						}
					],
					"trackingId": "tracking-id"
				}`),
			},
			webexError: &WebexError{
				StatusCode: http.StatusBadRequest,
				Message:    "Message not found",
				Errors: []WebexErrorDetail{
					{
						Description: "Message not found",
					},
				},
				TrackingId: "tracking-id",
			},
		},
		{
			name:         "Error — message identifier is not specified",
			messageId:    "", // is not specified
			errorMessage: "the message identifier is not specified",
		},
	}

	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				testRequest, err := testing_tools.NewRequestFrom(request)
				assert.Check(t, cmp.Nil(err))
				assert.Check(t, testing_tools.CompareRequests(testRequest, testCase.getMessageHTTPRequest))

				err = testCase.getMessageHTTPResponse.WriteTo(response)
				assert.Check(t, cmp.Nil(err))
			})

			mockWebexAPIServer := httptest.NewServer(handler)
			defer mockWebexAPIServer.Close()

			webexClient := NewClient(authToken, WithWebexURL(mockWebexAPIServer.URL))
			response, webexErr, err := webexClient.GetMessage(ctx, testCase.messageId)

			switch {
			case testCase.errorMessage != "":
				assert.Assert(t, cmp.Nil(response))
				assert.Assert(t, cmp.Nil(webexErr))
				assert.ErrorContains(t, err, testCase.errorMessage)

			case testCase.webexError != nil:
				assert.Assert(t, cmp.Nil(response))
				assert.NilError(t, err)
				assert.DeepEqual(t, testCase.webexError, webexErr)

			default:
				assert.Assert(t, cmp.Nil(webexErr))
				assert.NilError(t, err)
				assert.DeepEqual(t, testCase.message, response)
			}
		})
	}
}

func TestClient_CreateMessage(t *testing.T) {
	const authToken = "auth_token"

	testCases := []struct {
		name                      string
		createMessageRequest      CreateMessageRequest
		createMessageHTTPRequest  testing_tools.Request
		createMessageHTTPResponse testing_tools.Response
		message                   *Message
		webexError                *WebexError
		errorMessage              string
	}{
		{
			name: "OK — message with attachment is created",
			createMessageRequest: CreateMessageRequest{
				RoomId:        "Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv",
				ParentId:      "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
				ToPersonId:    "Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
				ToPersonEmail: "john.smith@example.org",
				Text:          "This message contains important information",
				Markdown:      "This message contains **important** information",
				Html:          "This message contains <p><strong>important</strong> information",
				Attachment: &Attachment{
					ContentType: "application/vnd.microsoft.card.adaptive",
					Content: testing_tools.MustMinifyJSON(t, `{
						"type": "AdaptiveCard",
						"version": "1.0",
						"body": [
							{
								"type": "TextBlock",
								"text": "Adaptive Cards",
								"size": "large"
							}
						],
						"actions": [
							{
								"type": "Action.OpenUrl",
								"url": "https://example.org/",
								"title": "Example page"
							}
						]
					}`),
				},
			},
			createMessageHTTPRequest: testing_tools.Request{
				Method: http.MethodPost,
				Path:   "/messages",
				Header: testing_tools.Header{
					"Authorization": fmt.Sprintf("Bearer %s", authToken),
					"Content-Type":  "application/json",
				},
				Body: []byte(`{
					"roomId": "Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv",
					"parentId": "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
					"toPersonId": "Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
					"toPersonEmail": "john.smith@example.org",
					"text": "This message contains important information",
					"markdown": "This message contains **important** information",
					"html": "This message contains <p><strong>important</strong> information",
					"attachments": {
						"contentType": "application/vnd.microsoft.card.adaptive",
						"content": {
							"type": "AdaptiveCard",
							"version": "1.0",
							"body": [
								{
									"type": "TextBlock",
									"text": "Adaptive Cards",
									"size": "large"
								}
							],
							"actions": [
								{
									"type": "Action.OpenUrl",
									"url": "https://example.org/",
									"title": "Example page"
								}
							]
						}
					}
				}`),
			},
			createMessageHTTPResponse: testing_tools.Response{
				StatusCode: http.StatusOK,
				Body: []byte(`{
					"id": "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
					"parentId": "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
					"roomId": "Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv",
					"toPersonId": "Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
					"toPersonEmail": "person@example.org",
					"roomType": "direct",
					"text": "This message contains important information",
					"markdown": "This message contains **important** information",
					"html": "This message contains <p><strong>important</strong> information",
					"personId": "Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
					"personEmail": "john.smith@example.org",
					"mentionedPeople": [
						"Y2lzY29zcGFyazovL3VzL1BFT1BMRS4eCPzf6CprcM4yzM0nBNHWynUhLTbxGaCaQ8gMxx4zc7YCIQS",
						"Y2lzY29zcGFyazovL3VzL1BFT1BMRSWy2fcisBE6wK1PdivUcsPRuMmdhxHwMvVNBXYWruTQ5Sv17TU"
					],
					"mentionedGroups": [
						"all"
					],
					"attachments": [
						{
							"contentType": "application/vnd.microsoft.card.adaptive",
							"content": {
								"type": "AdaptiveCard",
								"version": "1.0",
								"body": [
									{
										"type": "TextBlock",
										"text": "Adaptive Cards",
										"size": "large"
									}
								],
								"actions": [
									{
										"type": "Action.OpenUrl",
										"url": "https://example.org/",
										"title": "Example page"
									}
								]
							}
						}
					],
					"created": "2024-01-03T12:03:42+00:00",
					"updated": "2024-01-03T13:03:42+00:00",
					"isVoiceClip": false
				}`),
			},
			message: &Message{
				Id:          "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
				ParentId:    "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
				RoomId:      "Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv",
				RoomType:    RoomTypeDirect,
				Text:        "This message contains important information",
				Markdown:    "This message contains **important** information",
				Html:        "This message contains <p><strong>important</strong> information",
				PersonId:    "Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
				PersonEmail: "john.smith@example.org",
				MentionedPeople: []string{
					"Y2lzY29zcGFyazovL3VzL1BFT1BMRS4eCPzf6CprcM4yzM0nBNHWynUhLTbxGaCaQ8gMxx4zc7YCIQS",
					"Y2lzY29zcGFyazovL3VzL1BFT1BMRSWy2fcisBE6wK1PdivUcsPRuMmdhxHwMvVNBXYWruTQ5Sv17TU",
				},
				MentionedGroups: []string{
					"all",
				},
				Attachments: []Attachment{
					{
						ContentType: "application/vnd.microsoft.card.adaptive",
						Content: testing_tools.MustMinifyJSON(t, `{
							"type": "AdaptiveCard",
							"version": "1.0",
							"body": [
								{
									"type": "TextBlock",
									"text": "Adaptive Cards",
									"size": "large"
								}
							],
							"actions": [
								{
									"type": "Action.OpenUrl",
									"url": "https://example.org/",
									"title": "Example page"
								}
							]
						}`),
					},
				},
				Created:     testing_tools.Inception,
				Updated:     testing_tools.HourAfter,
				IsVoiceClip: false,
			},
		},
		{
			name: "OK — message with file is created",
			createMessageRequest: CreateMessageRequest{
				RoomId:        "Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv",
				ParentId:      "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
				ToPersonId:    "Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
				ToPersonEmail: "john.smith@example.org",
				Text:          "This message contains important information",
				Markdown:      "This message contains **important** information",
				Html:          "This message contains <p><strong>important</strong> information",
				File: &File{
					Name:    "attachment.txt",
					Content: bytes.NewBufferString("file_content"),
				},
			},
			createMessageHTTPRequest: testing_tools.Request{
				Method: http.MethodPost,
				Path:   "/messages",
				Header: testing_tools.Header{
					"Authorization": fmt.Sprintf("Bearer %s", authToken),
					"Content-Type":  "multipart/form-data; boundary=xxxMultipartBoundaryxxx",
				},
				Body: []byte(`--xxxMultipartBoundaryxxx
Content-Disposition: form-data; name="roomId"

Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv
--xxxMultipartBoundaryxxx
Content-Disposition: form-data; name="parentId"

Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH
--xxxMultipartBoundaryxxx
Content-Disposition: form-data; name="toPersonId"

Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd
--xxxMultipartBoundaryxxx
Content-Disposition: form-data; name="toPersonEmail"

john.smith@example.org
--xxxMultipartBoundaryxxx
Content-Disposition: form-data; name="text"

This message contains important information
--xxxMultipartBoundaryxxx
Content-Disposition: form-data; name="markdown"

This message contains **important** information
--xxxMultipartBoundaryxxx
Content-Disposition: form-data; name="html"

This message contains <p><strong>important</strong> information
--xxxMultipartBoundaryxxx
Content-Disposition: form-data; name="files"; filename="attachment.txt"
Content-Type: application/octet-stream

file_content
--xxxMultipartBoundaryxxx--
`),
			},
			createMessageHTTPResponse: testing_tools.Response{
				StatusCode: http.StatusOK,
				Body: []byte(`{
					"id": "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
					"parentId": "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
					"roomId": "Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv",
					"roomType": "direct",
					"text": "This message contains important information",
					"markdown": "This message contains **important** information",
					"html": "This message contains <p><strong>important</strong> information",
					"files": [
						"https://webexapis.com/v1/contents/attachment.txt"
					],
					"personId": "Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
					"personEmail": "john.smith@example.org",
					"mentionedPeople": [
						"Y2lzY29zcGFyazovL3VzL1BFT1BMRS4eCPzf6CprcM4yzM0nBNHWynUhLTbxGaCaQ8gMxx4zc7YCIQS",
						"Y2lzY29zcGFyazovL3VzL1BFT1BMRSWy2fcisBE6wK1PdivUcsPRuMmdhxHwMvVNBXYWruTQ5Sv17TU"
					],
					"mentionedGroups": [
						"all"
					],
					"created": "2024-01-03T12:03:42.000Z",
					"updated": "2024-01-03T13:03:42.000Z",
					"isVoiceClip": false
				}`),
			},
			message: &Message{
				Id:       "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
				ParentId: "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
				RoomId:   "Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv",
				RoomType: RoomTypeDirect,
				Text:     "This message contains important information",
				Markdown: "This message contains **important** information",
				Html:     "This message contains <p><strong>important</strong> information",
				Files: []string{
					"https://webexapis.com/v1/contents/attachment.txt",
				},
				PersonId:    "Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
				PersonEmail: "john.smith@example.org",
				MentionedPeople: []string{
					"Y2lzY29zcGFyazovL3VzL1BFT1BMRS4eCPzf6CprcM4yzM0nBNHWynUhLTbxGaCaQ8gMxx4zc7YCIQS",
					"Y2lzY29zcGFyazovL3VzL1BFT1BMRSWy2fcisBE6wK1PdivUcsPRuMmdhxHwMvVNBXYWruTQ5Sv17TU",
				},
				MentionedGroups: []string{
					"all",
				},
				Created:     testing_tools.Inception,
				Updated:     testing_tools.HourAfter,
				IsVoiceClip: false,
			},
		},
		{
			name: "Error — unable to create message",
			createMessageRequest: CreateMessageRequest{
				ToPersonId: "Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
				Text:       "This message contains important information",
			},
			createMessageHTTPRequest: testing_tools.Request{
				Method: http.MethodPost,
				Path:   "/messages",
				Header: testing_tools.Header{
					"Authorization": fmt.Sprintf("Bearer %s", authToken),
					"Content-Type":  "application/json",
				},
				Body: []byte(`{
					"toPersonId": "Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
					"text": "This message contains important information"
				}`),
			},
			createMessageHTTPResponse: testing_tools.Response{
				StatusCode: http.StatusBadRequest,
				Body: []byte(`{
					"message": "Unable to create message",
					"errors": [
						{
							"description": "Unable to create message"
						}
					],
					"trackingId": "tracking-id"
				}`),
			},
			webexError: &WebexError{
				StatusCode: http.StatusBadRequest,
				Message:    "Unable to create message",
				Errors: []WebexErrorDetail{
					{
						Description: "Unable to create message",
					},
				},
				TrackingId: "tracking-id",
			},
		},
		{
			name: "Error — message with file and attachment",
			createMessageRequest: CreateMessageRequest{
				File:       &File{},
				Attachment: &Attachment{},
			},
			errorMessage: "a message cannot contain both a file and an attachment",
		},
	}

	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			webhookCreationHandler := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				testRequest, err := testing_tools.NewRequestFrom(request)
				assert.Check(t, cmp.Nil(err))
				assert.Check(t, testing_tools.CompareRequests(testCase.createMessageHTTPRequest, testRequest))

				err = testCase.createMessageHTTPResponse.WriteTo(response)
				assert.Check(t, cmp.Nil(err))
			})

			mockWebexAPIServer := httptest.NewServer(webhookCreationHandler)
			defer mockWebexAPIServer.Close()

			webexClient := NewClient(authToken, WithWebexURL(mockWebexAPIServer.URL))
			response, webexErr, err := webexClient.CreateMessage(ctx, testCase.createMessageRequest)

			switch {
			case testCase.errorMessage != "":
				assert.Assert(t, cmp.Nil(response))
				assert.Assert(t, cmp.Nil(webexErr))
				assert.ErrorContains(t, err, testCase.errorMessage)

			case testCase.webexError != nil:
				assert.Assert(t, cmp.Nil(response))
				assert.NilError(t, err)
				assert.DeepEqual(t, testCase.webexError, webexErr)

			default:
				assert.Assert(t, cmp.Nil(webexErr))
				assert.NilError(t, err)
				assert.DeepEqual(t, testCase.message, response)
			}
		})
	}
}
