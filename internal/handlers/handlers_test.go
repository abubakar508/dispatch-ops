package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abubakar508/dispatch-ops/internal/models"
	"github.com/abubakar508/dispatch-ops/internal/osrm"
)

type stubService struct {
	result models.OptimizationResult
	err    error
}

func (s *stubService) Optimize(ctx context.Context, req models.RouteRequest) (models.OptimizationResult, error) {
	return s.result, s.err
}

type stubGeocoder struct {
	results []models.Place
	err     error
}

func (s *stubGeocoder) Search(ctx context.Context, query string, limit int) ([]models.Place, error) {
	return s.results, s.err
}

func newTestHandler(svc RouteOptimizer) *Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(svc, &stubGeocoder{}, logger)
}

func TestHealth(t *testing.T) {
	h := newTestHandler(&stubService{})
	rec := httptest.NewRecorder()
	h.Health(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "ok") {
		t.Fatal("expected ok status")
	}
}

func TestOptimizeSuccess(t *testing.T) {
	svc := &stubService{result: models.OptimizationResult{
		StopCount:        2,
		OptimizedOrder:   []int{0, 1, 2, 3},
		OrderedStops:     []models.OrderedStop{{Position: 1, Role: models.RoleStart}},
		OptimizedMetrics: models.RouteMetrics{DistanceMeters: 5000, DurationSeconds: 600},
		Geometry:         [][]float64{{36.8, -1.2}},
	}}
	h := newTestHandler(svc)

	body := mustJSON(t, models.RouteRequest{
		Start: models.Coordinate{Lat: -1.28, Lng: 36.81},
		Stops: []models.Coordinate{{Lat: -1.26, Lng: 36.85}},
		End:   models.Coordinate{Lat: -1.30, Lng: 36.78},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/optimize", body)
	rec := httptest.NewRecorder()

	h.Optimize(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result models.OptimizationResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("expected valid JSON body: %v", err)
	}
	if result.StopCount != 2 {
		t.Fatalf("expected stop count 2, got %d", result.StopCount)
	}
}

func TestOptimizeServiceUnavailable(t *testing.T) {
	h := newTestHandler(&stubService{err: osrm.ErrServiceUnavailable})

	body := mustJSON(t, models.RouteRequest{
		Start: models.Coordinate{Lat: -1.28, Lng: 36.81},
		Stops: []models.Coordinate{{Lat: -1.26, Lng: 36.85}},
		End:   models.Coordinate{Lat: -1.30, Lng: 36.78},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/optimize", body)
	rec := httptest.NewRecorder()

	h.Optimize(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	var resp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected error JSON: %v", err)
	}
	if !strings.Contains(resp.Error, "temporarily unavailable") {
		t.Fatalf("expected friendly message, got %q", resp.Error)
	}
}

func TestOptimizeBadJSON(t *testing.T) {
	h := newTestHandler(&stubService{})
	req := httptest.NewRequest(http.MethodPost, "/api/optimize", strings.NewReader("{not-json"))
	rec := httptest.NewRecorder()

	h.Optimize(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRouterRoutes(t *testing.T) {
	h := newTestHandler(&stubService{result: models.OptimizationResult{StopCount: 1}})
	router := NewRouter(h, "")
	server := httptest.NewServer(router)
	defer server.Close()

	res, err := http.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatalf("healthz request failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /healthz, got %d", res.StatusCode)
	}
	if res.Header.Get("Access-Control-Allow-Origin") == "" {
		t.Fatal("expected CORS header")
	}
}

func TestDeliveriesSuccess(t *testing.T) {
	svc := &stubService{result: models.OptimizationResult{StopCount: 1, OptimizedOrder: []int{0, 1, 2}}}
	h := newTestHandler(svc)

	body := mustJSON(t, map[string]any{
		"points": []models.Coordinate{
			{Lat: -1.28, Lng: 36.81},
			{Lat: -1.26, Lng: 36.85},
			{Lat: -1.30, Lng: 36.78},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/deliveries", body)
	rec := httptest.NewRecorder()

	h.Deliveries(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDeliveriesTooFewPoints(t *testing.T) {
	h := newTestHandler(&stubService{})
	body := mustJSON(t, map[string]any{
		"points": []models.Coordinate{{Lat: -1.28, Lng: 36.81}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/deliveries", body)
	rec := httptest.NewRecorder()

	h.Deliveries(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestSearchSuccess(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	geocoder := &stubGeocoder{results: []models.Place{
		{Name: "Westlands", Coordinate: models.Coordinate{Lat: -1.2683, Lng: 36.8109}},
	}}
	h := New(&stubService{}, geocoder, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=westlands", nil)
	rec := httptest.NewRecorder()
	h.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Westlands") {
		t.Fatal("expected search result in body")
	}
}

func mustJSON(t *testing.T, v any) io.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return bytes.NewReader(b)
}
