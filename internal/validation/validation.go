package validation

import (
	"errors"
	"fmt"
	"math"

	"github.com/abubakar508/dispatch-ops/internal/models"
)

var (
	ErrInvalidLatitude   = errors.New("latitude must be between -90 and 90")
	ErrInvalidLongitude  = errors.New("longitude must be between -180 and 180")
	ErrInsufficientStops = errors.New("a route requires a start, an end, and at least one delivery stop")
	ErrTooManyStops      = errors.New("route exceeds the maximum supported number of stops")
	ErrTooFewPoints      = errors.New("at least three coordinates are required: a start, one delivery, and an end")
)

const MaxStops = 100

func ValidateCoordinate(c models.Coordinate) error {
	if math.IsNaN(c.Lat) || c.Lat < -90 || c.Lat > 90 {
		return fmt.Errorf("%w: got %v", ErrInvalidLatitude, c.Lat)
	}
	if math.IsNaN(c.Lng) || c.Lng < -180 || c.Lng > 180 {
		return fmt.Errorf("%w: got %v", ErrInvalidLongitude, c.Lng)
	}
	return nil
}

func ValidateRouteRequest(r models.RouteRequest) error {
	if err := ValidateCoordinate(r.Start); err != nil {
		return fmt.Errorf("start coordinate invalid: %w", err)
	}
	if err := ValidateCoordinate(r.End); err != nil {
		return fmt.Errorf("end coordinate invalid: %w", err)
	}
	if len(r.Stops) < 1 {
		return ErrInsufficientStops
	}
	if len(r.Stops) > MaxStops {
		return fmt.Errorf("%w: %d stops requested, limit is %d", ErrTooManyStops, len(r.Stops), MaxStops)
	}
	for i, stop := range r.Stops {
		if err := ValidateCoordinate(stop); err != nil {
			return fmt.Errorf("stop %d invalid: %w", i+1, err)
		}
	}
	return nil
}

func RouteRequestFromPoints(points []models.Coordinate) (models.RouteRequest, error) {
	if len(points) < 3 {
		return models.RouteRequest{}, ErrTooFewPoints
	}
	req := models.RouteRequest{
		Start: points[0],
		Stops: append([]models.Coordinate(nil), points[1:len(points)-1]...),
		End:   points[len(points)-1],
	}
	if err := ValidateRouteRequest(req); err != nil {
		return models.RouteRequest{}, err
	}
	return req, nil
}
