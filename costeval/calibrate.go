package costeval

import (
	"fmt"
	"path/filepath"

	"github.com/qw4990/tidb-cost-calibrator/utils"
)

func CostCalibrate() {
	var rs utils.Records
	dataDir := "./data"
	recordFile := filepath.Join(dataDir, "tpch_clustered-2-true-records.json")
	utils.Must(utils.ReadFrom(recordFile, &rs))
	whiteList := []string{
		//"TableScan", "BCastJoin",
		"Agg", "Scan", "Sel", "HashJoin", "MergeJoin",
	}
	rs = filterByLabel(rs, whiteList)

	factors := map[string]float64{
		"tikv_scan_factor":       40.7,
		"tikv_desc_scan_factor":  61.05,
		"tiflash_scan_factor":    11.6,
		"tidb_cpu_factor":        49.9,
		"tikv_cpu_factor":        49.9,
		"tiflash_cpu_factor":     2.4,
		"tidb_kv_net_factor":     3.96,
		"tidb_flash_net_factor":  2.2,
		"tiflash_mpp_net_factor": 1,
		"tidb_mem_factor":        0.198,
		"tikv_mem_factor":        0.198,
		"tiflash_mem_factor":     0.05,
		"tidb_disk_factor":       198,
		"tidb_request_factor":    0,
	}
	updateCost(rs, factors)

	utils.DrawCostRecordsTo(rs, "./data/calibrate.png", "linear")
}

func updateCost(rs utils.Records, factors map[string]float64) {
	for i := range rs {
		ff := factors
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
