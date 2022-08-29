package costeval

import (
	"fmt"
	"strings"

	"github.com/qw4990/tidb-cost-calibrator/utils"
)

type pattern struct {
	sql      string
	rangeCol string
	rangeMax int
	label    string
}

func gen4Patterns(ins utils.Instance, ps []pattern, n int) utils.Queries {
	var qs utils.Queries
	for _, p := range ps {
		qs = append(qs, gen4Pattern(ins, p, n)...)
	}
	return qs
}

func gen4Pattern(ins utils.Instance, p pattern, n int) utils.Queries {
	var qs utils.Queries
	minV, maxV := getColRange4Pattern(ins, p)
	for i := 0; i < n; i++ {
		l, r := randRange(minV, maxV, i, n)

		// handle joins specially
		cond := fmt.Sprintf("%v>=%v and %v<=%v", p.rangeCol, l, p.rangeCol, r)
		if strings.Contains(p.label, "Join") ||
			strings.Contains(p.label, "HJ") ||
			strings.Contains(p.label, "BCJ") {
			cond = fmt.Sprintf("t1.%v>=%v and t1.%v<=%v and t2.%v>=%v and t2.%v<=%v",
				p.rangeCol, l, p.rangeCol, r, p.rangeCol, l, p.rangeCol, r)
		}

		qs = append(qs, utils.Query{
			SQL:   fmt.Sprintf(p.sql, cond),
			Label: p.label,
		})
	}
	return qs
}

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

func gen4Templates(ts []template, n int) utils.Queries {
	var qs utils.Queries
	for _, t := range ts {
		qs = append(qs, gen4Template(t, n)...)
	}
	return qs
}

func gen4Template(t template, n int) utils.Queries {
	var qs utils.Queries
	for i := 0; i < n; i++ {
		sql := t.sql
		for _, item := range t.items {
			l, r := randRange(item.minVal, item.maxVal, i, n)
			cond := fmt.Sprintf("%v.%v>=%v and %v.%v<=%v",
				item.table, item.column, l, item.table, item.column, r)
			sql = strings.Replace(sql, "#", cond, 1)
		}
		qs = append(qs, utils.Query{
			SQL:   sql,
			Label: t.label,
		})
	}
	return qs
}

func getColRange4Pattern(ins utils.Instance, p pattern) (l, r int) {
	utils.MustReadOneLine(ins, fmt.Sprintf(`select min(%v), max(%v) from %v`, p.rangeCol, p.rangeCol, "t"), &l, &r)
	if r-l > p.rangeMax {
		r = l + p.rangeMax
	}
	return
}

func randRange(minVal, maxVal, iter, totalRepeat int) (int, int) {
	step := (maxVal - minVal + 1) / totalRepeat
	l := 1
	r := step * (iter + 1)
	if r > maxVal {
		r = maxVal
	}
	return l, r
}
