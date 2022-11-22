package regression

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/qw4990/tidb-cost-calibrator/utils"
)

func RegDetect() {
	opt := utils.Option{
		Addr:     "127.0.0.1",
		Port:     4000,
		User:     "root",
		Password: "",
		Label:    "",
	}
	ins := utils.MustConnectTo(opt)
	qs := getTPCHQueries("regression/tpch/queries")
	ins.MustExec("use tpch")
	compare(qs, ins, false, "tidb,tikv")
}

func getTPCHQueries(dir string) []string {
	var qs []string
	qs = append(qs, "")
	for i := 1; i <= 22; i++ {
		fpath := path.Join(dir, fmt.Sprintf("%v.sql", i))
		content, err := ioutil.ReadFile(fpath)
		utils.Must(err)
		qs = append(qs, cleanQuery(string(content)))
	}
	qs[15] = "" // skip q15
	return qs
}

func cleanQuery(q string) string {
	lines := strings.Split(q, "\n")
	for i := range lines {
		if idx := strings.Index(lines[i], "--"); idx != -1 {
			lines[i] = lines[i][:idx]
		}
		lines[i] = strings.TrimSpace(lines[i])
	}
	var tmp []string
	for _, l := range lines {
		if len(l) > 0 {
			tmp = append(tmp, l)
		}
	}
	return strings.Join(tmp, "\n")
}

type Plan [][]string

func getPlan(query string, db utils.Instance) Plan {
	query = "explain format=verbose " + query
	rows := db.MustQuery(query)
	var p Plan
	for rows.Next() {
		var id, estRows, estCost, task, acc, op string
		utils.Must(rows.Scan(&id, &estRows, &estCost, &task, &acc, &op))
		p = append(p, []string{id, estRows, estCost, task, acc, op})
	}
	return p
}

func cmpPlan(q1, q2 Plan) (same bool) {
	if len(q1) != len(q2) {
		return false
	}
	for i := range q1 {
		if q1[i][0] != q2[i][0] {
			return false
		}
	}
	return true
}

func printPlan(q Plan) {
	for _, row := range q {
		line := strings.Join(row, "\t")
		fmt.Println(line)
	}
}

func compare(queries []string, db utils.Instance, analyze bool, engines string) {
	vPlans := make([][]Plan, 3)
	vPlans[1] = append(vPlans[1], nil)
	vPlans[2] = append(vPlans[2], nil)
	for _, costVer := range []int{1, 2} {
		db.MustExec(fmt.Sprintf("set @@tidb_isolation_read_engines='%v'", engines))
		db.MustExec(fmt.Sprintf("set @@tidb_cost_model_version=%v", costVer))

		for i := 1; i < len(queries); i++ {
			if queries[i] == "" { // skip
				vPlans[costVer] = append(vPlans[costVer], nil)
				continue
			}
			fmt.Printf("get plan for q%v on model%v\n", i, costVer)
			vPlans[costVer] = append(vPlans[costVer], getPlan(queries[i], db))
		}
	}

	for i := 1; i < len(queries); i++ {
		fmt.Printf("q%v ============================================================= \n", i)
		if vPlans[1][i] == nil {
			fmt.Printf("SKIP: q%v\n", i)
			continue
		}
		if same := cmpPlan(vPlans[1][i], vPlans[2][i]); same {
			//fmt.Println("SAME: ", queries[i])
			//printPlan(vPlans[1][i])
		} else {
			fmt.Printf("DIFF: q%v\n", i)
			printPlan(vPlans[1][i])
			fmt.Println("-------------------------------------")
			printPlan(vPlans[2][i])
			fmt.Println("\n\n\n")
		}
	}
}
