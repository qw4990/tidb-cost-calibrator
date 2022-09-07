package costeval

import "github.com/qw4990/tidb-cost-calibrator/utils"

func genTPCHQueries(n int) utils.Queries {
	return gen4Templates([]template{
		{
			`select
    l_returnflag,
    l_linestatus,
    sum(l_quantity) as sum_qty,
    sum(l_extendedprice) as sum_base_price,
    sum(l_extendedprice * (1 - l_discount)) as sum_disc_price,
    sum(l_extendedprice * (1 - l_discount) * (1 + l_tax)) as sum_charge,
    avg(l_quantity) as avg_qty,
    avg(l_extendedprice) as avg_price,
    avg(l_discount) as avg_disc,
    count(*) as count_order
from
    lineitem
where
    l_shipdate <= date_sub('1998-12-01', interval 108 day)
    and #
group by
    l_returnflag,
    l_linestatus
order by
    l_returnflag,
    l_linestatus`,
			[]tempitem{{"", "l_orderkey", 1, 30000000}},
			"q1",
		},
		{
			`select
    l_orderkey,
    sum(l_extendedprice * (1 - l_discount)) as revenue,
    o_orderdate,
    o_shippriority
from
    customer,
    orders,
    lineitem
where
    c_mktsegment = 'AUTOMOBILE'
    and c_custkey = o_custkey
    and l_orderkey = o_orderkey
    and o_orderdate < '1995-03-13'
    and l_shipdate > '1995-03-13'
    and #
    and #
    and #
group by
    l_orderkey,
    o_orderdate,
    o_shippriority
order by
    revenue desc,
    o_orderdate
    limit 10`,
			[]tempitem{
				{"", "c_custkey", 1, 5000000},
				{"", "o_orderkey", 1, 50000000},
				{"", "l_orderkey", 1, 200000000},
			},
			"q3",
		},
	}, n)
}
