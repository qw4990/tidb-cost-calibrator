package costeval

import "github.com/qw4990/tidb-cost-calibrator/utils"

func genTPCHQueries2(n int) utils.Queries {
	var qs utils.Queries
	qs = append(qs, genTPCHScan(n)...)
	qs = append(qs, genTPCHAgg(n)...)
	qs = append(qs, genTPCHJoin(n)...)
	qs = append(qs, genTPCHSort(n)...)
	return qs
}

func genTPCHScan(n int) utils.Queries {
	return gen4Templates([]template{
		// TiKV Table Scan
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ l_orderkey from lineitem where # `,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`TableScan1`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ l_orderkey, l_partkey, l_suppkey, l_linenumber, ` +
				`l_quantity, l_extendedprice, l_discount, l_tax, l_shipdate from lineitem where # `,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`TableScan2`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ * from lineitem where # `,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`TableScan3`,
		},
		// MPP Scan
		{
			`select /*+ read_from_storage(tiflash[lineitem]) */ l_orderkey from lineitem where # `,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPPScan1`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]) */ l_orderkey, l_partkey, l_suppkey, l_linenumber, ` +
				`l_quantity, l_extendedprice, l_discount, l_tax, l_shipdate from lineitem where # `,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPPScan2`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]) */ * from lineitem where # `,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPPScan3`,
		},
	}, n)
}

func genTPCHAgg(n int) utils.Queries {
	return gen4Templates([]template{
		// TiDB Agg
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`HashAgg1`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`HashAgg2`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`HashAgg3`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`HashAgg4`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`HashAgg5`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`HashAgg6`,
		},
		// TiDB StreamAgg
		{
			`select /*+ read_from_storage(tikv[lineitem]), stream_agg() */ count(*) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`StreamAgg1`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), stream_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`StreamAgg2`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), stream_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`StreamAgg3`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), stream_agg() */ count(*) from lineitem where # group by l_orderkey`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`StreamAgg4`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), stream_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where # group by l_orderkey`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`StreamAgg5`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), stream_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where # group by l_orderkey`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`StreamAgg6`,
		},
		// MPP Agg
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_1phase_agg() */ count(*) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPP1PhaseAgg1`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_1phase_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPP1PhaseAgg2`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_1phase_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPP1PhaseAgg3`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_2phase_agg() */ count(*) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPP2PhaseAgg1`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_2phase_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPP2PhaseAgg2`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_2phase_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPP2PhaseAgg3`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_tidb_agg() */ count(*) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPPTiDBAgg1`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_tidb_agg() */ count(*) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPPTiDBAgg2`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_tidb_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPPTiDBAgg3`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]), mpp_tidb_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPPTiDBAgg4`,
		},
	}, n)
}

func genTPCHJoin(n int) utils.Queries {
	return gen4Templates([]template{
		// HashJoin
		{
			`select /*+ read_from_storage(tikv[lineitem, part]), hash_join(lineitem, part) */ l_orderkey, p_partkey from lineitem, part where l_partkey = p_partkey and # and #`,
			[]tempitem{
				{"", "l_orderkey", 1, 10},
				{"", "p_partkey", 1, 10},
			},
			`HashJoin1`,
		},
		{
			`select /*+ read_from_storage(tikv[customer, orders]), hash_join(customer, orders) */ c_custkey, o_orderkey from customer, orders where c_custkey = o_custkey where # and #`,
			[]tempitem{
				{"", "c_custkey", 1, 10},
				{"", "o_orderkey", 1, 10},
			},
			`HashJoin2`,
		},
		// MergeJoin
		{
			`select /*+ read_from_storage(tikv[lineitem, part]), merge_join(lineitem, part) */ l_orderkey, p_partkey from lineitem, part where l_partkey = p_partkey and # and #`,
			[]tempitem{
				{"", "l_orderkey", 1, 10},
				{"", "p_partkey", 1, 10},
			},
			`MergeJoin`,
		},
		// IndexJoin
		{
			`select /*+ read_from_storage(tikv[lineitem, part]), tidb_inlj(lineitem, part) */ l_orderkey, p_partkey from lineitem, part where l_partkey = p_partkey and # and #`,
			[]tempitem{
				{"", "l_orderkey", 1, 10},
				{"", "p_partkey", 1, 10},
			},
			`MergeJoin`,
		},
		// ShuffleJoin
		{
			`select /*+ read_from_storage(tiflash[lineitem, part]), shuffle_join(lineitem, part) */ l_orderkey, p_partkey from lineitem, part where l_partkey = p_partkey and # and #`,
			[]tempitem{
				{"", "l_orderkey", 1, 10},
				{"", "p_partkey", 1, 10},
			},
			`ShuffleJoin1`,
		},
		{
			`select /*+ read_from_storage(tiflash[customer, orders]), shuffle_join(customer, orders) */ c_custkey, o_orderkey from customer, orders where c_custkey = o_custkey where # and #`,
			[]tempitem{
				{"", "c_custkey", 1, 10},
				{"", "o_orderkey", 1, 10},
			},
			`ShuffleJoin2`,
		},
		// BCastJoin
		{
			`select /*+ read_from_storage(tiflash[customer, nation]), broadcast_join(nation) */ c_custkey, n_nationkey from customer, nation where c_nationkey = n_nationkey and #`,
			[]tempitem{{"", "c_custkey", 1, 10}},
			`BCastJoin1`,
		},
		{
			`select /*+ read_from_storage(tiflash[supplier, nation]), broadcast_join(nation) */ * from supplier, nation where s_nationkey = n_nationkey and #`,
			[]tempitem{{"", "s_suppkey", 1, 10}},
			`BCastJoin2`,
		},
	}, n)
}

func genTPCHSort(n int) utils.Queries {
	return gen4Templates([]template{
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_shipdate`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`TiDBSort1`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`TiDBSort2`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_shipdate limit 100`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`TiDBTopN1`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_returnflag, l_linestatus limit 100`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`TiDBTopN2`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_shipdate`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPPSort1`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPPSort2`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_shipdate limit 100`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPPTopN1`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]) */ l_orderkey, l_shipmode from lineitem where # order by l_returnflag, l_linestatus limit 100`,
			[]tempitem{{"", "l_orderkey", 1, 10}},
			`MPPTopN2`,
		},
	}, n)
}
