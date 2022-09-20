package costeval

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestRegression1(t *testing.T) {
	x := randMatrix(10, 2, 10)
	w := randMatrix(1, 2, 10)[0]

	var y []float64
	for i := range x {
		y = append(y, prodVector(x[i], w))
	}
	estW := regression(x, y)
	fmt.Println(w)
	fmt.Println(estW)
}

func TestRegression2(t *testing.T) {
	x := randMatrix(10, 4, 10)
	w := randMatrix(1, 4, 10)[0]

	var y []float64
	for i := range x {
		y = append(y, prodVector(x[i], w))
	}
	estW := regression(x, y)
	fmt.Println(w)
	fmt.Println(estW)
}

func randMatrix(row, col int, scale float64) (m [][]float64) {
	for i := 0; i < row; i++ {
		var v []float64
		for j := 0; j < col; j++ {
			v = append(v, rand.Float64()*scale)
		}
		m = append(m, v)
	}
	return
}

func prodVector(v1, v2 []float64) (v float64) {
	for i := range v1 {
		v += v1[i] * v2[i]
	}
	return v
}
