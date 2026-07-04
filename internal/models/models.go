package models

import "fmt"

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (c Coordinate) String() string {
	return fmt.Sprintf("%.6f,%.6f", c.Lat, c.Lng)
}

type Waypoint struct {
	Coordinate
	Label string `json:"label,omitempty"`
}

type RouteRequest struct {
	Start Coordinate   `json:"start"`
	Stops []Coordinate `json:"stops"`
	End   Coordinate   `json:"end"`
}

func (r RouteRequest) Points() []Coordinate {
	points := make([]Coordinate, 0, len(r.Stops)+2)
	points = append(points, r.Start)
	points = append(points, r.Stops...)
	points = append(points, r.End)
	return points
}

type RouteMetrics struct {
	DistanceMeters  float64 `json:"distance_meters"`
	DurationSeconds float64 `json:"duration_seconds"`
}

type ModeEstimate struct {
	Mode            string  `json:"mode"`
	DistanceMeters  float64 `json:"distance_meters"`
	DurationSeconds float64 `json:"duration_seconds"`
}

type OrderedStop struct {
	Index      int        `json:"index"`
	Position   int        `json:"position"`
	Coordinate Coordinate `json:"coordinate"`
	Role       string     `json:"role"`
	H3Cell     string     `json:"h3_cell"`
}

type H3Cluster struct {
	Cell        string      `json:"cell"`
	Center      Coordinate  `json:"center"`
	Boundary    [][]float64 `json:"boundary"`
	StopIndexes []int       `json:"stop_indexes"`
	Count       int         `json:"count"`
}

type OptimizationResult struct {
	OriginalOrder      []int          `json:"original_order"`
	OptimizedOrder     []int          `json:"optimized_order"`
	OrderedStops       []OrderedStop  `json:"ordered_stops"`
	OriginalMetrics    RouteMetrics   `json:"original_metrics"`
	OptimizedMetrics   RouteMetrics   `json:"optimized_metrics"`
	ModeEstimates      []ModeEstimate `json:"mode_estimates"`
	DistanceSaved      float64        `json:"distance_saved_meters"`
	TimeSaved          float64        `json:"time_saved_seconds"`
	EfficiencyPercent  float64        `json:"efficiency_percent"`
	OptimizationMillis int64          `json:"optimization_millis"`
	StopCount          int            `json:"stop_count"`
	Geometry           [][]float64    `json:"geometry"`
	OriginalGeometry   [][]float64    `json:"original_geometry"`
	DurationMatrix     [][]float64    `json:"duration_matrix"`
	DistanceMatrix     [][]float64    `json:"distance_matrix"`
	H3Resolution       int            `json:"h3_resolution"`
	H3Clusters         []H3Cluster    `json:"h3_clusters"`
}

type Place struct {
	DisplayName string     `json:"display_name"`
	Name        string     `json:"name"`
	Coordinate  Coordinate `json:"coordinate"`
	Category    string     `json:"category"`
}

const (
	RoleStart = "start"
	RoleStop  = "stop"
	RoleEnd   = "end"

	ModeCar        = "car"
	ModeMotorcycle = "motorcycle"
	ModeBike       = "bike"
)
