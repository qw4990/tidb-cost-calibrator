package ossinsight

import (
	"fmt"
	"os"
	"strings"

	"github.com/qw4990/tidb-cost-calibrator/utils"
)

func initSchema(db utils.Instance, schemaFile string) {
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
		db.MustExec(sql)
	}
}

func importStats(db utils.Instance, statsFile string) {
	db.MustExec(fmt.Sprintf("load stats '%v'", statsFile))
}

func Regression() {
	opt := utils.Option{
		Addr:     "172.16.5.173",
		Port:     4000,
		User:     "root",
		Password: "",
		Label:    "",
	}
	ins := utils.MustConnectTo(opt)

	initSchema(ins, "")
	importStats(ins, "")
}
