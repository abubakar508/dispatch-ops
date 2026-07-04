package optimizer

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrEmptyMatrix     = errors.New("optimizer: duration matrix is empty")
	ErrMatrixNotSquare = errors.New("optimizer: duration matrix must be square")
	ErrTooFewNodes     = errors.New("optimizer: at least three nodes are required to optimize")
)

type Problem struct {
	DurationMatrix [][]float64
	Start          int
	End            int
}

type Solution struct {
	Order         []int
	TotalDuration float64
	Iterations    int
}

type Solver interface {
	Solve(ctx context.Context, p Problem) (Solution, error)
}

func validate(p Problem) error {
	n := len(p.DurationMatrix)
	if n == 0 {
		return ErrEmptyMatrix
	}
	for _, row := range p.DurationMatrix {
		if len(row) != n {
			return ErrMatrixNotSquare
		}
	}
	if n < 3 {
		return ErrTooFewNodes
	}
	if p.Start < 0 || p.Start >= n {
		return fmt.Errorf("optimizer: start index %d out of range", p.Start)
	}
	if p.End < 0 || p.End >= n {
		return fmt.Errorf("optimizer: end index %d out of range", p.End)
	}
	if p.Start == p.End {
		return errors.New("optimizer: start and end must be distinct nodes")
	}
	return nil
}

func routeDuration(matrix [][]float64, order []int) float64 {
	var total float64
	for i := 0; i+1 < len(order); i++ {
		total += matrix[order[i]][order[i+1]]
	}
	return total
}
