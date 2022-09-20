package costeval

import (
	"fmt"
	"path/filepath"

	"github.com/qw4990/tidb-cost-calibrator/utils"
	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

func CostRegression() {
	var rs utils.Records
	dataDir := "./data"
	recordFile := filepath.Join(dataDir, "tpch_clustered-2-true-records.json")
	utils.Must(utils.ReadFrom(recordFile, &rs))

	rs = filterByLabel(rs, []string{""})

}

// x * |w| == y
func regression(x [][]float64, y []float64) (w []float64) {
	tmpX := make([]float64, 0, len(x)*len(x[0]))
	for i := range x {
		for j := range x[i] {
			tmpX = append(tmpX, x[i][j])
		}
	}
	tx := tensor.New(tensor.WithShape(len(x), len(x[0])), tensor.WithBacking(tmpX))
	ty := tensor.New(tensor.WithShape(len(y)), tensor.WithBacking(y))

	// construct the NN graph
	g := gorgonia.NewGraph()
	xNode := gorgonia.NodeFromAny(g, tx, gorgonia.WithName("x"))
	yNode := gorgonia.NodeFromAny(g, ty, gorgonia.WithName("y"))

	weights := gorgonia.NewVector(g, gorgonia.Float64,
		gorgonia.WithName("w"),
		gorgonia.WithShape(xNode.Shape()[1]),
		gorgonia.WithInit(gorgonia.Uniform(0, 10)))

	absWeights := mustG(gorgonia.Abs(weights))

	predNode := mustG(gorgonia.Mul(xNode, absWeights))
	var predicated gorgonia.Value
	gorgonia.Read(predNode, &predicated)

	diff := mustG(gorgonia.Sub(predNode, yNode))
	absDiff := mustG(gorgonia.Abs(diff))
	relativeDiff := mustG(gorgonia.Div(absDiff, yNode))
	loss := mustG(gorgonia.Mean(relativeDiff))
	_, err := gorgonia.Grad(loss, weights)
	if err != nil {
		panic(fmt.Sprintf("Failed to backpropagate: %v", err))
	}

	// training
	solver := gorgonia.NewAdamSolver(gorgonia.WithLearnRate(0.01))
	model := []gorgonia.ValueGrad{weights}
	machine := gorgonia.NewTapeMachine(g, gorgonia.BindDualValues(weights))
	defer machine.Close()

	fmt.Println("init weights: ", weights.Value())
	iter := 2000
	for i := 0; i < iter; i++ {
		if err := machine.RunAll(); err != nil {
			panic(fmt.Sprintf("Error during iteration: %v: %v\n", i, err))
		}

		if err := solver.Step(model); err != nil {
			panic(err)
		}

		machine.Reset()
		lossV := loss.Value().Data().(float64)
		if i%200 == 0 {
			fmt.Printf("weights: %v, Iter: %v Loss: %.6f\n",
				weights.Value(), i, lossV)
		}
	}

	return weights.Value().Data().([]float64)
}

func mustG(n *gorgonia.Node, err error) *gorgonia.Node {
	if err != nil {
		panic(err)
	}
	return n
}
