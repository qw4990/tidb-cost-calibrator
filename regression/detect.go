package regression

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/qw4990/tidb-cost-calibrator/utils"
)

func parseResultFile(rf string) (caseName string, execTime time.Duration, failReason string) {
	data, err := os.ReadFile(rf)
	utils.Must(err)
	lines := strings.Split(string(data), "\n")

	tikv, tiflash := false, false
	for _, l := range lines {
		if strings.Contains(l, "TIME=") {
			tmp := strings.Split(l, ",")
			caseName = strings.TrimSpace(tmp[0])
			tmp = strings.Split(tmp[2], "=")
			tStr := strings.TrimSpace(tmp[1])
			execTime, err = time.ParseDuration(tStr)
			utils.Must(err)
		}
		if strings.Contains(l, "FAIL") {
			failReason = "FAIL"
		}
		if strings.Contains(l, "TIMEOUT") {
			execTime = time.Second * 3
			fmt.Println(">>>> TIMEOUT ", rf)
			//failReason = "TIMEMOUT"
		}
		if strings.Contains(l, "tikv") {
			tikv = true
		}
		if strings.Contains(l, "tiflash") {
			tiflash = true
		}
	}

	if tikv && tiflash {
		//fmt.Println("---> mix plan >> ", caseName)
	}

	return
}

func planAnalyze(dir string, engines []string) {
	result := make(map[string]map[string]time.Duration) // engine:case-name:exec-time
	caseNames := make(map[string]struct{})
	for _, engine := range engines {
		result[engine] = make(map[string]time.Duration)
		resultFile := utils.ReadDirFiles(dir, fmt.Sprintf("-%v.result", engine))
		for _, rf := range resultFile {
			caseName, execTime, failReason := parseResultFile(rf)
			if failReason != "" {
				fmt.Println("FAIL CASE: ", caseName, engine, failReason)
				continue
			}
			caseNames[caseName] = struct{}{}
			result[engine][caseName] = execTime
		}
	}

	for caseName := range caseNames {
		if result["mix"][caseName] < result["ap"][caseName] && (float64(result["mix"][caseName])/float64(result["ap"][caseName]) < 0.85) {
			fmt.Println(">>> GOOD ", caseName, result["mix"][caseName], result["ap"][caseName])
		}
		if result["mix"][caseName] > result["ap"][caseName] && (float64(result["ap"][caseName])/float64(result["mix"][caseName]) < 0.85) {
			fmt.Println(">>> BAD ", caseName, result["mix"][caseName], result["ap"][caseName])
		}
	}

	totTime := make(map[string]time.Duration)
	numValidCases := 0
	for caseName := range caseNames {
		valid := true
		for _, engine := range engines {
			if _, ok := result[engine][caseName]; !ok {
				valid = false
			}
		}
		if !valid {
			fmt.Println("skip ", caseName)
			continue
		}
		numValidCases += 1
		for _, engine := range engines {
			totTime[engine] += result[engine][caseName]
		}
	}

	fmt.Println(numValidCases, totTime)
}

func GetPlans() {
	settings := []string{
		"set max_execution_time=180000,tidb_cost_model_version=2,tidb_isolation_read_engines='tidb,tikv'",
		//"set max_execution_time=180000,tidb_cost_model_version=2,tidb_isolation_read_engines='tidb,tiflash'",
		//"set max_execution_time=180000,tidb_cost_model_version=2,tidb_isolation_read_engines='tidb,tikv,tiflash'",
	}
	alias := []string{
		"tp",
		"ap",
		"mix",
	}
	workload := "tpcds"
	resultFileDir := fmt.Sprintf("regression/%v/plans/", workload)
	planAnalyze(resultFileDir, alias)
	return

	opt := utils.Option{
		Addr:     "tidb-1-peer",
		Port:     4000,
		User:     "root",
		Password: "",
		Label:    "",
	}
	ins := utils.MustConnectTo(opt)
	ins.SetLogThreshold(0)
	qs := getQueries("regression/tpcds/queries")
	ins.MustExec("use tpcds50")

	//l, r := 1, 99
	xs := []int{2, 4, 5, 11, 14, 18, 22, 23, 27, 36, 39, 59, 67, 70, 77, 80, 86}
	for _, i := range xs {
		q := qs[i]
		//for i, q := range qs {
		//	if i < l || i > r {
		//		continue
		//	}
		for k, setting := range settings {
			ins.MustExec(setting)
			fmt.Println(">>> run ", alias[k], i)
			p, t, err := getPlan(q, ins, true)
			fmt.Println(">>> ", alias[k], i, t, err)

			var result bytes.Buffer
			result.WriteString(fmt.Sprintf("query_%v, ENGINE=%v, TIME=%v\n\n", i, alias[k], t))
			if err != nil {
				if strings.Contains(err.Error(), "context canceled") ||
					strings.Contains(err.Error(), "Query execution was interrupted") {
					fmt.Println("timeout, wait 30s")
					time.Sleep(time.Second * 30)
					result.WriteString("TIMEOUT")
				} else {
					fmt.Println("fail ", err)
					result.WriteString("FAIL " + err.Error())
				}
			} else {
				result.WriteString(planStr(p))
			}

			resultFile := path.Join(resultFileDir, fmt.Sprintf("query_%v-%v.result", i, alias[k]))
			utils.Must(os.WriteFile(resultFile, result.Bytes(), 0666))
		}
	}
}

