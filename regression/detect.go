package regression

import (
	"fmt"
	"os"
	"path"
	"strconv"
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
	qs := getQueries("regression/tpch/queries")
	ins.MustExec("use tpch")
	delete(qs, 15) // skip q15
	compare(qs, ins, false, "tikv")
}

func getQueries(dir string) map[int]string {
	q := make(map[int]string)
	for _, f := range utils.ReadDirFiles(dir, ".sql") {
		// f: q23.sql
		base := strings.Split(f, ".")[0]
		numStr := base[1:]
		num, err := strconv.Atoi(numStr)
		utils.Must(err)
		content, err := os.ReadFile(path.Join(dir, f))
		utils.Must(err)
		q[num] = cleanQuery(string(content))
	}
	return q
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

func compare(qs map[int]string, db utils.Instance, analyze bool, engines string) {
	switch engines {
	case "tikv":
		engines = "tidb,tikv"
	case "tiflash":
		engines = "tidb,tiflash"
	case "hybrid":
		engines = "tidb,tikv,tiflash"
	default:
		panic(engines)
	}

	vPlans := make([]map[int]Plan, 0, 3)
	vPlans = append(vPlans, nil, make(map[int]Plan), make(map[int]Plan))
	for _, costVer := range []int{1, 2} {
		db.MustExec(fmt.Sprintf("set @@tidb_isolation_read_engines='%v'", engines))
		db.MustExec(fmt.Sprintf("set @@tidb_cost_model_version=%v", costVer))

		for no, q := range qs {
			fmt.Printf("get plan for q%v on model%v\n", no, costVer)
			vPlans[costVer][no] = getPlan(q, db)
		}
	}

	for no := range qs {
		fmt.Printf("q%v ============================================================= \n", no)
		if same := cmpPlan(vPlans[1][no], vPlans[2][no]); same {
			//fmt.Println("SAME: ", queries[i])
			//printPlan(vPlans[1][i])
		} else {
			fmt.Printf("DIFF: q%v\n", no)
			printPlan(vPlans[1][no])
			fmt.Println("-------------------------------------")
			printPlan(vPlans[2][no])
			fmt.Println("\n\n\n")
		}
	}
}
