package utils

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ExplainAnalyzeResult struct {
	RootOperator    string
	PlanCost        float64
	TimeMS          float64
	OperatorActRows map[string]float64
	OperatorEstRows map[string]float64
	RawPlan         []string
}

func ParseExplainAnalyzeResultsWithRows(rs *sql.Rows) *ExplainAnalyzeResult {
	//+------------------------+----------+-----------+---------+-----------+---------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+-------------------------------------------------+---------+------+
	//| id                     | estRows  | estCost   | actRows | task      | access object | execution info                                                                                                                                                                                                                                              | operator info                                   | memory  | disk |
	//+------------------------+----------+-----------+---------+-----------+---------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+-------------------------------------------------+---------+------+
	//| TableReader_6          | 10000.00 | 52418.00  | 10000   | root      |               | time:9.93ms, loops:11, cop_task: {num: 1, max: 9.75ms, proc_keys: 10000, tot_proc: 5ms, tot_wait: 2ms, rpc_num: 1, rpc_time: 9.7ms, copr_cache: disabled}                                                                                                   | data:TableRangeScan_5                           | 78.5 KB | N/A  |
	//| └─TableRangeScan_5     | 10000.00 | 705020.00 | 10000   | cop[tikv] | table:t       | tikv_task:{time:5ms, loops:14}, scan_detail: {total_process_keys: 10000, total_process_keys_size: 270000, total_keys: 10500, rocksdb: {delete_skipped_count: 0, key_skipped_count: 10499, block: {cache_hit_count: 33, read_count: 0, read_byte: 0 Bytes}}} | range:[1,10000], keep order:false, stats:pseudo | N/A     | N/A  |
	//+------------------------+----------+-----------+---------+-----------+---------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+-------------------------------------------------+---------+------+
	var id, estRows, actRows, estCost, task, access, execInfo, opInfo, mem, disk string
	r := &ExplainAnalyzeResult{OperatorActRows: make(map[string]float64), OperatorEstRows: make(map[string]float64)}

	for rs.Next() {
		Must(rs.Scan(&id, &estRows, &estCost, &actRows, &task, &access, &execInfo, &opInfo, &mem, &disk))
		operator, id := parseOperatorName(id)
		r.OperatorEstRows[id] = mustStr2Float(estRows)
		r.OperatorActRows[id] = mustStr2Float(actRows)
		if r.RootOperator == "" {
			r.RootOperator = operator
			r.TimeMS = parseTimeFromExecInfo(execInfo)
			r.PlanCost = mustStr2Float(estCost)
		}
		r.RawPlan = append(r.RawPlan, strings.Join([]string{id, estRows, actRows, estCost, task, access, execInfo, opInfo, mem, disk}, "\t"))
	}
	Must(rs.Close())
	return r
}

func parseTimeFromExecInfo(execInfo string) (timeMS float64) {
	// time:252.5ms, loops:99, index_task: {total_time: 13
	timeField := strings.Split(execInfo, ",")[0]
	timeField = strings.Split(timeField, ":")[1]
	dur, err := time.ParseDuration(timeField)
	if err != nil {
		panic(fmt.Sprintf("invalid time %v", timeField))
	}
	return float64(dur) / float64(time.Millisecond)
}

func mustStr2Float(str string) float64 {
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func parseOperatorName(str string) (name, nameWithID string) {
	//	├─IndexRangeScan_5(Build)
	begin := 0
	for begin < len(str) {
		if str[begin] < 'A' || str[begin] > 'Z' { // not a upper letter
			begin++
		} else {
			break
		}
	}
	end0 := begin
	for end0 < len(str) {
		if str[end0] != '_' {
			end0++
		} else {
			break
		}
	}

	end1 := end0 + 1
	for end1 < len(str) {
		if str[end1] >= '0' && str[end1] <= '9' {
			end1++
		} else {
			break
		}
	}
	return str[begin:end0], str[begin:end1]
}
