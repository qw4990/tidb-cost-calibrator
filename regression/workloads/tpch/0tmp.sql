-- +----------------+------------+----------------+---------------------+--------------+-----------+
-- | Db_name        | Table_name | Partition_name | Update_time         | Modify_count | Row_count |
-- +----------------+------------+----------------+---------------------+--------------+-----------+
-- | tpch_clustered | nation     |                | 2022-09-08 05:04:13 |            0 |        25 |
-- | tpch_clustered | region     |                | 2022-09-08 05:04:13 |            0 |         5 |
-- | tpch_clustered | part       |                | 2022-09-08 05:02:46 |            0 |   4000000 |
-- | tpch_clustered | supplier   |                | 2022-09-08 05:02:30 |            0 |    200000 |
-- | tpch_clustered | partsupp   |                | 2022-09-08 05:02:26 |            0 |  16000000 |
-- | tpch_clustered | customer   |                | 2022-09-08 05:04:12 |            0 |   3000000 |
-- | tpch_clustered | orders     |                | 2022-09-08 05:04:01 |            0 |  30000000 |
-- | tpch_clustered | lineitem   |                | 2022-09-08 05:01:53 |       458992 | 119994608 |
-- +----------------+------------+----------------+---------------------+--------------+-----------+


-- q18
select
    c_name,
    c_custkey,
    o_orderkey,
    o_orderdate,
    o_totalprice,
    sum(l_quantity)
from
    customer,
    orders,
    lineitem
where
    o_orderkey in (
    select
        l_orderkey
    from
        lineitem
    where
        l_orderkey >= ? and l_orderkey <= ?
    group by
        l_orderkey having
        sum(l_quantity) > 314
    )
    and c_custkey = o_custkey
    and o_orderkey = l_orderkey
    and l_orderkey >= ? and l_orderkey <= ?
    and c_custkey >= ? and c_custkey <= ?
    and o_orderkey >= ? and o_orderkey <= ?
group by
    c_name,
    c_custkey,
    o_orderkey,
    o_orderdate,
    o_totalprice
order by
    o_totalprice desc,
    o_orderdate
    limit 100;


-- q17
select
    sum(l_extendedprice) / 7.0 as avg_yearly
from
    lineitem,
    part
where
    p_partkey = l_partkey
    and p_brand = 'Brand#44'
    and p_container = 'WRAP PKG'
    and l_quantity < (
        select 0.2 * avg(l_quantity)
        from lineitem
        where l_partkey = p_partkey
        and l_orderkey >= ? and l_orderkey <= ?)
    and p_partkey >= ? and p_partkey <= ?
    and l_orderkey >= ? and l_orderkey <= ?
;


-- q14
select
    100.00 * sum(case when p_type like 'PROMO%'
        then l_extendedprice * (1 - l_discount)
        else 0 end) / sum(l_extendedprice * (1 - l_discount)) as promo_revenue
from
    lineitem,
    part
where
    l_partkey = p_partkey
    and l_shipdate >= '1996-12-01'
    and l_shipdate < date_add('1996-12-01', interval '1' month),
    and l_orderkey >= ? and l_orderkey <= ?
    and p_partkey >= ? and p_partkey <= ?;



-- q13
select
    c_count,
    count(*) as custdist
from
    (
        select
            c_custkey,
            count(o_orderkey) as c_count
        from
            customer left outer join orders on c_custkey = o_custkey
            and o_comment not like '%pending%deposits%'
            and c_custkey >= ? and c_custkey <= ?
            and o_orderkey >= ? and o_orderkey <= ?
        group by
            c_custkey
    ) c_orders
group by
    c_count
order by
    custdist desc,
    c_count desc;



-- q12
select
    l_shipmode,
    sum(case
            when o_orderpriority = '1-URGENT'
                or o_orderpriority = '2-HIGH'
                then 1
            else 0
        end) as high_line_count,
    sum(case
            when o_orderpriority <> '1-URGENT'
                and o_orderpriority <> '2-HIGH'
                then 1
            else 0
        end) as low_line_count
from
    orders,
    lineitem
where
    o_orderkey = l_orderkey
    and l_shipmode in ('RAIL', 'FOB')
    and l_commitdate < l_receiptdate
    and l_shipdate < l_commitdate
    and l_receiptdate >= '1997-01-01'
    and l_receiptdate < date_add('1997-01-01', interval '1' year)
    and o_orderkey >= ? and o_orderkey <= ?
    and l_orderkey >= ? and l_orderkey <= ?
group by
    l_shipmode
order by
    l_shipmode;




-- q10
select
    /*+ read_from_storage(tiflash[customer, orders, lineitem, nation]) */
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
    and c_custkey >= 1 and c_custkey <= 15000
    and o_orderkey >= 1 and o_orderkey <= 150000
    and l_orderkey >= 1 and l_orderkey <= 600000
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
    limit 20;



-- q6
select
    sum(l_extendedprice * l_discount) as revenue
from
    lineitem
where
    l_shipdate >= '1994-01-01'
    and l_shipdate < date_add('1994-01-01', interval '1' year)
    and l_discount between 0.06 - 0.01 and 0.06 + 0.01
    and l_quantity < 24
    and l_orderkey >= 1 and l_orderkey <= 100000000;



-- q5
select
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
    and c_custkey >= 1 and c_custkey <= 1500000
    and o_orderkey >= 1 and o_orderkey <= 15000000
    and l_orderkey >= 1 and l_orderkey <= 60000000
    and s_suppkey >= 1 and s_suppkey <= 100000
group by
    n_name
order by
    revenue desc;


-- q3
select
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
    and c_custkey >= 1 and c_custkey <= 5000000
    and o_orderkey >= 1 and o_orderkey <= 50000000
    and l_orderkey >= 1 and l_orderkey <= 200000000
group by
    l_orderkey,
    o_orderdate,
    o_shippriority
order by
    revenue desc,
    o_orderdate
    limit 10;



-- q1
select
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
    and l_orderkey >= 1 and l_orderkey <= 30000000
group by
    l_returnflag,
    l_linestatus
order by
    l_returnflag,
    l_linestatus;
