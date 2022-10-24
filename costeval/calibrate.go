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
		"",
	}
	rs = filterByLabel(rs, whiteList)

	factors := map[string]float64{
		"tikv_scan_factor":       100,
		"tikv_desc_scan_factor":  150,
		"tiflash_scan_factor":    5,
		"tidb_cpu_factor":        30,
		"tikv_cpu_factor":        30,
		"tiflash_cpu_factor":     5,
		"tidb_kv_net_factor":     8,
		"tidb_flash_net_factor":  4,
		"tiflash_mpp_net_factor": 4,
		"tidb_mem_factor":        1,
		"tikv_mem_factor":        1,
		"tiflash_mem_factor":     1,
		"tidb_disk_factor":       1000,
		"tidb_request_factor":    9500000,
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
