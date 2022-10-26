package costeval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/minio/minio-go/v7/pkg/credentials"
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
	costEval(ins, &evalOpt{"synthetic", 2, 3, 5, true, 1})
	//costEval(ins, &evalOpt{"synthetic", 1, 3, 5, true, 1})
	//costEval(ins, &evalOpt{"tpch_clustered", 2, 1, 5, true, 1})
	//costEval(ins, &evalOpt{"tpch_clustered", 1, 1, 5, true, 1})
}

type evalOpt struct {
	ceType       string
	dbName       string
	costModelVer int
	repeatTimes  int
	numPerQuery  int
	concurrent1  bool
	scaleFactor  float64
}

func (opt *evalOpt) name() string {
	return fmt.Sprintf("%v-%v-%v", opt.ceType, opt.costModelVer, opt.concurrent1)
}

func (opt *evalOpt) genInitSQLs() []string {
	return []string{
		fmt.Sprintf(`use %v`, opt.dbName),
		`set @@tidb_distsql_scan_concurrency=1`,
		`set @@tidb_executor_concurrency=1`,
		`set @@tidb_opt_tiflash_concurrency_factor=1`,
		`set @@tidb_enable_new_cost_interface=1`,
		fmt.Sprintf(`set @@tidb_cost_model_version=%v`, opt.costModelVer),
	}
}

func costEval(ins utils.Instance, opt *evalOpt, minioOption utils.MinioOption) {
	info("start cost evaluation, ceType=%v, ver=%v", opt.ceType, opt.costModelVer)
	var qs utils.Queries
	dataDir := "./data"
	if !utils.PathExist(dataDir) {
		if err := os.Mkdir(dataDir, 0755); err != nil {
			panic(err)
		}
	}
	queryFile := filepath.Join(dataDir, fmt.Sprintf("%v-queries.json", opt.ceType))
	if err := utils.ReadFrom(queryFile, &qs); err != nil {
		switch opt.ceType {
		case "synthetic":
			qs = genSYNQueries(opt.numPerQuery, opt.scaleFactor)
		case "tpch_clustered":
			qs = genTPCHQueries2(opt.numPerQuery, opt.scaleFactor)
		default:
			panic(fmt.Sprintf("unknown DB/Workload %v", opt.ceType))
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

	var tmp utils.Records
	for _, r := range rs {
		//if !strings.Contains(r.Label, "MPPScan") {
		//	continue
		//}
		tmp = append(tmp, r)
	}
	rs = tmp

	sort.Slice(rs, func(i, j int) bool {
		return rs[i].TimeMS < rs[j].TimeMS
	})

	for _, r := range rs {
		info("record %vms \t %.2f \t %v \t %v", r.TimeMS, r.Cost, r.Label, r.SQL)
	}
	utils.DrawCostRecordsTo(rs, fmt.Sprintf("./data/%v-scatter.png", opt.name()), "linear")

	if len(minioOption.Endpoint) != 0 {
		upload(minioOption)
	}
}

func upload(minioOption utils.MinioOption) {
	minioClient, err := minio.New(minioOption.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioOption.ID, minioOption.Secret, ""),
		Secure: false,
	})
	if err != nil {
		panic(err)
	}
	dir := "/cost-calibrator/" + time.Now().Format("2006-01-02-15-04") + "/"
	fs, err := os.ReadDir("./data")
	if err != nil {
		panic(err)
	}
	for _, f := range fs {
		if f.IsDir() {
			continue
		}

		object := filepath.Join(dir, f.Name())
		fp := filepath.Join("./data", f.Name())
		info("upload %s", object)
		if _, err := minioClient.FPutObject(context.Background(), "planner", object, fp, minio.PutObjectOptions{}); err != nil {
			panic(err)
		}
	}
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

			if k == 0 && opt.costModelVer == 2 { // parse factor weights
				weights = cleanCostWeights(parseCostWeights(ins))
			}
		}
		sort.Float64s(execTimes)
		info(">> exec time %v", ms2Dur(execTimes))
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
	// Warning | 1105 | factor weights: {"tidb_kv_net_factor":23472,"tidb_request_factor":1,"tikv_scan_factor":10854.458609780715}
	var warn, weightStr string
	var errno int
	for rs.Next() {
		var tmp string
		utils.Must(rs.Scan(&warn, &errno, &tmp))
		if strings.Contains(tmp, "factor weights:") {
			weightStr = tmp
		}
	}
	if weightStr == "" {
		panic("cannot get weights")
	}
	idx := strings.Index(weightStr, "weights:")
	data := strings.TrimSpace(weightStr[idx+len("weights:"):])
	utils.Must(json.Unmarshal([]byte(data), &weights))
	return weights
}

func cleanCostWeights(w map[string]float64) map[string]float64 {
	ret := make(map[string]float64)
	for k, v := range w {
		if v > 0 {
			ret[k] = v
		}
	}
	return ret
}

func info(format string, args ...interface{}) {
	fmt.Printf("[cost-eval] %v\n", fmt.Sprintf(format, args...))
}

func ms2Dur(times []float64) []time.Duration {
	var durs []time.Duration
	for _, t := range times {
		durs = append(durs, time.Millisecond*time.Duration(t))
	}
	return durs
}
