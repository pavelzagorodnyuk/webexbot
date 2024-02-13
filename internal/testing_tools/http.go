package testing_tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"gotest.tools/v3/assert/cmp"
)

// Request represents an HTTP request intended for testing
type Request struct {
	Method string
	Path   string
	Header Header
	Body   []byte
}

// NewRequestFrom creates a new Request from the http.Request
func NewRequestFrom(request *http.Request) (Request, error) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return Request{}, fmt.Errorf("unable to read the request body : %w", err)
	}

	return Request{
		Method: request.Method,
		Path:   request.RequestURI,
		Header: newHeaderFrom(request.Header),
		Body:   body,
	}, nil
}

// CompareRequests compares two requests and succeeds if they are equal
func CompareRequests(requestX, requestY Request) cmp.Comparison {
	return func() cmp.Result {
		var comparisons []cmp.Comparison

		methodComparison := cmp.Equal(requestX.Method, requestY.Method)
		comparisons = append(comparisons, methodComparison)

		pathComparison := cmp.Equal(requestX.Path, requestY.Path)
		comparisons = append(comparisons, pathComparison)

		headerComparison := compareHeaders(requestX.Header, requestY.Header)
		comparisons = append(comparisons, headerComparison)

		bodyComparison := compareBodies(requestX.Body, requestY.Body)
		comparisons = append(comparisons, bodyComparison)

		return executeComparisons(comparisons)
	}
}

func compareHeaders(headerX, headerY Header) cmp.Comparison {
	return func() cmp.Result {
		// TODO: implement me
		return cmp.ResultSuccess
	}
}

func compareBodies(bodyX, bodyY []byte) cmp.Comparison {
	return func() cmp.Result {
		minifiedBodyX, err := MinifyJSON(bodyX)
		if err != nil {
			return cmp.ResultFromError(err)
		}

		minifiedBodyY, err := MinifyJSON(bodyY)
		if err != nil {
			return cmp.ResultFromError(err)
		}

		return cmp.DeepEqual(minifiedBodyX, minifiedBodyY)()
	}
}

func executeComparisons(comparisons []cmp.Comparison) cmp.Result {
	for _, compariseFunc := range comparisons {
		result := compariseFunc()
		if !result.Success() {
			return result
		}
	}
	return cmp.ResultSuccess
}

// MinifyJSON removes spaces from the JSON data and escapes some characters
func MinifyJSON(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return []byte{}, nil
	}

	bufferForCompression := new(bytes.Buffer)
	err := json.Compact(bufferForCompression, src)
	if err != nil {
		return nil, fmt.Errorf("unable to compact JSON : %w", err)
	}

	bufferForEscape := new(bytes.Buffer)
	json.HTMLEscape(bufferForEscape, bufferForCompression.Bytes())

	return bufferForEscape.Bytes(), nil
}

// MustMinifyJSON acts like MinifyJSON, but accepts JSON data in string format and calls t.Fatal() if any error
// occurred during minimization
func MustMinifyJSON(t *testing.T, src string) []byte {
	minifiedJSON, err := MinifyJSON([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	return minifiedJSON
}

// Header represents an HTTP header intended for testing
type Header map[string]string

// newHeaderFrom creates a new Header from the http.Header
func newHeaderFrom(httpHeader http.Header) Header {
	header := make(Header)
	for key := range httpHeader {
		value := httpHeader.Get(key)
		header[key] = value
	}
	return header
}

// Response represents an HTTP response intended for testing
type Response struct {
	StatusCode int
	Body       []byte
}

// WriteTo writes the HTTP response to the response writer
func (r *Response) WriteTo(responseWriter http.ResponseWriter) error {
	responseWriter.WriteHeader(r.StatusCode)

	minifiedBody, err := MinifyJSON(r.Body)
	if err != nil {
		return err
	}
	_, err = responseWriter.Write(minifiedBody)
	return err
}
