package services

import (
	"context"
	"errors"
	"testing"

	"github.com/abubakar508/dispatch-ops/internal/geo"
	"github.com/abubakar508/dispatch-ops/internal/models"
	"github.com/abubakar508/dispatch-ops/internal/optimizer"
	"github.com/abubakar508/dispatch-ops/internal/osrm"
)

func newTestService(client osrm.Client) *RouteService {
	return NewRouteService(client, optimizer.NewLocalSearchSolver(), geo.NewIndexer(9))
}

type stubOSRM struct {
	table    osrm.TableResult
	route    osrm.RouteResult
	tableErr error
	routeErr error
	calls    int
}

func (s *stubOSRM) Table(ctx context.Context, points []models.Coordinate) (osrm.TableResult, error) {
	return s.table, s.tableErr
}

func (s *stubOSRM) Route(ctx context.Context, points []models.Coordinate) (osrm.RouteResult, error) {
	s.calls++
	if s.routeErr != nil {
		return osrm.RouteResult{}, s.routeErr
	}
	duration := 600.0 - float64(s.calls)*100
	return osrm.RouteResult{
		Geometry: [][]float64{{36.81, -1.28}, {36.82, -1.29}},
		Metrics:  models.RouteMetrics{DistanceMeters: 5000 - float64(s.calls)*500, DurationSeconds: duration},
	}, nil
}

func sampleRequest() models.RouteRequest {
	return models.RouteRequest{
		Start: models.Coordinate{Lat: -1.2864, Lng: 36.8172},
		Stops: []models.Coordinate{
			{Lat: -1.2683, Lng: 36.8109},
			{Lat: -1.2635, Lng: 36.8506},
		},
		End: models.Coordinate{Lat: -1.2999, Lng: 36.7820},
	}
}

func TestOptimizeHappyPath(t *testing.T) {
	stub := &stubOSRM{
		table: osrm.TableResult{
			Durations: [][]float64{
				{0, 300, 600, 900},
				{300, 0, 200, 500},
				{600, 200, 0, 400},
				{900, 500, 400, 0},
			},
		},
	}
	svc := newTestService(stub)

	result, err := svc.Optimize(context.Background(), sampleRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StopCount != 2 {
		t.Fatalf("expected 2 stops, got %d", result.StopCount)
	}
	if result.OptimizedOrder[0] != 0 {
		t.Fatalf("expected start index first, got %d", result.OptimizedOrder[0])
	}
	if result.OptimizedOrder[len(result.OptimizedOrder)-1] != 3 {
		t.Fatalf("expected end index last, got %d", result.OptimizedOrder[len(result.OptimizedOrder)-1])
	}
	if len(result.OrderedStops) != 4 {
		t.Fatalf("expected 4 ordered stops, got %d", len(result.OrderedStops))
	}
	if result.OrderedStops[0].Role != models.RoleStart {
		t.Fatalf("expected first role start, got %s", result.OrderedStops[0].Role)
	}
	if result.OrderedStops[3].Role != models.RoleEnd {
		t.Fatalf("expected last role end, got %s", result.OrderedStops[3].Role)
	}
}

func TestOptimizeValidationError(t *testing.T) {
	svc := newTestService(&stubOSRM{})
	req := sampleRequest()
	req.Stops = nil

	_, err := svc.Optimize(context.Background(), req)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestOptimizeRejectsMismatchedMatrix(t *testing.T) {
	stub := &stubOSRM{
		table: osrm.TableResult{
			Durations: [][]float64{
				{0, 1, 2, 3, 4},
				{1, 0, 2, 3, 4},
				{2, 2, 0, 3, 4},
				{3, 3, 3, 0, 4},
				{4, 4, 4, 4, 0},
			},
		},
	}
	svc := newTestService(stub)

	_, err := svc.Optimize(context.Background(), sampleRequest())
	if err == nil {
		t.Fatal("expected error for matrix/point size mismatch")
	}
}

func TestOptimizePropagatesOSRMError(t *testing.T) {
	stub := &stubOSRM{tableErr: osrm.ErrServiceUnavailable}
	svc := newTestService(stub)

	_, err := svc.Optimize(context.Background(), sampleRequest())
	if !errors.Is(err, osrm.ErrServiceUnavailable) {
		t.Fatalf("expected ErrServiceUnavailable, got %v", err)
	}
}
