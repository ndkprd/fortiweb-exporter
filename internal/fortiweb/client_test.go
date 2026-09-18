package fortiweb

import (
	"context"
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
	const wantUsername, wantPassword = "admin", "changeme"

	var gotMethod, gotPath string
	var gotUsername, gotPassword string
	var gotOK bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotUsername, gotPassword, gotOK = r.BasicAuth()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(validStatusResponse))
	}))
	defer server.Close()

	client := NewClient(server.URL, wantUsername, wantPassword, false)
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
	if !gotOK {
		t.Fatal("request had no Basic Auth credentials")
	}
	if gotUsername != wantUsername || gotPassword != wantPassword {
		t.Errorf("Basic Auth = (%q, %q), want (%q, %q)", gotUsername, gotPassword, wantUsername, wantPassword)
	}
}

func TestGetSystemResourceStatus_ParsesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(validStatusResponse))
	}))
	defer server.Close()

	client := NewClient(server.URL, "admin", "changeme", false)
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

	client := NewClient(server.URL, "admin", "wrong-password", false)
	_, err := client.GetSystemResourceStatus(context.Background())
	if err == nil {
		t.Fatal("GetSystemResourceStatus() returned nil error, want non-nil for a 401 response")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error = %q, want it to mention status code 401", err.Error())
	}
}

func TestGetSystemResourceStatus_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{not valid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "admin", "changeme", false)
	_, err := client.GetSystemResourceStatus(context.Background())
	if err == nil {
		t.Fatal("GetSystemResourceStatus() returned nil error, want non-nil for a malformed JSON body")
	}
}
