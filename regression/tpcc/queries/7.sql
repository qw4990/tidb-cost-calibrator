SELECT count(c_id) namecnt
FROM customer
WHERE c_w_id = 1 AND c_d_id = 1
  AND c_last = 1;