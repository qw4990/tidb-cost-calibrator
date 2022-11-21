package ossinsight

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/qw4990/tidb-cost-calibrator/utils"
)

func initSchema(db utils.Instance, schemaDir string) {
	fmt.Println("=============================== init schema =====================================")
	db.MustExec("CREATE DATABASE IF NOT EXISTS gharchive_dev")
	db.MustExec("USE gharchive_dev")

	schemaFiles := readDirFiles(schemaDir, ".sql")
	sort.Slice(schemaFiles, func(i, j int) bool {
		if strings.Contains(schemaFiles[i], "tiflash_replicas") {
			return false
		}
		return true
	})
	for _, f := range schemaFiles {
		data, err := os.ReadFile(f)
		utils.Must(err)
		sqls := strings.Split(string(data), ";") // multiple SQLs
		for _, sql := range sqls {
			sql = strings.TrimSpace(sql)
			if sql == "" {
				continue
			}
			fmt.Println(sql)
			fmt.Println("------------------------------------------------------------------------")
			db.MustExec(sql)
		}
	}
	fmt.Println("=============================== init schema end =====================================")
}

func importStats(db utils.Instance, statsDir string) {
	fmt.Println("=============================== import stats =====================================")
	statsFiles := readDirFiles(statsDir, ".json")
	for _, f := range statsFiles {
		fmt.Println(f)
		mysql.RegisterLocalFile(f)
		db.MustExec(fmt.Sprintf("load stats '%v'", f))
	}
	fmt.Println("=============================== import stats end =====================================")
}

func readDirFiles(dir, suffix string) []string {
	entries, err := os.ReadDir(dir)
	utils.Must(err)

	var fs []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), suffix) {
			continue
		}
		fs = append(fs, path.Join(dir, e.Name()))
	}
	return fs
}

func explain(db utils.Instance, query string) string {
	rows := db.MustQuery("explain " + query)
	return utils.BeautifulPlan(utils.ParseExplain(rows))
}

func regression(db utils.Instance, queryDir string) {
	db.MustExec("USE gharchive_dev")
	db.MustExec("SET tidb_cost_model_version=2")
	queryFiles := readDirFiles(queryDir, ".sql")
	for _, f := range queryFiles {
		data, err := os.ReadFile(f)
		utils.Must(err)
		plan := explain(db, string(data))
		planFile := strings.Replace(f, ".sql", "_plan.txt", 1)
		utils.Must(os.WriteFile(planFile, []byte(plan), 0666))
	}
}

func filter(keys, white, black []string) []string {
	var ret []string
	for _, k := range keys {
		ok := false
		for _, w := range white {
			if strings.Contains(k, w) {
				ok = true
			}
		}
		for _, b := range black {
			if strings.Contains(k, b) {
				ok = false
			}
		}
		if ok {
			ret = append(ret, k)
		}
	}
	return ret
}

func Regression() {
	opt := utils.Option{
		Addr:     "127.0.0.1",
		Port:     4000,
		User:     "root",
		Password: "",
		Label:    "",
	}
	ins := utils.MustConnectTo(opt)

	//initSchema(ins, "ossinsight/schema")
	//importStats(ins, "ossinsight/stats")
	regression(ins, "ossinsight/query")
}

// Benchmark executes all selected OSSInsight queries and collects their execution times.
func Benchmark() {
	bench, explainAnalyze, analyzeResult := true, false, false
	benchQueryDir := "/Users/zhangyuanjia/Workspace/go/src/github.com/qw4990/tidb-cost-calibrator/ossinsight/bench_query"
	engines := []string{"mix", "ap"}

	if bench {
		opt := utils.Option{}
		ins := utils.MustConnectTo(opt)
		defer func() {
			utils.Must(ins.Close())
		}()
		ins.SetLogThreshold(0)

		queryFiles := readDirFiles(benchQueryDir, ".sql")
		whiteKeys := []string{"trending-repos-past-3-months-All"}
		blackKeys := []string{"collection"}
		queryFiles = filter(queryFiles, whiteKeys, blackKeys)

		ins.MustExec("use gharchive_dev")
		ins.MustExec("set tidb_cost_model_version=2")
		ins.MustExec("set tidb_stats_load_sync_wait=9999")
		ins.MustExec("set max_execution_time=180000")
		for _, queryFile := range queryFiles {
			benchmark(queryFile, ins, engines, explainAnalyze)
		}
	}

	if analyzeResult {
		benchanalyze(benchQueryDir, engines)
	}
}