func RegDetect() {
	opt := utils.Option{
		Addr:     "127.0.0.1",
		Port:     4000,
		User:     "root",
		Password: "",
		Label:    "",
	}
	ins := utils.MustConnectTo(opt)
	ins.SetLogThreshold(0)
	var qs map[int]string
	workload := "tpcds"
	switch workload {
	case "tpch":
		qs = getQueries("regression/tpch/queries")
		ins.MustExec("use tpch")
		delete(qs, 15) // skip q15
	case "tpcds":
		qs = getQueries("regression/tpcds/queries")
		ins.MustExec("use tpcds50")
	default:
		panic(workload)
	}

	resultFileDir := fmt.Sprintf("regression/%v/result/", workload)
	utils.CleanDir(resultFileDir)
	settings := []string{
		"set tidb_cost_model_version=2,tidb_isolation_read_engines='tidb,tiflash'",
		"set tidb_cost_model_version=2,tidb_isolation_read_engines='tidb,tikv,tiflash'",
	}
	plans := make([]map[int]Plan, len(settings))
	for i, setting := range settings {
		ins.MustExec(setting)
		plans[i] = getPlans(qs, ins, false)
	}

	for no := range plans[0] {
		if len(settings) == 1 {
			// just print Plans
			p := plans[0][no]
			resultFile := path.Join(resultFileDir, fmt.Sprintf("q%v.plan", no))
			utils.Must(os.WriteFile(resultFile, []byte(planStr(p)), 0666))
		} else if len(settings) == 2 {
			// compare Plans
			p1 := plans[0][no]
			p2 := plans[1][no]
			same := cmpPlan(p1, p2)
			if same {
				resultFile := path.Join(resultFileDir, fmt.Sprintf("q%v.plan", no))
				utils.Must(os.WriteFile(resultFile, []byte("SAME\n"+planStr(p1)), 0666))
			} else {
				resultFile := path.Join(resultFileDir, fmt.Sprintf("q%v_d.plan", no))
				var result bytes.Buffer
				result.WriteString("DIFF\n")
				result.WriteString(">> " + settings[0] + "\n")
				result.WriteString(planStr(p1))
				result.WriteString("\n\n\n")
				result.WriteString(">> " + settings[1] + "\n")
				result.WriteString(planStr(p2))
				utils.Must(os.WriteFile(resultFile, result.Bytes(), 0666))
				fmt.Printf("DIFF q%v\n", no)
			}
		}
	}
}

func getPlans(qs map[int]string, db utils.Instance, analyze bool) map[int]Plan {
	plans := make(map[int]Plan)
	for no, q := range qs {
		fmt.Printf(">>> get q%v plan\n", no)
		plans[no], _, _ = getPlan(q, db, analyze)
	}
	return plans
}

func getQueries(dir string) map[int]string {
	q := make(map[int]string)
	for _, f := range utils.ReadDirFiles(dir, ".sql") {
		// f: q23.sql or 23.sql or query_23.sql
		base := strings.Split(path.Base(f), ".")[0]
		var numStr string
		if strings.HasPrefix(base, "query_") {
			numStr = base[len("query_"):]
		} else if base[0] == 'q' {
			numStr = base[1:]
		} else {
			numStr = base
		}
		num, err := strconv.Atoi(numStr)
		utils.Must(err)
		content, err := os.ReadFile(f)
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

func getPlan(query string, db utils.Instance, analyze bool) (Plan, time.Duration, error) {
	if !analyze {
		query = "explain format=verbose " + query
		begin := time.Now()
		rows, err := db.Query(query)
		cost := time.Since(begin)
		var p Plan
		for rows.Next() {
			var id, estRows, estCost, task, acc, op string
			utils.Must(rows.Scan(&id, &estRows, &estCost, &task, &acc, &op))
			p = append(p, []string{id, estRows, estCost, task, acc, op})
		}
		return p, cost, err
	} else {
		query = "explain analyze format=verbose " + query
		begin := time.Now()
		rows, err := db.Query(query)
		cost := time.Since(begin)
		var p Plan
		if err != nil {
			return p, cost, err
		}
		for rows.Next() {
			//| id | estRows | estCost | actRows | task | access object | execution info | operator info | memory  | disk |
			var id, estRows, estCost, actRows, task, acc, exec, op, mem, disk string
			utils.Must(rows.Scan(&id, &estRows, &estCost, &actRows, &task, &acc, &exec, &op, &mem, &disk))
			p = append(p, []string{id, estRows, estCost, actRows, task, acc, exec, op, mem, disk})
		}
		return p, cost, err
	}
}

func cmpPlan(q1, q2 Plan) (same bool) {
	if len(q1) != len(q2) {
		return false
	}
	for i := range q1 {
		op1 := strings.Split(q1[i][0], "_")[0]
		op2 := strings.Split(q2[i][0], "_")[0]
		if op1 != op2 {
			return false
		}
	}
	return true
}

func planStr(q Plan) string {
	var buf bytes.Buffer
	for _, row := range q {
		line := strings.Join(row, "\t")
		buf.WriteString(line + "\n")
	}
	return buf.String()
}
