package main

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

// CallAPI sends an HTTP request to the specified URL with the given method and body.
// It returns the response body as a byte slice and any error encountered.
func CallAPI(method, url string, body []byte, headers map[string]string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}