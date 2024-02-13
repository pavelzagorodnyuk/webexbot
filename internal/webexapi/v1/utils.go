package webexapi

import "net/http"

// isSuccessful checks if the HTTP response is successful
func isSuccessful(response *http.Response) bool {
	return response.StatusCode > 199 && response.StatusCode < 300
}
