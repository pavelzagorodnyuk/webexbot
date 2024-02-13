package webexapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// WebexError represents an error from the Webex API
type WebexError struct {
	StatusCode int
	Message    string
	Errors     []WebexErrorDetail
	TrackingId string
}

// WebexErrorDetail is a part of the WebexError which contains more detailed information
type WebexErrorDetail struct {
	Description string
}

// Error returns a string representation of the error
func (e *WebexError) Error() string {
	return fmt.Sprintf("status code: %d, message: '%s', tracking ID: '%s'", e.StatusCode, e.Message, e.TrackingId)
}

// fetchWebexErrorFrom fetches a Webex API error from the HTTP response. Please note that this function does not close
// the response body after reading.
func fetchWebexErrorFrom(response *http.Response) (*WebexError, error) {
	webexError := new(WebexError)
	err := json.NewDecoder(response.Body).Decode(webexError)
	if err != nil {
		return nil, fmt.Errorf("unable to decode the response body : %w", err)
	}

	webexError.StatusCode = response.StatusCode
	return webexError, nil
}
