package services

import (
	"context"
	"fmt"
	"time"

	"github.com/abubakar508/dispatch-ops/internal/geo"
	"github.com/abubakar508/dispatch-ops/internal/models"
	"github.com/abubakar508/dispatch-ops/internal/optimizer"
	"github.com/abubakar508/dispatch-ops/internal/osrm"
	"github.com/abubakar508/dispatch-ops/internal/validation"
)

type RouteService struct {
	osrm    osrm.Client
	solver  optimizer.Solver
	indexer *geo.Indexer
}

func NewRouteService(client osrm.Client, solver optimizer.Solver, indexer *geo.Indexer) *RouteService {
	return &RouteService{osrm: client, solver: solver, indexer: indexer}
}

func (s *RouteService) Optimize(ctx context.Context, req models.RouteRequest) (models.OptimizationResult, error) {
	if err := validation.ValidateRouteRequest(req); err != nil {
		return models.OptimizationResult{}, fmt.Errorf("route service: %w", err)
	}

	points := req.Points()
	startIndex := 0
	endIndex := len(points) - 1

	table, err := s.osrm.Table(ctx, points)
	if err != nil {
		return models.OptimizationResult{}, fmt.Errorf("route service: build travel matrix: %w", err)
	}
	if len(table.Durations) != len(points) {
		return models.OptimizationResult{}, fmt.Errorf("route service: travel matrix size %d does not match %d points: %w",
			len(table.Durations), len(points), osrm.ErrInvalidResponse)
	}

	originalOrder := sequentialOrder(len(points))

	start := time.Now()
	solution, err := s.solver.Solve(ctx, optimizer.Problem{
		DurationMatrix: table.Durations,
		Start:          startIndex,
		End:            endIndex,
	})
	if err != nil {
		return models.OptimizationResult{}, fmt.Errorf("route service: optimize order: %w", err)
	}
	optimizationMillis := time.Since(start).Milliseconds()

	optimizedGeometry, optimizedMetrics, err := s.geometryFor(ctx, points, solution.Order)
	if err != nil {
		return models.OptimizationResult{}, fmt.Errorf("route service: optimized geometry: %w", err)
	}

	originalGeometry, originalMetrics, err := s.geometryFor(ctx, points, originalOrder)
	if err != nil {
		return models.OptimizationResult{}, fmt.Errorf("route service: original geometry: %w", err)
	}

	distanceSaved := originalMetrics.DistanceMeters - optimizedMetrics.DistanceMeters
	timeSaved := originalMetrics.DurationSeconds - optimizedMetrics.DurationSeconds

	efficiency := 0.0
	if originalMetrics.DurationSeconds > 0 {
		efficiency = (timeSaved / originalMetrics.DurationSeconds) * 100
	}
	if efficiency < 0 {
		efficiency = 0
	}

	orderedStops, err := s.buildOrderedStops(points, solution.Order, startIndex, endIndex)
	if err != nil {
		return models.OptimizationResult{}, fmt.Errorf("route service: index stops: %w", err)
	}

	clusters, err := s.indexer.Cluster(points)
	if err != nil {
		return models.OptimizationResult{}, fmt.Errorf("route service: cluster stops: %w", err)
	}

	return models.OptimizationResult{
		OriginalOrder:      originalOrder,
		OptimizedOrder:     solution.Order,
		OrderedStops:       orderedStops,
		OriginalMetrics:    originalMetrics,
		OptimizedMetrics:   optimizedMetrics,
		ModeEstimates:      estimateModes(optimizedMetrics),
		DistanceSaved:      distanceSaved,
		TimeSaved:          timeSaved,
		EfficiencyPercent:  efficiency,
		OptimizationMillis: optimizationMillis,
		StopCount:          len(req.Stops),
		Geometry:           optimizedGeometry,
		OriginalGeometry:   originalGeometry,
		DurationMatrix:     table.Durations,
		DistanceMatrix:     table.Distances,
		H3Resolution:       s.indexer.Resolution(),
		H3Clusters:         clusters,
	}, nil
}

func (s *RouteService) geometryFor(ctx context.Context, points []models.Coordinate, order []int) ([][]float64, models.RouteMetrics, error) {
	ordered := make([]models.Coordinate, len(order))
	for i, idx := range order {
		ordered[i] = points[idx]
	}

	result, err := s.osrm.Route(ctx, ordered)
	if err != nil {
		return nil, models.RouteMetrics{}, err
	}
	return result.Geometry, result.Metrics, nil
}

func sequentialOrder(n int) []int {
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	return order
}

func (s *RouteService) buildOrderedStops(points []models.Coordinate, order []int, startIndex, endIndex int) ([]models.OrderedStop, error) {
	stops := make([]models.OrderedStop, len(order))
	for pos, idx := range order {
		role := models.RoleStop
		switch idx {
		case startIndex:
			role = models.RoleStart
		case endIndex:
			role = models.RoleEnd
		}

		cell, err := s.indexer.Cell(points[idx])
		if err != nil {
			return nil, err
		}

		stops[pos] = models.OrderedStop{
			Index:      idx,
			Position:   pos + 1,
			Coordinate: points[idx],
			Role:       role,
			H3Cell:     cell,
		}
	}
	return stops, nil
}
