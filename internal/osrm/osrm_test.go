package osrm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/abubakar508/dispatch-ops/internal/models"
)

var testPoints = []models.Coordinate{
	{Lat: -1.2864, Lng: 36.8172},
	{Lat: -1.2683, Lng: 36.8109},
	{Lat: -1.2999, Lng: 36.7820},
}

func TestTableSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/table/v1/driving/") {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"Ok","durations":[[0,120,240],[120,0,180],[240,180,0]],"distances":[[0,1000,2000],[1000,0,1500],[2000,1500,0]]}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, 5*time.Second)
	result, err := client.Table(context.Background(), testPoints)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Durations) != 3 {
		t.Fatalf("expected 3x3 matrix, got %d rows", len(result.Durations))
	}
	if result.Durations[0][1] != 120 {
		t.Fatalf("expected duration 120, got %v", result.Durations[0][1])
	}
}

func TestTableUpstreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, 5*time.Second)
	_, err := client.Table(context.Background(), testPoints)
	if !errors.Is(err, ErrServiceUnavailable) {
		t.Fatalf("expected ErrServiceUnavailable, got %v", err)
	}
}

func TestTableRateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, 5*time.Second)
	_, err := client.Table(context.Background(), testPoints)
	if !errors.Is(err, ErrServiceUnavailable) {
		t.Fatalf("expected ErrServiceUnavailable, got %v", err)
	}
}

func TestRouteSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"Ok","routes":[{"distance":3500,"duration":420,"geometry":{"coordinates":[[36.8172,-1.2864],[36.8109,-1.2683],[36.7820,-1.2999]]}}]}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, 5*time.Second)
	result, err := client.Route(context.Background(), testPoints)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Metrics.DistanceMeters != 3500 {
		t.Fatalf("expected distance 3500, got %v", result.Metrics.DistanceMeters)
	}
	if len(result.Geometry) != 3 {
		t.Fatalf("expected 3 geometry points, got %d", len(result.Geometry))
	}
}

func TestRouteNoRoute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"NoRoute","message":"impossible"}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, 5*time.Second)
	_, err := client.Route(context.Background(), testPoints)
	if !errors.Is(err, ErrNoRoute) {
		t.Fatalf("expected ErrNoRoute, got %v", err)
	}
}

func TestTableTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, 10*time.Millisecond)
	_, err := client.Table(context.Background(), testPoints)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestTableRequiresTwoPoints(t *testing.T) {
	client := NewHTTPClient("http://example.com", time.Second)
	_, err := client.Table(context.Background(), testPoints[:1])
	if err == nil {
		t.Fatal("expected error for single point")
	}
}
