package webexapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pavelzagorodnyuk/webexbot/internal/testing_tools"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestClient_GetAttachmentAction(t *testing.T) {
	const authToken = "auth_token"

	testCases := []struct {
		name                            string
		attachmentActionId              string
		getAttachmentActionHTTPRequest  testing_tools.Request
		getAttachmentActionHTTPResponse testing_tools.Response
		attachmentAction                *AttachmentAction
		webexError                      *WebexError
		errorMessage                    string
	}{
		{
			name:               "OK — attachment action is received",
			attachmentActionId: "Y2lzY29zcGFyazovL3VzL0FUVEFDSE1FTlRfQUNUSU9OLzhiZWMxYTMwLWM5MTAtMTFlgS1iZmExLWJmMmMwNGIzY2ItOQ",
			getAttachmentActionHTTPRequest: testing_tools.Request{
				Method: http.MethodGet,
				Path:   "/attachment/actions/Y2lzY29zcGFyazovL3VzL0FUVEFDSE1FTlRfQUNUSU9OLzhiZWMxYTMwLWM5MTAtMTFlgS1iZmExLWJmMmMwNGIzY2ItOQ",
				Header: testing_tools.Header{
					"Authorization": fmt.Sprintf("Bearer %s", authToken),
					"Content-Type":  "application/json",
				},
			},
			getAttachmentActionHTTPResponse: testing_tools.Response{
				StatusCode: http.StatusOK,
				Body: []byte(`{
					"id":"Y2lzY29zcGFyazovL3VzL0FUVEFDSE1FTlRfQUNUSU9OLzhiZWMxYTMwLWM5MTAtMTFlgS1iZmExLWJmMmMwNGIzY2ItOQ",
					"type":"submit",
					"messageId":"Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
					"inputs":{
						"Name": "John Smith",
    					"Email": "john.smith@example.org"
					},
					"personId":"Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
					"roomId":"Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv",
					"created":"2024-01-03T12:03:42.000Z"
				}`),
			},
			attachmentAction: &AttachmentAction{
				Id:        "Y2lzY29zcGFyazovL3VzL0FUVEFDSE1FTlRfQUNUSU9OLzhiZWMxYTMwLWM5MTAtMTFlgS1iZmExLWJmMmMwNGIzY2ItOQ",
				PersonId:  "Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
				RoomId:    "Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv",
				Type:      AttachmentActionTypeSubmit,
				MessageId: "Y2lzY29zcGFyazovL3VzL01FU1NBR0UvQiUtNcR1vgzLrZshe5A3N98XAJSB3wDQFnajqSiO9KqlAKtH",
				Inputs: testing_tools.MustMinifyJSON(t, `{
					"Name": "John Smith",
					"Email": "john.smith@example.org"
				}`),
				Created: testing_tools.Inception,
			},
		},
		{
			name:               "Error — not found",
			attachmentActionId: "Y2lzY29zcGFyazovL3VzL0FUVEFDSE1FTlRfQUNUSU9OLzhiZWMxYTMwLWM5MTAtMTFlgS1iZmExLWJmMmMwNGIzY2ItOQ",
			getAttachmentActionHTTPRequest: testing_tools.Request{
				Method: http.MethodGet,
				Path:   "/attachment/actions/Y2lzY29zcGFyazovL3VzL0FUVEFDSE1FTlRfQUNUSU9OLzhiZWMxYTMwLWM5MTAtMTFlgS1iZmExLWJmMmMwNGIzY2ItOQ",
				Header: testing_tools.Header{
					"Authorization": fmt.Sprintf("Bearer %s", authToken),
					"Content-Type":  "application/json",
				},
			},
			getAttachmentActionHTTPResponse: testing_tools.Response{
				StatusCode: http.StatusBadRequest,
				Body: []byte(`{
					"message": "Attachment action not found",
					"errors": [
						{
							"description": "Attachment action not found"
						}
					],
					"trackingId": "tracking-id"
				}`),
			},
			webexError: &WebexError{
				StatusCode: http.StatusBadRequest,
				Message:    "Attachment action not found",
				Errors: []WebexErrorDetail{
					{
						Description: "Attachment action not found",
					},
				},
				TrackingId: "tracking-id",
			},
		},
		{
			name:               "Error — attachment action identifier is not specified",
			attachmentActionId: "", // is not specified
			errorMessage:       "the attachment action identifier is not specified",
		},
	}

	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				testRequest, err := testing_tools.NewRequestFrom(request)
				assert.Check(t, cmp.Nil(err))
				assert.Check(t, testing_tools.CompareRequests(testRequest, testCase.getAttachmentActionHTTPRequest))

				err = testCase.getAttachmentActionHTTPResponse.WriteTo(response)
				assert.Check(t, cmp.Nil(err))
			})

			mockWebexAPIServer := httptest.NewServer(handler)
			defer mockWebexAPIServer.Close()

			webexClient := NewClient(authToken, WithWebexURL(mockWebexAPIServer.URL))
			response, webexErr, err := webexClient.GetAttachmentAction(ctx, testCase.attachmentActionId)

			switch {
			case testCase.errorMessage != "":
				assert.Assert(t, response == nil)
				assert.Assert(t, webexErr == nil)
				assert.ErrorContains(t, err, testCase.errorMessage)

			case testCase.webexError != nil:
				assert.Assert(t, response == nil)
				assert.Assert(t, err == nil)
				assert.DeepEqual(t, testCase.webexError, webexErr)

			default:
				assert.Assert(t, webexErr == nil)
				assert.Assert(t, err == nil)
				assert.DeepEqual(t, testCase.attachmentAction, response)
			}
		})
	}
}
