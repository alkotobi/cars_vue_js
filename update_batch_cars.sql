-- Update cars_stock.is_batch to 1 for all cars that have an associated sell_bill with is_batch_sell = 1
-- This query links cars_stock to sell_bill via the id_sell field and updates the is_batch field accordingly

UPDATE cars_stock cs
INNER JOIN sell_bill sb ON cs.id_sell = sb.id
SET cs.is_batch = 1
WHERE sb.is_batch_sell = 1
AND cs.hidden = 0;

-- Optional: Show the count of cars that will be updated (run this first to preview)
-- SELECT COUNT(*) as cars_to_update
-- FROM cars_stock cs
-- INNER JOIN sell_bill sb ON cs.id_sell = sb.id
-- WHERE sb.is_batch_sell = 1
-- AND cs.hidden = 0;

-- Optional: Show details of cars that will be updated (run this first to preview)
-- SELECT 
--     cs.id as car_id,
--     cs.vin,
--     sb.id as sell_bill_id,
--     sb.bill_ref,
--     sb.is_batch_sell,
--     cs.is_batch as current_is_batch
-- FROM cars_stock cs
-- INNER JOIN sell_bill sb ON cs.id_sell = sb.id
-- WHERE sb.is_batch_sell = 1
-- AND cs.hidden = 0; 