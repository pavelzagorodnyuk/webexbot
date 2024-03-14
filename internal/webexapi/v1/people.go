package webexapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Person struct {
	Id     string   `json:"id,omitempty"`
	Emails []string `json:"emails,omitempty"`
	// TODO: implement the remaining fields
}

func (c *client) GetMyOwnDetails(ctx context.Context) (*Person, *WebexError, error) {
	httpRequest, err := c.newHTTPRequestToGetMyOwnDetails(ctx)
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

	response := new(Person)
	err = json.NewDecoder(httpResponse.Body).Decode(response)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to decode the response body : %w", err)
	}
	return response, nil, nil
}

func (c *client) newHTTPRequestToGetMyOwnDetails(ctx context.Context) (*http.Request, error) {
	fullURL, err := url.JoinPath(c.webexURL, "/people/me")
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
