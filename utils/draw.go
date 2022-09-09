package utils

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
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

func DrawCostRecordsTo(r Records, f, scale string) {
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

	for label, r := range labeledRecords {
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

func genGlyphStyleForLabel(labels []string) map[string]draw.GlyphStyle {
	sort.Strings(labels)
	styles := make(map[string]draw.GlyphStyle)
	shapes := map[string]draw.GlyphDrawer{
		"scan": draw.RingGlyph{},
		"agg":  draw.TriangleGlyph{},
		"join": draw.SquareGlyph{},
	}
	colorCnt := make(map[string]int)
	for _, l := range labels {
		l = strings.ToLower(l)
		ok := false
		for op, shape := range shapes {
			if strings.Contains(l, op) {
				cnt := colorCnt[op]
				currentColor := plotutil.DarkColors[(cnt % len(plotutil.DarkColors))]
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
