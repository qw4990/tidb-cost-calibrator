SELECT /*+ TIDB_INLJ(order_line,stock) */
    COUNT(DISTINCT (s_i_id)) stock_count
FROM order_line, stock
WHERE ol_w_id = 1 AND ol_d_id = 1
  AND ol_o_id < 100 AND ol_o_id >= 100 - 20
  AND s_w_id = 1 AND s_i_id = ol_i_id
  AND s_quantity < 1;