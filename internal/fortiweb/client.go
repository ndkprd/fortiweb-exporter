// Package fortiweb provides a client for the FortiWeb REST API.
package fortiweb

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const requestTimeout = 10 * time.Second

// SystemResourceStatus holds the fields returned by
// GET /api/v2.0/system/status.systemresource.
type SystemResourceStatus struct {
	CPU           float64 `json:"cpu"`
	Mem           float64 `json:"mem"`
	DiskUsage     float64 `json:"diskUsage"`
	SessionCount  int64   `json:"sessionCount"`
	ConnCntPerSec int64   `json:"connCntPerSec"`
	LogDisk       string  `json:"logDisk"`
	DBStatus      string  `json:"dbStatus"`
}

type systemResourceStatusEnvelope struct {
	Results SystemResourceStatus `json:"results"`
}

// Client is an HTTP client for a single FortiWeb appliance's REST API,
// authenticating with HTTP Basic Auth.
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// NewClient builds a Client for the FortiWeb appliance at baseURL.
// insecureSkipVerify controls whether the appliance's TLS certificate is
// verified; FortiWeb devices commonly ship with self-signed certificates.
func NewClient(baseURL, username, password string, insecureSkipVerify bool) *Client {
	httpClient := &http.Client{Timeout: requestTimeout}
	if insecureSkipVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		username:   username,
		password:   password,
		httpClient: httpClient,
	}
}

// GetSystemResourceStatus fetches CPU, memory, disk, session, and connection
// metrics from the FortiWeb appliance.
func (c *Client) GetSystemResourceStatus(ctx context.Context) (*SystemResourceStatus, error) {
	resp, err := c.doStatusRequest(ctx)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fortiweb: unexpected status %d: %s", resp.StatusCode, readSnippet(resp.Body))
	}

	var envelope systemResourceStatusEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("fortiweb: decode system resource status: %w", err)
	}

	return &envelope.Results, nil
}

func (c *Client) doStatusRequest(ctx context.Context) (*http.Response, error) {
	url := c.baseURL + "/api/v2.0/system/status.systemresource"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("fortiweb: build request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fortiweb: request system resource status: %w", err)
	}

	return resp, nil
}

// readSnippet returns a bounded prefix of r's contents for use in error
// messages, without risking an unbounded read of an oversized response body.
func readSnippet(r io.Reader) string {
	const maxSnippetBytes = 4096

	body, err := io.ReadAll(io.LimitReader(r, maxSnippetBytes))
	if err != nil {
		return "<unreadable response body>"
	}
	return string(body)
}
