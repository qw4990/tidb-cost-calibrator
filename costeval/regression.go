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

func handleFixedFactors(rs utils.Records, fixedFactors map[string]float64) utils.Records {
	for i, r := range rs {
		rr := r.Clone()
		for fn, fv := range fixedFactors {
			if w, ok := r.Weights[fn]; ok && w > 0 {
				tot := w * fv
				rr.Cost -= tot
				rr.TimeMS -= r.TimeMS * (tot / r.Cost)
				delete(rr.Weights, fn)
			}
		}
		if rr.Cost < 0 {
			panic("invalid fixed factors")
		}
		rs[i] = rr
	}
	return rs
}

func CostRegression() {
	fmt.Println("============== prepare data ===============")
	var rs utils.Records
	dataDir := "./data"
	recordFile := filepath.Join(dataDir, "tpch_clustered-2-true-records.json")
	utils.Must(utils.ReadFrom(recordFile, &rs))
	rs = filterByLabel(rs, []string{"HashAgg", "StreamAgg", "PhaseAgg"})
	//rs = scaleByLabel(rs, map[string]int{"PhaseAgg": 2})

	fmt.Println("============== handle fixed factor ===============")
	rs = handleFixedFactors(rs, map[string]float64{
		//"tiflash_mem_factor": 0,
		//"tidb_cpu_factor":    0,
	})
	for _, r := range rs {
		fmt.Println("> ", r.Label, r.Cost, r.TimeMS)
	}

	fmt.Println("============== training ===============")
	x, y, idx2Name := prepareData(rs)
	w := regression(x, y)
	factor := make(map[string]float64)
	for i := range w {
		factor[idx2Name[i]] = w[i]
	}

	fmt.Println("============== normalize factors ===============")
	minFactor := 1e20
	for _, v := range factor {
		minFactor = math.Min(v, minFactor)
	}
	for i := range factor {
		factor[i] /= minFactor
	}
	sort.Strings(idx2Name)
	for _, name := range idx2Name {
		fmt.Printf("%-40s\t%.2f\n", name, factor[name])
	}

	fmt.Println("============== draw ===============")
	updateCost(rs, factor)
	utils.DrawCostRecordsTo(rs, "./data/regression.png", "linear")
}

func prepareData(rs utils.Records) (x [][]float64, y []float64, idx2Name []string) {
	name2Idx := make(map[string]int)
	wIdxCnt := 0
	factorWeights := make(map[string]float64)
	factorCounts := make(map[string]int)
	for i := range rs {
		for name := range rs[i].Weights {
			if _, ok := name2Idx[name]; !ok {
				name2Idx[name] = wIdxCnt
				idx2Name = append(idx2Name, name)
				wIdxCnt++
			}
		}
	}
	for i := range rs {
		v := make([]float64, wIdxCnt)
		for name, val := range rs[i].Weights {
			idx := name2Idx[name]
			v[idx] = val
			factorWeights[name] += val
			factorCounts[name]++
		}
		x = append(x, v)
		y = append(y, rs[i].TimeMS)
	}

	for name, weight := range factorWeights {
		fmt.Printf("%-40s\t%v\t%.2f\n", name, factorCounts[name], weight)
	}

	return
}

// x * |w| == y, the returned w is non-negative
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

	solver := gorgonia.NewVanillaSolver(gorgonia.WithLearnRate(0.2))
	model := []gorgonia.ValueGrad{weights}
	machine := gorgonia.NewTapeMachine(g, gorgonia.BindDualValues(weights))
	defer machine.Close()

	// training
	fmt.Println("init weights: ", weights.Value())
	iter := 20000
	for i := 0; i < iter; i++ {
		if err := machine.RunAll(); err != nil {
			panic(fmt.Sprintf("Error during iteration: %v: %v\n", i, err))
		}

		if err := solver.Step(model); err != nil {
			panic(err)
		}

		machine.Reset()
		lossV := loss.Value().Data().(float64)
		if i%10000 == 0 {
			fmt.Printf("Iter: %v Loss: %.6f\n", i, lossV)
		}
	}

	// post-process the results
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
