package fortiweb

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const validStatusResponse = `{
  "results": {
    "cpu": 17,
    "mem": 78,
    "logDisk": "Available",
    "dbStatus": "Available",
    "diskUsage": 79,
    "sessionCount": 21051,
    "connCntPerSec": 482
  }
}`

func TestGetSystemResourceStatus_AuthHeaderAndPath(t *testing.T) {
	const wantUsername, wantPassword, wantVDOM = "admin", "changeme", "root"

	var gotMethod, gotPath, gotAuthHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuthHeader = r.Header.Get("Authorization")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(validStatusResponse))
	}))
	defer server.Close()

	client := NewClient(server.URL, wantUsername, wantPassword, wantVDOM, false)
	if _, err := client.GetSystemResourceStatus(context.Background()); err != nil {
		t.Fatalf("GetSystemResourceStatus() returned unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("request method = %q, want %q", gotMethod, http.MethodGet)
	}
	const wantPath = "/api/v2.0/system/status.systemresource"
	if gotPath != wantPath {
		t.Errorf("request path = %q, want %q", gotPath, wantPath)
	}

	decoded, err := base64.StdEncoding.DecodeString(gotAuthHeader)
	if err != nil {
		t.Fatalf("Authorization header %q is not valid base64: %v", gotAuthHeader, err)
	}

	var gotPayload authHeaderPayload
	if err := json.Unmarshal(decoded, &gotPayload); err != nil {
		t.Fatalf("decoded Authorization %q is not valid JSON: %v", decoded, err)
	}
	wantPayload := authHeaderPayload{Username: wantUsername, Password: wantPassword, VDOM: wantVDOM}
	if gotPayload != wantPayload {
		t.Errorf("decoded Authorization payload = %+v, want %+v", gotPayload, wantPayload)
	}
}

func TestGetSystemResourceStatus_ParsesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(validStatusResponse))
	}))
	defer server.Close()

	client := NewClient(server.URL, "admin", "changeme", "root", false)
	status, err := client.GetSystemResourceStatus(context.Background())
	if err != nil {
		t.Fatalf("GetSystemResourceStatus() returned unexpected error: %v", err)
	}

	want := SystemResourceStatus{
		CPU:           17,
		Mem:           78,
		DiskUsage:     79,
		SessionCount:  21051,
		ConnCntPerSec: 482,
		LogDisk:       "Available",
		DBStatus:      "Available",
	}
	if *status != want {
		t.Errorf("GetSystemResourceStatus() = %+v, want %+v", *status, want)
	}
}

func TestGetSystemResourceStatus_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid credentials"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "admin", "wrong-password", "root", false)
	_, err := client.GetSystemResourceStatus(context.Background())
	if err == nil {
		t.Fatal("GetSystemResourceStatus() returned nil error, want non-nil for a 401 response")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error = %q, want it to mention status code 401", err.Error())
	}
}

func TestResourceCountMethods(t *testing.T) {
	tests := []struct {
		name     string
		call     func(*Client, context.Context) (int, error)
		wantPath string
	}{
		{"GetServerPolicyCount", (*Client).GetServerPolicyCount, "/api/v2.0/cmdb/server-policy/policy"},
		{"GetContentRoutingCount", (*Client).GetContentRoutingCount, "/api/v2.0/cmdb/server-policy/http-content-routing-policy"},
		{"GetServerPoolCount", (*Client).GetServerPoolCount, "/api/v2.0/cmdb/server-policy/server-pool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"results":[{},{},{}]}`))
			}))
			defer server.Close()

			client := NewClient(server.URL, "admin", "changeme", "root", false)
			got, err := tt.call(client, context.Background())
			if err != nil {
				t.Fatalf("%s(context.Background()) returned unexpected error: %v", tt.name, err)
			}
			if got != 3 {
				t.Errorf("%s() = %d, want 3", tt.name, got)
			}
			if gotPath != tt.wantPath {
				t.Errorf("request path = %q, want %q", gotPath, tt.wantPath)
			}
		})
	}
}

func TestResourceCountMethods_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid credentials"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "admin", "wrong-password", "root", false)
	if _, err := client.GetServerPolicyCount(context.Background()); err == nil {
		t.Fatal("GetServerPolicyCount() returned nil error, want non-nil for a 401 response")
	}
}

func TestGetSystemResourceStatus_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{not valid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "admin", "changeme", "root", false)
	_, err := client.GetSystemResourceStatus(context.Background())
	if err == nil {
		t.Fatal("GetSystemResourceStatus() returned nil error, want non-nil for a malformed JSON body")
	}
}
