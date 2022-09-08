package costeval

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/qw4990/tidb-cost-calibrator/utils"
)

func CostEval() {
	opt := utils.Option{
		Addr:     "172.16.5.173",
		Port:     4000,
		User:     "root",
		Password: "",
		Label:    "",
	}
	ins := utils.MustConnectTo(opt)
	//costEval(ins, &evalOpt{"synthetic", 2, 3, 5, true})
	costEval(ins, &evalOpt{"tpch_clustered", 2, 3, 5, true})
}

type evalOpt struct {
	db           string
	costModelVer int
	repeatTimes  int
	numPerQuery  int
	concurrent1  bool
}

func (opt *evalOpt) name() string {
	return fmt.Sprintf("%v-%v-%v", opt.db, opt.costModelVer, opt.concurrent1)
}

func (opt *evalOpt) genInitSQLs() []string {
	return []string{
		fmt.Sprintf(`use %v`, opt.db),
		`set @@tidb_distsql_scan_concurrency=1`,
		`set @@tidb_executor_concurrency=1`,
		`set @@tidb_opt_tiflash_concurrency_factor=1`,
		`set @@tidb_enable_new_cost_interface=1`,
		fmt.Sprintf(`set @@tidb_cost_model_version=%v`, opt.costModelVer),
	}
}

func costEval(ins utils.Instance, opt *evalOpt) {
	info("start cost evaluation, db=%v, ver=%v", opt.db, opt.costModelVer)
	var qs utils.Queries
	dataDir := "./data"
	queryFile := filepath.Join(dataDir, fmt.Sprintf("%v-queries.json", opt.db))
	if err := utils.ReadFrom(queryFile, &qs); err != nil {
		switch opt.db {
		case "synthetic":
			qs = genSYNQueries(opt.numPerQuery)
		case "tpch_clustered":
			qs = genTPCHQueries2(opt.numPerQuery)
		default:
			panic(fmt.Sprintf("unknown DB/Workload %v", opt.db))
		}
		utils.SaveTo(queryFile, qs)
		info("generate %v queries successfully", len(qs))
	} else {
		info("read %v queries successfully", len(qs))
	}

	for i := range qs {
		qs[i].SQL = `explain analyze format='true_card_cost' ` + qs[i].SQL
	}

	var rs utils.Records
	recordFile := filepath.Join(dataDir, fmt.Sprintf("%v-records.json", opt.name()))
	if err := utils.ReadFrom(recordFile, &rs); err != nil {
		rs = runEvalQueries(ins, opt, qs)
		utils.SaveTo(recordFile, rs)
		info("generate %v records successfully", len(rs))
	} else {
		info("read %v records successfully", len(rs))
	}

	sort.Slice(rs, func(i, j int) bool {
		return rs[i].TimeMS < rs[j].TimeMS
	})

	for _, r := range rs {
		info("record %vms \t %.2f \t %v \t %v", r.TimeMS, r.Cost, r.Label, r.SQL)
	}
	utils.DrawCostRecordsTo(rs, fmt.Sprintf("./data/%v-scatter.png", opt.name()))
}

func runEvalQueries(ins utils.Instance, opt *evalOpt, qs utils.Queries) utils.Records {
	for _, sql := range opt.genInitSQLs() {
		ins.MustExec(sql)
		fmt.Printf("%v;\n", sql)
	}

	var rs utils.Records
	beginAt := time.Now()
	for i, q := range qs {
		info("run %v/%v tot=%v, q=%v", i, len(qs), q.SQL, time.Since(beginAt))
		var cost float64
		weights := make(map[string]float64)
		var execTimes []float64
		for k := 0; k < opt.repeatTimes; k++ {
			rs := ins.MustQuery(q.SQL)
			r := utils.ParseExplainAnalyzeResultsWithRows(rs)
			if k == 0 {
				cost = r.PlanCost
			} else if cost != r.PlanCost { // the plan changes
				panic(fmt.Sprintf("q=%v, cost=%v, new-cost=%v", q.SQL, cost, r.PlanCost))
			}
			execTimes = append(execTimes, r.TimeMS)

			// TODO: support TPCH workload as well
			if k == 0 && opt.costModelVer == 2 && opt.db == "synthetic" { // parse factor weights
				weights = parseCostWeights(ins)
			}
		}
		sort.Float64s(execTimes)
		info(">> exec time %v", execTimes)
		t := execTimes[(len(execTimes)-1)/2]
		rs = append(rs, utils.Record{
			Cost:    cost,
			TimeMS:  t,
			Label:   q.Label,
			SQL:     q.SQL,
			Weights: weights,
		})
	}
	return rs
}

func parseCostWeights(ins utils.Instance) map[string]float64 {
	weights := make(map[string]float64)
	rs := ins.MustQuery("show warnings")
	defer func() {
		utils.Must(rs.Close())
	}()
	// | Warning | 1105 | valid cost trace result: tidb_opt_cpu_factor_v2=1.0000, tidb_opt_copcpu_factor_v2=0.0000, ... |
	var warn, weightStr string
	var errno int
	if !rs.Next() {
		panic("cannot get weights")
	}
	utils.Must(rs.Scan(&warn, &errno, &weightStr))
	if strings.Contains(weightStr, "invalid") {
		return nil
	}
	idx := strings.Index(weightStr, "result: ")
	if idx == -1 {
		panic("unexpected: " + weightStr)
	}
	tmp := strings.Split(weightStr[idx+len("result: "):], ",")
	for _, kv := range tmp {
		kvList := strings.Split(kv, "=")
		k := strings.TrimSpace(kvList[0])
		v, err := strconv.ParseFloat(strings.TrimSpace(kvList[1]), 64)
		if err != nil {
			panic("unexpected: " + weightStr)
		}
		weights[k] = v
	}
	return weights
}

func info(format string, args ...interface{}) {
	fmt.Printf("[cost-eval] %v\n", fmt.Sprintf(format, args...))
}
