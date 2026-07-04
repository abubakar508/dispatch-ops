package optimizer

import (
	"context"
	"fmt"
)

type LocalSearchSolver struct {
	maxIterations int
}

func NewLocalSearchSolver() *LocalSearchSolver {
	return &LocalSearchSolver{maxIterations: 2000}
}

func (s *LocalSearchSolver) Solve(ctx context.Context, p Problem) (Solution, error) {
	if err := validate(p); err != nil {
		return Solution{}, err
	}

	interior := interiorNodes(p)
	if len(interior) == 0 {
		order := []int{p.Start, p.End}
		return Solution{Order: order, TotalDuration: routeDuration(p.DurationMatrix, order)}, nil
	}

	order := s.nearestNeighbor(p, interior)
	order, iterations, err := s.twoOptImprove(ctx, p.DurationMatrix, order)
	if err != nil {
		return Solution{}, err
	}

	return Solution{
		Order:         order,
		TotalDuration: routeDuration(p.DurationMatrix, order),
		Iterations:    iterations,
	}, nil
}

func interiorNodes(p Problem) []int {
	nodes := make([]int, 0, len(p.DurationMatrix))
	for i := range p.DurationMatrix {
		if i != p.Start && i != p.End {
			nodes = append(nodes, i)
		}
	}
	return nodes
}

func (s *LocalSearchSolver) nearestNeighbor(p Problem, interior []int) []int {
	matrix := p.DurationMatrix
	remaining := make(map[int]bool, len(interior))
	for _, n := range interior {
		remaining[n] = true
	}

	order := make([]int, 0, len(interior)+2)
	order = append(order, p.Start)
	current := p.Start

	for len(remaining) > 0 {
		next := -1
		best := 0.0
		for candidate := range remaining {
			d := matrix[current][candidate]
			if next == -1 || d < best {
				best = d
				next = candidate
			}
		}
		order = append(order, next)
		delete(remaining, next)
		current = next
	}

	order = append(order, p.End)
	return order
}

func (s *LocalSearchSolver) twoOptImprove(ctx context.Context, matrix [][]float64, order []int) ([]int, int, error) {
	best := make([]int, len(order))
	copy(best, order)
	bestCost := routeDuration(matrix, best)

	iterations := 0
	improved := true

	for improved && iterations < s.maxIterations {
		if err := ctx.Err(); err != nil {
			return nil, iterations, fmt.Errorf("optimizer: %w", err)
		}
		improved = false
		for i := 1; i < len(best)-2; i++ {
			for k := i + 1; k < len(best)-1; k++ {
				iterations++
				candidate := twoOptSwap(best, i, k)
				candidateCost := routeDuration(matrix, candidate)
				if candidateCost+1e-9 < bestCost {
					best = candidate
					bestCost = candidateCost
					improved = true
				}
			}
		}
	}

	return best, iterations, nil
}

func twoOptSwap(order []int, i, k int) []int {
	result := make([]int, 0, len(order))
	result = append(result, order[:i]...)
	for j := k; j >= i; j-- {
		result = append(result, order[j])
	}
	result = append(result, order[k+1:]...)
	return result
}
