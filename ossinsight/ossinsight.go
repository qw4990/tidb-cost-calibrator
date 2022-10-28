package ossinsight

import (
	"fmt"
	"os"
	"strings"

	"github.com/qw4990/tidb-cost-calibrator/utils"
)

func initSchema(db utils.Instance, schemaFile string) {
	fmt.Println("=============================== init schema =====================================")
	data, err := os.ReadFile(schemaFile)
	if err != nil {
		panic(err)
	}
	sqls := strings.Split(string(data), ";")
	for _, sql := range sqls {
		sql = strings.TrimSpace(sql)
		if len(sql) == 0 {
			continue
		}
		fmt.Println(sql)
		fmt.Println("------------------------------------------------------------------------")
		db.MustExec(sql)
	}
	fmt.Println("=============================== init schema end =====================================")
}

func importStats(db utils.Instance, statsFile string) {
	db.MustExec(fmt.Sprintf("load stats '%v'", statsFile))
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
		fs = append(fs, e.Name())
	}
	return fs
}

func explain(db utils.Instance, query string) (plan []string) {
	rows := db.MustQuery("explain " + query)
	for rows.Next() {
		utils.Must(rows.Scan())
		plan = append(plan, "")
	}
	return
}

func regression(db utils.Instance, queryDir string) {
	queryFiles := readDirFiles(queryDir, ".sql")
	for _, f := range queryFiles {
		data, err := os.ReadFile(f)
		utils.Must(err)
		plan := explain(db, string(data))
		planFile := strings.Replace(f, ".sql", "_plan.txt", 1)
		utils.Must(os.WriteFile(planFile, []byte(strings.Join(plan, "\n")), 0666))
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

	//initSchema(ins, "ossinsight/schema.sql")
	importStats(ins, "ossinsight/stats")
}
