-- Query to find duplicated VIN numbers in cars_stock table
-- This query shows VINs that appear more than once

SELECT 
    vin,
    COUNT(*) as duplicate_count,
    GROUP_CONCAT(id ORDER BY id) as car_ids
FROM cars_stock 
WHERE vin IS NOT NULL 
    AND vin != '' 
    AND hidden = 0
GROUP BY vin 
HAVING COUNT(*) > 1
ORDER BY duplicate_count DESC, vin;

-- Alternative query showing more details for each duplicated VIN
SELECT 
    cs1.id,
    cs1.vin,
    cs1.notes,
    cs1.date_sell,
    cs1.price_cell,
    cs1.freight,
    cs1.date_loding,
    cs1.hidden,
    cs1.is_used_car,
    cs1.is_big_car,
    cs1.container_ref,
    cs1.is_batch
FROM cars_stock cs1
INNER JOIN (
    SELECT vin
    FROM cars_stock 
    WHERE vin IS NOT NULL 
        AND vin != '' 
        AND hidden = 0
    GROUP BY vin 
    HAVING COUNT(*) > 1
) cs2 ON cs1.vin = cs2.vin
WHERE cs1.hidden = 0
ORDER BY cs1.vin, cs1.id;

-- Query to count total duplicates
SELECT 
    COUNT(*) as total_duplicate_vin_groups
FROM (
    SELECT vin
    FROM cars_stock 
    WHERE vin IS NOT NULL 
        AND vin != '' 
        AND hidden = 0
    GROUP BY vin 
    HAVING COUNT(*) > 1
) duplicates; 

-- Query to count cars purchased before a given time that haven't arrived at port or don't have export license
-- Replace '2024-01-01' with your desired purchase date threshold

SELECT 
    COUNT(*) as total_cars_not_ready
FROM cars_stock cs
INNER JOIN buy_details bd ON cs.id_buy_details = bd.id
INNER JOIN buy_bill bb ON bd.id_buy_bill = bb.id
WHERE 
    bb.date_buy < '2024-01-01'  -- Replace with your desired purchase date threshold
    AND (
        -- Cars that haven't arrived at port (no loading date)
        cs.date_loding IS NULL 
        OR 
        -- Cars that don't have export license
        cs.export_lisence_ref IS NULL 
        OR 
        cs.export_lisence_ref = ''
    )
    AND cs.hidden = 0  -- Exclude hidden cars
    AND cs.date_send_documents IS NULL;  -- Exclude sold cars

-- Alternative query with more detailed breakdown:
SELECT 
    COUNT(*) as total_cars_not_ready,
    SUM(CASE WHEN cs.date_loding IS NULL THEN 1 ELSE 0 END) as cars_not_loaded,
    SUM(CASE WHEN cs.export_lisence_ref IS NULL OR cs.export_lisence_ref = '' THEN 1 ELSE 0 END) as cars_no_export_license,
    SUM(CASE WHEN cs.date_loding IS NULL AND (cs.export_lisence_ref IS NULL OR cs.export_lisence_ref = '') THEN 1 ELSE 0 END) as cars_both_issues
FROM cars_stock cs
INNER JOIN buy_details bd ON cs.id_buy_details = bd.id
INNER JOIN buy_bill bb ON bd.id_buy_bill = bb.id
WHERE 
    bb.date_buy < '2024-01-01'  -- Replace with your desired purchase date threshold
    AND (
        cs.date_loding IS NULL 
        OR 
        cs.export_lisence_ref IS NULL 
        OR 
        cs.export_lisence_ref = ''
    )
    AND cs.hidden = 0
    AND cs.date_send_documents IS NULL; 