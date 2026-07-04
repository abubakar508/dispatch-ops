package validation

import (
	"errors"
	"testing"

	"github.com/abubakar508/dispatch-ops/internal/models"
)

func TestValidateCoordinate(t *testing.T) {
	tests := []struct {
		name    string
		coord   models.Coordinate
		wantErr error
	}{
		{"valid nairobi", models.Coordinate{Lat: -1.2921, Lng: 36.8219}, nil},
		{"lat too high", models.Coordinate{Lat: 95, Lng: 0}, ErrInvalidLatitude},
		{"lat too low", models.Coordinate{Lat: -95, Lng: 0}, ErrInvalidLatitude},
		{"lng too high", models.Coordinate{Lat: 0, Lng: 200}, ErrInvalidLongitude},
		{"lng too low", models.Coordinate{Lat: 0, Lng: -200}, ErrInvalidLongitude},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCoordinate(tt.coord)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateRouteRequest(t *testing.T) {
	valid := models.RouteRequest{
		Start: models.Coordinate{Lat: -1.28, Lng: 36.81},
		Stops: []models.Coordinate{{Lat: -1.26, Lng: 36.85}},
		End:   models.Coordinate{Lat: -1.30, Lng: 36.78},
	}
	if err := ValidateRouteRequest(valid); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}

	noStops := valid
	noStops.Stops = nil
	if err := ValidateRouteRequest(noStops); !errors.Is(err, ErrInsufficientStops) {
		t.Fatalf("expected ErrInsufficientStops, got %v", err)
	}

	tooMany := valid
	tooMany.Stops = make([]models.Coordinate, MaxStops+1)
	for i := range tooMany.Stops {
		tooMany.Stops[i] = models.Coordinate{Lat: -1.2, Lng: 36.8}
	}
	if err := ValidateRouteRequest(tooMany); !errors.Is(err, ErrTooManyStops) {
		t.Fatalf("expected ErrTooManyStops, got %v", err)
	}

	badStop := valid
	badStop.Stops = []models.Coordinate{{Lat: 200, Lng: 0}}
	if err := ValidateRouteRequest(badStop); !errors.Is(err, ErrInvalidLatitude) {
		t.Fatalf("expected ErrInvalidLatitude, got %v", err)
	}
}
