package handler

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndexHandler_ServeHTTP_ListsTargetsSorted(t *testing.T) {
	h := NewIndexHandler([]string{"fwb-02.example.com", "fwb-01.example.com"}, MetricsPath, ProbePath)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()

	firstIdx := strings.Index(body, "fwb-01.example.com")
	secondIdx := strings.Index(body, "fwb-02.example.com")
	if firstIdx == -1 || secondIdx == -1 {
		t.Fatalf("body missing target names:\n%s", body)
	}
	if firstIdx > secondIdx {
		t.Errorf("targets not sorted alphabetically:\n%s", body)
	}

	if !strings.Contains(body, `href="/probe?target=fwb-01.example.com"`) {
		t.Errorf("body missing probe link for fwb-01.example.com:\n%s", body)
	}
	if !strings.Contains(body, `href="/probe?target=fwb-02.example.com"`) {
		t.Errorf("body missing probe link for fwb-02.example.com:\n%s", body)
	}
	if !strings.Contains(body, `href="/metrics"`) {
		t.Errorf("body missing link to the exporter's own /metrics:\n%s", body)
	}
}

func TestIndexHandler_ServeHTTP_NonRootPath404s(t *testing.T) {
	h := NewIndexHandler([]string{"fwb-01.example.com"}, MetricsPath, ProbePath)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestIndexHandler_ServeHTTP_NoTargets(t *testing.T) {
	h := NewIndexHandler(nil, MetricsPath, ProbePath)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `href="/metrics"`) {
		t.Errorf("body missing link to the exporter's own /metrics:\n%s", rec.Body.String())
	}
}
