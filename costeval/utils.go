package costeval

import (
	"fmt"
	"math"
	"strings"

	"github.com/qw4990/tidb-cost-calibrator/utils"
)

type tempitem struct {
	table  string
	column string
	minVal int
	maxVal int
}

type template struct {
	sql   string     // `select * from t where # and #`
	items []tempitem // [{t, a, 0, 100}, {t, b, 0, 10000}]
	label string     // `TableScan`
}

func genQueries(n int, scale float64, geners ...func(n int, scale float64) utils.Queries) (qs utils.Queries) {
	for _, g := range geners {
		qs = append(qs, g(n, scale)...)
	}
	return
}

func gen4Templates(ts []template, n int, scale float64) utils.Queries {
	var qs utils.Queries
	for _, t := range ts {
		qs = append(qs, gen4Template(t, n, scale)...)
	}
	return qs
}

func gen4Template(t template, n int, scale float64) utils.Queries {
	var qs utils.Queries
	for i := 0; i < n; i++ {
		sql := t.sql
		for _, item := range t.items {
			maxVal := int(scale * float64(item.maxVal))
			l, r := randRangeLinear(item.minVal, maxVal, i, n)
			var cond string
			if item.table != "" {
				cond = fmt.Sprintf("%v.%v>=%v and %v.%v<=%v",
					item.table, item.column, l, item.table, item.column, r)
			} else {
				cond = fmt.Sprintf("%v>=%v and %v<=%v", item.column, l, item.column, r)
			}
			sql = strings.Replace(sql, "#", cond, 1)
		}
		qs = append(qs, utils.Query{
			SQL:   sql,
			Label: t.label,
		})
	}
	return qs
}

func randRangeLog10(minVal, maxVal, iter, totalRepeat int) (int, int) {
	gap := float64(maxVal - minVal)
	rangeWidth := math.Pow(10, (math.Log10(gap)/float64(totalRepeat))*float64(iter+1))
	l := minVal
	r := l + int(rangeWidth)
	if r > maxVal {
		r = maxVal
	}
	return l, r
}

func randRangeLinear(minVal, maxVal, iter, totalRepeat int) (int, int) {
	step := (maxVal - minVal + 1) / totalRepeat
	l := 1
	r := step * (iter + 1)
	if r > maxVal {
		r = maxVal
	}
	return l, r
}

func filterByLabel(rs utils.Records, whiteList []string) utils.Records {
	var ret utils.Records
	for _, r := range rs {
		pass := false
		for _, label := range whiteList {
			if strings.Contains(r.Label, label) {
				pass = true
				break
			}
		}
		if pass {
			ret = append(ret, r)
		}
	}
	return ret
}

func scaleByLabel(rs utils.Records, scale map[string]int) utils.Records {
	var ret utils.Records
	for _, r := range rs {
		repeat := 1
		for k, v := range scale {
			if strings.Contains(r.Label, k) {
				repeat = v
				break
			}
		}
		for i := 0; i < repeat; i++ {
			ret = append(ret, r)
		}
	}
	return ret
}
