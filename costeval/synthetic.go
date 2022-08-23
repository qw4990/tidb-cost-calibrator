package costeval

import (
	"fmt"

	"github.com/qw4990/tidb-cost-calibrator/utils"
)

func genSYNQueries(ins utils.Instance, db string, n int) utils.Queries {
	ins.MustExec(fmt.Sprintf("use %v", db))

	var qs utils.Queries
	qs = append(qs, genSYNScans(ins, n)...)
	qs = append(qs, genSYNAgg(ins, n)...)
	return qs
}

func genSYNScans(ins utils.Instance, n int) utils.Queries {
	ps := []pattern{
		// TiKV Table Scan
		{`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a from t where %v`,
			"a", 20000000, "TableScan"},
		//{`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a, c from t where %v`,
		//	"a", 1000, "WideTableScan"},
		//{`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a from t where %v order by a desc`,
		//	"a", 1000, "DescTableScan"},
		//
		//// TiKV Index Scan
		//{`select /*+ use_index(t, b), read_from_storage(tikv[t]) */ b from t where %v`,
		//	"b", 1000, "IndexScan"},
		//{`select /*+ use_index(t, bc), read_from_storage(tikv[t]) */ b, c from t where %v`,
		//	"b", 1000, "DescIndexScan"},
		//{`select /*+ use_index(t, b), read_from_storage(tikv[t]) */ b from t where %v order by b desc`,
		//	"b", 1000, "DescIndexScan"},
		//
		//// MPP Scan
		//{`SELECT /*+ read_from_storage(tiflash[t]) */ a FROM t WHERE %v`,
		//	"a", 1000, "MPPScan"},
	}
	return gen4Patterns(ins, ps, n)
}

func genSYNAgg(ins utils.Instance, n int) utils.Queries {
	ps := []pattern{
		// TiKV StreamAgg
		{`select /*+ read_from_storage(tikv[t]), stream_agg(), use_index(t, b) */ sum(a) from t where %v`,
			"b", 20000000, "StreamAgg1"},
		{`select /*+ read_from_storage(tikv[t]), stream_agg(), use_index(t, b) */ sum(a) from t where %v`,
			"b", 20000000, "StreamAgg2"},

		// TiKV HashAgg
		{`select /*+ read_from_storage(tikv[t]), use_index(t, b), hash_agg() */ count(1) from t where %v`,
			"b", 15000000, "HashAgg1"}, // agg without group-by
		{`select /*+ read_from_storage(tikv[t]), use_index(t, b), hash_agg() */ count(b), b from t where %v group by b`,
			"b", 4000000, "HashAgg2"}, // agg with group-by

		// MPP HashAgg
		{`select /*+ read_from_storage(tiflash[t]), mpp_tidb_agg() */ count(a) from t where %v`,
			"b", 12000000, "MPPTiDBAgg1"}, // mpp tidb agg without group-by
		{`select /*+ read_from_storage(tiflash[t]), mpp_tidb_agg() */ count(a), b from t where %v group by b`,
			"b", 4000000, "MPPTiDBAgg2"}, // mpp tidb agg with group-by
		{`select /*+ read_from_storage(tiflash[t]), mpp_1phase_agg() */ count(a), b from t where %v group by b`,
			"b", 15000000, "MPP1PhaseAgg"},
		{`select /*+ read_from_storage(tiflash[t]), mpp_2phase_agg() */ count(a), b from t where %v group by b`,
			"b", 15000000, "MPP2PhaseAgg"},
	}
	return gen4Patterns(ins, ps, n)
}

func genSYNJoin(ins utils.Instance, n int) utils.Queries {
	// TODO:
	return nil
}
