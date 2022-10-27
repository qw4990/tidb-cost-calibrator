package costeval

import "github.com/qw4990/tidb-cost-calibrator/utils"

func genSYNQueries2(n int, scale float64) utils.Queries { // baseline: 3s
	return genQueries(n, scale,
		genSYNKVScan2,  // tikv_scan, tikv_net factors
		genSYNKVDScan2, // tikv_desc_scan
		genSYNKVCPU2,   // tikv_cpu
		// tidb_cpu
		genSYNAgg, genSYNJoin, genSYNSort)
}

func genSYNKVScan2(n int, scale float64) utils.Queries {
	return gen4Templates([]template{
		{ // table scan
			`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a from t where #`,
			[]tempitem{{"t", "a", 0, 5000000}},
			"KVScan1",
		},
		{ // wide table scan
			`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a, c from t where #`,
			[]tempitem{{"t", "a", 0, 2000000}},
			"KVScan2",
		},
		{ // index scan
			`select /*+ use_index(t, b), read_from_storage(tikv[t]) */ b from t where #`,
			[]tempitem{{"t", "b", 0, 4500000}},
			"KVScan3",
		},
		{ // wide index scan
			`select /*+ use_index(t, bc), read_from_storage(tikv[t]) */ b, c from t where #`,
			[]tempitem{{"t", "b", 0, 2000000}},
			"KVScan4",
		},
	}, n, scale)
}

func genSYNKVDScan2(n int, scale float64) utils.Queries {
	return gen4Templates([]template{
		{ // desc table scan
			`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a from t where # order by a desc`,
			[]tempitem{{"t", "a", 0, 3500000}},
			"KVDScan1",
		},
		{ // wide desc table scan
			`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a, c from t where # order by a desc`,
			[]tempitem{{"t", "a", 0, 3500000}},
			"KVDScan2",
		},
		{ // desc index scan
			`select /*+ use_index(t, b), read_from_storage(tikv[t]) */ b from t where # order by b desc`,
			[]tempitem{{"t", "b", 0, 3500000}},
			"KVDScan3",
		},
		{ // wide desc index scan
			`select /*+ use_index(t, bc), read_from_storage(tikv[t]) */ b, c from t where # order by b desc`,
			[]tempitem{{"t", "b", 0, 3500000}},
			"KVDScan4",
		},
	}, n, scale)
}

func genSYNKVCPU2(n int, scale float64) utils.Queries {
	return gen4Templates([]template{
		{
			`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a from t where # and b<0`,
			[]tempitem{{"t", "a", 0, 5000000}},
			"KVCPU1",
		},
		{
			`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a from t where # and b>0 and b+1>0 and b+2>0 and d<0`,
			[]tempitem{{"t", "a", 0, 5000000}},
			"KVCPU2",
		},
		{
			`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ sum(a) from t where #`,
			[]tempitem{{"t", "a", 0, 5000000}},
			"KVCPU3",
		},
		{
			`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ sum(a), avg(a), sum(b), avg(b) from t where #`,
			[]tempitem{{"t", "a", 0, 5000000}},
			"KVCPU3",
		},
	}, n, scale)
}

func genSYNDBCPU2(n int, scale float64) utils.Queries {
	return gen4Templates([]template{
		{ // table scan
			`select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ a from t where #`,
			[]tempitem{{"t", "a", 0, 5000000}},
			"KVScan1",
		},
	}, n, scale)
}

// select /*+ use_index(t, primary), read_from_storage(tikv[t]) */ sum(a) from t where a<1000
