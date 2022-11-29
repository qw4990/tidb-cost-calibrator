SELECT c_id
FROM customer
WHERE c_w_id = 1 AND
        c_d_id = 1 AND c_last = 1
ORDER BY c_first;