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

func TestClient_GetMyOwnDetails(t *testing.T) {
	const authToken = "auth_token"

	testCases := []struct {
		name                        string
		getMyOwnDetailsHTTPRequest  testing_tools.Request
		getMyOwnDetailsHTTPResponse testing_tools.Response
		person                      *Person
		webexError                  *WebexError
	}{
		{
			name: "OK — details are received",
			getMyOwnDetailsHTTPRequest: testing_tools.Request{
				Method: http.MethodGet,
				Path:   "/people/me",
				Header: testing_tools.Header{
					"Authorization": fmt.Sprintf("Bearer %s", authToken),
					"Content-Type":  "application/json",
				},
			},
			getMyOwnDetailsHTTPResponse: testing_tools.Response{
				StatusCode: http.StatusOK,
				Body: []byte(`{
					"id": "Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
					"emails": [
					  "john.smith@example.org"
					]
				}`),
			},
			person: &Person{
				Id: "Y2lzY29zcGFyazovL3VzL1BFT1BMRSawpYVDJ7hv3F7126S7dvqBsSDoH4XqOZM4OyCotKJtgDvhAdd",
				Emails: []string{
					"john.smith@example.org",
				},
			},
		},
		{
			name: "Error — internal server error",
			getMyOwnDetailsHTTPRequest: testing_tools.Request{
				Method: http.MethodGet,
				Path:   "/people/me",
				Header: testing_tools.Header{
					"Authorization": fmt.Sprintf("Bearer %s", authToken),
					"Content-Type":  "application/json",
				},
			},
			getMyOwnDetailsHTTPResponse: testing_tools.Response{
				StatusCode: http.StatusInternalServerError,
				Body: []byte(`{
					"message": "Internal server error",
					"errors": [
						{
							"description": "Internal server error"
						}
					],
					"trackingId": "tracking-id"
				}`),
			},
			webexError: &WebexError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Internal server error",
				Errors: []WebexErrorDetail{
					{
						Description: "Internal server error",
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
				assert.Check(t, testing_tools.CompareRequests(testRequest, testCase.getMyOwnDetailsHTTPRequest))

				err = testCase.getMyOwnDetailsHTTPResponse.WriteTo(response)
				assert.Check(t, cmp.Nil(err))
			})

			mockWebexAPIServer := httptest.NewServer(handler)
			defer mockWebexAPIServer.Close()

			webexClient := NewClient(authToken, WithWebexURL(mockWebexAPIServer.URL))
			response, webexErr, err := webexClient.GetMyOwnDetails(ctx)
			assert.NilError(t, err)

			switch {
			case testCase.webexError != nil:
				assert.Assert(t, cmp.Nil(response))
				assert.DeepEqual(t, testCase.webexError, webexErr)

			default:
				assert.Assert(t, cmp.Nil(webexErr))
				assert.DeepEqual(t, testCase.person, response)
			}
		})
	}
}
