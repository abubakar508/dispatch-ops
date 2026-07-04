package services

import "github.com/abubakar508/dispatch-ops/internal/models"

type modeProfile struct {
	mode           string
	durationFactor float64
	distanceFactor float64
}

var travelProfiles = []modeProfile{
	{mode: models.ModeCar, durationFactor: 1.0, distanceFactor: 1.0},
	{mode: models.ModeMotorcycle, durationFactor: 0.82, distanceFactor: 0.98},
	{mode: models.ModeBike, durationFactor: 3.1, distanceFactor: 0.92},
}

func estimateModes(driving models.RouteMetrics) []models.ModeEstimate {
	estimates := make([]models.ModeEstimate, 0, len(travelProfiles))
	for _, p := range travelProfiles {
		estimates = append(estimates, models.ModeEstimate{
			Mode:            p.mode,
			DistanceMeters:  driving.DistanceMeters * p.distanceFactor,
			DurationSeconds: driving.DurationSeconds * p.durationFactor,
		})
	}
	return estimates
}
