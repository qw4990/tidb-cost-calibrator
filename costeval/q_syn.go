package costeval

import (
	"github.com/qw4990/tidb-cost-calibrator/utils"
)

func genSYNQueries(n int) utils.Queries {
	var qs utils.Queries
	qs = append(qs, genSYNScan(n)...)
	qs = append(qs, genSYNAgg(n)...)
	qs = append(qs, genSYNJoin(n)...)
	qs = append(qs, genSYNSort(n)...)
	return qs
}

func genSYNSort(n int) utils.Queries {
	return gen4Templates([]template{
		{
			`select /*+ use_index(t, bc), read_from_storage(tikv[t]) */ b from t where # order by c`,
			[]tempitem{{"t", "b", 0, 2500000}},
			"Sort",
		},
		{`select /*+ use_index(t, bc), read_from_storage(tikv[t]) */ b from t where # order by c limit 1000`,
			[]tempitem{{"t", "b", 0, 5000000}},
			"TopN",
		},
	}, n)
}

func genSYNScan(n int) utils.Queries {
	return gen4Templates([]template{
		// TiKV Table Scan
		{
			`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a from t where #`,
			[]tempitem{{"t", "a", 0, 10000000}},
			"TableScan",
		},
		{
			`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a, c from t where #`,
			[]tempitem{{"t", "a", 0, 3000000}},
			"WideTableScan",
		},
		{
			`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a from t where # order by a desc`,
			[]tempitem{{"t", "a", 0, 6000000}},
			"DescTableScan",
		},

		// TiKV Index Scan
		{
			`select /*+ use_index(t, b), read_from_storage(tikv[t]) */ b from t where #`,
			[]tempitem{{"t", "b", 0, 10000000}},
			"IndexScan",
		},
		{
			`select /*+ use_index(t, bc), read_from_storage(tikv[t]) */ b, c from t where #`,
			[]tempitem{{"t", "b", 0, 3000000}},
			"DescIndexScan",
		},
		{
			`select /*+ use_index(t, b), read_from_storage(tikv[t]) */ b from t where # order by b desc`,
			[]tempitem{{"t", "b", 0, 6000000}},
			"DescIndexScan",
		},

		// TiKV Index Lookup
		{
			`select /*+ use_index(t, b), read_from_storage(tikv[t]) */ b, d from t where #`,
			[]tempitem{{"t", "b", 0, 250000}},
			"IndexLookup",
		},

		// MPP Scan
		{
			`select /*+ read_from_storage(tiflash[t]) */ a from t where #`,
			[]tempitem{{"t", "a", 0, 20000000}},
			"MPPScan",
		},
	}, n)
}

func genSYNAgg(n int) utils.Queries {
	return gen4Templates([]template{
		// TiKV StreamAgg
		{
			`select /*+ read_from_storage(tikv[t]), stream_agg(), use_index(t, b) */ sum(a) from t where #`,
			[]tempitem{{"t", "b", 0, 8000000}},
			"StreamAgg1",
		},
		{
			`select /*+ read_from_storage(tikv[t]), stream_agg(), use_index(t, b) */ sum(a) from t where # group by b`,
			[]tempitem{{"t", "b", 0, 5000000}},
			"StreamAgg2",
		},

		// TiKV HashAgg
		{
			`select /*+ read_from_storage(tikv[t]), use_index(t, b), hash_agg() */ count(1) from t where #`,
			[]tempitem{{"t", "b", 0, 10000000}},
			"HashAgg1",
		}, // agg without group-by
		{
			`select /*+ read_from_storage(tikv[t]), use_index(t, b), hash_agg() */ count(b), b from t where # group by b`,
			[]tempitem{{"t", "b", 0, 2200000}},
			"HashAgg2",
		}, // agg with group-by

		// MPP HashAgg
		{
			`select /*+ read_from_storage(tiflash[t]), mpp_tidb_agg() */ count(a) from t where #`,
			[]tempitem{{"t", "b", 0, 20000000}},
			"MPPTiDBAgg1",
		}, // mpp tidb agg without group-by
		{
			`select /*+ read_from_storage(tiflash[t]), mpp_tidb_agg() */ count(a), b from t where # group by b`,
			[]tempitem{{"t", "b", 0, 2000000}},
			"MPPTiDBAgg2",
		}, // mpp tidb agg with group-by
		{
			`select /*+ read_from_storage(tiflash[t]), mpp_1phase_agg() */ count(a), b from t where # group by b`,
			[]tempitem{{"t", "b", 0, 8000000}},
			"MPP1PhaseAgg",
		},
		{
			`select /*+ read_from_storage(tiflash[t]), mpp_2phase_agg() */ count(a), b from t where # group by b`,
			[]tempitem{{"t", "b", 0, 8000000}},
			"MPP2PhaseAgg",
		},
	}, n)
}

func genSYNJoin(n int) utils.Queries {
	return gen4Templates([]template{
		// TiKV Join
		{
			`select /*+ use_index(t1, b), use_index(t2, b), tidb_hj(t1, t2), read_from_storage(tikv[t1, t2]) */ t1.b, t2.b from t t1, t t2 where t1.b=t2.b and # and #`,
			[]tempitem{
				{"t1", "a", 0, 2500000},
				{"t2", "a", 0, 2500000},
			},
			"HashJoin",
		},
		{
			`select /*+ use_index(t1, b), use_index(t2, b), tidb_smj(t1, t2), read_from_storage(tikv[t1, t2]) */ t1.b, t2.b from t t1, t t2 where t1.b=t2.b and # and #`,
			[]tempitem{
				{"t1", "a", 0, 8000000},
				{"t2", "a", 0, 8000000},
			},
			"MergeJoin",
		},
		{
			`select /*+ TIDB_INLJ(t1, t2) */ t2.b from t t1, t t2 where t1.b = t2.b and # and #`,
			[]tempitem{
				{"t1", "b", 0, 500000},
				{"t2", "b", 0, 500000},
			},
			"IndexJoin",
		},

		// MPP Join
		{
			`SELECT /*+ read_from_storage(tiflash[t1, t2]), shuffle_join(t1, t2) */ t1.b, t2.b FROM t t1, t t2 WHERE t1.b=t2.b and # and #`,
			[]tempitem{
				{"t1", "a", 0, 5000000},
				{"t2", "a", 0, 5000000},
			},
			"ShuffleJoin",
		},
		{
			`SELECT /*+ read_from_storage(tiflash[t1, t2]), broadcast_join(t1) */ t1.b, t2.b FROM t t1, t t2 WHERE t1.b=t2.b and # and #`,
			[]tempitem{
				{"t1", "a", 0, 1000000},
				{"t2", "a", 5000000, 10000000},
			},
			"BroadcastJoin",
		},
	}, n)
}
