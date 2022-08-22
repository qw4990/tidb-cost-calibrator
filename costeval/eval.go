package costeval

import (
	"fmt"
	"path/filepath"

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
	costEval(ins, &evalOpt{"synthetic", 2, 2, 5})
}

type evalOpt struct {
	db           string
	costModelVer int
	repeatTimes  int
	numPerQuery  int
}

func (opt *evalOpt) genInitSQLs() []string {
	return []string{
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
		qs = genSYNQueries(ins, "synthetic", opt.numPerQuery)
		utils.SaveTo(queryFile, qs)
		info("generate %v queries successfully", len(qs))
	} else {
		info("read %v queries successfully", len(qs))
	}

	var rs utils.Records
	recordFile := filepath.Join(dataDir, fmt.Sprintf("%v-%v-records.json", opt.db, opt.costModelVer))
	if err := utils.ReadFrom(recordFile, &rs); err != nil {
		rs = runCostEvalQueries(ins, opt.db, qs, opt.InitSQLs(), opt.processRepeat, opt.processTimeLimitMS)
		utils.SaveTo(recordFile, rs)
		info("generate %v records successfully", len(rs))
	} else {
		info("read %v records successfully", len(rs))
	}
}

func info(format string, args ...interface{}) {
	fmt.Printf("[cost-eval] %v\n", fmt.Sprintf(format, args...))
}
