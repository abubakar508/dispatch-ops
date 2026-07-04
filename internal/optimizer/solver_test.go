package optimizer

import (
	"context"
	"errors"
	"testing"
)

func TestSolveRejectsInvalidMatrix(t *testing.T) {
	solver := NewLocalSearchSolver()
	ctx := context.Background()

	if _, err := solver.Solve(ctx, Problem{DurationMatrix: nil}); !errors.Is(err, ErrEmptyMatrix) {
		t.Fatalf("expected ErrEmptyMatrix, got %v", err)
	}

	nonSquare := [][]float64{{0, 1}, {1}}
	if _, err := solver.Solve(ctx, Problem{DurationMatrix: nonSquare}); !errors.Is(err, ErrMatrixNotSquare) {
		t.Fatalf("expected ErrMatrixNotSquare, got %v", err)
	}

	tooFew := [][]float64{{0, 1}, {1, 0}}
	if _, err := solver.Solve(ctx, Problem{DurationMatrix: tooFew}); !errors.Is(err, ErrTooFewNodes) {
		t.Fatalf("expected ErrTooFewNodes, got %v", err)
	}
}

func TestSolvePreservesStartAndEnd(t *testing.T) {
	matrix := [][]float64{
		{0, 10, 15, 20, 25},
		{10, 0, 35, 25, 30},
		{15, 35, 0, 30, 5},
		{20, 25, 30, 0, 15},
		{25, 30, 5, 15, 0},
	}

	solver := NewLocalSearchSolver()
	sol, err := solver.Solve(context.Background(), Problem{DurationMatrix: matrix, Start: 0, End: 4})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sol.Order[0] != 0 {
		t.Fatalf("expected start at index 0, got %d", sol.Order[0])
	}
	if sol.Order[len(sol.Order)-1] != 4 {
		t.Fatalf("expected end at last position, got %d", sol.Order[len(sol.Order)-1])
	}
	if len(sol.Order) != len(matrix) {
		t.Fatalf("expected %d nodes, got %d", len(matrix), len(sol.Order))
	}

	seen := make(map[int]bool)
	for _, n := range sol.Order {
		if seen[n] {
			t.Fatalf("node %d visited more than once", n)
		}
		seen[n] = true
	}
}

func TestSolveFindsBetterThanNaive(t *testing.T) {
	matrix := [][]float64{
		{0, 5, 100, 100, 8},
		{5, 0, 6, 100, 100},
		{100, 6, 0, 7, 100},
		{100, 100, 7, 0, 9},
		{8, 100, 100, 9, 0},
	}

	solver := NewLocalSearchSolver()
	naive := []int{0, 1, 2, 3, 4}
	naiveCost := routeDuration(matrix, naive)

	sol, err := solver.Solve(context.Background(), Problem{DurationMatrix: matrix, Start: 0, End: 4})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sol.TotalDuration > naiveCost {
		t.Fatalf("optimized cost %.1f worse than naive %.1f", sol.TotalDuration, naiveCost)
	}
}

func TestSolveCancelledContext(t *testing.T) {
	matrix := [][]float64{
		{0, 1, 2, 3},
		{1, 0, 4, 5},
		{2, 4, 0, 6},
		{3, 5, 6, 0},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	solver := NewLocalSearchSolver()
	if _, err := solver.Solve(ctx, Problem{DurationMatrix: matrix, Start: 0, End: 3}); err == nil {
		t.Fatal("expected context cancellation error")
	}
}
