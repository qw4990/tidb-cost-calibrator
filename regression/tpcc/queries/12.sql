SELECT ol_d_id, SUM(ol_amount)
FROM order_line
WHERE (ol_w_id, ol_d_id, ol_o_id)
          IN ((1,1,1), (2,2,2))
GROUP BY ol_d_id;