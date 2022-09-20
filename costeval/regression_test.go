package costeval

import (
	"fmt"
	"testing"
)

func TestRegression(t *testing.T) {
	x := [][]float64{
		{1, 1},
		{2, 3},
		{3, 4},
		{4, 2},
	}
	y := make([]float64, len(x))
	for i := range x {
		y[i] = x[i][0]*2 + x[i][1]*5
	}
	w := regression(x, y)
	fmt.Println(w)
}
