package costeval

import "github.com/qw4990/tidb-cost-calibrator/utils"

func genTPCHQueries2(n int) utils.Queries {
	return
}

func genTPCHScan(n int) utils.Queries {
	return gen4Templates([]template{
		// TiKV Table Scan
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ l_orderkey from lineitem where # `,
			[]tempitem{{"", "l_orderkey", 1, 1}},
			`TableScan1`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ l_orderkey, l_partkey, l_suppkey, l_linenumber, ` +
				`l_quantity, l_extendedprice, l_discount, l_tax, l_shipdate from lineitem where # `,
			[]tempitem{{"", "l_orderkey", 1, 1}},
			`TableScan2`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]) */ * from lineitem where # `,
			[]tempitem{{"", "l_orderkey", 1, 1}},
			`TableScan3`,
		},
		// MPP Scan
		{
			`select /*+ read_from_storage(tiflash[lineitem]) */ l_orderkey from lineitem where # `,
			[]tempitem{{"", "l_orderkey", 1, 1}},
			`MPPScan1`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]) */ l_orderkey, l_partkey, l_suppkey, l_linenumber, ` +
				`l_quantity, l_extendedprice, l_discount, l_tax, l_shipdate from lineitem where # `,
			[]tempitem{{"", "l_orderkey", 1, 1}},
			`MPPScan2`,
		},
		{
			`select /*+ read_from_storage(tiflash[lineitem]) */ * from lineitem where # `,
			[]tempitem{{"", "l_orderkey", 1, 1}},
			`MPPScan3`,
		},
	}, n)
}

func genTPCHAgg(n int) utils.Queries {
	return gen4Templates([]template{
		// TiDB Agg
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 1}},
			`HashAgg1`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 1}},
			`HashAgg2`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where #`,
			[]tempitem{{"", "l_orderkey", 1, 1}},
			`HashAgg3`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 1}},
			`HashAgg4`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*), sum(l_quantity), ` +
				`sum(l_extendedprice), sum(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 1}},
			`HashAgg5`,
		},
		{
			`select /*+ read_from_storage(tikv[lineitem]), hash_agg() */ count(*), sum(l_quantity), sum(l_extendedprice), ` +
				`sum(l_discount), avg(l_quantity), avg(l_extendedprice), avg(l_discount) from lineitem where # ` +
				`group by l_returnflag, l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 1}},
			`HashAgg6`,
		},
		// TODO: TiDB StreamAgg
		// MPP Agg
	}, n)
}

func genTPCHJoin(n int) utils.Queries {

}

func genTPCHSort(n int) utils.Queries {

}
