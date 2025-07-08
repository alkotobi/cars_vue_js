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