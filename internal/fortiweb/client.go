// Package fortiweb provides a client for the FortiWeb REST API.
package fortiweb

import (
	"context"
	"crypto/tls"
	"encoding/base64"
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

// resourceListEnvelope is the shape of FortiWeb's "Get List" endpoints:
// {"results": [...]}. Callers only need the count, so entries are left
// undecoded.
type resourceListEnvelope struct {
	Results []json.RawMessage `json:"results"`
}

const (
	serverPolicyListPath      = "/api/v2.0/cmdb/server-policy/policy"
	contentRoutingListPath    = "/api/v2.0/cmdb/server-policy/http-content-routing-policy"
	serverPoolListPath        = "/api/v2.0/cmdb/server-policy/server-pool"
	protectionProfileListPath = "/api/v2.0/cmdb/waf/web-protection-profile.inline-protection"
	allowedHostsListPath      = "/api/v2.0/cmdb/server-policy/allow-hosts"
	virtualServerListPath     = "/api/v2.0/cmdb/server-policy/vserver"
	signatureListPath         = "/api/v2.0/cmdb/waf/signature"
	fortiGuardStatusPath      = "/api/v2.0/system/config.fortiguard"
)

const licenseExpiredDateLayout = "2006-01-02"

// LicenseStatus is a single FortiGuard license service's expiry and
// validity, as reported by GET /api/v2.0/system/config.fortiguard.
type LicenseStatus struct {
	Service string
	Expiry  time.Time
	Valid   bool
}

// FortiGuardStatus holds registration and per-service license status from
// GET /api/v2.0/system/config.fortiguard.
type FortiGuardStatus struct {
	IsRegistered bool
	Licenses     []LicenseStatus
}

// licenseServiceEntry is the shape shared by every per-service license
// object in the config.fortiguard response: {"expired": "YYYY-MM-DD",
// "is_valid": bool, ...extra fields ignored}.
type licenseServiceEntry struct {
	Expired string `json:"expired"`
	IsValid bool   `json:"is_valid"`
}

type fortiGuardResultsEnvelope struct {
	Results struct {
		Registration struct {
			IsRegistered bool `json:"is_registered"`
		} `json:"registration"`
		SecurityService           licenseServiceEntry `json:"securityService"`
		AntivirusService          licenseServiceEntry `json:"antivirusService"`
		ReputationService         licenseServiceEntry `json:"reputationService"`
		CredentialStuffingDefense licenseServiceEntry `json:"credentialStuffingDefense"`
		SbclService               licenseServiceEntry `json:"sbclService"`
		DlpSignature              licenseServiceEntry `json:"dlpSignature"`
		GeodbService              licenseServiceEntry `json:"geodbService"`
		FuzzyWebshellService      licenseServiceEntry `json:"fuzzyWebshellService"`
		ThreatAnalytics           licenseServiceEntry `json:"ThreatAnalytics"`
		AdvancedBotProtection     licenseServiceEntry `json:"AdvancedBotProtection"`
	} `json:"results"`
}

// authHeaderPayload is the JSON object FortiWeb expects to be base64-encoded
// into the Authorization header value.
type authHeaderPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
	VDOM     string `json:"vdom"`
}

// Client is an HTTP client for a single FortiWeb appliance's REST API,
// authenticating with FortiWeb's Authorization header scheme: the header
// value is the base64 encoding of an authHeaderPayload JSON object — not
// standard HTTP Basic Auth.
type Client struct {
	baseURL    string
	username   string
	password   string
	vdom       string
	httpClient *http.Client
}

// NewClient builds a Client for the FortiWeb appliance at baseURL, scoped to
// the given vdom (use "root" if the appliance doesn't use virtual domains).
// insecureSkipVerify controls whether the appliance's TLS certificate is
// verified; FortiWeb devices commonly ship with self-signed certificates.
func NewClient(baseURL, username, password, vdom string, insecureSkipVerify bool) *Client {
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
		vdom:       vdom,
		httpClient: httpClient,
	}
}

// authHeaderValue builds the value of the Authorization header FortiWeb
// expects: the base64 encoding of {"username":...,"password":...,"vdom":...}.
func (c *Client) authHeaderValue() (string, error) {
	payload, err := json.Marshal(authHeaderPayload{
		Username: c.username,
		Password: c.password,
		VDOM:     c.vdom,
	})
	if err != nil {
		return "", fmt.Errorf("fortiweb: encode auth header: %w", err)
	}
	return base64.StdEncoding.EncodeToString(payload), nil
}

