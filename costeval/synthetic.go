package costeval

import (
	"fmt"

	"github.com/qw4990/tidb-cost-calibrator/utils"
)

func genSYNQueries(ins utils.Instance, db string, n int) utils.Queries {
	ins.MustExec(fmt.Sprintf("use %v", db))

	var qs utils.Queries
	qs = append(qs, genSYNAgg(ins, n)...)
	return qs
}

func genSYNScans(ins utils.Instance, n int) utils.Queries {
	// TODO:
	return nil
}

func genSYNAgg(ins utils.Instance, n int) utils.Queries {
	ps := []pattern{
		// TiKV SA
		{`select /*+ read_from_storage(tikv[t]), stream_agg(), use_index(t, b) */ sum(a) from t where %v`,
			"b", 2000000, "StreamAgg1"},
		{`select /*+ read_from_storage(tikv[t]), stream_agg(), use_index(t, b) */ sum(a) from t where %v`,
			"b", 2000000, "StreamAgg2"},

		// TiKV HA
		{`select /*+ read_from_storage(tikv[t]), use_index(t, b), hash_agg() */ count(1) from t where %v`,
			"b", 10000000, "HashAgg1"}, // agg without group-by
		{`select /*+ read_from_storage(tikv[t]), use_index(t, b), hash_agg() */ count(b), b from t where %v group by b`,
			"b", 3000000, "HashAgg2"}, // agg with group-by

		// MPP HA
		{`select /*+ read_from_storage(tiflash[t]), mpp_tidb_agg() */ count(a) from t where %v`,
			"b", 3000000, "MPPTiDBAgg1"}, // mpp tidb agg without group-by
		{`select /*+ read_from_storage(tiflash[t]), mpp_tidb_agg() */ count(a), b from t where %v group by b`,
			"b", 3000000, "MPPTiDBAgg2"}, // mpp tidb agg with group-by
		{`select /*+ read_from_storage(tiflash[t]), mpp_1phase_agg() */ count(a), b from t where %v group by b`,
			"b", 3000000, "MPP1PhaseAgg"},
		{`select /*+ read_from_storage(tiflash[t]), mpp_2phase_agg() */ count(a), b from t where %v group by b`,
			"b", 3000000, "MPP2PhaseAgg"},
	}
	return gen4Patterns(ins, ps, n)
}

func genSYNJoin(ins utils.Instance, n int) utils.Queries {
	// TODO:
	return nil
}
