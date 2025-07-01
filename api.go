package main

import (
	"bytes"
	"compress/gzip"
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

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gzReader.Close()
		reader = gzReader
	}

	respBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}