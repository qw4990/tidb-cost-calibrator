package utils

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type log10Nor struct{}

func (*log10Nor) Normalize(min, max, x float64) float64 {
	logMin := math.Log10(min)
	return (math.Log10(x) - logMin) / (math.Log10(max) - logMin)
}

type log10Tick struct{}

func (*log10Tick) Ticks(min, max float64) (ts []plot.Tick) {
	base := math.Pow10(int(math.Ceil(math.Max(1, math.Log10(max)-math.Log10(min)) / 6)))
	for min < max {
		label := fmt.Sprintf("%.0f", min)
		if min > 1e4 {
			label = fmt.Sprintf("%.0e", min)
		}
		ts = append(ts, plot.Tick{
			Value: min,
			Label: label,
		})
		min *= base
	}
	return
}

func (r Records) Len() int {
	return len(r)
}

func (r Records) XY(k int) (x, y float64) {
	return r[k].Cost, r[k].TimeMS
}

func minMaxNormCost(r Records) {
	minCost, maxCost := -1.0, -1.0
	for _, v := range r {
		if minCost < 0 || v.Cost < minCost {
			minCost = v.Cost
		}
		if maxCost < 0 || v.Cost > maxCost {
			maxCost = v.Cost
		}
	}
	for i := range r {
		r[i].Cost = (r[i].Cost - minCost) / (maxCost - minCost)
	}
}

func DrawCostRecordsTo(r Records, f, scale string) {
	//minMaxNormCost(r)

	p := plot.New()
	p.Title.Text = "cost model accuracy scatter plot"
	p.X.Label.Text = fmt.Sprintf("cost estimation [%v scale]", scale)
	p.Y.Label.Text = fmt.Sprintf("actual exec-time(ms) [%v scale]", scale)
	fontSize := vg.Length(25)
	p.Title.TextStyle.Font.Size = fontSize
	p.X.Tick.Label.Font.Size = fontSize
	p.X.Label.TextStyle.Font.Size = fontSize
	p.Y.Label.TextStyle.Font.Size = fontSize
	p.Y.Tick.Label.Font.Size = fontSize

	var maxX, maxY float64
	for i := 0; i < r.Len(); i++ {
		x, y := r.XY(i)
		maxX = math.Max(maxX, x)
		maxY = math.Max(maxY, y)
	}

	switch scale {
	case "linear":
		p.X.Max = maxX * 1.2
		p.Y.Max = maxY * 1.05
	case "log10":
		p.X.Min = 1
		p.Y.Min = 1
		p.X.Max = maxX * 1.2
		p.Y.Max = maxY * 1.05
		p.X.Scale = new(log10Nor)
		p.Y.Scale = new(log10Nor)
		p.X.Tick.Marker = new(log10Tick)
		p.Y.Tick.Marker = new(log10Tick)
	default:
		panic(fmt.Sprintf("unknonw scale %v", scale))
	}

	labeledRecords := make(map[string]Records)
	for _, record := range r {
		labeledRecords[record.Label] = append(labeledRecords[record.Label], record)
	}
	var labels []string
	for l := range labeledRecords {
		labels = append(labels, l)
	}
	labelStyle := genGlyphStyleForLabel(labels)

	sort.Slice(labels, func(i, j int) bool {
		var ki, kj int
		for k, op := range opTypes {
			if strings.Contains(strings.ToLower(labels[i]), op) {
				ki = k
			}
			if strings.Contains(strings.ToLower(labels[j]), op) {
				kj = k
			}
		}
		if ki != kj {
			return ki < kj
		}
		return labels[i] < labels[j]
	})

	for _, label := range labels {
		r := labeledRecords[label]
		s, err := plotter.NewScatter(r)
		if err != nil {
			panic(err)
		}
		s.GlyphStyle = labelStyle[label]
		p.Add(s)
		p.Legend.Add(label, s)
		p.Legend.TextStyle.Font.Size = fontSize
	}

	err := p.Save(1200, 1200, f)
	if err != nil {
		panic(err)
	}
}

var (
	opTypes = []string{"scan", "agg"}
	shapes  = map[string]draw.GlyphDrawer{
		"scan": draw.BoxGlyph{},
		"agg":  draw.PyramidGlyph{},
		"sort": draw.PlusGlyph{},
		"join": draw.RingGlyph{},
	}
	// https://www.rapidtables.com/web/color/RGB_Color.html
	colors = map[string][]color.Color{
		"scan": {rgb(102, 0, 0), rgb(153, 0, 0), rgb(204, 0, 0), rgb(255, 0, 0),
			rgb(255, 102, 102), rgb(255, 153, 153), rgb(255, 204, 204)},
		"agg": {rgb(51, 102, 0), rgb(76, 153, 0), rgb(102, 204, 0), rgb(128, 255, 0),
			rgb(153, 255, 51), rgb(178, 255, 102), rgb(229, 255, 204)},
		"sort": {rgb(51, 0, 102), rgb(76, 0, 153), rgb(102, 0, 204), rgb(127, 0, 255)},
		"join": {rgb(102, 0, 51), rgb(153, 0, 76), rgb(204, 0, 102),
			rgb(255, 0, 127), rgb(255, 51, 153), rgb(255, 102, 178), rgb(255, 153, 204)},
	}
)

func rgb(r, g, b uint8) color.Color {
	return color.RGBA{r, g, b, 255}
}

func genGlyphStyleForLabel(labels []string) map[string]draw.GlyphStyle {
	sort.Strings(labels)
	styles := make(map[string]draw.GlyphStyle)
	colorCnt := make(map[string]int)
	for _, l := range labels {
		ok := false
		for op, shape := range shapes {
			if strings.Contains(strings.ToLower(l), op) {
				cnt := colorCnt[op]
				currentColor := colors[op][(cnt % len(colors[op]))]
				colorCnt[op] = cnt + 1
				styles[l] = draw.GlyphStyle{
					Radius: 4,
					Shape:  shape,
					Color:  currentColor,
				}
				ok = true
				break
			}
		}
		if !ok {
			panic(fmt.Sprintf("cannot get style for %v", l))
		}
	}
	return styles
}
