package costeval

import "github.com/qw4990/tidb-cost-calibrator/utils"

//CREATE TABLE `t` (
//`a` int(11) NOT NULL,
//`b` int(11) DEFAULT NULL,
//`c` varchar(128) DEFAULT NULL,
//`d` int(11) DEFAULT NULL,
//PRIMARY KEY (`a`) /*T![clustered_index] CLUSTERED */,
//KEY `b` (`b`),
//KEY `bc` (`b`,`c`)
//)

func genSYNQueries(n int, scale float64) utils.Queries { // baseline: 3s
	return genQueries(n, scale, genSYNScan, genSYNAgg, genSYNJoin, genSYNSort)
}

func genSYNSort(n int, scale float64) utils.Queries {
	return gen4Templates([]template{
		{
			`select /*+ use_index(t, bc), read_from_storage(tikv[t]) */ b from t where # order by c`,
			[]tempitem{{"t", "b", 0, 1600000}},
			"Sort1",
		},
		{`select /*+ use_index(t, bc), read_from_storage(tikv[t]) */ b from t where # order by c limit 1000`,
			[]tempitem{{"t", "b", 0, 3000000}},
			"Sort2",
		},
	}, n, scale)
}

func genSYNScan(n int, scale float64) utils.Queries {
	return gen4Templates([]template{
		// TiKV Table Scan
		{
			`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a from t where #`,
			[]tempitem{{"t", "a", 0, 5000000}},
			"TableScan1",
		},
		{
			`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a, c from t where #`,
			[]tempitem{{"t", "a", 0, 2000000}},
			"TableScan2",
		},
		{
			`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a from t where # order by a desc`,
			[]tempitem{{"t", "a", 0, 3500000}},
			"TableScan3",
		},

		// TiKV Index Scan
		{
			`select /*+ use_index(t, b), read_from_storage(tikv[t]) */ b from t where #`,
			[]tempitem{{"t", "b", 0, 4500000}},
			"IndexScan1",
		},
		{
			`select /*+ use_index(t, bc), read_from_storage(tikv[t]) */ b, c from t where #`,
			[]tempitem{{"t", "b", 0, 2000000}},
			"IndexScan2",
		},
		{
			`select /*+ use_index(t, b), read_from_storage(tikv[t]) */ b from t where # order by b desc`,
			[]tempitem{{"t", "b", 0, 3500000}},
			"IndexScan3",
		},

		// TiKV Index Lookup
		{
			`select /*+ use_index(t, b), read_from_storage(tikv[t]) */ b, d from t where #`,
			[]tempitem{{"t", "b", 0, 140000}},
			"IndexLookup",
		},
	}, n, scale)
}

func genSYNAgg(n int, scale float64) utils.Queries {
	return gen4Templates([]template{
		// TiKV StreamAgg
		{
			`select /*+ read_from_storage(tikv[t]), stream_agg(), use_index(t, b) */ sum(a) from t where #`,
			[]tempitem{{"t", "b", 0, 4000000}},
			"StreamAgg1",
		},
		{
			`select /*+ read_from_storage(tikv[t]), stream_agg(), use_index(t, b) */ sum(a) from t where # group by b`,
			[]tempitem{{"t", "b", 0, 3500000}},
			"StreamAgg2",
		},

		// TiKV HashAgg
		{
			`select /*+ read_from_storage(tikv[t]), use_index(t, b), hash_agg() */ count(1) from t where #`,
			[]tempitem{{"t", "b", 0, 5000000}},
			"HashAgg1",
		}, // agg without group-by
		{
			`select /*+ read_from_storage(tikv[t]), use_index(t, b), hash_agg() */ count(b), b from t where # group by b`,
			[]tempitem{{"t", "b", 0, 1400000}},
			"HashAgg2",
		}, // agg with group-by
	}, n, scale)
}

func genSYNJoin(n int, scale float64) utils.Queries {
	return gen4Templates([]template{
		// TiKV Join
		{
			`select /*+ use_index(t1, b), use_index(t2, b), tidb_hj(t1, t2), read_from_storage(tikv[t1, t2]) */ t1.b, t2.b from t t1, t t2 where t1.b=t2.b and # and #`,
			[]tempitem{
				{"t1", "b", 0, 1400000},
				{"t2", "b", 0, 1400000},
			},
			"HashJoin",
		},
		{
			`select /*+ use_index(t1, b), use_index(t2, b), tidb_smj(t1, t2), read_from_storage(tikv[t1, t2]) */ t1.b, t2.b from t t1, t t2 where t1.b=t2.b and # and #`,
			[]tempitem{
				{"t1", "b", 0, 3500000},
				{"t2", "b", 0, 3500000},
			},
			"MergeJoin",
		},
		{
			`select /*+ TIDB_INLJ(t1, t2) */ t2.b from t t1, t t2 where t1.b = t2.b and # and #`,
			[]tempitem{
				{"t1", "b", 0, 220000},
				{"t2", "b", 0, 220000},
			},
			"IndexJoin",
		},
	}, n, scale)
}
