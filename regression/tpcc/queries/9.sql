SELECT o_id, o_carrier_id, o_entry_d
FROM orders
WHERE o_w_id = 1 AND o_d_id = 1
  AND o_c_id = 1
ORDER BY o_id DESC
    LIMIT 1;