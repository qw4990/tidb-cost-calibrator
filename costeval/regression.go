package costeval

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"

	"github.com/qw4990/tidb-cost-calibrator/utils"
	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

func CostRegression() {
	var rs utils.Records
	dataDir := "./data"
	recordFile := filepath.Join(dataDir, "tpch_clustered-2-true-records.json")
	utils.Must(utils.ReadFrom(recordFile, &rs))

	rs = filterByLabel(rs, []string{"Agg"})

	nameIdx := make(map[string]int)
	var idxName []string
	wIdxCnt := 0
	for i := range rs {
		for name := range rs[i].Weights {
			if _, ok := nameIdx[name]; !ok {
				nameIdx[name] = wIdxCnt
				idxName = append(idxName, name)
				wIdxCnt++
			}
		}
	}
	var x [][]float64
	var y []float64
	for i := range rs {
		v := make([]float64, wIdxCnt)
		for name, val := range rs[i].Weights {
			idx := nameIdx[name]
			v[idx] = val
		}
		x = append(x, v)
		y = append(y, rs[i].TimeMS)
	}

	fmt.Println("========================== training ======================")
	for i := range x {
		fmt.Println(x[i])
	}
	fmt.Println(y)
	fmt.Println("========================================")
	w := regression(x, y)
	factor := make(map[string]float64)
	for i := range w {
		factor[idxName[i]] = w[i]
	}
	minFactor := 1e20
	for _, v := range factor {
		minFactor = math.Min(v, minFactor)
	}
	for i := range factor {
		factor[i] /= minFactor
	}
	fmt.Println("=============== norm factor =====================")
	sort.Strings(idxName)
	for _, name := range idxName {
		fmt.Printf("%v: %v\n", name, factor[name])
	}
	fmt.Println("=============== norm factor =====================")
	updateCost(rs, factor)

	utils.DrawCostRecordsTo(rs, "./data/regression.png", "linear")
}

// x * |w| == y
func regression(x [][]float64, y []float64) (w []float64) {
	// max norm
	var maxX, maxY float64
	for i := range x {
		for j := range x[i] {
			maxX = math.Max(maxX, x[i][j])
		}
	}
	for i := range y {
		maxY = math.Max(maxY, y[i])
	}
	for i := range x {
		for j := range x[i] {
			x[i][j] /= maxX
		}
	}
	for i := range y {
		y[i] /= maxY
	}

	// construct tensors
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
		gorgonia.WithInit(gorgonia.Uniform(0, 1)))

	absWeights := mustG(gorgonia.Abs(weights))

	predNode := mustG(gorgonia.Mul(xNode, absWeights))
	var predicated gorgonia.Value
	gorgonia.Read(predNode, &predicated)

	diff := mustG(gorgonia.Abs(mustG(gorgonia.Sub(predNode, yNode))))
	relativeDiff := mustG(gorgonia.Div(diff, yNode))
	loss := mustG(gorgonia.Mean(relativeDiff))
	_, err := gorgonia.Grad(loss, weights)
	if err != nil {
		panic(fmt.Sprintf("Failed to backpropagate: %v", err))
	}

	// training
	solver := gorgonia.NewVanillaSolver(gorgonia.WithLearnRate(0.5))
	model := []gorgonia.ValueGrad{weights}
	machine := gorgonia.NewTapeMachine(g, gorgonia.BindDualValues(weights))
	defer machine.Close()

	fmt.Println("init weights: ", weights.Value())
	iter := 200000
	for i := 0; i < iter; i++ {
		if err := machine.RunAll(); err != nil {
			panic(fmt.Sprintf("Error during iteration: %v: %v\n", i, err))
		}

		if err := solver.Step(model); err != nil {
			panic(err)
		}

		machine.Reset()
		lossV := loss.Value().Data().(float64)
		if i%100 == 0 {
			fmt.Printf("weights: %v, Iter: %v Loss: %.6f\n",
				weights.Value(), i, lossV)
		}
	}

	w = absWeights.Value().Data().([]float64)
	for i := range w {
		w[i] = w[i] * maxX / maxY
	}
	return w
}

func mustG(n *gorgonia.Node, err error) *gorgonia.Node {
	if err != nil {
		panic(err)
	}
	return n
}
