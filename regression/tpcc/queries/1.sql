set tidb_cost_model_version=2;

explain format=verbose
SELECT /*+ merge_join(customer, warehouse) */ c_discount, c_last, c_credit, w_tax
FROM customer, warehouse
WHERE w_id = 2 AND c_w_id = w_id AND
        c_d_id = 10 AND c_id = 1808;

explain format=verbose
SELECT /*+ tidb_inlj(warehouse) */ c_discount, c_last, c_credit, w_tax
FROM customer, warehouse
WHERE w_id = 2 AND c_w_id = w_id AND
        c_d_id = 10 AND c_id = 1808;

explain format=verbose SELECT c_discount, c_last, c_credit, w_tax
FROM customer, warehouse
WHERE w_id = 2 AND c_w_id = w_id AND
        c_d_id = 7 AND c_id = 2154;
1, 9, 2499
    
    2, 7, 2154
;
    
prepare stmt from 'SELECT c_discount, c_last, c_credit, w_tax
                   FROM customer, warehouse
                   WHERE w_id = ? AND c_w_id = w_id AND
                           c_d_id = ? AND c_id = ?';
set @a=1,@b=9,@c=2499;
execute stmt using @a,@b,@c;
    