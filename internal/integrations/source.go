package integrations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const maxResponseBytes = 5 << 20

// Errors returned here never include the request URL or headers, because a
// source URL may carry credentials in its query string and these errors reach
// API clients and logs.
var errRequestFailed = errors.New("source request failed")

type HTTP struct {
	URL    string
	APIKey string
	Client *http.Client
}

func NewHTTP(url, apiKey string) *HTTP {
	return &HTTP{
		URL:    url,
		APIKey: apiKey,
		Client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (h *HTTP) Records(ctx context.Context) ([]json.RawMessage, error) {
	if h.URL == "" {
		return nil, errors.New("source URL is not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.URL, nil)
	if err != nil {
		return nil, errRequestFailed
	}
	if h.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.APIKey)
	}
	client := h.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errRequestFailed
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("source returned status %d", resp.StatusCode)
	}
	var records []json.RawMessage
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(&records); err != nil {
		return nil, errors.New("source returned an invalid response")
	}
	return records, nil
}