func benchanalyze(dir string, engines []string) {
	result := make(map[string]map[string]time.Duration) // engine:case-name:exec-time
	caseNames := make(map[string]struct{})
	for _, engine := range engines {
		result[engine] = make(map[string]time.Duration)
		resultFile := readDirFiles(dir, fmt.Sprintf("-%v.result", engine))
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

	// mix = min(tp, ap, mix)
	for caseName := range caseNames {
		if _, ok := result["mix"][caseName]; !ok {
			continue
		}
		minT := result["mix"][caseName]
		for _, engine := range engines {
			if t, ok := result[engine][caseName]; ok && t < minT {
				minT = t
			}
		}
		result["mix"][caseName] = minT
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
		fmt.Println("---> mix plan >> ", caseName)
	}

	return
}

func benchmark(queryFile string, ins utils.Instance, engines []string, explainAnalyze bool) {
	_, fileName := filepath.Split(queryFile)
	caseName := strings.Split(fileName, ".")[0]

	for _, engineType := range engines {
		fmt.Println(">>> ", queryFile, engineType)
		switch engineType {
		case "mix":
			ins.MustExec("set @@tidb_isolation_read_engines='tidb,tikv,tiflash'")
		case "ap":
			ins.MustExec("set @@tidb_isolation_read_engines='tidb,tiflash'")
		case "tp":
			ins.MustExec("set @@tidb_isolation_read_engines='tidb,tikv'")
		default:
			panic(engineType)
		}

		specifiedFile := strings.Replace(queryFile, ".sql", fmt.Sprintf("-%v.sql", engineType), 1)
		if utils.PathExist(specifiedFile) {
			fmt.Println("use ", specifiedFile)
			queryFile = specifiedFile
		}

		data, err := os.ReadFile(queryFile)
		utils.Must(err)
		var q string
		if explainAnalyze {
			q = "explain analyze " + string(data)
		} else {
			q = "explain format=verbose " + string(data)
		}

		begin := time.Now()
		rs, err := ins.Query(q)
		execTime := time.Since(begin)
		fmt.Println(execTime)
		var plan string
		if err != nil {
			if strings.Contains(err.Error(), "context canceled") ||
				strings.Contains(err.Error(), "Query execution was interrupted") {
				fmt.Println("timeout, wait 20s")
				time.Sleep(time.Second * 20)
				plan = "TIMEOUT"
			} else {
				fmt.Println("fail ", err)
				plan = "FAIL " + err.Error()
			}
			fmt.Println(plan)
		} else {
			for rs.Next() {
				if explainAnalyze {
					//| id | estRows | actRows | task | access object | execution info | operator info | memory | disk |
					var id, est, act, task, obj, exe, op, mem, disk string
					utils.Must(rs.Scan(&id, &est, &act, &task, &obj, &exe, &op, &mem, &disk))
					plan = fmt.Sprintf("%v\n%v", plan, strings.Join([]string{id, est, act, task, obj, exe, op, mem, disk}, "\t"))
				} else {
					//| id | estRows | cost | task | access object | operator info |
					var id, est, cost, task, obj, op string
					utils.Must(rs.Scan(&id, &est, &cost, &task, &obj, &op))
					plan = fmt.Sprintf("%v\n%v", plan, strings.Join([]string{id, est, cost, task, obj, op}, "\t"))

				}
			}
			utils.Must(rs.Close())
		}

		// save result
		outputFile := strings.Replace(queryFile, ".sql", fmt.Sprintf("-%v.result", engineType), 1)
		result := fmt.Sprintf("%v, ENGINE=%v, TIME=%v\n\n", caseName, engineType, execTime)
		result += plan
		utils.Must(os.WriteFile(outputFile, []byte(result), 0666))
	}
}
