package costeval

import (
	"fmt"
	"math"

	"github.com/apache/arrow/go/arrow/tensor"
	"github.com/qw4990/tidb-cost-calibrator/utils"
	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

func regression(x [][]float64, y []float64) (w []float64) {
	
}

func regressionCostFactors(rs utils.Records) CostFactors {
	var scale [NumFactors]float64
	rs, scale = normalize(rs)
	x, y := convert2XY(rs)
	g := gorgonia.NewGraph()
	xNode := gorgonia.NodeFromAny(g, x, gorgonia.WithName("x"))
	yNode := gorgonia.NodeFromAny(g, y, gorgonia.WithName("y"))

	costFactor := gorgonia.NewVector(g, gorgonia.Float64,
		gorgonia.WithName("cost-factor"),
		gorgonia.WithShape(xNode.Shape()[1]),
		gorgonia.WithInit(func(dt tensor.Dtype, s ...int) interface{} {
			switch dt {
			case tensor.Float64: // (CPU, CopCPU, Net, Scan, DescScan, Mem, Seek)
				return []float64{0, 0, 0, 0, 0, 0, 0}
			default:
				panic("invalid type")
			}
			return nil
		}))
	//gorgonia.WithInit(gorgonia.Zeroes()))
	//gorgonia.WithInit(gorgonia.Uniform(0, 300)))
	//strictFactor, err := gorgonia.LeakyRelu(costFactor, 0)
	//if err != nil {
	//	panic(err)
	//}

	pred := must(gorgonia.Mul(xNode, costFactor))
	var predicated gorgonia.Value
	gorgonia.Read(pred, &predicated)

	diff := must(gorgonia.Abs(must(gorgonia.Sub(pred, yNode))))
	relativeDiff := must(gorgonia.Div(diff, yNode))
	loss := must(gorgonia.Mean(relativeDiff))
	_, err := gorgonia.Grad(loss, costFactor)
	if err != nil {
		panic(fmt.Sprintf("Failed to backpropagate: %v", err))
	}

	solver := gorgonia.NewAdamSolver(gorgonia.WithLearnRate(0.00001))
	model := []gorgonia.ValueGrad{costFactor}

	machine := gorgonia.NewTapeMachine(g, gorgonia.BindDualValues(costFactor))
	defer machine.Close()

	fmt.Println("init theta: ", costFactor.Value())

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
		if i%1000 == 0 {
			fmt.Printf("theta: %v, Iter: %v Loss: %.6f\n",
				costFactor.Value(),
				i,
				lossV)
		}
	}

	var fv CostFactors
	for i := range fv {
		fv[i] = costFactor.Value().Data().([]float64)[i]
	}

	// scale factors
	minFV := 1e9
	for i := range fv {
		fv[i] /= scale[i]
		if fv[i] == 0 {
			continue
		}
		if math.Abs(fv[i]) < minFV {
			minFV = math.Abs(fv[i])
		}
	}
	for i := range fv {
		fv[i] /= minFV
	}

	return fv
}

func convert2XY(rs Records) (*tensor.Dense, *tensor.Dense) {
	by := make([]float64, 0, len(rs))
	for _, r := range rs {
		by = append(by, r.TimeMS)
	}
	y := tensor.New(tensor.WithShape(len(rs)), tensor.WithBacking(by))

	bx := make([]float64, 0, len(rs)*len(rs[0].CostWeights))
	for _, r := range rs {
		for _, w := range r.CostWeights {
			bx = append(bx, w)
		}
	}
	x := tensor.New(tensor.WithShape(len(rs), len(rs[0].CostWeights)), tensor.WithBacking(bx))

	return x, y
}

func must(n *gorgonia.Node, err error) *gorgonia.Node {
	if err != nil {
		panic(err)
	}
	return n
}

func one(size int) []float64 {
	one := make([]float64, size)
	for i := 0; i < size; i++ {
		one[i] = 1.0
	}
	return one
}
