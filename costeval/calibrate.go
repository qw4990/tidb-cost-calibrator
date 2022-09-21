package costeval

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/qw4990/tidb-cost-calibrator/utils"
)

func CostCalibrate() {
	var rs utils.Records
	dataDir := "./data"
	recordFile := filepath.Join(dataDir, "tpch_clustered-2-true-records.json")
	utils.Must(utils.ReadFrom(recordFile, &rs))
	whiteList := []string{
		"",
	}
	rs = filterByLabel(rs, whiteList)

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
	updateCost(rs, factors)

	utils.DrawCostRecordsTo(rs, "./data/calibrate.png", "linear")
}

func updateCost(rs utils.Records, factors map[string]float64) {
	mppScanfactors := make(map[string]float64)
	for k, v := range factors {
		mppScanfactors[k] = v
	}
	mppScanfactors["tidb_opt_mpp_network_factor_v2"] = mppScanfactors["tidb_opt_network_factor_v2"] / 3

	for i := range rs {
		ff := factors
		if strings.Contains(rs[i].Label, "MPPScan") {
			ff = mppScanfactors
		}

		if rs[i].Weights == nil {
			fmt.Println(">>> skip >> ", rs[i].Label, rs[i].SQL)
		}
		var cost float64
		for k, w := range rs[i].Weights {
			if _, ok := ff[k]; !ok {
				panic(fmt.Sprintf("no %v", k))
			}
			cost += w * ff[k]
		}
		rs[i].Cost = cost
	}
}
