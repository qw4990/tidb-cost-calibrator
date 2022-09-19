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

	factors := map[string]float64{
		"tidb_opt_concurrency_factor_v2":        3,
		"tidb_opt_copcpu_factor_v2":             30,
		"tidb_opt_cpu_factor_v2":                30,
		"tidb_opt_tiflash_cpu_factor_v2":        5,
		"tidb_opt_hash_table_factor_v2":         1000,
		"tidb_opt_tiflash_hash_table_factor_v2": 5,
		"tidb_opt_scan_factor_v2":               20,
		"tidb_opt_desc_factor_v2":               150,
		"tidb_opt_network_factor_v2":            8,
		"tidb_opt_seek_factor_v2":               9500000,
		"tidb_opt_tiflash_scan_factor_v2":       5,
		"tidb_opt_memory_factor_v2":             0.001,
		"tidb_opt_disk_factor_v2":               1.5,
	}

	whiteList := []string{
		"Scan",
		"PhaseAgg",
	}

	var rs2 utils.Records
	for i := range rs {
		if strings.Contains(rs[i].Label, "TableScan") {
			factors["tidb_opt_network_factor_v2"] = 1.8
		} else if strings.Contains(rs[i].Label, "MPPScan") {
			factors["tidb_opt_network_factor_v2"] = 0.8
		} else {
			factors["tidb_opt_network_factor_v2"] = 0.8
		}

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
			cost += w * factors[k]
		}
		rs[i].Cost = cost
		rs2 = append(rs2, rs[i])
	}

	utils.DrawCostRecordsTo(rs2, "./data/calibrate.png", "linear")
}
