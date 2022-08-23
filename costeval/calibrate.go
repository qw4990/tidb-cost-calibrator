package costeval

import (
	"path/filepath"

	"github.com/qw4990/tidb-cost-calibrator/utils"
)

func CostCalibrate() {
	var rs utils.Records
	dataDir := "./data"
	recordFile := filepath.Join(dataDir, "synthetic-2-records.json")
	utils.Must(utils.ReadFrom(recordFile, &rs))

	factors := map[string]float64{
		"tidb_opt_concurrency_factor_v2":        3,
		"tidb_opt_copcpu_factor_v2":             30,
		"tidb_opt_cpu_factor_v2":                30,
		"tidb_opt_tiflash_cpu_factor_v2":        5,
		"tidb_opt_hash_table_factor_v2":         100,
		"tidb_opt_tiflash_hash_table_factor_v2": 30,
		"tidb_opt_scan_factor_v2":               100,
		"tidb_opt_desc_factor_v2":               150,
		"tidb_opt_network_factor_v2":            8,
		"tidb_opt_seek_factor_v2":               9500000,
		"tidb_opt_tiflash_scan_factor_v2":       5,
		"tidb_opt_memory_factor_v2":             0.001,
		"tidb_opt_disk_factor_v2":               1.5,
	}

	for i := range rs {
		var cost float64
		for k, w := range rs[i].Weights {
			cost += w * factors[k]
		}
		rs[i].Cost = cost
	}

	utils.DrawCostRecordsTo(rs, "./data/calibrate.png")
}
