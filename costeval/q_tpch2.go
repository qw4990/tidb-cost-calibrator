package costeval

import "github.com/qw4990/tidb-cost-calibrator/utils"

func genTPCHQueries2(n int, scale float64) utils.Queries {
	return genQueries(n, scale, genTPCHScan, genTPCHAgg)
	//return genQueries(n, scale, genTPCHScan, genTPCHAgg, genTPCHSort, genTPCHJoin)
}

func genTPCHScan(n int, scale float64) utils.Queries {
	return gen4Templates([]template{
		// TiKV Table Scan
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ l_orderkey from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 15000000}},
			`TableScan1`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ l_orderkey, l_partkey, l_suppkey, ` +
				`l_linenumber, l_quantity, l_extendedprice from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 6000000}},
			`TableScan2`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ l_orderkey, l_partkey, l_suppkey, ` +
				`l_linenumber, l_quantity, l_extendedprice, l_discount, l_tax, l_returnflag, ` +
				`l_linestatus, l_shipdate, l_commitdate from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 4000000}},
			`TableScan3`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ * from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 3000000}},
			`TableScan4`,
		},
		// MPP Scan
		{
			`select /*+ read_from_storage(tiflash[lineitem]) */ l_orderkey from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 80000000}},
			`MPPScan1`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]) */ l_orderkey, l_partkey, l_suppkey, ` +
				`l_linenumber, l_quantity, l_extendedprice from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 20000000}},
			`MPPScan2`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]) */ l_orderkey, l_partkey, l_suppkey, ` +
				`l_linenumber, l_quantity, l_extendedprice, l_discount, l_tax, l_returnflag, ` +
				`l_linestatus, l_shipdate, l_commitdate from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 10000000}},
			`MPPScan3`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]) */ * from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 7500000}},
			`MPPScan4`,
		},
	}, n, scale)
}

func genTPCHSel(n int, scale float64) utils.Queries {
	return gen4Templates([]template{
		// TiDB

		// TiFlash
		{
			`select /*+ */ `,
			nil,
			"TiFlashSel",
		},
	}, n, scale)
}

func genTPCHAgg(n int, scale float64) utils.Queries {
	return gen4Templates([]template{
		// TiDB Agg
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 15000000}},
			`HashAgg1`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 8000000}},
			`HashAgg2`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 7000000}},
			`HashAgg3`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 8000000}},
			`HashAgg4`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 6000000}},
			`HashAgg5`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 5000000}},
			`HashAgg6`,
		},
		// TiDB StreamAgg
		{
			`select /*+ read_from_storage(tikv[lineitem]), stream_agg() */ count(*) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 15000000}},
			`StreamAgg1`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), stream_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 9000000}},
			`StreamAgg2`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), stream_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 8000000}},
			`StreamAgg3`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), stream_agg() */ count(*) from lineitem where # group by l_orderkey`,
			[]tempitem{{"", "l_orderkey", 1, 15000000}},
			`StreamAgg4`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), stream_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where # group by l_orderkey`,
			[]tempitem{{"", "l_orderkey", 1, 8000000}},
			`StreamAgg5`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), stream_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where # group by l_orderkey`,
			[]tempitem{{"", "l_orderkey", 1, 6000000}},
			`StreamAgg6`,
		},
		// MPP Agg
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_1phase_agg() */ count(*) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 45000000}},
			`MPP1PhaseAgg1`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_1phase_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 35000000}},
			`MPP1PhaseAgg2`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_1phase_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 30000000}},
			`MPP1PhaseAgg3`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_2phase_agg() */ count(*) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 55000000}},
			`MPP2PhaseAgg1`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_2phase_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 50000000}},
			`MPP2PhaseAgg2`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_2phase_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 40000000}},
			`MPP2PhaseAgg3`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_tidb_agg() */ count(*) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 90000000}},
			`MPPTiDBAgg1`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_tidb_agg() */ count(*) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 60000000}},
			`MPPTiDBAgg2`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_tidb_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 50000000}},
			`MPPTiDBAgg3`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_tidb_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 50000000}},
			`MPPTiDBAgg4`,
		},
	}, n, scale)
}

