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

func TestClient_CreateWebhook(t *testing.T) {
	const authToken = "auth_token"

	testCases := []struct {
		name                      string
		createWebhookRequest      CreateWebhookRequest
		createWebhookHTTPRequest  testing_tools.Request
		createWebhookHTTPResponse testing_tools.Response
		createWebhookResponse     *Webhook
		webexError                *WebexError
	}{
		{
			name: "OK — webhook is created",
			createWebhookRequest: CreateWebhookRequest{
				Name:      "Message webhook",
				TargetUrl: "https://example.org/webhooks",
				Resource:  Messages,
				Event:     ResourceCreated,
				Filter:    "roomId=Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv",
				Secret:    "7SpNFV2s2vPBLjajG8Jvl1qweZKgM7",
				OwnedBy:   "org",
			},
			createWebhookHTTPRequest: testing_tools.Request{
				Method: http.MethodPost,
				Path:   "/webhooks",
				Header: testing_tools.Header{
					"Authorization": fmt.Sprintf("Bearer %s", authToken),
					"Content-Type":  "application/json",
				},
				Body: []byte(`{
					"name": "Message webhook",
					"targetUrl": "https://example.org/webhooks",
					"resource": "messages",
					"event": "created",
					"filter": "roomId=Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv",
					"secret": "7SpNFV2s2vPBLjajG8Jvl1qweZKgM7",
					"ownedBy": "org"
				}`),
			},
			createWebhookHTTPResponse: testing_tools.Response{
				StatusCode: http.StatusOK,
				Body: []byte(`{
					"id": "Y2lzY29zcGFyazovL3VzL1dFQkhPT0svMGQwYWIyN2ItNmJmZS00Y3NiLWI5YWIhMGIxMGU3YTYzN2Fj",
					"name": "Message webhook",
					"targetUrl": "https://example.org/webhooks",
					"resource": "messages",
					"event": "created",
					"filter": "roomId=Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv",
					"secret": "7SpNFV2s2vPBLjajG8Jvl1qweZKgM7",
					"status": "active",
					"created": "2024-01-03T12:03:42.000Z",
					"ownedBy": "org"
				}`),
			},
			createWebhookResponse: &Webhook{
				Id:        "Y2lzY29zcGFyazovL3VzL1dFQkhPT0svMGQwYWIyN2ItNmJmZS00Y3NiLWI5YWIhMGIxMGU3YTYzN2Fj",
				Name:      "Message webhook",
				TargetUrl: "https://example.org/webhooks",
				Resource:  Messages,
				Event:     ResourceCreated,
				Filter:    "roomId=Y2lzY29zcGFyazovL3VzL1JPT00vYmJjR1ITrDCPtsLaXSZhPgvx6sntZDoDpEn3FWRZqlHMGjNv",
				Secret:    "7SpNFV2s2vPBLjajG8Jvl1qweZKgM7",
				Status:    WebhookStatusActive,
				Created:   testing_tools.Inception,
				OwnedBy:   "org",
			},
		},
		{
			name: "Error — unable to create webhook",
			createWebhookRequest: CreateWebhookRequest{
				Name:      "Message webhook",
				TargetUrl: "https://example.org/webhooks",
				Resource:  Messages,
				Event:     ResourceCreated,
			},
			createWebhookHTTPRequest: testing_tools.Request{
				Method: http.MethodPost,
				Path:   "/webhooks",
				Header: testing_tools.Header{
					"Authorization": fmt.Sprintf("Bearer %s", authToken),
					"Content-Type":  "application/json",
				},
				Body: []byte(`{
					"name": "Message webhook",
					"targetUrl": "https://example.org/webhooks",
					"resource": "messages",
					"event": "created"
				}`),
			},
			createWebhookHTTPResponse: testing_tools.Response{
				StatusCode: http.StatusBadRequest,
				Body: []byte(`{
					"message": "Unable to create webhook",
					"errors": [
						{
							"description": "Unable to create webhook"
						}
					],
					"trackingId": "tracking-id"
				}`),
			},
			webexError: &WebexError{
				StatusCode: http.StatusBadRequest,
				Message:    "Unable to create webhook",
				Errors: []WebexErrorDetail{
					{
						Description: "Unable to create webhook",
					},
				},
				TrackingId: "tracking-id",
			},
		},
	}

	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				testRequest, err := testing_tools.NewRequestFrom(request)
				assert.Check(t, cmp.Nil(err))
				assert.Check(t, testing_tools.CompareRequests(testRequest, testCase.createWebhookHTTPRequest))

				err = testCase.createWebhookHTTPResponse.WriteTo(response)
				assert.Check(t, cmp.Nil(err))
			})

			mockWebexAPIServer := httptest.NewServer(handler)
			defer mockWebexAPIServer.Close()

			webexClient := NewClient(authToken, WithWebexURL(mockWebexAPIServer.URL))
			response, webexErr, err := webexClient.CreateWebhook(ctx, testCase.createWebhookRequest)
			assert.NilError(t, err)

			switch {
			case testCase.webexError != nil:
				assert.Assert(t, cmp.Nil(response))
				assert.DeepEqual(t, testCase.webexError, webexErr)

			default:
				assert.Assert(t, cmp.Nil(webexErr))
				assert.DeepEqual(t, testCase.createWebhookResponse, response)
			}
		})
	}
}
