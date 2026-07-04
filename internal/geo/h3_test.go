package geo

import (
	"testing"

	"github.com/abubakar508/dispatch-ops/internal/models"
)

var nairobi = models.Coordinate{Lat: -1.2921, Lng: 36.8219}

func TestCellIsDeterministic(t *testing.T) {
	ix := NewIndexer(9)
	a, err := ix.Cell(nairobi)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, _ := ix.Cell(nairobi)
	if a == "" || a != b {
		t.Fatalf("expected stable cell, got %q and %q", a, b)
	}
}

func TestResolutionAffectsCell(t *testing.T) {
	coarse, _ := NewIndexer(5).Cell(nairobi)
	fine, _ := NewIndexer(11).Cell(nairobi)
	if coarse == fine {
		t.Fatal("expected different cells at different resolutions")
	}
}

func TestClusterGroupsByCell(t *testing.T) {
	ix := NewIndexer(7)
	coords := []models.Coordinate{
		{Lat: -1.2921, Lng: 36.8219},
		{Lat: -1.2922, Lng: 36.8220},
		{Lat: -1.3182, Lng: 36.8283},
	}
	clusters, err := ix.Cluster(coords)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clusters) == 0 || len(clusters) > len(coords) {
		t.Fatalf("unexpected cluster count %d", len(clusters))
	}
	for _, c := range clusters {
		if c.Count == 0 || len(c.Boundary) < 4 {
			t.Fatalf("invalid cluster: count=%d boundary=%d", c.Count, len(c.Boundary))
		}
	}
}

func TestDedupeCollapsesSameCell(t *testing.T) {
	ix := NewIndexer(7)
	coords := []models.Coordinate{
		{Lat: -1.2921, Lng: 36.8219},
		{Lat: -1.29215, Lng: 36.82195},
	}
	deduped := ix.Dedupe(coords)
	if len(deduped) != 1 {
		t.Fatalf("expected 1 deduped coordinate, got %d", len(deduped))
	}
}

func TestInvalidResolutionFallsBack(t *testing.T) {
	ix := NewIndexer(99)
	if ix.Resolution() != DefaultResolution {
		t.Fatalf("expected fallback resolution %d, got %d", DefaultResolution, ix.Resolution())
	}
}