func genTPCHJoin(n int, scale float64) utils.Queries {
	return gen4Templates([]template{
		// HashJoin
		{
			`select /*+ read_from_storage(tikv[lineitem, part]), hash_join(lineitem, part) */ l_orderkey, p_partkey from lineitem, part where l_partkey = p_partkey and # and #`,
			[]tempitem{
				{"", "l_orderkey", 1, 3000000},
				{"", "p_partkey", 1, 3000000},
			},
			`HashJoin1`,
		},
		{
			`select /*+ read_from_storage(tikv[customer, orders]), hash_join(customer, orders) */ c_custkey, o_orderkey from customer, orders where c_custkey = o_custkey and # and #`,
			[]tempitem{
				{"", "c_custkey", 1, 3000000},
				{"", "o_orderkey", 1, 3000000},
			},
			`HashJoin2`,
		},
		// MergeJoin
		{
			`select /*+ read_from_storage(tikv[lineitem, part]), merge_join(lineitem, part) */ l_orderkey, p_partkey from lineitem, part where l_partkey = p_partkey and # and #`,
			[]tempitem{
				{"", "l_orderkey", 1, 3000000},
				{"", "p_partkey", 1, 3000000},
			},
			`MergeJoin1`,
		},
		// ShuffleJoin
		{
			`select /*+ read_from_storage(tiflash[lineitem, part]), shuffle_join(lineitem, part) */ l_orderkey, p_partkey from lineitem, part where l_partkey = p_partkey and # and #`,
			[]tempitem{
				{"", "l_orderkey", 1, 25000000},
				{"", "p_partkey", 1, 25000000},
			},
			`ShuffleJoin1`,
		},
		{
			`select /*+ read_from_storage(tiflash[customer, orders]), shuffle_join(customer, orders) */ c_custkey, o_orderkey from customer, orders where c_custkey = o_custkey and # and #`,
			[]tempitem{
				{"", "c_custkey", 1, 3000000},
				{"", "o_orderkey", 1, 90000000},
			},
			`ShuffleJoin2`,
		},
		// BCastJoin
		{
			`select /*+ read_from_storage(tiflash[customer, nation]), broadcast_join(nation) */ c_custkey, n_nationkey from customer, nation where c_nationkey = n_nationkey and #`,
			[]tempitem{{"", "c_custkey", 1, 3000000}},
			`BCastJoin1`,
		},
		{
			`select /*+ read_from_storage(tiflash[supplier, nation]), broadcast_join(nation) */ * from supplier, nation where s_nationkey = n_nationkey and #`,
			[]tempitem{{"", "s_suppkey", 1, 200000}},
			`BCastJoin2`,
		},
	}, n, scale)
}

func genTPCHSort(n int, scale float64) utils.Queries {
	return gen4Templates([]template{
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_shipdate`,
			[]tempitem{{"", "l_orderkey", 1, 3000000}},
			`TiDBSort1`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 4000000}},
			`TiDBSort2`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_shipdate limit 100`,
			[]tempitem{{"", "l_orderkey", 1, 9000000}},
			`TiDBSort3`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_returnflag, l_linestatus limit 100`,
			[]tempitem{{"", "l_orderkey", 1, 8000000}},
			`TiDBSort4`,
		},
		// NOTE: not supported
		//{
		//	`select /*+ read_from_storage(tiflash[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_shipdate`,
		//	[]tempitem{{"", "l_orderkey", 1, 10}},
		//	`MPPSort1`,
		//},
		//{
		//	`select /*+ read_from_storage(tiflash[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_returnflag, l_linestatus`,
		//	[]tempitem{{"", "l_orderkey", 1, 10}},
		//	`MPPSort2`,
		//},
		//{
		//	`select /*+ read_from_storage(tiflash[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_shipdate limit 100`,
		//	[]tempitem{{"", "l_orderkey", 1, 10}},
		//	`MPPTopN1`,
		//},
		//{
		//	`select /*+ read_from_storage(tiflash[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_returnflag, l_linestatus limit 100`,
		//	[]tempitem{{"", "l_orderkey", 1, 10}},
		//	`MPPTopN2`,
		//},
	}, n, scale)
}
