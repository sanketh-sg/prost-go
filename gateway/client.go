package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// HTTPClient wraps HTTP operations for calling downstream services
type HTTPClient struct {
    client *http.Client
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient() *HTTPClient {
    return &HTTPClient{
        client: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

// Request makes HTTP request to downstream service
func (hc *HTTPClient) Request(ctx context.Context, method, url string, headers map[string]string, body interface{}) ([]byte, error) {
    var bodyReader io.Reader

    if body != nil {
        bodyBytes, err := json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal body: %w", err)
        }
        bodyReader = bytes.NewReader(bodyBytes)
    }

    req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    // Add headers
    req.Header.Set("Content-Type", "application/json")
    for k, v := range headers {
        req.Header.Set(k, v)
    }

    resp, err := hc.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, fmt.Errorf("service returned status %d: %s", resp.StatusCode, string(respBody))
    }

    return respBody, nil
}

// GET makes GET request
func (hc *HTTPClient) GET(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
    return hc.Request(ctx, http.MethodGet, url, headers, nil)
}

// POST makes POST request
func (hc *HTTPClient) POST(ctx context.Context, url string, headers map[string]string, body interface{}) ([]byte, error) {
    return hc.Request(ctx, http.MethodPost, url, headers, body)
}

// PUT makes PUT request
func (hc *HTTPClient) PUT(ctx context.Context, url string, headers map[string]string, body interface{}) ([]byte, error) {
    return hc.Request(ctx, http.MethodPut, url, headers, body)
}

// DELETE makes DELETE request
func (hc *HTTPClient) DELETE(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
    return hc.Request(ctx, http.MethodDelete, url, headers, nil)
}