package costeval

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/qw4990/tidb-cost-calibrator/utils"
)

func CostCalibrate() {
	var rs utils.Records
	dataDir := "./data"
	recordFile := filepath.Join(dataDir, "tpch_clustered-2-true-records.json")
	utils.Must(utils.ReadFrom(recordFile, &rs))

	factors := map[string]float64{
		"tidb_opt_cpu_factor_v2":                30,
		"tidb_opt_copcpu_factor_v2":             30,
		"tidb_opt_tiflash_cpu_factor_v2":        5,
		"tidb_opt_hash_table_factor_v2":         10,
		"tidb_opt_cop_hash_table_factor_v2":     10,
		"tidb_opt_tiflash_hash_table_factor_v2": 5,
		"tidb_opt_scan_factor_v2":               20,
		"tidb_opt_tiflash_scan_factor_v2":       8,
		"tidb_opt_network_factor_v2":            8,
		"tidb_opt_mpp_network_factor_v2":        4,
	}

	whiteList := []string{
		"",
	}

	var rs2 utils.Records
	for i := range rs {
		pass := false
		for _, w := range whiteList {
			if strings.Contains(rs[i].Label, w) {
				pass = true
			}
		}
		if !pass {
			continue
		}

		if rs[i].Weights == nil {
			fmt.Println(">>> skip >> ", rs[i].Label, rs[i].SQL)
		}

		var cost float64
		for k, w := range rs[i].Weights {
			if _, ok := factors[k]; !ok {
				panic(fmt.Sprintf("no %v", k))
			}
			cost += w * factors[k]
		}
		rs[i].Cost = cost
		rs2 = append(rs2, rs[i])
	}

	sort.Slice(rs2, func(i, j int) bool {
		return (rs2[i].Cost / rs2[i].TimeMS) < (rs2[j].Cost / rs2[j].TimeMS)
	})
	for _, r := range rs2 {
		fmt.Println(r.Label, r.Cost, r.TimeMS)
	}

	utils.DrawCostRecordsTo(rs2, "./data/calibrate.png", "linear")
}