// GetSystemResourceStatus fetches CPU, memory, disk, session, and connection
// metrics from the FortiWeb appliance.
func (c *Client) GetSystemResourceStatus(ctx context.Context) (*SystemResourceStatus, error) {
	resp, err := c.doGetRequest(ctx, "/api/v2.0/system/status.systemresource")
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

// GetServerPolicyCount returns the number of server-policy objects configured
// on the client's vdom.
func (c *Client) GetServerPolicyCount(ctx context.Context) (int, error) {
	return c.getResourceCount(ctx, serverPolicyListPath)
}

// GetContentRoutingCount returns the number of content-routing-policy objects
// configured on the client's vdom.
func (c *Client) GetContentRoutingCount(ctx context.Context) (int, error) {
	return c.getResourceCount(ctx, contentRoutingListPath)
}

// GetServerPoolCount returns the number of server-pool objects configured on
// the client's vdom.
func (c *Client) GetServerPoolCount(ctx context.Context) (int, error) {
	return c.getResourceCount(ctx, serverPoolListPath)
}

// GetProtectionProfileCount returns the number of protection-profile objects
// configured on the client's vdom.
func (c *Client) GetProtectionProfileCount(ctx context.Context) (int, error) {
	return c.getResourceCount(ctx, protectionProfileListPath)
}

// GetAllowedHostsCount returns the number of allowed-hosts objects configured
// on the client's vdom.
func (c *Client) GetAllowedHostsCount(ctx context.Context) (int, error) {
	return c.getResourceCount(ctx, allowedHostsListPath)
}

// GetVirtualServerCount returns the number of virtual-server objects
// configured on the client's vdom.
func (c *Client) GetVirtualServerCount(ctx context.Context) (int, error) {
	return c.getResourceCount(ctx, virtualServerListPath)
}

// GetSignatureCount returns the number of signature objects configured on
// the client's vdom.
func (c *Client) GetSignatureCount(ctx context.Context) (int, error) {
	return c.getResourceCount(ctx, signatureListPath)
}

// GetFortiGuardStatus fetches FortiGuard registration and per-service
// license expiry/validity from the FortiWeb appliance.
func (c *Client) GetFortiGuardStatus(ctx context.Context) (*FortiGuardStatus, error) {
	resp, err := c.doGetRequest(ctx, fortiGuardStatusPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fortiweb: unexpected status %d: %s", resp.StatusCode, readSnippet(resp.Body))
	}

	var envelope fortiGuardResultsEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("fortiweb: decode fortiguard status: %w", err)
	}

	entries := []struct {
		service string
		entry   licenseServiceEntry
	}{
		{"security", envelope.Results.SecurityService},
		{"antivirus", envelope.Results.AntivirusService},
		{"reputation", envelope.Results.ReputationService},
		{"credential_stuffing_defense", envelope.Results.CredentialStuffingDefense},
		{"sbcl", envelope.Results.SbclService},
		{"dlp_signature", envelope.Results.DlpSignature},
		{"geodb", envelope.Results.GeodbService},
		{"fuzzy_webshell", envelope.Results.FuzzyWebshellService},
		{"threat_analytics", envelope.Results.ThreatAnalytics},
		{"advanced_bot_protection", envelope.Results.AdvancedBotProtection},
	}

	licenses := make([]LicenseStatus, len(entries))
	for i, e := range entries {
		expiry, err := time.Parse(licenseExpiredDateLayout, e.entry.Expired)
		if err != nil {
			return nil, fmt.Errorf("fortiweb: parse %s license expiry %q: %w", e.service, e.entry.Expired, err)
		}
		licenses[i] = LicenseStatus{Service: e.service, Expiry: expiry, Valid: e.entry.IsValid}
	}

	return &FortiGuardStatus{
		IsRegistered: envelope.Results.Registration.IsRegistered,
		Licenses:     licenses,
	}, nil
}

func (c *Client) getResourceCount(ctx context.Context, path string) (int, error) {
	resp, err := c.doGetRequest(ctx, path)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("fortiweb: unexpected status %d: %s", resp.StatusCode, readSnippet(resp.Body))
	}

	var envelope resourceListEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return 0, fmt.Errorf("fortiweb: decode resource list %s: %w", path, err)
	}

	return len(envelope.Results), nil
}

func (c *Client) doGetRequest(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("fortiweb: build request: %w", err)
	}

	authHeader, err := c.authHeaderValue()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fortiweb: request %s: %w", path, err)
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
