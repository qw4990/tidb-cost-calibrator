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
		{
			`select
    n_name,
    sum(l_extendedprice * (1 - l_discount)) as revenue
from
    customer,
    orders,
    lineitem,
    supplier,
    nation,
    region
where
    c_custkey = o_custkey
    and l_orderkey = o_orderkey
    and l_suppkey = s_suppkey
    and c_nationkey = s_nationkey
    and s_nationkey = n_nationkey
    and n_regionkey = r_regionkey
    and r_name = 'MIDDLE EAST'
    and o_orderdate >= '1994-01-01'
    and o_orderdate < date_add('1994-01-01', interval '1' year)
    and #
    and #
    and #
    and #
group by
    n_name
order by
    revenue desc`,
			[]tempitem{
				{"", "c_custkey", 1, 1500000},
				{"", "o_orderkey", 1, 15000000},
				{"", "l_orderkey", 1, 60000000},
				{"", "s_suppkey", 1, 100000},
			},
			"q5",
		},
		{`select
    sum(l_extendedprice * l_discount) as revenue
from
    lineitem
where
    l_shipdate >= '1994-01-01'
    and l_shipdate < date_add('1994-01-01', interval '1' year)
    and l_discount between 0.06 - 0.01 and 0.06 + 0.01
    and l_quantity < 24
    and l_orderkey >= ? and l_orderkey <= ?`,
			[]tempitem{{"", "l_orderkey", 1, 100000000}},
			"q6",
		},
		{
			`select
    c_custkey,
    c_name,
    sum(l_extendedprice * (1 - l_discount)) as revenue,
    c_acctbal,
    n_name,
    c_address,
    c_phone,
    c_comment
from
    customer,
    orders,
    lineitem,
    nation
where
    c_custkey = o_custkey
    and l_orderkey = o_orderkey
    and o_orderdate >= '1993-08-01'
    and o_orderdate < date_add('1993-08-01', interval '3' month)
    and l_returnflag = 'R'
    and c_nationkey = n_nationkey
    and #
    and #
    and #
group by
    c_custkey,
    c_name,
    c_acctbal,
    c_phone,
    n_name,
    c_address,
    c_comment
order by
    revenue desc
    limit 20`,
			[]tempitem{
				{"", "c_custkey", 1, 1},
				{"", "o_orderkey", 1, 1},
				{"", "l_orderkey", 1, 1},
			},
			"q10",
		},
	}, n)
}
