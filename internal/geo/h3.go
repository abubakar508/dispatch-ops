package geo

import (
	"fmt"

	"github.com/abubakar508/dispatch-ops/internal/models"
	"github.com/uber/h3-go/v4"
)

const DefaultResolution = 9

type Indexer struct {
	resolution int
}

func NewIndexer(resolution int) *Indexer {
	if resolution < 0 || resolution > 15 {
		resolution = DefaultResolution
	}
	return &Indexer{resolution: resolution}
}

func (ix *Indexer) Resolution() int {
	return ix.resolution
}

func (ix *Indexer) Cell(c models.Coordinate) (string, error) {
	cell, err := h3.LatLngToCell(h3.NewLatLng(c.Lat, c.Lng), ix.resolution)
	if err != nil {
		return "", fmt.Errorf("geo: index coordinate: %w", err)
	}
	return cell.String(), nil
}

func (ix *Indexer) Cluster(coords []models.Coordinate) ([]models.H3Cluster, error) {
	order := make([]string, 0, len(coords))
	grouped := make(map[string][]int)

	for i, c := range coords {
		cell, err := ix.Cell(c)
		if err != nil {
			return nil, err
		}
		if _, seen := grouped[cell]; !seen {
			order = append(order, cell)
		}
		grouped[cell] = append(grouped[cell], i)
	}

	clusters := make([]models.H3Cluster, 0, len(order))
	for _, cellStr := range order {
		cluster, err := buildCluster(cellStr, grouped[cellStr])
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster)
	}
	return clusters, nil
}

func buildCluster(cellStr string, indexes []int) (models.H3Cluster, error) {
	cell := h3.Cell(h3.IndexFromString(cellStr))
	if !cell.IsValid() {
		return models.H3Cluster{}, fmt.Errorf("geo: parse cell %q", cellStr)
	}

	center, err := cell.LatLng()
	if err != nil {
		return models.H3Cluster{}, fmt.Errorf("geo: cell center: %w", err)
	}

	boundary, err := cell.Boundary()
	if err != nil {
		return models.H3Cluster{}, fmt.Errorf("geo: cell boundary: %w", err)
	}

	ring := make([][]float64, 0, len(boundary)+1)
	for _, pt := range boundary {
		ring = append(ring, []float64{pt.Lng, pt.Lat})
	}
	if len(ring) > 0 {
		ring = append(ring, ring[0])
	}

	return models.H3Cluster{
		Cell:        cellStr,
		Center:      models.Coordinate{Lat: center.Lat, Lng: center.Lng},
		Boundary:    ring,
		StopIndexes: indexes,
		Count:       len(indexes),
	}, nil
}

func (ix *Indexer) Dedupe(coords []models.Coordinate) []models.Coordinate {
	seen := make(map[string]bool, len(coords))
	result := make([]models.Coordinate, 0, len(coords))
	for _, c := range coords {
		cell, err := ix.Cell(c)
		if err != nil || cell == "" {
			result = append(result, c)
			continue
		}
		if seen[cell] {
			continue
		}
		seen[cell] = true
		result = append(result, c)
	}
	return result
}
