package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mable/audience-builder/backend/internal/api"
	"github.com/mable/audience-builder/backend/internal/evaluator"
	"github.com/mable/audience-builder/backend/internal/model"
	"github.com/mable/audience-builder/backend/internal/store"
)

func setupTestServer(t *testing.T) http.Handler {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test store: %v", err)
	}
	if err := db.SeedData(); err != nil {
		t.Fatalf("failed to seed test store: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	eval := evaluator.New(db)
	return api.NewServer(logger, eval)
}

func TestHealthEndpoint(t *testing.T) {
	handler := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed parsing health response: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", body["status"])
	}
}

func TestPreviewEndpoint_Success(t *testing.T) {
	handler := setupTestServer(t)

	payload := `{
		"name": "Viewed but not purchased",
		"asOf": "2026-09-29T00:00:00.000Z",
		"conditions": [
			{
				"eventType": "product_view",
				"operator": "at_least",
				"count": 2,
				"withinDays": 7
			},
			{
				"eventType": "purchase",
				"operator": "exactly",
				"count": 0,
				"withinDays": 7
			}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/v1/audiences/preview", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp model.PreviewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed decoding preview response: %v", err)
	}

	if resp.Name != "Viewed but not purchased" {
		t.Errorf("expected name 'Viewed but not purchased', got %q", resp.Name)
	}
	if resp.Total != 4 {
		t.Fatalf("expected total 4, got %d", resp.Total)
	}
	if len(resp.Members) != 4 {
		t.Fatalf("expected 4 members, got %d", len(resp.Members))
	}
}

func TestPreviewEndpoint_ValidationError(t *testing.T) {
	handler := setupTestServer(t)

	// Missing name and negative count
	payload := `{
		"name": "",
		"asOf": "2026-09-29T00:00:00.000Z",
		"conditions": [
			{
				"eventType": "product_view",
				"operator": "at_least",
				"count": -1,
				"withinDays": 7
			}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/v1/audiences/preview", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
	}

	var errResp model.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed decoding error response: %v", err)
	}

	if errResp.Error != "validation_error" {
		t.Errorf("expected error code 'validation_error', got %q", errResp.Error)
	}
	if len(errResp.Details) < 2 {
		t.Errorf("expected at least 2 validation details, got %d", len(errResp.Details))
	}
}

func TestPreviewEndpoint_MalformedJSON(t *testing.T) {
	handler := setupTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/audiences/preview", bytes.NewBufferString("{ invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	handler := setupTestServer(t)

	req := httptest.NewRequest(http.MethodOptions, "/v1/audiences/preview", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for OPTIONS, got %d", rec.Code)
	}

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Allow-Origin '*', got %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}
