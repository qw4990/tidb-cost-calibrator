package ossinsight

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

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
	queryFiles := readDirFiles(queryDir, ".sql")
	for _, f := range queryFiles {
		data, err := os.ReadFile(f)
		utils.Must(err)
		plan := explain(db, string(data))
		planFile := strings.Replace(f, ".sql", "_plan.txt", 1)
		utils.Must(os.WriteFile(planFile, plan, 0666))
	}
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
